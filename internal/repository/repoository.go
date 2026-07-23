package repository

import (
	"errors"

	"gopricemon/internal/domain"
)

var (
	ErrPlatformAlreadyExists = errors.New("platform already exists")
	ErrPlatformNotFound      = errors.New("platform not found")
)

type Repository interface {
	SetOffer(itemID int, offer domain.Offer) error
	SetZeroCount(itemID, platformID int) (bool, error)
	GetPlatform(platformID int) (domain.Platform, error)
	GetOrCreateItem(categoryID int, name string) (int, error)
	GetSummary(categoryID, offset, limit int, maxAge *int) ([]domain.ItemSummary, error)
	GetPlatforms() ([]domain.Platform, error)
	CreatePlatform(name string) (domain.Platform, error)
	DeletePlatform(platformID int) error
	RegeneratePlatformToken(platformID int) (string, error)
}
