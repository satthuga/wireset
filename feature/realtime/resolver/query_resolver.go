package resolver

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"github.com/pkg/errors"
)

type QueryResolver struct {
}

var QueryResolverWireset = wire.NewSet(
	wire.Bind(new(IdentityResolver), new(*QueryResolver)),
	wire.Struct(new(QueryResolver), "*"),
)

func (r *QueryResolver) Resolve(ctx *fiber.Ctx) (*Identity, error) {

	username := ctx.Query("username")
	if username == "" {
		return nil, errors.New("userName is required")
	}

	roomID := ctx.Query("roomID")
	if roomID == "" {
		return nil, errors.New("room is required")
	}

	return &Identity{
		Username: username,
		Room:     roomID,
	}, nil
}
