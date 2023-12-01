package pubsub

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/garsue/watermillzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

// NewCommandBus creates a new command bus.
func NewCommandBus(publisher message.Publisher, logger *zap.Logger) (*cqrs.CommandBus, error) {
	commandBus, err := cqrs.NewCommandBusWithConfig(publisher, cqrs.CommandBusConfig{
		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
			return params.CommandName, nil
		},
		OnSend: func(params cqrs.CommandBusOnSendParams) error {
			params.Message.Metadata.Set("sent_at", time.Now().String())
			return nil
		},
		Marshaler: cqrs.JSONMarshaler{},
		Logger:    watermillzap.NewLogger(logger.Named("cqrsFacade")),
	})

	return commandBus, err
}

// NewEventBus creates a new event bus.
func NewEventBus(publisher message.Publisher, logger *zap.Logger) (*cqrs.EventBus, error) {
	eventBus, err := cqrs.NewEventBusWithConfig(publisher, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return params.EventName, nil
		},
		OnPublish: func(params cqrs.OnEventSendParams) error {
			params.Message.Metadata.Set("published_at", time.Now().String())
			return nil
		},
		Marshaler: cqrs.JSONMarshaler{},
		Logger:    watermillzap.NewLogger(logger.Named("cqrsFacade")),
	})

	return eventBus, err
}

func NewEventProcessor(router *message.Router, subscriber message.Subscriber, logger *zap.Logger) (*cqrs.EventProcessor, error) {
	return cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return params.EventName, nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return subscriber, nil
			},

			OnHandle: func(params cqrs.EventProcessorOnHandleParams) error {
				err := params.Handler.Handle(params.Message.Context(), params.Event)
				return errors.Wrap(err, "error handling event")
			},

			Marshaler: cqrs.JSONMarshaler{},
			Logger:    watermillzap.NewLogger(logger.Named("eventGroupProcessor")),
		},
	)
}

// NewCommandProcessor creates a new command processor.
func NewCommandProcessor(router *message.Router, subscriber message.Subscriber, logger *zap.Logger) (*cqrs.CommandProcessor, error) {
	return cqrs.NewCommandProcessorWithConfig(
		router,
		cqrs.CommandProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
				return params.CommandName, nil
			},
			SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return subscriber, nil
			},

			OnHandle: func(params cqrs.CommandProcessorOnHandleParams) error {
				err := params.Handler.Handle(params.Message.Context(), params.Command)
				return errors.Wrap(err, "error handling command")
			},

			Marshaler: cqrs.JSONMarshaler{},
			Logger:    watermillzap.NewLogger(logger.Named("commandProcessor")),
		},
	)
}
