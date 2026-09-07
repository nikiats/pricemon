package handler

import (
	"net/http"
)

type item struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type deal struct {
	Item         item     `json:"item"`
	BuyPlatform  platform `json:"buyPlatform"`
	SellPlatform platform `json:"sellPlatform"`
	BuyPrice     int64    `json:"buyPrice"`
	SellPrice    int64    `json:"sellPrice"`
}

func (h *Handler) ListDeals(w http.ResponseWriter, r *http.Request) {
	found, err := h.service.ListDeals(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	deals := make([]deal, 0, len(found))
	for _, source := range found {
		deals = append(deals, deal{
			Item:         item{ID: source.Item.ID, Name: source.Item.Name},
			BuyPlatform:  platform{ID: source.BuyPlatform.ID, Name: source.BuyPlatform.Name},
			SellPlatform: platform{ID: source.SellPlatform.ID, Name: source.SellPlatform.Name},
			BuyPrice:     source.BuyPrice,
			SellPrice:    source.SellPrice,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"deals": deals})
}
