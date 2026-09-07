package service

import (
	"context"
	"errors"
	"fmt"

	"pricemon/internal/pricemon/model"
	"pricemon/internal/pricemon/repository"
)

func (s *Service) ApplyOfferChanges(ctx context.Context, platformID int64, changes model.OfferChanges) (model.OfferBatchResult, error) {
	prepared, err := prepareChanges(changes)
	if err != nil {
		return model.OfferBatchResult{}, err
	}

	result, err := s.repo.ApplyOfferChanges(ctx, platformID, prepared)
	switch {
	case errors.Is(err, repository.ErrNoRelation):
		return model.OfferBatchResult{}, ErrUnknownPlatform
	case err != nil:
		return model.OfferBatchResult{}, fmt.Errorf("apply offer changes: %w", err)
	}

	return result, nil
}

func prepareChanges(changes model.OfferChanges) (model.OfferChanges, error) {
	prepared := model.OfferChanges{
		Set:    make([]model.Offer, 0, len(changes.Set)),
		Delete: make([]model.OfferKey, 0, len(changes.Delete)),
	}
	seen := make(map[model.OfferKey]struct{}, len(changes.Set)+len(changes.Delete))

	for _, offer := range changes.Set {
		key, err := prepareKey(offer.OfferKey, seen)
		if err != nil {
			return model.OfferChanges{}, err
		}
		if offer.Price < 0 {
			return model.OfferChanges{}, ErrInvalidOffer
		}

		prepared.Set = append(prepared.Set, model.Offer{OfferKey: key, Price: offer.Price})
	}

	for _, key := range changes.Delete {
		key, err := prepareKey(key, seen)
		if err != nil {
			return model.OfferChanges{}, err
		}

		prepared.Delete = append(prepared.Delete, key)
	}

	return prepared, nil
}

func prepareKey(key model.OfferKey, seen map[model.OfferKey]struct{}) (model.OfferKey, error) {
	name, err := normalizeName(key.ItemName)
	if err != nil || !key.Side.Valid() {
		return model.OfferKey{}, ErrInvalidOffer
	}

	key.ItemName = name
	if _, duplicate := seen[key]; duplicate {
		return model.OfferKey{}, ErrDuplicateOffer
	}
	seen[key] = struct{}{}

	return key, nil
}
