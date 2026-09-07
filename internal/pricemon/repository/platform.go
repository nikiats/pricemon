package repository

import (
	"context"
	"fmt"

	"pricemon/internal/pricemon/model"
	"pricemon/internal/pricemon/repository/db"
)

func (r *Repository) ListPlatforms(ctx context.Context) ([]model.Platform, error) {
	rows, err := r.q.ListPlatforms(ctx)
	if err != nil {
		return nil, fmt.Errorf("list platforms: %w", err)
	}

	platforms := make([]model.Platform, 0, len(rows))
	for _, row := range rows {
		platforms = append(platforms, toPlatform(row))
	}

	return platforms, nil
}

func (r *Repository) CreatePlatform(ctx context.Context, name string) (model.Platform, error) {
	row, err := r.q.CreatePlatform(ctx, name)
	if err != nil {
		if known := classify(err); known != nil {
			return model.Platform{}, known
		}
		return model.Platform{}, fmt.Errorf("create platform: %w", err)
	}

	return toPlatform(row), nil
}

func (r *Repository) DeletePlatform(ctx context.Context, id int64) error {
	return r.tx(ctx, func(q *db.Queries) error {
		items, err := q.ListItemsOfPlatform(ctx, id)
		if err != nil {
			return fmt.Errorf("list items of platform: %w", err)
		}

		affected, err := q.DeletePlatform(ctx, id)
		if err != nil {
			return fmt.Errorf("delete platform: %w", err)
		}
		if affected == 0 {
			return ErrNotFound
		}

		for _, itemID := range items {
			if err := refreshBestDeal(ctx, q, itemID); err != nil {
				return err
			}
		}

		return nil
	})
}

func toPlatform(row db.Platform) model.Platform {
	return model.Platform{
		ID:   row.ID,
		Name: row.Name,
	}
}
