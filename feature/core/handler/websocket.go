package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/wire"
	"log"
)

type WebsocketHandler struct {
}

func NewWebsocketHandler() *WebsocketHandler {
	return &WebsocketHandler{}
}

var WebsocketHandlerWireset = wire.NewSet(NewWebsocketHandler)

func (s *WebsocketHandler) Register(fiberApp *fiber.App) {
	fiberApp.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}

		return c.JSON(map[string]interface{}{
			"token": "df",
		})
	})

	fiberApp.Get("/ws/:id", websocket.New(func(conn *websocket.Conn) {
		var (
			mt  int
			msg []byte
			err error
		)
		// add conn to pool

		for {
			if mt, msg, err = conn.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)

			if err = conn.WriteMessage(mt, msg); err != nil {
				log.Println("write:", err)
				break
			}
		}

	}))
}
