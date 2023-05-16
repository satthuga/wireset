package event

import (
	"api/pkg/configsvc"
	model2 "api/pkg/model"
	"api/pkg/pubsub"
	"api/pkg/shopifysvc"
	"context"
	"firebase.google.com/go/auth"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ShopInstalledHandler struct {
	logger     *zap.Logger
	eventBus   *cqrs.EventBus
	commandBus *cqrs.CommandBus
	shopifySvc *shopifysvc.ShopifyService
	configSvc  *configsvc.ConfigService
	authClient *auth.Client
}

// this handler used to init the wallet when user registered

func NewShopInstalledHandler(
	logger *zap.Logger,
	shopifySvc *shopifysvc.ShopifyService,
	configSvc *configsvc.ConfigService,
	registry *pubsub.HandlerRegistry,
	authClient *auth.Client,
) *ShopInstalledHandler {

	handler := &ShopInstalledHandler{
		logger:     logger,
		shopifySvc: shopifySvc,
		configSvc:  configSvc,
		authClient: authClient,
	}

	registry.AddEventHandler(handler)

	return handler
}

func (h *ShopInstalledHandler) HandlerName() string {
	return "install-webhook"
}

func (h *ShopInstalledHandler) NewEvent() interface{} {
	return &model2.ShopInstalledEvt{}
}

func (h *ShopInstalledHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.eventBus = eventBus
	h.commandBus = commandBus
}

func (h *ShopInstalledHandler) Handle(ctx context.Context, event interface{}) error {
	evt := event.(*model2.ShopInstalledEvt)

	user, err := h.createUser(ctx, evt)
	if err != nil {
		return errors.WithMessage(err, "create user failed")
	}

	_ = user
	// TODO send email

	installWebhookCmd := &model2.InstallWebhookCmd{
		MyshopifyDomain: evt.MyshopifyDomain,
		AccessToken:     evt.AccessToken,
	}

	if err := h.commandBus.Send(ctx, installWebhookCmd); err != nil {
		return errors.WithMessage(err, "install webhook failed")
	}

	return nil
}

// createUser
func (h *ShopInstalledHandler) createUser(ctx context.Context, evt *model2.ShopInstalledEvt) (*auth.UserRecord, error) {
	params := (&auth.UserToCreate{}).
		Email("user@example.com").
		EmailVerified(false).
		PhoneNumber("+15555550100").
		Password("secretPassword").
		DisplayName("John Doe").
		PhotoURL("http://www.example.com/12345678/photo.png").
		Disabled(false)

	u, err := h.authClient.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return u, nil
}
