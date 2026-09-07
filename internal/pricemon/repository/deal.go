package repository

import (
	"context"
	"fmt"

	"pricemon/internal/pricemon/model"
)

func (r *Repository) ListDeals(ctx context.Context) ([]model.Deal, error) {
	rows, err := r.q.ListDeals(ctx)
	if err != nil {
		return nil, fmt.Errorf("list deals: %w", err)
	}

	deals := make([]model.Deal, 0, len(rows))
	for _, row := range rows {
		deals = append(deals, model.Deal{
			Item:         model.Item{ID: row.ItemID, Name: row.ItemName},
			BuyPlatform:  model.Platform{ID: row.BuyPlatformID, Name: row.BuyPlatformName},
			SellPlatform: model.Platform{ID: row.SellPlatformID, Name: row.SellPlatformName},
			BuyPrice:     row.BuyPrice,
			SellPrice:    row.SellPrice,
		})
	}

	return deals, nil
}
