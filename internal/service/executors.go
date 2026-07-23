package service

import (
	"errors"
	"strings"

	"gopricemon/internal/domain"
)

var (
	ErrInvalidInventory      = errors.New("invalid inventory")
	ErrInsufficientInventory = errors.New("insufficient inventory")
)

type inventoryKey struct {
	categoryID int
	itemName   string
}

func (s *Service) ReplaceInventory(items []domain.InventoryItem) error {
	seen := make(map[inventoryKey]struct{}, len(items))

	for i := range items {
		items[i].ItemName = strings.TrimSpace(items[i].ItemName)
		if items[i].CategoryID < 1 || items[i].ItemName == "" || items[i].Quantity < 0 {
			return ErrInvalidInventory
		}

		key := inventoryKey{items[i].CategoryID, items[i].ItemName}
		if _, exists := seen[key]; exists {
			return ErrInvalidInventory
		}
		seen[key] = struct{}{}
	}

	return s.repo.ReplaceInventory(items)
}

func (s *Service) ChangeInventory(items []domain.InventoryDelta) error {
	seen := make(map[inventoryKey]struct{}, len(items))

	for i := range items {
		items[i].ItemName = strings.TrimSpace(items[i].ItemName)
		if items[i].CategoryID < 1 || items[i].ItemName == "" || items[i].Delta == 0 {
			return ErrInvalidInventory
		}

		key := inventoryKey{items[i].CategoryID, items[i].ItemName}
		if _, exists := seen[key]; exists {
			return ErrInvalidInventory
		}
		seen[key] = struct{}{}
	}

	updated, err := s.repo.ChangeInventory(items)
	if err != nil {
		return err
	}
	if !updated {
		return ErrInsufficientInventory
	}

	return nil
}
