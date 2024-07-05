package shopifyapp

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/realtime/registry"
	"github.com/aiocean/wireset/feature/shopifyapp/api"
	"github.com/aiocean/wireset/feature/shopifyapp/command"
	"github.com/aiocean/wireset/feature/shopifyapp/event"
	"github.com/aiocean/wireset/feature/shopifyapp/middleware"
	"github.com/aiocean/wireset/feature/shopifyapp/models"
	"github.com/aiocean/wireset/feature/shopifyapp/ws"
	"github.com/aiocean/wireset/fiberapp"
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	wire.Struct(new(FeatureCore), "*"),

	command.NewInstallWebhookHandler,
	command.NewSetShopStateHandler,
	wire.Struct(new(event.CreateUserHandler), "*"),
	wire.Struct(new(event.WelcomeHandler), "*"),
	wire.Struct(new(event.OnUserConnectedHandler), "*"),
	wire.Struct(new(event.OnCheckedInHandler), "*"),

	wire.Struct(new(ws.FetchActivateSubscriptionHandler), "*"),
	wire.Struct(new(ws.CreateSubscriptionHandler), "*"),

	middleware.NewAuthzController,

	wire.Struct(new(api.AuthHandler), "*"),
	wire.Struct(new(api.WebhookHandler), "*"),
	wire.Struct(new(api.GdprHandler), "*"),
)

type FeatureCore struct {
	InstallWebhookCmdHandler *command.InstallWebhookHandler
	SetShopStateCmdHandler   *command.SetShopStateHandler

	FetchPlanWsHandler        *ws.FetchActivateSubscriptionHandler
	CreateSubscriptionHandler *ws.CreateSubscriptionHandler

	ShopInstalledEvtHandler *event.CreateUserHandler
	WelcomeEvtHandler       *event.WelcomeHandler
	OnUserConnectedHandler  *event.OnUserConnectedHandler
	OnCheckedInHandler      *event.OnCheckedInHandler

	AuthzMiddleware *middleware.ShopifyAuthzMiddleware

	AuthHandler    *api.AuthHandler
	WebhookHandler *api.WebhookHandler
	GdprHandler    *api.GdprHandler

	EventProcessor   *cqrs.EventProcessor
	CommandProcessor *cqrs.CommandProcessor
	HttpRegistry     *fiberapp.Registry
	WsRegistry       *registry.HandlerRegistry
}

func (f *FeatureCore) Init() error {

	if err := f.CommandProcessor.AddHandlers(
		f.InstallWebhookCmdHandler,
		f.SetShopStateCmdHandler,
	); err != nil {
		return err
	}

	if err := f.EventProcessor.AddHandlers(
		f.ShopInstalledEvtHandler,
		f.WelcomeEvtHandler,
		f.OnUserConnectedHandler,
		f.OnCheckedInHandler,
	); err != nil {
		return err
	}

	f.HttpRegistry.AddHttpMiddleware("/", f.AuthzMiddleware.Handle)

	f.HttpRegistry.AddHttpHandlers(
		&fiberapp.HttpHandler{
			Method:   fiber.MethodGet,
			Path:     "/auth/shopify/login-callback",
			Handlers: []fiber.Handler{f.AuthHandler.LoginCallback},
		},
		&fiberapp.HttpHandler{
			Method:   fiber.MethodGet,
			Path:     "/auth/shopify/checkin",
			Handlers: []fiber.Handler{f.AuthHandler.Checkin},
		},
		&fiberapp.HttpHandler{
			Method:   fiber.MethodGet,
			Path:     "/webhook/shopify/app-uninstalled",
			Handlers: []fiber.Handler{f.WebhookHandler.Uninstalled},
		},
		&fiberapp.HttpHandler{
			Method:   fiber.MethodPost,
			Path:     "/gdpr/customers/data_request",
			Handlers: []fiber.Handler{f.GdprHandler.CustomerDataRequest},
		},
		&fiberapp.HttpHandler{
			Method:   fiber.MethodPost,
			Path:     "/gdpr/customers/redact",
			Handlers: []fiber.Handler{f.GdprHandler.CustomerRedact},
		},
		&fiberapp.HttpHandler{
			Method:   fiber.MethodPost,
			Path:     "/gdpr/shop/redact",
			Handlers: []fiber.Handler{f.GdprHandler.ShopRedact},
		},
	)

	f.WsRegistry.AddWebsocketHandler(
		&registry.WebsocketHandler{
			Topic:   models.TopicFetchActivateSubscription.String(),
			Handler: f.FetchPlanWsHandler.Handle,
		},
		&registry.WebsocketHandler{
			Topic:   models.TopicCreateSubscription.String(),
			Handler: f.CreateSubscriptionHandler.Handle,
		},
	)
	return nil
}
