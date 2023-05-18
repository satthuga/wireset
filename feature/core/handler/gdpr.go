package handler

import "github.com/gofiber/fiber/v2"

type GdprHandler struct {
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

func (g *GdprHandler) Register(fiberApp *fiber.App) {
	fiberApp.Post("/gdpr/customers/data_request", g.customerDataRequest)
	fiberApp.Post("/gdpr/customers/redact", g.customerRedact)
	fiberApp.Post("/gdpr/shop/redact", g.shopRedact)
}
