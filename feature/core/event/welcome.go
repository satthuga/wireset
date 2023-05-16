package event

import (
	"api/pkg/model"
	"api/pkg/pubsub"
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"go.uber.org/zap"
)

type WelcomeHandler struct {
	logger     *zap.Logger
	eventBus   *cqrs.EventBus
	commandBus *cqrs.CommandBus
}

// this handler used to init the wallet when user registered

func NewWelcomeHandler(
	logger *zap.Logger,
	registry *pubsub.HandlerRegistry,
) *WelcomeHandler {

	handler := &WelcomeHandler{
		logger: logger,
	}

	registry.AddEventHandler(handler)

	return handler
}

func (h *WelcomeHandler) HandlerName() string {
	return "send-welcome-email"
}

func (h *WelcomeHandler) NewEvent() interface{} {
	return &model.ShopInstalledEvt{}
}

func (h *WelcomeHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.eventBus = eventBus
	h.commandBus = commandBus
}

func (h *WelcomeHandler) Handle(ctx context.Context, event interface{}) error {
	cmd := event.(*model.ShopInstalledEvt)
	_ = cmd
	// TODO send email
	return nil
}
