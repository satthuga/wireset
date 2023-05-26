package event

import (
	"context"
	"firebase.google.com/go/auth"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/configsvc"
	model "github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/shopifysvc"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type CreateUserHandler struct {
	Logger     *zap.Logger
	EventBus   *cqrs.EventBus
	CommandBus *cqrs.CommandBus
	ShopifySvc *shopifysvc.ShopifyService
	ConfigSvc  *configsvc.ConfigService
	AuthClient *auth.Client
}

func NewCreateUserHandler(
	logger *zap.Logger,
	shopifySvc *shopifysvc.ShopifyService,
	configSvc *configsvc.ConfigService,
	authClient *auth.Client,
) *CreateUserHandler {
	return &CreateUserHandler{
		Logger:     logger,
		ShopifySvc: shopifySvc,
		ConfigSvc:  configSvc,
		AuthClient: authClient,
	}
}

func (h *CreateUserHandler) HandlerName() string {
	return "install-webhook"
}

func (h *CreateUserHandler) NewEvent() interface{} {
	return &model.ShopInstalledEvt{}
}

func (h *CreateUserHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.EventBus = eventBus
	h.CommandBus = commandBus
}

func (h *CreateUserHandler) Handle(ctx context.Context, event interface{}) error {
	evt := event.(*model.ShopInstalledEvt)

	_, err := h.createUser(ctx, evt)
	if err != nil {
		return errors.WithMessage(err, "create user failed")
	}

	installWebhookCmd := &model.InstallWebhookCmd{
		MyshopifyDomain: evt.MyshopifyDomain,
		AccessToken:     evt.AccessToken,
	}

	if err := h.CommandBus.Send(ctx, installWebhookCmd); err != nil {
		return errors.WithMessage(err, "install webhook failed")
	}

	return nil
}

// createUser
func (h *CreateUserHandler) createUser(ctx context.Context, evt *model.ShopInstalledEvt) (*auth.UserRecord, error) {

	shopID := repository.NormalizeShopID(evt.ShopID)
	password := "LKoiu987(*&okj2oiuasdfOIUasdf@Dfsadf"
	params := (&auth.UserToCreate{}).
		UID(shopID).
		Email(shopID + "@aiodecor.aiocean.io").
		EmailVerified(true).
		Password(password).
		DisplayName(evt.MyshopifyDomain).
		Disabled(false)

	u, err := h.AuthClient.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return u, nil
}
