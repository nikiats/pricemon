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
			itemID, touched, err := applyChange(ctx, q, platformID, change, &result)
			if err != nil {
				return err
			}

			if touched && !slices.Contains(items, itemID) {
				items = append(items, itemID)
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

func applyChange(ctx context.Context, q *db.Queries, platformID int64, change offerChange, result *model.OfferBatchResult) (int64, bool, error) {
	if change.remove {
		removed, err := q.DeleteOffer(ctx, db.DeleteOfferParams{
			PlatformID: platformID,
			Name:       change.key.ItemName,
			Side:       db.OfferSide(change.key.Side),
		})
		if err != nil {
			return 0, false, fmt.Errorf("delete offer: %w", err)
		}
		if len(removed) == 0 {
			return 0, false, nil
		}

		result.Deleted++

		return removed[0], true, nil
	}

	itemID, err := q.UpsertItem(ctx, change.key.ItemName)
	if err != nil {
		return 0, false, fmt.Errorf("upsert item: %w", err)
	}

	created, err := q.UpsertOffer(ctx, db.UpsertOfferParams{
		PlatformID: platformID,
		ItemID:     itemID,
		Side:       db.OfferSide(change.key.Side),
		Price:      change.price,
	})
	if err != nil {
		if known := classify(err); known != nil {
			return 0, false, known
		}
		return 0, false, fmt.Errorf("upsert offer: %w", err)
	}

	if created {
		result.Created++
	} else {
		result.Replaced++
	}

	return itemID, true, nil
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
		if a.key.ItemName != b.key.ItemName {
			return cmp.Compare(a.key.ItemName, b.key.ItemName)
		}
		return cmp.Compare(a.key.Side, b.key.Side)
	})

	return ordered
}
