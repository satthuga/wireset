package handler

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/repository"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	ShopRepo *repository.ShopRepository
	EventBus *cqrs.EventBus
	FiberApp *fiber.App
}

func (s *WebhookHandler) Uninstalled(c *fiber.Ctx) error {
	s.EventBus.Publish(c.UserContext(), &model.ShopUninstalledEvt{
		MyshopifyDomain: c.Query("shop"),
	})

	return c.SendStatus(http.StatusOK)
}
