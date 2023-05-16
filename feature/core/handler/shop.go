package handler

import (
	"api/pkg/repository"
	"github.com/gofiber/fiber/v2"
	"net/http"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/google/wire"
)

var ShopHandlerWireset = wire.NewSet(NewShopHandler)

type ShopHandler struct {
	shopRepo *repository.ShopRepository
	cqrsSvc  *cqrs.Facade
}

func NewShopHandler(
	shopRepo *repository.ShopRepository,
	cqrsSvc *cqrs.Facade,
	fiberApp *fiber.App,
) *ShopHandler {
	h := &ShopHandler{
		shopRepo: shopRepo,
		cqrsSvc:  cqrsSvc,
	}

	h.Register(fiberApp)

	return h
}

func (ctrl *ShopHandler) Register(fiberApp *fiber.App) {
	shopGroup := fiberApp.Group("/shop")
	{
		shopGroup.Get("/:id", ctrl.GetDetails)
	}
}

func (ctrl *ShopHandler) GetDetails(ctx *fiber.Ctx) error {

	currentShop := ctx.Query("shop")
	if currentShop == "" {
		return fiber.NewError(http.StatusInternalServerError, "shop not found")
	}

	return ctx.JSON(map[string]interface{}{
		"shop": currentShop,
	})
}
