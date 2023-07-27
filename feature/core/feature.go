package core

import (
	"github.com/aiocean/wireset/feature/core/command"
	"github.com/aiocean/wireset/feature/core/event"
	"github.com/aiocean/wireset/feature/core/handler"
	"github.com/aiocean/wireset/feature/core/middleware"
	"github.com/aiocean/wireset/fiberapp"
	"github.com/aiocean/wireset/pubsub"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	wire.Struct(new(FeatureCore), "*"),

	command.NewInstallWebhookHandler,
	command.NewSetShopStateHandler,
	event.NewCreateUserHandler,
	event.NewWelcomeHandler,

	middleware.NewAuthzController,

	wire.Struct(new(handler.AuthHandler), "*"),
	wire.Struct(new(handler.WebhookHandler), "*"),
	handler.NewWebsocketHandler,
	wire.Struct(new(handler.GdprHandler), "*"),
)

type FeatureCore struct {
	InstallWebhookCmdHandler *command.InstallWebhookHandler
	SetShopStateCmdHandler   *command.SetShopStateHandler

	ShopInstalledEvtHandler *event.CreateUserHandler
	WelcomeEvtHandler       *event.WelcomeHandler

	authzMiddleware *middleware.ShopifyAuthzMiddleware

	AuthHandler      *handler.AuthHandler
	WebhookHandler   *handler.WebhookHandler
	WebsocketHandler *handler.WebsocketHandler
	GdprHandler      *handler.GdprHandler

	PubsubRegistry *pubsub.HandlerRegistry
	HttpRegistry   *fiberapp.Registry
}

func (f *FeatureCore) Init() error {
	f.PubsubRegistry.AddCommandHandler(f.InstallWebhookCmdHandler)
	f.PubsubRegistry.AddCommandHandler(f.SetShopStateCmdHandler)

	f.PubsubRegistry.AddEventHandler(f.ShopInstalledEvtHandler)
	f.PubsubRegistry.AddEventHandler(f.WelcomeEvtHandler)

	f.HttpRegistry.AddHttpMiddleware("/ws", f.WebsocketHandler.CheckUpgrade)
	f.HttpRegistry.AddHttpMiddleware("/", f.authzMiddleware.Handle)

	f.HttpRegistry.AddHttpHandlers([]*fiberapp.HttpHandler{
		{
			Method:  fiber.MethodGet,
			Path:    "/auth/shopify/login-callback",
			Handler: f.AuthHandler.LoginCallback,
		},
		{
			Method:  fiber.MethodGet,
			Path:    "/auth/shopify/checkin",
			Handler: f.AuthHandler.Checkin,
		},
		{
			Method:  fiber.MethodGet,
			Path:    "/webhook/shopify/app-uninstalled",
			Handler: f.WebhookHandler.Uninstalled,
		},
		{
			Method:  fiber.MethodGet,
			Path:    "/ws/:id",
			Handler: websocket.New(f.WebsocketHandler.Handle),
		},
		{
			Method:  fiber.MethodPost,
			Path:    "/gdpr/customers/data_request",
			Handler: f.GdprHandler.CustomerDataRequest,
		},
		{
			Method:  fiber.MethodPost,
			Path:    "/gdpr/customers/redact",
			Handler: f.GdprHandler.CustomerRedact,
		},
		{
			Method:  fiber.MethodPost,
			Path:    "/gdpr/shop/redact",
			Handler: f.GdprHandler.ShopRedact,
		},
	})
	return nil
}
