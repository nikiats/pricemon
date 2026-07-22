package service

import (
	"crypto/subtle"
	"errors"
	"strings"

	"github.com/shopspring/decimal"

	"gopricemon/internal/domain"
	"gopricemon/internal/repository"
)

var (
	ErrInvalidCategoryID     = errors.New("invalid category ID")
	ErrInvalidOffer          = errors.New("invalid offer")
	ErrInvalidPagination     = errors.New("invalid pagination")
	ErrInvalidPlatformToken  = errors.New("invalid platform token")
	ErrInvalidPlatformID     = errors.New("invalid platform ID")
	ErrInvalidPlatformName   = errors.New("invalid platform name")
	ErrPlatformAlreadyExists = errors.New("platform already exists")
)

type Service struct {
	repo repository.Repository
}

func NewService(repository repository.Repository) *Service {
	return &Service{
		repo: repository,
	}
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

func (s *Service) SetOffer(offer domain.Offer) error {
	offer.ItemName = strings.TrimSpace(offer.ItemName)
	offer.PlatformToken = strings.TrimSpace(offer.PlatformToken)
	if offer.CategoryID < 1 ||
		offer.ItemName == "" ||
		offer.PlatformID < 1 ||
		offer.PlatformToken == "" ||
		offer.Count < 0 ||
		offer.Price.LessThanOrEqual(decimal.Zero) ||
		(offer.Side != domain.SideSell && offer.Side != domain.SideBuy) {
		return ErrInvalidOffer
	}

	platform, err := s.repo.GetPlatform(offer.PlatformID)
	if errors.Is(err, repository.ErrPlatformNotFound) {
		return ErrInvalidPlatformToken
	}
	if err != nil {
		return err
	}
	if subtle.ConstantTimeCompare([]byte(offer.PlatformToken), []byte(platform.Token)) != 1 {
		return ErrInvalidPlatformToken
	}

	itemID, err := s.repo.GetOrCreateItem(offer.CategoryID, offer.ItemName)
	if err != nil {
		return err
	}

	return s.repo.SetOffer(itemID, offer)
}

func (s *Service) SetZeroCount(categoryID int, itemName string, platformID int, platformToken string) (bool, error) {
	itemName = strings.TrimSpace(itemName)
	platformToken = strings.TrimSpace(platformToken)
	if categoryID < 1 || itemName == "" || platformID < 1 || platformToken == "" {
		return false, ErrInvalidOffer
	}

	platform, err := s.repo.GetPlatform(platformID)
	if errors.Is(err, repository.ErrPlatformNotFound) {
		return false, ErrInvalidPlatformToken
	}
	if err != nil {
		return false, err
	}
	if subtle.ConstantTimeCompare([]byte(platformToken), []byte(platform.Token)) != 1 {
		return false, ErrInvalidPlatformToken
	}

	itemID, err := s.repo.GetOrCreateItem(categoryID, itemName)
	if err != nil {
		return false, err
	}

	return s.repo.SetZeroCount(itemID, platformID)
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

func (s *Service) RegeneratePlatformToken(platformID int) (string, error) {
	if platformID < 1 {
		return "", ErrInvalidPlatformID
	}

	return s.repo.RegeneratePlatformToken(platformID)
}
