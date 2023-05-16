package handler

import (
	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/pubsub"
	"github.com/aiocean/wireset/repository"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
)

var WebhookHandlerWireset = wire.NewSet(NewWebhookHandler)

type WebhookHandler struct {
	shopRepo *repository.ShopRepository
	pubsub   *pubsub.Pubsub
}

func NewWebhookHandler(
	shopRepo *repository.ShopRepository,
	pubsub *pubsub.Pubsub,
	fiberApp *fiber.App,
) *WebhookHandler {
	h := &WebhookHandler{
		shopRepo: shopRepo,
		pubsub:   pubsub,
	}

	shopGroup := fiberApp.Group("/webhook")
	{
		shopGroup.Get("/shopify/app-uninstalled", h.Uninstalled)
	}

	return h
}

func (s *WebhookHandler) Uninstalled(c *fiber.Ctx) error {
	s.pubsub.Send(c.UserContext(), &model.ShopUninstalledEvt{
		MyshopifyDomain: c.Query("shop"),
	})

	return c.SendStatus(http.StatusOK)
}
