package command

import (
	"api/pkg/model"
	"api/pkg/pubsub"
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type ExampleHandler struct {
	eventBus   *cqrs.EventBus
	commandBus *cqrs.CommandBus
}

func NewExampleHandler(
	registry *pubsub.HandlerRegistry,
) *ExampleHandler {
	handler := &ExampleHandler{}
	registry.AddCommandHandler(handler)
	return handler
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

func (h *ExampleHandler) Handle(ctx context.Context, cmdItf interface{}) error {
	cmd := cmdItf.(*model.ExampleCmd)
	fmt.Println("ExampleCmd", cmd)
	return nil
}
