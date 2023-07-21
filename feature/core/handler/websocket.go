package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"log"
	"sync"
)

type WebsocketHandler struct {
	connections map[string]*websocket.Conn
	mutex       sync.Mutex
}

func NewWebsocketHandler() *WebsocketHandler {
	return &WebsocketHandler{}
}

func (s *WebsocketHandler) CheckUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}

	return c.JSON(map[string]interface{}{
		"token": "df",
	})
}

func (s *WebsocketHandler) Handle(conn *websocket.Conn) {
	var (
		mt  int
		msg []byte
		err error
	)

	//add conn to pool
	s.mutex.Lock()
	s.connections["TODO"] = conn
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		delete(s.connections, "TODO")
		s.mutex.Unlock()
	}()

	for {
		if mt, msg, err = conn.ReadMessage(); err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", msg)

		// broadcast message to all connected sockets
		s.mutex.Lock()
		for _, c := range s.connections {
			if err = c.WriteMessage(mt, msg); err != nil {
				log.Println("write:", err)
				break
			}
		}
		s.mutex.Unlock()
	}
}
