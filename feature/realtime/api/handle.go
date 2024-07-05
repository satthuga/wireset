package api

import (
	"context"
	"github.com/aiocean/wireset/feature/realtime/models"
	"github.com/aiocean/wireset/feature/realtime/room"
	"github.com/gofiber/contrib/websocket"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

func (h *WebsocketHandler) Handle(conn *websocket.Conn) {
	roomID := conn.Locals(roomIDKey).(string)
	currentRoom, err := h.RoomManager.GetRoom(roomID)
	logger := h.Logger.With(zap.String("room", roomID))

	if err != nil && errors.Is(err, room.ErrRoomNotFound) {
		logger.Info("Room not found, create new room")
		currentRoom, err = h.RoomManager.AddNewRoom(roomID)
		if err != nil {
			h.handleError(conn, logger, err, "failed to create new room")
			return
		}

		logger.Info("New rom is created")
	} else if err != nil {
		h.handleError(conn, logger, err, "failed to get room")
		return
	}

	logger.Info("Current room", zap.String(roomIDKey, roomID))
	username := conn.Locals(usernameKey).(string)

	if currentRoom.IsMemberExists(username) {
		h.handleError(conn, logger, errors.New("username already exists"), "username already exists")
		return
	}

	logger.Info("New user want to join group", zap.String(usernameKey, username))

	if err := currentRoom.AddMember(username, conn); err != nil {
		h.handleError(conn, logger, err, "failed to add user to room")
		return
	}
	logger.Info("Member Joined", zap.String(usernameKey, username), zap.String(roomIDKey, roomID))

	if err := h.EventBus.Publish(context.Background(), &models.UserJoinedEvt{
		UserName: username,
		RoomID:   roomID,
	}); err != nil {
		logger.Error("failed to publish user joined event", zap.Error(err))
	}
	// do some clean up
	defer h.OnDisconnect(conn, currentRoom, username)

	// reuse variable for avoiding memory allocation, but it's not a good practice, use it carefully to avoid memory leak, race condition, etc.
	var buf []byte
	for {
		if _, buf, err = conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("failed to read message, unexpected close error", zap.Error(err))
			} else {
				logger.Info("failed to read message", zap.Error(err))
			}
			return
		}

		result := gjson.GetManyBytes(buf, "topic", "payload")
		if result[0].Type == gjson.Null {
			logger.Error("topic is required")
			continue
		}

		if result[1].Type == gjson.Null {
			logger.Error("payload is required")
			continue
		}

		// it's safe to read the value of result[0] and result[1] from other goroutines, because we will never modify the value of result[0] and result[1]
		if err := h.Registry.Handle(result[0].String(), conn, &result[1]); err != nil {
			logger.Error("failed to handle message", zap.Error(err))
			continue
		}
	}
}
