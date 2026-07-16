package repository

import (
	"gopricemon/internal/domain"
)

type Repository interface {
	SetOffer(offer domain.Offer) error
	GetItems(categoryID int) ([]domain.Item, error)
}
