package room

import "github.com/gofiber/contrib/websocket"

type Member struct {
	Name       string
	connection *websocket.Conn
}

func NewMember(name string, conn *websocket.Conn) *Member {
	return &Member{
		Name:       name,
		connection: conn,
	}
}

func (m *Member) Send(message interface{}) error {
	if messageBytes, ok := message.([]byte); ok {
		return m.connection.WriteMessage(websocket.TextMessage, messageBytes)
	}

	// check if it's string
	if messageString, ok := message.(string); ok {
		return m.connection.WriteMessage(websocket.TextMessage, []byte(messageString))
	}

	return m.connection.WriteJSON(message)
}

func (m *Member) Close() error {
	return m.connection.Close()
}

func (m *Member) ReadMessage() (int, []byte, error) {
	return m.connection.ReadMessage()
}

func (m *Member) WriteMessage(messageType int, data []byte) error {
	return m.connection.WriteMessage(messageType, data)
}
