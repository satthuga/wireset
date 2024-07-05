// Package wsresolver provides functionality to resolve WebSocket identity.
// Read more about the resolver package here: /feature/realtime/resolver/contract.go
package wsresolver

import (
	"errors"
	"github.com/aiocean/wireset/feature/realtime/resolver"
	"github.com/aiocean/wireset/shopifysvc"
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
)

const DefaultUsername = "default"

var JwtIdentityResolverset = wire.NewSet(
	wire.Bind(new(resolver.IdentityResolver), new(*JwtIdentityResolver)),
	wire.Struct(new(JwtIdentityResolver), "*"),
)

type JwtIdentityResolver struct {
	ShopifyConfig *shopifysvc.Config
}

func (j JwtIdentityResolver) Resolve(c *fiber.Ctx) (*resolver.Identity, error) {
	myshopifyDomain, ok := c.Locals("myshopifyDomain").(string)
	if !ok {
		return nil, errors.New("myshopifyDomain not found in context")
	}

	return &resolver.Identity{
		Username: DefaultUsername,
		Room:     myshopifyDomain,
	}, nil

}
