package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type SequentialTaskStatus string

const (
	SequentialTaskStatusNotStarted SequentialTaskStatus = "NOT STARTED"
	SequentialTaskStatusBuying     SequentialTaskStatus = "BUYING"
	SequentialTaskStatusSelling    SequentialTaskStatus = "SELLING"
	SequentialTaskStatusCompleted  SequentialTaskStatus = "COMPLETED"
	SequentialTaskStatusFailed     SequentialTaskStatus = "FAILED"
)

type SequentialTask struct {
	ID             int
	InboxEventID   string
	CategoryID     int
	ItemID         int
	PlatformSellID int
	PlatformBuyID  int
	SellPrice      decimal.Decimal
	BuyPrice       decimal.Decimal
	BuyTaskID      *int
	SellTaskID     *int
	Status         SequentialTaskStatus
	Error          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
