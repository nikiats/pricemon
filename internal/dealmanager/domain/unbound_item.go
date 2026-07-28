package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type UnboundItem struct {
	ID               int
	ItemID           int
	PurchasePrice    decimal.Decimal
	PurchasedAt      time.Time
	Sold             bool
	BuyTaskID        *int
	SequentialTaskID *int
}
