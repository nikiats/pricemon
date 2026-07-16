package service

import (
	"gopricemon/internal/domain"
	"gopricemon/internal/repository"
)

type Service struct {
	repo repository.Repository
}

func NewService(repository repository.Repository) *Service {
	return &Service{
		repo: repository,
	}
}

func (s *Service) GetItems(categoryID int) ([]domain.Item, error) {
	return s.repo.GetItems(categoryID)
}

func (s *Service) SetOffer(offer domain.Offer) error {
	return s.repo.SetOffer(offer)
}
