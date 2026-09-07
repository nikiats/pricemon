package repository

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"pricemon/internal/pricemon/model"
	"pricemon/internal/pricemon/repository/db"
)

type offerChange struct {
	key    model.OfferKey
	price  int64
	remove bool
}

func (r *Repository) ApplyOfferChanges(ctx context.Context, platformID int64, changes model.OfferChanges) (model.OfferBatchResult, error) {
	ordered := orderChanges(changes)

	var result model.OfferBatchResult

	err := r.tx(ctx, func(q *db.Queries) error {
		result = model.OfferBatchResult{}
		items := make([]int64, 0, len(ordered))

		for _, change := range ordered {
			if err := applyChange(ctx, q, platformID, change, &result); err != nil {
				return err
			}

			if len(items) == 0 || items[len(items)-1] != change.key.ItemID {
				items = append(items, change.key.ItemID)
			}
		}

		for _, itemID := range items {
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

func applyChange(ctx context.Context, q *db.Queries, platformID int64, change offerChange, result *model.OfferBatchResult) error {
	if change.remove {
		removed, err := q.DeleteOffer(ctx, db.DeleteOfferParams{
			PlatformID: platformID,
			ItemID:     change.key.ItemID,
			Side:       db.OfferSide(change.key.Side),
		})
		if err != nil {
			return fmt.Errorf("delete offer: %w", err)
		}

		if removed > 0 {
			result.Deleted++
		}

		return nil
	}

	created, err := q.UpsertOffer(ctx, db.UpsertOfferParams{
		PlatformID: platformID,
		ItemID:     change.key.ItemID,
		Side:       db.OfferSide(change.key.Side),
		Price:      change.price,
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

	return nil
}

func orderChanges(changes model.OfferChanges) []offerChange {
	ordered := make([]offerChange, 0, len(changes.Set)+len(changes.Delete))

	for _, offer := range changes.Set {
		ordered = append(ordered, offerChange{key: offer.OfferKey, price: offer.Price})
	}
	for _, key := range changes.Delete {
		ordered = append(ordered, offerChange{key: key, remove: true})
	}

	slices.SortFunc(ordered, func(a, b offerChange) int {
		if a.key.ItemID != b.key.ItemID {
			return cmp.Compare(a.key.ItemID, b.key.ItemID)
		}
		return cmp.Compare(a.key.Side, b.key.Side)
	})

	return ordered
}
