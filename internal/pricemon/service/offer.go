package service

import (
	"context"
	"errors"
	"fmt"

	"pricemon/internal/pricemon/model"
	"pricemon/internal/pricemon/repository"
)

func (s *Service) ApplyOfferChanges(ctx context.Context, platformID int64, changes model.OfferChanges) (model.OfferBatchResult, error) {
	items, err := groupByItem(changes)
	if err != nil {
		return model.OfferBatchResult{}, err
	}

	var total model.OfferBatchResult
	for name, itemChanges := range items {
		result, err := s.repo.ApplyItemChanges(ctx, platformID, name, itemChanges)
		switch {
		case errors.Is(err, repository.ErrNoRelation):
			return model.OfferBatchResult{}, ErrUnknownPlatform
		case err != nil:
			return model.OfferBatchResult{}, fmt.Errorf("apply item changes: %w", err)
		}

		total.Created += result.Created
		total.Replaced += result.Replaced
		total.Deleted += result.Deleted
	}

	return total, nil
}

func groupByItem(changes model.OfferChanges) (map[string][]model.OfferChange, error) {
	items := make(map[string][]model.OfferChange)

	for _, offer := range changes.Set {
		change := model.OfferChange{Side: offer.Side, Price: offer.Price}
		if err := addChange(items, offer.ItemName, change); err != nil {
			return nil, err
		}
	}
	for _, key := range changes.Delete {
		change := model.OfferChange{Side: key.Side, Delete: true}
		if err := addChange(items, key.ItemName, change); err != nil {
			return nil, err
		}
	}

	return items, nil
}

func addChange(items map[string][]model.OfferChange, name string, change model.OfferChange) error {
	name, err := normalizeName(name)
	if err != nil || !change.Side.Valid() || change.Price < 0 {
		return ErrInvalidOffer
	}

	for _, existing := range items[name] {
		if existing.Side == change.Side {
			return ErrDuplicateOffer
		}
	}

	items[name] = append(items[name], change)
	return nil
}
