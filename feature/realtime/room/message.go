package room

type Message struct {
	Sender    string      `json:"sender,omitempty"`
	Recipient string      `json:"recipient"`
	Type      string      `json:"type,omitempty"`
	Meta      interface{} `json:"meta,omitempty"`
	Message   interface{} `json:"message,omitempty"`
}
