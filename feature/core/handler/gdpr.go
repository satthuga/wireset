package handler

import "github.com/gofiber/fiber/v2"

type GdprHandler struct {
}

func NewGdprHandler(
	fiberApp *fiber.App,
) *GdprHandler {

	handler := &GdprHandler{}

	fiberApp.Post("/gdpr/customers/data_request", handler.customerDataRequest)
	fiberApp.Post("/gdpr/customers/redact", handler.customerRedact)
	fiberApp.Post("/gdpr/shop/redact", handler.shopRedact)

	return handler
}

func (g *GdprHandler) customerDataRequest(ctx *fiber.Ctx) error {

	return ctx.SendStatus(200)
}

func (g *GdprHandler) customerRedact(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}

func (g *GdprHandler) shopRedact(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}
