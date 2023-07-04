package notify

import (
	"github.com/aiocean/wireset/feature/notify/event"
	"github.com/aiocean/wireset/pubsub"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	wire.Struct(new(FeatureNotify), "*"),
	event.NewNotifyDiscordOnInstallHandler,
)

type FeatureNotify struct {
	HandlerRegistry               *pubsub.HandlerRegistry
	NotifyDiscordOnInstallHandler *event.NotifyDiscordOnInstallHandler
}

func (f *FeatureNotify) Init() error {
	f.HandlerRegistry.AddEventHandler(f.NotifyDiscordOnInstallHandler)
	return nil
}
