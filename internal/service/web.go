package service

import (
	"errors"
	"strings"

	"github.com/shopspring/decimal"

	"gopricemon/internal/domain"
	"gopricemon/internal/repository"
)

var (
	ErrInvalidCategoryID     = errors.New("invalid category ID")
	ErrInvalidPagination     = errors.New("invalid pagination")
	ErrInvalidPlatformID     = errors.New("invalid platform ID")
	ErrInvalidPlatformName   = errors.New("invalid platform name")
	ErrPlatformAlreadyExists = errors.New("platform already exists")
	ErrPlatformNotFound      = errors.New("platform not found")
	ErrInvalidExecutorID     = errors.New("invalid executor ID")
	ErrInvalidExecutorName   = errors.New("invalid executor name")
	ErrExecutorAlreadyExists = errors.New("executor already exists")
	ErrExecutorNotFound      = errors.New("executor not found")
	ErrInvalidMinimumProfit  = errors.New("invalid minimum profit")
	ErrInvalidSummaryAge     = errors.New("invalid maximum summary age")
)

type Service struct {
	repo                repository.Repository
	taskLeaseMaxSeconds int
}

func NewService(repository repository.Repository, taskLeaseMaxSeconds int) *Service {
	return &Service{repo: repository, taskLeaseMaxSeconds: taskLeaseMaxSeconds}
}

func (s *Service) GetSummary(categoryID, offset, limit int, maxAge *int) (domain.SummaryPage, error) {
	if categoryID < 1 {
		return domain.SummaryPage{}, ErrInvalidCategoryID
	}
	if offset < 0 || limit < 1 || limit > 100 {
		return domain.SummaryPage{}, ErrInvalidPagination
	}
	if maxAge != nil && *maxAge < 0 {
		return domain.SummaryPage{}, ErrInvalidPagination
	}

	items, err := s.repo.GetSummary(categoryID, offset, limit+1, maxAge)
	if err != nil {
		return domain.SummaryPage{}, err
	}

	page := domain.SummaryPage{Items: items}
	if len(items) <= limit {
		return page, nil
	}

	page.Items = items[:limit]
	nextOffset := offset + limit
	page.NextOffset = &nextOffset

	return page, nil
}

func (s *Service) GetInventory() ([]domain.InventorySummary, error) {
	return s.repo.GetInventory()
}

func (s *Service) SetTradeSettings(minimumProfit decimal.Decimal, maximumSummaryAgeSecs int) error {
	if minimumProfit.IsNegative() {
		return ErrInvalidMinimumProfit
	}
	if maximumSummaryAgeSecs < 1 {
		return ErrInvalidSummaryAge
	}

	return s.repo.SetTradeSettings(minimumProfit, maximumSummaryAgeSecs)
}

func (s *Service) GetPlatforms() ([]domain.Platform, error) {
	return s.repo.GetPlatforms()
}

func (s *Service) CreatePlatform(name string) (domain.Platform, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Platform{}, ErrInvalidPlatformName
	}

	platform, err := s.repo.CreatePlatform(name)
	if errors.Is(err, repository.ErrPlatformAlreadyExists) {
		return domain.Platform{}, ErrPlatformAlreadyExists
	}

	return platform, err
}

func (s *Service) DeletePlatform(platformID int) error {
	if platformID < 1 {
		return ErrInvalidPlatformID
	}

	err := s.repo.DeletePlatform(platformID)
	if errors.Is(err, repository.ErrPlatformNotFound) {
		return ErrPlatformNotFound
	}

	return err
}

func (s *Service) RegeneratePlatformToken(platformID int) (string, error) {
	if platformID < 1 {
		return "", ErrInvalidPlatformID
	}

	return s.repo.RegeneratePlatformToken(platformID)
}

func (s *Service) GetExecutors() ([]domain.Executor, error) {
	return s.repo.GetExecutors()
}

func (s *Service) CreateExecutor(name string) (domain.Executor, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Executor{}, ErrInvalidExecutorName
	}

	executor, err := s.repo.CreateExecutor(name)
	if errors.Is(err, repository.ErrExecutorAlreadyExists) {
		return domain.Executor{}, ErrExecutorAlreadyExists
	}

	return executor, err
}

func (s *Service) DeleteExecutor(executorID int) error {
	if executorID < 1 {
		return ErrInvalidExecutorID
	}

	err := s.repo.DeleteExecutor(executorID)
	if errors.Is(err, repository.ErrExecutorNotFound) {
		return ErrExecutorNotFound
	}

	return err
}

func (s *Service) RegenerateExecutorToken(executorID int) (string, error) {
	if executorID < 1 {
		return "", ErrInvalidExecutorID
	}

	token, err := s.repo.RegenerateExecutorToken(executorID)
	if errors.Is(err, repository.ErrExecutorNotFound) {
		return "", ErrExecutorNotFound
	}

	return token, err
}
