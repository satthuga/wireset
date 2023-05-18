package command

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/model"
)

type ExampleHandler struct {
	eventBus   *cqrs.EventBus
	commandBus *cqrs.CommandBus
}

func NewExampleHandler() *ExampleHandler {
	return &ExampleHandler{}
}

func (h *ExampleHandler) HandlerName() string {
	return "ExampleHandler"
}

func (h *ExampleHandler) NewCommand() interface{} {
	return &model.ExampleCmd{}
}

func (h *ExampleHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.eventBus = eventBus
	h.commandBus = commandBus
}

func (h *ExampleHandler) Handle(ctx context.Context, cmd interface{}) error {
	fmt.Println("ExampleCmd", cmd)
	return nil
}
