package service

import (
	"context"
	"errors"
	"fmt"

	"pricemon/internal/pricemon/model"
	"pricemon/internal/pricemon/repository"
)

func (s *Service) ApplyOfferChanges(ctx context.Context, platformID int64, changes model.OfferChanges) (model.OfferBatchResult, error) {
	if err := validateChanges(changes); err != nil {
		return model.OfferBatchResult{}, err
	}

	result, err := s.repo.ApplyOfferChanges(ctx, platformID, changes)
	switch {
	case errors.Is(err, repository.ErrNoRelation):
		return model.OfferBatchResult{}, ErrUnknownItem
	case err != nil:
		return model.OfferBatchResult{}, fmt.Errorf("apply offer changes: %w", err)
	}

	return result, nil
}

func validateChanges(changes model.OfferChanges) error {
	seen := make(map[model.OfferKey]struct{}, len(changes.Set)+len(changes.Delete))

	for _, offer := range changes.Set {
		if !offer.Side.Valid() || offer.ItemID < 1 || offer.Price < 0 {
			return ErrInvalidOffer
		}
		if _, duplicate := seen[offer.OfferKey]; duplicate {
			return ErrDuplicateOffer
		}
		seen[offer.OfferKey] = struct{}{}
	}

	for _, key := range changes.Delete {
		if !key.Side.Valid() || key.ItemID < 1 {
			return ErrInvalidOffer
		}
		if _, duplicate := seen[key]; duplicate {
			return ErrDuplicateOffer
		}
		seen[key] = struct{}{}
	}

	return nil
}
