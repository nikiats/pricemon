package domain

import "encoding/json"

type OutboxMessage struct {
	ID      string
	Payload json.RawMessage
}
