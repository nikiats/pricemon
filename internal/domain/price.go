package domain

import "github.com/shopspring/decimal"

type OfferSide string

const (
	SideSell OfferSide = "S"
	SideBuy  OfferSide = "B"
)

type Offer struct {
	ID         int
	ItemID     int
	PlatformID int
	Price      decimal.Decimal
	Side       OfferSide
}

type Platform struct {
	ID   int
	Name string
}
