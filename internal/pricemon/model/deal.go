package model

type Deal struct {
	Item         Item
	BuyPlatform  Platform
	SellPlatform Platform
	BuyPrice     int64
	SellPrice    int64
}

type BestPair struct {
	ItemID         int64
	BuyPlatformID  int64
	SellPlatformID int64
	BuyPrice       int64
	SellPrice      int64
}

func (p BestPair) Profit() int64 {
	return p.SellPrice - p.BuyPrice
}

func FindBestPair(itemID int64, offers []PlatformOffer) (BestPair, bool) {
	var best BestPair
	found := false

	for _, ask := range offers {
		if ask.Side != SideSell {
			continue
		}

		for _, bid := range offers {
			if bid.Side != SideBuy || bid.PlatformID == ask.PlatformID {
				continue
			}

			pair := BestPair{
				ItemID:         itemID,
				BuyPlatformID:  ask.PlatformID,
				SellPlatformID: bid.PlatformID,
				BuyPrice:       ask.Price,
				SellPrice:      bid.Price,
			}
			if pair.Profit() <= 0 {
				continue
			}
			if found && pair.Profit() <= best.Profit() {
				continue
			}

			best = pair
			found = true
		}
	}

	return best, found
}
