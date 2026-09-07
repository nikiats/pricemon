package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"pricemon/internal/pricemon/model"
	"pricemon/internal/pricemon/repository"
)

const maxNameLength = 200

func (s *Service) ListPlatforms(ctx context.Context) ([]model.Platform, error) {
	platforms, err := s.repo.ListPlatforms(ctx)
	if err != nil {
		return nil, fmt.Errorf("list platforms: %w", err)
	}

	return platforms, nil
}

func (s *Service) CreatePlatform(ctx context.Context, name string) (model.Platform, error) {
	name, err := normalizeName(name)
	if err != nil {
		return model.Platform{}, err
	}

	platform, err := s.repo.CreatePlatform(ctx, name)
	switch {
	case errors.Is(err, repository.ErrDuplicate):
		return model.Platform{}, ErrPlatformNameTaken
	case err != nil:
		return model.Platform{}, fmt.Errorf("create platform: %w", err)
	}

	return platform, nil
}

func (s *Service) DeletePlatform(ctx context.Context, id int64) error {
	err := s.repo.DeletePlatform(ctx, id)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return ErrNotFound
	case err != nil:
		return fmt.Errorf("delete platform: %w", err)
	}

	return nil
}

func normalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > maxNameLength {
		return "", ErrInvalidName
	}

	return name, nil
}
