package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Item struct {
	ID         int
	Name       string
	CategoryID int
}

type ItemSummary struct {
	ID               int
	Name             string
	CategoryID       int
	SellPrice        *decimal.Decimal
	SellCount        *int
	SellPlatformName *string
	SellURL          *string
	BuyPrice         *decimal.Decimal
	BuyCount         *int
	BuyPlatformName  *string
	BuyURL           *string
	ActualAt         *time.Time
}

type SummaryPage struct {
	Items      []ItemSummary
	NextOffset *int
}

type Category struct {
	ID   int
	Name string
}
