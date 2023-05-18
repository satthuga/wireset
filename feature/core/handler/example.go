package handler

import (
	"github.com/gofiber/fiber/v2"
)

type ExampleHandler struct {
}

func (g *ExampleHandler) Example(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}

func (g *ExampleHandler) Register(fiberApp *fiber.App) {
	fiberApp.All("/metrics", g.Example)
}
