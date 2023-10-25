package pubsub

import (
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	NewCommandProcessor,
	NewEventProcessor,
	NewCommandBus,
	NewEventBus,
	NewRouter,
)
