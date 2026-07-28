package domain

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type InboxEventStatus string

const (
	InboxEventStatusPending    InboxEventStatus = "PENDING"
	InboxEventStatusProcessing InboxEventStatus = "PROCESSING"
	InboxEventStatusProcessed  InboxEventStatus = "PROCESSED"
	InboxEventStatusFailed     InboxEventStatus = "FAILED"
)

type InboxEvent struct {
	ID          string
	Payload     json.RawMessage
	Status      InboxEventStatus
	Error       *string
	ReceivedAt  time.Time
	ProcessedAt *time.Time
	UpdatedAt   time.Time
}

type ItemSummaryPayload struct {
	CategoryID     int             `json:"categoryID"`
	ItemID         int             `json:"itemID"`
	PlatformSellID int             `json:"platformSellID"`
	PlatformBuyID  int             `json:"platformBuyID"`
	SellPrice      decimal.Decimal `json:"sellPrice"`
	BuyPrice       decimal.Decimal `json:"buyPrice"`
	ActualAt       time.Time       `json:"actualAt"`
	ExpiresAt      time.Time       `json:"expiresAt"`
}
