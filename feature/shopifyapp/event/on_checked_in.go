package event

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/shopifysvc"
	"go.uber.org/zap"
)

type OnCheckedInHandler struct {
	Logger        *zap.Logger
	EventBus      *cqrs.EventBus
	CommandBus    *cqrs.CommandBus
	ShopifyConfig *shopifysvc.Config
	ShopRepo      *repository.ShopRepository
	ShopifySvc    *shopifysvc.ShopifyService
	TokenRepo     *repository.TokenRepository
}

func (h *OnCheckedInHandler) HandlerName() string {
	return "OnCheckedInHandler"
}

func (h *OnCheckedInHandler) NewEvent() interface{} {
	return &model.ShopCheckedInEvt{}
}

func (h *OnCheckedInHandler) Handle(ctx context.Context, event interface{}) error {
	evt := event.(*model.ShopCheckedInEvt)
	accessTokenResponse, err := shopifysvc.ExchangeAccessToken(evt.MyshopifyDomain, h.ShopifyConfig.ClientId, h.ShopifyConfig.ClientSecret, evt.SessionToken)
	if err != nil {
		h.Logger.Error("failed to exchange access token", zap.Error(err))
		return err
	}

	// create shopify client
	shopify := h.ShopifySvc.GetShopifyClient(evt.MyshopifyDomain, accessTokenResponse.AccessToken)

	shopDetails, err := shopify.GetShopDetails()
	if err != nil {
		h.Logger.Error("failed to get shop details", zap.Error(err))
		return err
	}

	// check if shop is exist
	isShopExists, err := h.ShopRepo.IsShopExists(ctx, shopDetails.ID)
	if err != nil {
		h.Logger.Error("failed to check if shop exists", zap.Error(err))
		return err
	}

	if !isShopExists {

		// create shop
		if err := h.ShopRepo.Create(ctx, shopDetails); err != nil {
			h.Logger.Error("failed to create shop", zap.Error(err))
			return err
		}

		shopInstalledEvt := &model.ShopInstalledEvt{
			ShopID:          shopDetails.ID,
			MyshopifyDomain: shopDetails.Domain,
			AccessToken:     accessTokenResponse.AccessToken,
		}

		if err := h.EventBus.Publish(ctx, shopInstalledEvt); err != nil {
			h.Logger.Error("failed to publish shop installed event", zap.Error(err))
			return err
		}
	}

	token := &model.ShopifyToken{
		ShopID:      shopDetails.ID,
		AccessToken: accessTokenResponse.AccessToken,
	}

	if err := h.TokenRepo.SaveAccessToken(ctx, token); err != nil {
		h.Logger.Error("failed to save access token", zap.Error(err))
		return err
	}

	h.Logger.Info("shop checked in", zap.String("shop_id", shopDetails.ID))

	return nil
}
