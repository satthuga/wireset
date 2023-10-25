package notify

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/notify/event"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	wire.Struct(new(FeatureNotify), "*"),
	event.NewNotifyDiscordOnInstallHandler,
)

type FeatureNotify struct {
	EvtProcessor                  *cqrs.EventProcessor
	NotifyDiscordOnInstallHandler *event.NotifyDiscordOnInstallHandler
}

func (f *FeatureNotify) Init() error {
	f.EvtProcessor.AddHandlers(f.NotifyDiscordOnInstallHandler)
	return nil
}
