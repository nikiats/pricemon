package model

type Side string

const (
	SideBuy  Side = "buy"
	SideSell Side = "sell"
)

func (s Side) Valid() bool {
	return s == SideBuy || s == SideSell
}

type OfferKey struct {
	ItemName string
	Side     Side
}

type Offer struct {
	OfferKey
	Price int64
}

type PlatformOffer struct {
	PlatformID int64
	Side       Side
	Price      int64
}

type OfferChanges struct {
	Set    []Offer
	Delete []OfferKey
}

type OfferBatchResult struct {
	Created  int
	Replaced int
	Deleted  int
}
