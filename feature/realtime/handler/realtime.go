package handler

import (
	"github.com/aiocean/wireset/feature/realtime/room"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

type WebsocketHandler struct {
	RoomManager *room.Manager
	Logger      *zap.Logger
}

func NewWebsocketHandler(
	Logger *zap.Logger,
	roomManager *room.Manager,
) *WebsocketHandler {
	return &WebsocketHandler{
		RoomManager: roomManager,
		Logger:      Logger,
	}
}

func (h *WebsocketHandler) SendDm(ctx *fiber.Ctx) error {
	roomID := ctx.Query("room")
	if roomID == "" {
		return errors.New("roomID is required")
	}

	username := ctx.Query("username")
	if username == "" {
		return errors.New("userName is required")
	}

	currentRoom, err := h.RoomManager.GetRoom(roomID)
	if err != nil && !errors.Is(err, room.ErrRoomNotFound) {
		h.Logger.Error("get room", zap.Error(err))
		return err
	}

	if currentRoom == nil {
		h.Logger.Info("room not found, so you can use this name", zap.String("roomID", roomID))
		return ctx.Next()
	}

	if currentRoom.IsMemberExists(username) {
		h.Logger.Error("member already exists", zap.String("userName", username), zap.String("roomID", roomID))
		return errors.New("username already exists")
	}

	recipient := gjson.GetBytes(ctx.Body(), "recipient").String()
	if recipient == "" {
		return errors.New("recipient is required")
	}

	if recipient == "system" {
		return ctx.SendStatus(fiber.StatusOK)
	}

	if err := currentRoom.SendMessageTo(recipient, ctx.Body()); err != nil {
		h.Logger.Error("send message", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]string{
			"error": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
// require "room" and "username" query params.
func (h *WebsocketHandler) Upgrade(ctx *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(ctx) {
		roomID := ctx.Query("room")
		if roomID == "" {
			return errors.New("roomID is required")
		}
		ctx.Locals("roomID", roomID)

		username := ctx.Query("username")
		if username == "" {
			return errors.New("userName is required")
		}
		ctx.Locals("username", username)

		currentRoom, err := h.RoomManager.GetRoom(roomID)
		if err != nil && !errors.Is(err, room.ErrRoomNotFound) {
			h.Logger.Error("get room", zap.Error(err))
			return err
		}

		if currentRoom == nil {
			h.Logger.Info("room not found, so you can use this name", zap.String("roomID", roomID))
			return ctx.Next()
		}

		if currentRoom.IsMemberExists(username) {
			h.Logger.Error("member already exists", zap.String("userName", username), zap.String("roomID", roomID))
			return errors.New("username already exists")
		}

		return ctx.Next()
	}

	return fiber.ErrUpgradeRequired
}

// OnDisconnect is called when a client disconnects from the server.
func (h *WebsocketHandler) OnDisconnect(ctx *websocket.Conn, room *room.Room, username string) {
	// delete member from room
	if err := room.DeleteMember(username); err != nil {
		h.Logger.Error("delete member", zap.String("username", username), zap.String("roomID", room.ID))
		return
	}

	h.Logger.Info("Websocket Closed")

	if room.IsEmpty() {
		if err := h.RoomManager.DeleteRoom(room.ID); err != nil {
			h.Logger.Error("delete room", zap.String("roomID", room.ID))
		}

		h.Logger.Info("Room Deleted")
		return
	}

	if err := room.BroadcastMessage(map[string]string{
		"type":   "members/left",
		"member": username,
	}); err != nil {
		h.Logger.Error("broadcast message", zap.Error(err))
	}

	h.Logger.Info("Member Left", zap.String("username", username), zap.String("roomID", room.ID))
}

func (h *WebsocketHandler) Handle(conn *websocket.Conn) {
	roomID := conn.Locals("roomID").(string)
	currentRoom, err := h.RoomManager.GetRoom(roomID)
	logger := h.Logger.Named("room: " + roomID)

	if err != nil && errors.Is(err, room.ErrRoomNotFound) {
		logger.Info("Room not found, create new room")
		currentRoom, err = h.RoomManager.AddNewRoom(roomID)
		if err != nil {
			logger.Error("failed to add room", zap.String("roomID", roomID))
			conn.WriteJSON(map[string]string{
				"error": "failed to add room: " + err.Error(),
			})
			return
		}

		logger.Info("New rom is created")
	} else if err != nil {
		logger.Error("failed to get room", zap.Error(err))
		conn.WriteJSON(map[string]string{
			"error": "failed to get room: " + err.Error(),
		})

		return
	}

	logger.Info("Current room", zap.String("roomID", roomID))
	username := conn.Locals("username").(string)

	if currentRoom.IsMemberExists(username) {
		logger.Error("member already exists", zap.String("username", username), zap.String("roomID", roomID))
		conn.WriteJSON(map[string]string{
			"error": "username already exists",
		})
		return
	}

	logger.Info("New user want to join group", zap.String("username", username))

	if err := currentRoom.AddMember(username, conn); err != nil {
		logger.Error("failed to add user to room", zap.String("username", username), zap.String("roomID", roomID))
		return
	}
	logger.Info("Member Joined", zap.String("username", username), zap.String("roomID", roomID))

	if err := currentRoom.BroadcastMessage(map[string]string{
		"type":   "members/join",
		"member": username,
	}); err != nil {
		logger.Error("broadcast message", zap.Error(err))
	}

	// do some clean up
	defer h.OnDisconnect(conn, currentRoom, username)

	// reuse variable for avoiding memory allocation, but it's not a good practice, use it carefully to avoid memory leak, race condition, etc.
	var buf []byte
	var recipient string
	for {
		if _, buf, err = conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("failed to read message, unexpected close error", zap.Error(err))
			} else {
				logger.Info("failed to read message", zap.Error(err))
			}
			return
		}

		// set sender
		buf, err = sjson.SetBytes(buf, "sender", username)
		if err != nil {
			logger.Error("failed to set sender", zap.Error(err))
			continue
		}

		recipient = gjson.GetBytes(buf, "recipient").String()
		if recipient != "system" {
			if err := currentRoom.SendMessageTo(recipient, buf); err != nil {
				logger.Error("send message", zap.Error(err))
				continue
			}

			continue
		}
	}
}
