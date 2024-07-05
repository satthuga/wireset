package resolver

import "github.com/gofiber/fiber/v2"

type Identity struct {
	Username string `json:"username"`
	Room     string `json:"room"`
}

type IdentityResolver interface {
	Resolve(ctx *fiber.Ctx) (*Identity, error)
}
