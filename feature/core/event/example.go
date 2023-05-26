package event

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/model"
	"go.uber.org/zap"
)

type ExampleHandler struct {
	logger     *zap.Logger
	eventBus   *cqrs.EventBus
	commandBus *cqrs.CommandBus
}

// this handler used to init the wallet when user registered

func NewExampleHandler(
	logger *zap.Logger,
) *ExampleHandler {
	handler := &ExampleHandler{
		logger: logger,
	}

	return handler
}

func (h *ExampleHandler) HandlerName() string {
	return "ExampleHandler"
}

func (h *ExampleHandler) NewEvent() interface{} {
	return &model.ShopInstalledEvt{}
}

func (h *ExampleHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.eventBus = eventBus
	h.commandBus = commandBus
}

func (h *ExampleHandler) Handle(ctx context.Context, event interface{}) error {
	cmd := event.(*model.ShopInstalledEvt)
	_ = cmd
	return nil
}
