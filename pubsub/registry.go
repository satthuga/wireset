package pubsub

import "github.com/ThreeDotsLabs/watermill/components/cqrs"

type HandlerRegistry struct {
	// Event
	eventHandlers []cqrs.EventHandler
	// Command
	commandHandlers []cqrs.CommandHandler
}

func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		eventHandlers:   []cqrs.EventHandler{},
		commandHandlers: []cqrs.CommandHandler{},
	}
}

// AddEventHandler adds event handler to registry
func (r *HandlerRegistry) AddEventHandler(handler cqrs.EventHandler) {
	r.eventHandlers = append(r.eventHandlers, handler)
}

// AddCommandHandler adds command handler to registry
func (r *HandlerRegistry) AddCommandHandler(handler cqrs.CommandHandler) {
	r.commandHandlers = append(r.commandHandlers, handler)
}

// GetCommandHandlerFactory GetEventHandlerFactory returns event handler factory
func (r *HandlerRegistry) GetCommandHandlerFactory() func(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) []cqrs.CommandHandler {
	return func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.CommandHandler {
		for _, handler := range r.commandHandlers {
			if setter, ok := handler.(MessageHandler); ok {
				setter.RegisterBus(cb, eb)
			}
		}

		return r.commandHandlers
	}
}

// GetEventHandlerFactory GetCommandHandlerFactory returns command handler factory
func (r *HandlerRegistry) GetEventHandlerFactory() func(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) []cqrs.EventHandler {
	return func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.EventHandler {
		for _, handler := range r.eventHandlers {
			if setter, ok := handler.(MessageHandler); ok {
				setter.RegisterBus(cb, eb)
			}
		}

		return r.eventHandlers
	}
}
