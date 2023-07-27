package pubsub

import (
	"context"
	"sync"

	"cloud.google.com/go/firestore"
	watermillFirestore "github.com/ThreeDotsLabs/watermill-firestore/pkg/firestore"
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

type Pubsub struct {
	mu              sync.Mutex
	facade          *cqrs.Facade
	registry        *HandlerRegistry
	logger          *zap.Logger
	router          *message.Router
	firestoreClient *firestore.Client
}

// NewFacade creates a new Pubsub.
func NewPubsub(
	zapLogger *zap.Logger,
	router *message.Router,
	registry *HandlerRegistry,
	firestoreClient *firestore.Client,
) (*Pubsub, error) {
	logger := zapLogger.With(zap.Strings("tags", []string{"Pubsub"}))
	facade := &Pubsub{
		mu:              sync.Mutex{},
		facade:          nil,
		registry:        registry,
		logger:          logger,
		router:          router,
		firestoreClient: firestoreClient,
	}
	return facade, nil
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

// publisherConstructor returns a publisher constructor.
func (f *Pubsub) createEventPublisher() (message.Publisher, error) {
	eventsPublisher, err := watermillFirestore.NewPublisher(
		watermillFirestore.PublisherConfig{
			CustomFirestoreClient: f.firestoreClient,
		},
		watermillzap.NewLogger(f.logger.Named("event-publisher")),
	)
	if err != nil {
		return nil, err
	}

	return eventsPublisher, nil
}

// createCommandSender returns a command sender.
func (f *Pubsub) commandSenderConstructor() (message.Publisher, error) {

	commandsPublisher, err := watermillFirestore.NewPublisher(
		watermillFirestore.PublisherConfig{
			CustomFirestoreClient: f.firestoreClient,
		},
		watermillzap.NewLogger(f.logger.Named("command-sender")),
	)
	if err != nil {
		return nil, err
	}

	return commandsPublisher, nil
}

// commandSubscriberConstructor returns a command subscriber.
func (f *Pubsub) commandSubscriberConstructor() cqrs.CommandsSubscriberConstructor {
	return func(topic string) (message.Subscriber, error) {
		return watermillFirestore.NewSubscriber(
			watermillFirestore.SubscriberConfig{
				GenerateSubscriptionName: func(topic string) string {
					return topic
				},
				CustomFirestoreClient: f.firestoreClient,
			},
			watermillzap.NewLogger(f.logger.Named("command-subscriber")),
		)
	}
}

// eventSubscriberConstructor returns an event subscriber.
func (f *Pubsub) eventSubscriberConstructor() cqrs.EventsSubscriberConstructor {
	return func(topic string) (message.Subscriber, error) {
		return watermillFirestore.NewSubscriber(
			watermillFirestore.SubscriberConfig{
				GenerateSubscriptionName: func(topic string) string {
					return topic
				},
				CustomFirestoreClient: f.firestoreClient,
			},
			watermillzap.NewLogger(f.logger.Named("event-subscriber")),
		)
	}
}

func (f *Pubsub) createFacade() (*cqrs.Facade, error) {
	eventsPublisher, err := f.createEventPublisher()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create event publisher")
	}

	commandsSender, err := f.commandSenderConstructor()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command sender")
	}

	cqrsFacade, err := cqrs.NewFacade(cqrs.FacadeConfig{
		GenerateCommandsTopic: func(commandName string) string {
			return commandName
		},
		GenerateEventsTopic: func(eventName string) string {
			return eventName
		},
		CommandsSubscriberConstructor: f.commandSubscriberConstructor(),
		EventsSubscriberConstructor:   f.eventSubscriberConstructor(),
		CommandEventMarshaler:         cqrs.JSONMarshaler{},
		CommandsPublisher:             commandsSender,
		EventsPublisher:               eventsPublisher,
		Router:                        f.router,
		Logger:                        watermillzap.NewLogger(f.logger.Named("cqrsFacade")),
		CommandHandlers:               f.registry.GetCommandHandlerFactory(),
		EventHandlers:                 f.registry.GetEventHandlerFactory(),
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to create cqrs pubsub")
	}

	return cqrsFacade, nil
}
