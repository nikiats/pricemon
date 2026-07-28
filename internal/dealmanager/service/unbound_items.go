package service

import (
	"time"

	"gopricemon/internal/dealmanager/domain"
)

func (s *TaskWorker) failSelling(task *domain.SequentialTask, message string) error {
	var purchasedAt *time.Time
	if task.BuyTaskID != nil {
		buyTask, err := s.client.GetTask(*task.BuyTaskID)
		if err != nil {
			return err
		}
		purchasedAt = buyTask.CompletedAt
	}

	task.Status = domain.SequentialTaskStatusFailed
	task.Error = &message
	return s.repo.FailSellingSequentialTask(*task, purchasedAt)
}
