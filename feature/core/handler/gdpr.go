package handler

import "github.com/gofiber/fiber/v2"

type GdprHandler struct {
}

func (g *GdprHandler) CustomerDataRequest(ctx *fiber.Ctx) error {

	return ctx.SendStatus(200)
}

func (g *GdprHandler) CustomerRedact(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}

func (g *GdprHandler) ShopRedact(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}
