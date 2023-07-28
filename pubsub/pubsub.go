package pubsub

import (
	"context"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/aiocean/wireset/configsvc"
	"sync"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/aiocean/wireset/pubsub/router"
	"github.com/garsue/watermillzap"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var DefaultWireset = wire.NewSet(
	NewPubsub,
	NewHandlerRegistry,
	router.DefaultWireset,
)

type Pubsub struct {
	mu         sync.Mutex
	facade     *cqrs.Facade
	registry   *HandlerRegistry
	logger     *zap.Logger
	router     *message.Router
	cfg        *configsvc.ConfigService
	subscriber message.Subscriber
	publisher  message.Publisher
}

// NewPubsub NewFacade creates a new Pubsub.
func NewPubsub(
	zapLogger *zap.Logger,
	router *message.Router,
	registry *HandlerRegistry,
	cfg *configsvc.ConfigService,
	subscriber *redisstream.Subscriber,
	publisher *redisstream.Publisher,
) (*Pubsub, error) {
	logger := zapLogger.Named("pubsub")
	facade := &Pubsub{
		mu:         sync.Mutex{},
		facade:     nil,
		registry:   registry,
		logger:     logger,
		router:     router,
		cfg:        cfg,
		subscriber: subscriber,
		publisher:  publisher,
	}

	return facade, nil
}

// Send publishes a message to the given topic.
func (f *Pubsub) Send(ctx context.Context, cmd interface{}) error {
	facade, err := f.getFacade()
	if err != nil {
		return err
	}

	return facade.CommandBus().Send(ctx, cmd)
}

// Publish triggers a message to the given topic.
func (f *Pubsub) Publish(ctx context.Context, evt interface{}) error {
	facade, err := f.getFacade()
	if err != nil {
		return err
	}

	return facade.EventBus().Publish(ctx, evt)
}

// Register
func (f *Pubsub) Register() error {
	_, err := f.getFacade()
	if err != nil {
		return errors.WithMessage(err, "failed to get pubsub")
	}

	return nil
}

// getFacade returns a facade.
func (f *Pubsub) getFacade() (*cqrs.Facade, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.facade != nil {
		return f.facade, nil
	}

	facade, err := f.createFacade()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create pubsub")
	}

	f.facade = facade

	return f.facade, nil
}

func (f *Pubsub) createFacade() (*cqrs.Facade, error) {
	cqrsFacade, err := cqrs.NewFacade(cqrs.FacadeConfig{
		GenerateCommandsTopic: func(commandName string) string {
			return commandName
		},
		GenerateEventsTopic: func(eventName string) string {
			return eventName
		},
		CommandsSubscriberConstructor: func(topic string) (message.Subscriber, error) {
			return f.subscriber, nil
		},
		EventsSubscriberConstructor: func(topic string) (message.Subscriber, error) {
			return f.subscriber, nil
		},
		CommandEventMarshaler: cqrs.JSONMarshaler{},
		CommandsPublisher:     f.publisher,
		EventsPublisher:       f.publisher,
		Router:                f.router,
		Logger:                watermillzap.NewLogger(f.logger.Named("cqrsFacade")),
		CommandHandlers:       f.registry.GetCommandHandlerFactory(),
		EventHandlers:         f.registry.GetEventHandlerFactory(),
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to create cqrs pubsub")
	}

	return cqrsFacade, nil
}
