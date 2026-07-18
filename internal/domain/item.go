package domain

import "github.com/shopspring/decimal"

type Item struct {
	ID         int
	Name       string
	CategoryID int
}

type OfferSummary struct {
	ID              int
	Name            string
	CategoryID      int
	LowestSellPrice *decimal.Decimal
	LowestSellCount *int
	HighestBuyPrice *decimal.Decimal
	HighestBuyCount *int
}

type OfferPage struct {
	Items       []OfferSummary
	NextAfterID *int
}

type Category struct {
	ID   int
	Name string
}
