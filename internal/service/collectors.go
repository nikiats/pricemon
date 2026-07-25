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
	ErrInvalidOffer         = errors.New("invalid offer")
	ErrInvalidPlatformToken = errors.New("invalid platform token")
)

func (s *Service) SetOffer(offer domain.Offer) error {
	return s.SetOffers([]domain.Offer{offer})
}

func (s *Service) SetOffers(offers []domain.Offer) error {
	platforms := make(map[int]domain.Platform)
	for i := range offers {
		offer := &offers[i]
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

		platform, ok := platforms[offer.PlatformID]
		if !ok {
			var err error
			platform, err = s.repo.GetPlatform(offer.PlatformID)
			if errors.Is(err, repository.ErrPlatformNotFound) {
				return ErrInvalidPlatformToken
			}
			if err != nil {
				return err
			}
			platforms[offer.PlatformID] = platform
		}
		if subtle.ConstantTimeCompare([]byte(offer.PlatformToken), []byte(platform.Token)) != 1 {
			return ErrInvalidPlatformToken
		}
	}

	return s.repo.SetOffers(offers)
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
