package domain

import (
	"encoding/json"
	"time"
)

type InboxEventStatus string

const (
	InboxEventStatusPending    InboxEventStatus = "PENDING"
	InboxEventStatusProcessing InboxEventStatus = "PROCESSING"
	InboxEventStatusProcessed  InboxEventStatus = "PROCESSED"
	InboxEventStatusFailed     InboxEventStatus = "FAILED"
)

type InboxEvent struct {
	ID            string
	Type          string
	Payload       json.RawMessage
	Status        InboxEventStatus
	Attempts      int
	Error         *string
	ReceivedAt    time.Time
	NextAttemptAt time.Time
	LeaseUntil    *time.Time
	ProcessedAt   *time.Time
	UpdatedAt     time.Time
}
