package repository

import (
	"context"
	"fmt"

	"pricemon/internal/pricemon/model"
	"pricemon/internal/pricemon/repository/db"
)

func (r *Repository) ApplyItemChanges(ctx context.Context, platformID int64, itemName string, changes []model.OfferChange) (model.OfferBatchResult, error) {
	var result model.OfferBatchResult

	err := r.tx(ctx, func(q *db.Queries) error {
		itemID, err := q.UpsertItem(ctx, itemName)
		if err != nil {
			return fmt.Errorf("upsert item: %w", err)
		}

		for _, change := range changes {
			if change.Delete {
				deleted, err := q.DeleteOffer(ctx, db.DeleteOfferParams{
					PlatformID: platformID,
					ItemID:     itemID,
					Side:       db.OfferSide(change.Side),
				})
				if err != nil {
					return fmt.Errorf("delete offer: %w", err)
				}
				result.Deleted += int(deleted)
				continue
			}

			created, err := q.UpsertOffer(ctx, db.UpsertOfferParams{
				PlatformID: platformID,
				ItemID:     itemID,
				Side:       db.OfferSide(change.Side),
				Price:      change.Price,
			})
			if err != nil {
				return classify(err)
			}

			if created {
				result.Created++
			} else {
				result.Replaced++
			}
		}

		return refreshBestDeal(ctx, q, itemID)
	})
	if err != nil {
		return model.OfferBatchResult{}, err
	}

	return result, nil
}
