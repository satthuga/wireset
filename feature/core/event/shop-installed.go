package event

import (
	"context"
	"firebase.google.com/go/auth"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/configsvc"
	model2 "github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/shopifysvc"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ShopInstalledHandler struct {
	Logger     *zap.Logger
	EventBus   *cqrs.EventBus
	CommandBus *cqrs.CommandBus
	ShopifySvc *shopifysvc.ShopifyService
	ConfigSvc  *configsvc.ConfigService
	AuthClient *auth.Client
}

func NewShopInstalledHandler(
	logger *zap.Logger,
	shopifySvc *shopifysvc.ShopifyService,
	configSvc *configsvc.ConfigService,
	authClient *auth.Client,
) *ShopInstalledHandler {
	return &ShopInstalledHandler{
		Logger:     logger,
		ShopifySvc: shopifySvc,
		ConfigSvc:  configSvc,
		AuthClient: authClient,
	}
}

func (h *ShopInstalledHandler) HandlerName() string {
	return "install-webhook"
}

func (h *ShopInstalledHandler) NewEvent() interface{} {
	return &model2.ShopInstalledEvt{}
}

func (h *ShopInstalledHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.EventBus = eventBus
	h.CommandBus = commandBus
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

	if err := h.CommandBus.Send(ctx, installWebhookCmd); err != nil {
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

	u, err := h.AuthClient.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return u, nil
}
