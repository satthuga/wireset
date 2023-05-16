package core

import (
	"github.com/aiocean/wireset/feature/core/command"
	event2 "github.com/aiocean/wireset/feature/core/event"
	handler2 "github.com/aiocean/wireset/feature/core/handler"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	NewFeatureCore,

	command.NewInstallWebhookHandler,

	event2.NewCheckinHandler,
	event2.NewShopInstalledHandler,
	event2.NewWelcomeHandler,

	handler2.NewAuthHandler,
	handler2.NewWebhookHandler,
	handler2.NewWebsocketHandler,
	handler2.NewGdprHandler,
)

type FeatureCore struct {
}

func NewFeatureCore(

	_ *command.InstallWebhookHandler,

	_ *event2.CheckinHandler,
	_ *event2.ShopInstalledHandler,
	_ *event2.WelcomeHandler,

	_ *handler2.AuthHandler,
	_ *handler2.WebhookHandler,
	_ *handler2.WebsocketHandler,
	_ *handler2.GdprHandler,

) *FeatureCore {
	return &FeatureCore{}
}

func (f *FeatureCore) GetName() string {
	return "core"
}

func (f *FeatureCore) Register() {

}
