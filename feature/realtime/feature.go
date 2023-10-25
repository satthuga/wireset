package realtime

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/realtime/command"
	"github.com/aiocean/wireset/feature/realtime/event"
	"github.com/aiocean/wireset/feature/realtime/handler"
	"github.com/aiocean/wireset/feature/realtime/room"
	"github.com/aiocean/wireset/fiberapp"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	wire.Struct(new(FeatureRealtime), "*"),
	handler.NewWebsocketHandler,
	room.NewRoomManager,
	wire.Struct(new(command.ExampleHandler), "*"),
	wire.Struct(new(event.ExampleHandler), "*"),
)

type FeatureRealtime struct {
	HttpRegistry     *fiberapp.Registry
	WebsocketHandler *handler.WebsocketHandler

	CommandProcessor *cqrs.CommandProcessor
	EventProcessor   *cqrs.EventProcessor

	ExampleCommandHandler *command.ExampleHandler
	ExampleEventHandler   *event.ExampleHandler
}

func (f *FeatureRealtime) Init() error {
	if err := f.CommandProcessor.AddHandlers(f.ExampleCommandHandler); err != nil {
		return err
	}
	if err := f.EventProcessor.AddHandlers(f.ExampleEventHandler); err != nil {
		return err
	}

	f.HttpRegistry.AddHttpMiddleware("/api/v1/ws", f.WebsocketHandler.Upgrade)
	f.HttpRegistry.AddHttpHandlers([]*fiberapp.HttpHandler{
		{
			Method: fiber.MethodGet,
			Path:   "/api/v1/ws",
			Handlers: []fiber.Handler{
				websocket.New(f.WebsocketHandler.Handle),
			},
		},
		{
			Method: fiber.MethodPost,
			Path:   "/api/v1/dm",
			Handlers: []fiber.Handler{
				f.WebsocketHandler.SendDm,
			},
		},
	})
	return nil
}
