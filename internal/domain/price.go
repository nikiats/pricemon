package domain

import "github.com/shopspring/decimal"

type OfferSide string

const (
	SideSell OfferSide = "S"
	SideBuy  OfferSide = "B"
)

type Offer struct {
	ItemID        int
	CategoryID    int
	ItemName      string
	PlatformID    int
	PlatformToken string
	Price         decimal.Decimal
	Count         int
	Side          OfferSide
	URL           *string
}

type Platform struct {
	ID    int
	Name  string
	Token string
}
