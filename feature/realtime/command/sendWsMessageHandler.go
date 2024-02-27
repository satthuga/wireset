package command

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/realtime/room"
	"go.uber.org/zap"
)

type SendWsMessageHandler struct {
	EventBus    *cqrs.EventBus
	CommandBus  *cqrs.CommandBus
	RoomManager *room.Manager
	Logger      *zap.Logger
}

type SendWsMessageCmd struct {
	RoomID   string `json:"room_id"`
	Username string `json:"username"`
	Payload  any    `json:"payload"`
}

func (h *SendWsMessageHandler) HandlerName() string {
	return "SendWsMessageHandler"
}

func (h *SendWsMessageHandler) NewCommand() interface{} {
	return &SendWsMessageCmd{}
}
func (h *SendWsMessageHandler) Handle(ctx context.Context, raw any) error {
	cmd, ok := raw.(*SendWsMessageCmd)
	if !ok {
		h.Logger.Error("Failed to cast raw to SendWsMessageCmd")
		return fmt.Errorf("failed to cast raw to SendWsMessageCmd")
	}

	toRoom, err := h.RoomManager.GetRoom(cmd.RoomID)
	if err != nil {
		h.Logger.Error("Failed to get room", zap.String("roomID", cmd.RoomID), zap.Error(err))
		return fmt.Errorf("failed to get room: %w", err)
	}

	if err := toRoom.SendMessageTo(cmd.Username, cmd.Payload); err != nil {
		h.Logger.Error("Failed to send message", zap.String("username", cmd.Username), zap.Error(err))
		return fmt.Errorf("failed to send message: %w", err)
	}

	h.Logger.Info("Message sent successfully", zap.String("username", cmd.Username), zap.String("roomID", cmd.RoomID))
	return nil
}
