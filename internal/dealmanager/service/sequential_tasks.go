package service

import (
	"gopricemon/internal/dealmanager/domain"
	"gopricemon/internal/dealmanager/repository"
)

type SequentialTasksService struct {
	repo repository.Repository
}

func NewSequentialTasksService(repo repository.Repository) *SequentialTasksService {
	return &SequentialTasksService{repo: repo}
}

func (s *SequentialTasksService) GetSequentialTasks(offset, limit int) (domain.SequentialTaskPage, error) {
	tasks, err := s.repo.GetSequentialTasks(offset, limit+1)
	if err != nil {
		return domain.SequentialTaskPage{}, err
	}

	page := domain.SequentialTaskPage{}
	if len(tasks) > limit {
		tasks = tasks[:limit]
		nextOffset := offset + limit
		page.NextOffset = &nextOffset
	}
	page.Items = make([]domain.SequentialTaskInfo, len(tasks))
	for i, task := range tasks {
		page.Items[i] = domain.SequentialTaskInfo{
			ID:                 task.ID,
			ItemID:             task.ItemID,
			PurchasePlatformID: task.PlatformSellID,
			PurchasePrice:      task.SellPrice,
			SalePlatformID:     task.PlatformBuyID,
			SalePrice:          task.BuyPrice,
			Status:             task.Status,
			Error:              task.Error,
			CreatedAt:          task.CreatedAt,
			FinishedAt:         task.FinishedAt,
		}
	}

	return page, nil
}
