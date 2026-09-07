package service

import (
	"context"
	"fmt"

	"pricemon/internal/pricemon/model"
)

func (s *Service) ListDeals(ctx context.Context) ([]model.Deal, error) {
	deals, err := s.repo.ListDeals(ctx)
	if err != nil {
		return nil, fmt.Errorf("list deals: %w", err)
	}

	return deals, nil
}
