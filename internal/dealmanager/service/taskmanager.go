package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"

	"gopricemon/internal/dealmanager/domain"
	"gopricemon/internal/dealmanager/gopricemon"
	"gopricemon/internal/dealmanager/repository"
)

type TaskClient interface {
	CreateTask(itemID, platformID int, actionType string, price decimal.Decimal, taskKey string) (int, error)
	GetTask(taskID int) (gopricemon.Task, error)
}

type TaskWorker struct {
	repo   repository.Repository
	client TaskClient
}

func NewTaskWorker(repo repository.Repository, client TaskClient) *TaskWorker {
	return &TaskWorker{
		repo:   repo,
		client: client,
	}
}

func (s *TaskWorker) ProcessTasks() error {
	tasks, err := s.repo.GetActiveSequentialTasks()
	if err != nil {
		return err
	}

	for i := range tasks {
		if err = s.processTask(&tasks[i]); err != nil {
			log.Printf("dealmanager task worker: task=%d: %v", tasks[i].ID, err)
		}
	}

	return nil
}

func (s *TaskWorker) processTask(task *domain.SequentialTask) error {
	switch task.Status {
	case domain.SequentialTaskStatusNotStarted:
		return s.startBuying(task)
	case domain.SequentialTaskStatusBuying:
		return s.processBuying(task)
	case domain.SequentialTaskStatusSelling:
		return s.processSelling(task)
	}

	return nil
}

func (s *TaskWorker) RunTaskWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.ProcessTasks(); err != nil {
				log.Printf("dealmanager task worker: %v", err)
			}
		}
	}
}

func (s *TaskWorker) startBuying(task *domain.SequentialTask) error {
	taskID, err := s.client.CreateTask(task.ItemID, task.PlatformSellID, "buy", task.SellPrice, taskKey(task.ID, "buy"))
	if err != nil {
		return err
	}

	task.BuyTaskID = &taskID
	task.Status = domain.SequentialTaskStatusBuying
	return s.repo.UpdateSequentialTask(*task)
}

func (s *TaskWorker) processBuying(task *domain.SequentialTask) error {
	if task.BuyTaskID == nil {
		return s.fail(task, "buy task ID is not set")
	}

	buyTask, err := s.client.GetTask(*task.BuyTaskID)
	if err != nil {
		return err
	}
	switch buyTask.Status {
	case gopricemon.TaskStatusFailed:
		return s.fail(task, taskError(buyTask.Error))
	case gopricemon.TaskStatusCompleted:
		return s.startSelling(task)
	}

	return nil
}

func (s *TaskWorker) processSelling(task *domain.SequentialTask) error {
	if task.SellTaskID == nil {
		return s.startSelling(task)
	}

	sellTask, err := s.client.GetTask(*task.SellTaskID)
	if err != nil {
		return err
	}
	switch sellTask.Status {
	case gopricemon.TaskStatusFailed:
		return s.failSelling(task, taskError(sellTask.Error))
	case gopricemon.TaskStatusCompleted:
		return s.repo.CompleteSellingSequentialTask(task.ID)
	}

	return nil
}

func (s *TaskWorker) startSelling(task *domain.SequentialTask) error {
	taskID, err := s.client.CreateTask(task.ItemID, task.PlatformBuyID, "sell", task.BuyPrice, taskKey(task.ID, "sell"))
	if err != nil {
		return err
	}

	task.SellTaskID = &taskID
	task.Status = domain.SequentialTaskStatusSelling
	return s.repo.UpdateSequentialTask(*task)
}

func (s *TaskWorker) fail(task *domain.SequentialTask, message string) error {
	task.Status = domain.SequentialTaskStatusFailed
	task.Error = &message
	return s.repo.UpdateSequentialTask(*task)
}

func taskError(errorText *string) string {
	if errorText == nil || *errorText == "" {
		return "gopricemon task failed"
	}

	return fmt.Sprintf("gopricemon task failed: %s", *errorText)
}

func taskKey(taskID int, actionType string) string {
	return fmt.Sprintf("dealmanager:%d:%s", taskID, actionType)
}
