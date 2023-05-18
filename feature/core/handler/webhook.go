package handler

import (
	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/pubsub"
	"github.com/aiocean/wireset/repository"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	ShopRepo *repository.ShopRepository
	Pubsub   *pubsub.Pubsub
	FiberApp *fiber.App
}

// Init
func (s *WebhookHandler) Register(fiberApp *fiber.App) {
	shopGroup := fiberApp.Group("/webhook")
	{
		shopGroup.Get("/shopify/app-uninstalled", s.Uninstalled)
	}
}

func (s *WebhookHandler) Uninstalled(c *fiber.Ctx) error {
	s.Pubsub.Send(c.UserContext(), &model.ShopUninstalledEvt{
		MyshopifyDomain: c.Query("shop"),
	})

	return c.SendStatus(http.StatusOK)
}
