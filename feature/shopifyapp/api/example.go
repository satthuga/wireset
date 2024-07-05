package api

import (
	"github.com/gofiber/fiber/v2"
)

type ExampleHandler struct {
}

func (g *ExampleHandler) Handle(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}
