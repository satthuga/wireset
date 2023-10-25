package shopifyapp

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/shopifyapp/command"
	"github.com/aiocean/wireset/feature/shopifyapp/event"
	"github.com/aiocean/wireset/feature/shopifyapp/handler"
	"github.com/aiocean/wireset/feature/shopifyapp/middleware"
	"github.com/aiocean/wireset/fiberapp"
	"github.com/gofiber/fiber/v2"
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
	wire.Struct(new(handler.GdprHandler), "*"),
)

type FeatureCore struct {
	InstallWebhookCmdHandler *command.InstallWebhookHandler
	SetShopStateCmdHandler   *command.SetShopStateHandler

	ShopInstalledEvtHandler *event.CreateUserHandler
	WelcomeEvtHandler       *event.WelcomeHandler

	AuthzMiddleware *middleware.ShopifyAuthzMiddleware

	AuthHandler    *handler.AuthHandler
	WebhookHandler *handler.WebhookHandler
	GdprHandler    *handler.GdprHandler

	EventProcessor   *cqrs.EventProcessor
	CommandProcessor *cqrs.CommandProcessor
	HttpRegistry     *fiberapp.Registry
}

func (f *FeatureCore) Init() error {
	f.CommandProcessor.AddHandlers(f.InstallWebhookCmdHandler)
	f.CommandProcessor.AddHandlers(f.SetShopStateCmdHandler)

	f.EventProcessor.AddHandlers(f.ShopInstalledEvtHandler)
	f.EventProcessor.AddHandlers(f.WelcomeEvtHandler)

	f.HttpRegistry.AddHttpMiddleware("/", f.AuthzMiddleware.Handle)

	f.HttpRegistry.AddHttpHandlers([]*fiberapp.HttpHandler{
		{
			Method:   fiber.MethodGet,
			Path:     "/auth/shopify/login-callback",
			Handlers: []fiber.Handler{f.AuthHandler.LoginCallback},
		},
		{
			Method:   fiber.MethodGet,
			Path:     "/auth/shopify/checkin",
			Handlers: []fiber.Handler{f.AuthHandler.Checkin},
		},
		{
			Method:   fiber.MethodGet,
			Path:     "/webhook/shopify/app-uninstalled",
			Handlers: []fiber.Handler{f.WebhookHandler.Uninstalled},
		},
		{
			Method:   fiber.MethodPost,
			Path:     "/gdpr/customers/data_request",
			Handlers: []fiber.Handler{f.GdprHandler.CustomerDataRequest},
		},
		{
			Method:   fiber.MethodPost,
			Path:     "/gdpr/customers/redact",
			Handlers: []fiber.Handler{f.GdprHandler.CustomerRedact},
		},
		{
			Method:   fiber.MethodPost,
			Path:     "/gdpr/shop/redact",
			Handlers: []fiber.Handler{f.GdprHandler.ShopRedact},
		},
	})
	return nil
}
