package core

import (
	"github.com/aiocean/wireset/feature/core/command"
	"github.com/aiocean/wireset/feature/core/event"
	"github.com/aiocean/wireset/feature/core/handler"
	"github.com/aiocean/wireset/pubsub"
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	wire.Struct(new(FeatureCore), "*"),

	command.NewInstallWebhookHandler,
	event.NewCheckinHandler,
	event.NewShopInstalledHandler,
	event.NewWelcomeHandler,

	wire.Struct(new(handler.AuthHandler), "*"),
	wire.Struct(new(handler.WebhookHandler), "*"),
	wire.Struct(new(handler.WebsocketHandler), "*"),
	wire.Struct(new(handler.GdprHandler), "*"),
	wire.Struct(new(handler.PrometheusHandler), "*"),
)

type FeatureCore struct {
	InstallWebhookCmdHandler *command.InstallWebhookHandler

	CheckinEvtHandler       *event.CheckinHandler
	ShopInstalledEvtHandler *event.ShopInstalledHandler
	WelcomeEvtHandler       *event.WelcomeHandler

	AuthHandler       *handler.AuthHandler
	WebhookHandler    *handler.WebhookHandler
	WebsocketHandler  *handler.WebsocketHandler
	GdprHandler       *handler.GdprHandler
	PrometheusHandler *handler.PrometheusHandler

	HandlerRegistry *pubsub.HandlerRegistry
	FiberApp        *fiber.App
}

func (f *FeatureCore) Init() error {
	f.HandlerRegistry.AddCommandHandler(f.InstallWebhookCmdHandler)

	f.HandlerRegistry.AddEventHandler(f.CheckinEvtHandler)
	f.HandlerRegistry.AddEventHandler(f.ShopInstalledEvtHandler)
	f.HandlerRegistry.AddEventHandler(f.WelcomeEvtHandler)

	f.AuthHandler.Register(f.FiberApp)
	f.WebhookHandler.Register(f.FiberApp)
	f.WebsocketHandler.Register(f.FiberApp)
	f.GdprHandler.Register(f.FiberApp)
	f.PrometheusHandler.Register(f.FiberApp)

	return nil
}
