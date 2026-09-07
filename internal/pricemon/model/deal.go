package model

type Deal struct {
	Item         Item
	BuyPlatform  Platform
	SellPlatform Platform
	BuyPrice     int64
	SellPrice    int64
}
