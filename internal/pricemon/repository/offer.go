package repository

import (
	"context"
	"fmt"

	"pricemon/internal/pricemon/model"
	"pricemon/internal/pricemon/repository/db"
)

func (r *Repository) ApplyOfferChanges(ctx context.Context, platformID int64, changes model.OfferChanges) (model.OfferBatchResult, error) {
	var result model.OfferBatchResult

	err := r.tx(ctx, func(q *db.Queries) error {
		result = model.OfferBatchResult{}
		items := make(map[int64]struct{}, len(changes.Set)+len(changes.Delete))

		for _, offer := range changes.Set {
			created, err := q.UpsertOffer(ctx, db.UpsertOfferParams{
				PlatformID: platformID,
				ItemID:     offer.ItemID,
				Side:       db.OfferSide(offer.Side),
				Price:      offer.Price,
			})
			if err != nil {
				if known := classify(err); known != nil {
					return known
				}
				return fmt.Errorf("upsert offer: %w", err)
			}

			if created {
				result.Created++
			} else {
				result.Replaced++
			}
			items[offer.ItemID] = struct{}{}
		}

		for _, key := range changes.Delete {
			affected, err := q.DeleteOffer(ctx, db.DeleteOfferParams{
				PlatformID: platformID,
				ItemID:     key.ItemID,
				Side:       db.OfferSide(key.Side),
			})
			if err != nil {
				return fmt.Errorf("delete offer: %w", err)
			}

			if affected > 0 {
				result.Deleted++
				items[key.ItemID] = struct{}{}
			}
		}

		for itemID := range items {
			if err := refreshBestDeal(ctx, q, itemID); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return model.OfferBatchResult{}, err
	}

	return result, nil
}
