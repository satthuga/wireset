package discord_notify

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/discord-notify/event"
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
	err := f.EvtProcessor.AddHandlers(f.NotifyDiscordOnInstallHandler)
	if err != nil {
		return err
	}
	return nil
}
