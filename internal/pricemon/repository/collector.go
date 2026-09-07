package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"pricemon/internal/pricemon/model"
	"pricemon/internal/pricemon/repository/db"
)

func (r *Repository) ListCollectors(ctx context.Context) ([]model.Collector, error) {
	rows, err := r.q.ListCollectors(ctx)
	if err != nil {
		return nil, fmt.Errorf("list collectors: %w", err)
	}

	return toCollectors(rows), nil
}

func (r *Repository) ListCollectorsByPlatform(ctx context.Context, platformID int64) ([]model.Collector, error) {
	rows, err := r.q.ListCollectorsByPlatform(ctx, platformID)
	if err != nil {
		return nil, fmt.Errorf("list collectors by platform: %w", err)
	}

	return toCollectors(rows), nil
}

func (r *Repository) CreateCollector(ctx context.Context, platformID int64, name string, keyHash []byte) (model.Collector, error) {
	row, err := r.q.CreateCollector(ctx, db.CreateCollectorParams{
		PlatformID: platformID,
		Name:       name,
		ApiKeyHash: keyHash,
	})
	if err != nil {
		if known := classify(err); known != nil {
			return model.Collector{}, known
		}
		return model.Collector{}, fmt.Errorf("create collector: %w", err)
	}

	return toCollector(row), nil
}

func (r *Repository) DeleteCollector(ctx context.Context, id string) error {
	collectorID, err := uuid.Parse(id)
	if err != nil {
		return ErrNotFound
	}

	affected, err := r.q.DeleteCollector(ctx, collectorID)
	if err != nil {
		return fmt.Errorf("delete collector: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *Repository) SetCollectorAPIKey(ctx context.Context, id string, keyHash []byte) (model.Collector, error) {
	collectorID, err := uuid.Parse(id)
	if err != nil {
		return model.Collector{}, ErrNotFound
	}

	row, err := r.q.SetCollectorAPIKey(ctx, db.SetCollectorAPIKeyParams{
		ID:         collectorID,
		ApiKeyHash: keyHash,
	})
	if err != nil {
		if known := classify(err); known != nil {
			return model.Collector{}, known
		}
		return model.Collector{}, fmt.Errorf("set collector api key: %w", err)
	}

	return toCollector(row), nil
}

func (r *Repository) GetCollectorByAPIKeyHash(ctx context.Context, keyHash []byte) (model.Collector, error) {
	row, err := r.q.GetCollectorByAPIKeyHash(ctx, keyHash)
	if err != nil {
		if known := classify(err); known != nil {
			return model.Collector{}, known
		}
		return model.Collector{}, fmt.Errorf("get collector by api key hash: %w", err)
	}

	return toCollector(row), nil
}

func toCollectors(rows []db.Collector) []model.Collector {
	collectors := make([]model.Collector, 0, len(rows))
	for _, row := range rows {
		collectors = append(collectors, toCollector(row))
	}

	return collectors
}

func toCollector(row db.Collector) model.Collector {
	return model.Collector{
		ID:         row.ID.String(),
		PlatformID: row.PlatformID,
		Name:       row.Name,
		CreatedAt:  row.CreatedAt.Time,
	}
}
