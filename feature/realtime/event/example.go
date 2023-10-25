package event

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/model"
	"go.uber.org/zap"
)

type ExampleHandler struct {
	Logger     *zap.Logger
	EventBus   *cqrs.EventBus
	CommandBus *cqrs.CommandBus
}

func (h *ExampleHandler) HandlerName() string {
	return "EventExampleHandler"
}

func (h *ExampleHandler) NewEvent() interface{} {
	return &model.ShopInstalledEvt{}
}

func (h *ExampleHandler) Handle(ctx context.Context, event interface{}) error {
	cmd := event.(*model.ShopInstalledEvt)
	_ = cmd
	return nil
}
