package domain

import "github.com/shopspring/decimal"

type Item struct {
	ID         int
	Name       string
	CategoryID int
}

type ItemSummary struct {
	ID                   int
	Name                 string
	CategoryID           int
	BestSellPrice        *decimal.Decimal
	BestSellCount        *int
	BestSellPlatformName *string
	BestBuyPrice         *decimal.Decimal
	BestBuyCount         *int
	BestBuyPlatformName  *string
}

type SummaryPage struct {
	Items      []ItemSummary
	NextOffset *int
}

type Category struct {
	ID   int
	Name string
}
