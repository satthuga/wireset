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

func (h *CreateUserHandler) HandlerName() string {
	return "install-webhook"
}

func (h *CreateUserHandler) NewEvent() interface{} {
	return &model.ShopInstalledEvt{}
}

func (h *CreateUserHandler) Handle(ctx context.Context, event interface{}) error {
	evt := event.(*model.ShopInstalledEvt)

	_, err := h.createUser(ctx, evt)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return nil
		}
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

var ErrUserAlreadyExists = errors.New("user already exists")

// createUser
func (h *CreateUserHandler) createUser(ctx context.Context, evt *model.ShopInstalledEvt) (*auth.UserRecord, error) {

	shopID, err := repository.NormalizeShopID(evt.ShopID)
	if err != nil {
		return nil, err
	}

	// Check if user already exists
	_, err = h.AuthClient.GetUser(ctx, shopID)
	if err == nil {
		// User already exists, return an error or handle this situation as needed
		return nil, ErrUserAlreadyExists
	}

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
		return nil, errors.WithMessage(err, "create user failed")
	}

	return u, nil
}
