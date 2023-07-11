package core

import (
	"github.com/aiocean/wireset/feature/core/command"
	"github.com/aiocean/wireset/feature/core/event"
	"github.com/aiocean/wireset/feature/core/handler"
	"github.com/aiocean/wireset/fiberapp"
	"github.com/aiocean/wireset/pubsub"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	wire.Struct(new(FeatureCore), "*"),

	command.NewInstallWebhookHandler,
	command.NewSetShopStateHandler,
	event.NewCreateUserHandler,
	event.NewWelcomeHandler,

	wire.Struct(new(handler.AuthHandler), "*"),
	wire.Struct(new(handler.WebhookHandler), "*"),
	handler.NewWebsocketHandler,
	wire.Struct(new(handler.GdprHandler), "*"),
	wire.Struct(new(handler.PrometheusHandler), "*"),
)

type FeatureCore struct {
	InstallWebhookCmdHandler *command.InstallWebhookHandler
	SetShopStateCmdHandler   *command.SetShopStateHandler

	ShopInstalledEvtHandler *event.CreateUserHandler
	WelcomeEvtHandler       *event.WelcomeHandler

	AuthHandler       *handler.AuthHandler
	WebhookHandler    *handler.WebhookHandler
	WebsocketHandler  *handler.WebsocketHandler
	GdprHandler       *handler.GdprHandler
	PrometheusHandler *handler.PrometheusHandler

	PubsubRegistry *pubsub.HandlerRegistry
	HttpRegistry   *fiberapp.Registry
}

func (f *FeatureCore) Init() error {
	f.PubsubRegistry.AddCommandHandler(f.InstallWebhookCmdHandler)
	f.PubsubRegistry.AddCommandHandler(f.SetShopStateCmdHandler)

	f.PubsubRegistry.AddEventHandler(f.ShopInstalledEvtHandler)
	f.PubsubRegistry.AddEventHandler(f.WelcomeEvtHandler)

	f.HttpRegistry.AddHttpHandler(f.AuthHandler)
	f.HttpRegistry.AddHttpHandler(f.WebhookHandler)
	f.HttpRegistry.AddHttpHandler(f.WebsocketHandler)
	f.HttpRegistry.AddHttpHandler(f.GdprHandler)
	f.HttpRegistry.AddHttpHandler(f.PrometheusHandler)

	return nil
}
