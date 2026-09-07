package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"

	"pricemon/internal/pricemon/model"
	"pricemon/internal/pricemon/repository"
)

const collectorKeyPrefix = "pmc_"

func generateAPIKey() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate api key: %w", err)
	}

	return collectorKeyPrefix + base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashAPIKey(key string) []byte {
	sum := sha256.Sum256([]byte(key))
	return sum[:]
}

func (s *Service) IsServiceToken(token string) bool {
	return subtle.ConstantTimeCompare([]byte(token), s.serviceToken) == 1
}

func (s *Service) ResolveCollector(ctx context.Context, apiKey string) (model.Collector, error) {
	collector, err := s.repo.GetCollectorByAPIKeyHash(ctx, hashAPIKey(apiKey))
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return model.Collector{}, ErrNotFound
	case err != nil:
		return model.Collector{}, fmt.Errorf("resolve collector: %w", err)
	}

	return collector, nil
}
