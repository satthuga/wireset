package pubsub

import "github.com/ThreeDotsLabs/watermill/components/cqrs"

type MessageHandler interface {
	RegisterBus(*cqrs.CommandBus, *cqrs.EventBus)
}
