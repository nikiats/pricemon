package repository

import (
	"errors"

	"gopricemon/internal/domain"
)

var (
	ErrPlatformAlreadyExists = errors.New("platform already exists")
	ErrPlatformNotFound      = errors.New("platform not found")
	ErrItemNotFound          = errors.New("item not found")
)

type Repository interface {
	SetOffer(offer domain.Offer) error
	GetPlatform(platformID int) (domain.Platform, error)
	GetItemID(categoryID int, name string) (int, error)
	CreateItem(categoryID int, name string) (int, error)
	GetSummary(categoryID, offset, limit int) ([]domain.ItemSummary, error)
	GetPlatforms() ([]domain.Platform, error)
	CreatePlatform(name string) (domain.Platform, error)
	RegeneratePlatformToken(platformID int) (string, error)
}
