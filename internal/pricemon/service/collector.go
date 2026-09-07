package service

import (
	"context"
	"errors"
	"fmt"

	"pricemon/internal/pricemon/model"
	"pricemon/internal/pricemon/repository"
)

func (s *Service) ListCollectors(ctx context.Context, platformID int64) ([]model.Collector, error) {
	if platformID == 0 {
		collectors, err := s.repo.ListCollectors(ctx)
		if err != nil {
			return nil, fmt.Errorf("list collectors: %w", err)
		}

		return collectors, nil
	}

	collectors, err := s.repo.ListCollectorsByPlatform(ctx, platformID)
	if err != nil {
		return nil, fmt.Errorf("list collectors by platform: %w", err)
	}

	return collectors, nil
}

func (s *Service) CreateCollector(ctx context.Context, platformID int64, name string) (model.CollectorWithKey, error) {
	name, err := normalizeName(name)
	if err != nil {
		return model.CollectorWithKey{}, err
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		return model.CollectorWithKey{}, err
	}

	collector, err := s.repo.CreateCollector(ctx, platformID, name, hashAPIKey(apiKey))
	switch {
	case errors.Is(err, repository.ErrNoRelation):
		return model.CollectorWithKey{}, ErrUnknownPlatform
	case err != nil:
		return model.CollectorWithKey{}, fmt.Errorf("create collector: %w", err)
	}

	return model.CollectorWithKey{Collector: collector, APIKey: apiKey}, nil
}

func (s *Service) DeleteCollector(ctx context.Context, id string) error {
	err := s.repo.DeleteCollector(ctx, id)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return ErrNotFound
	case err != nil:
		return fmt.Errorf("delete collector: %w", err)
	}

	return nil
}

func (s *Service) RotateCollectorAPIKey(ctx context.Context, id string) (string, error) {
	apiKey, err := generateAPIKey()
	if err != nil {
		return "", err
	}

	_, err = s.repo.SetCollectorAPIKey(ctx, id, hashAPIKey(apiKey))
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return "", ErrNotFound
	case err != nil:
		return "", fmt.Errorf("rotate collector api key: %w", err)
	}

	return apiKey, nil
}
