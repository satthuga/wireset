package models

type WebsocketTopic string

// Stringer interface
func (t WebsocketTopic) String() string {
	return string(t)
}

type WebsocketMessage struct {
	Topic   WebsocketTopic `json:"topic"`
	Payload any            `json:"payload"`
}

const TopicError WebsocketTopic = "error"

type ErrorPayload struct {
	Message string `json:"message"`
}
