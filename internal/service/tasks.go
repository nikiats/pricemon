package service

import (
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"gopricemon/internal/domain"
	"gopricemon/internal/repository"
)

var (
	ErrInvalidExecutorToken   = errors.New("invalid executor token")
	ErrInvalidTaskID          = errors.New("invalid task ID")
	ErrInvalidLeaseToken      = errors.New("invalid lease token")
	ErrInvalidLeaseDuration   = errors.New("invalid lease duration")
	ErrInvalidTaskResult      = errors.New("invalid task result")
	ErrTaskLeaseExpired       = errors.New("task lease expired or belongs to another executor")
	ErrTaskResultReported     = errors.New("task result already reported")
	ErrTaskBelongsToExecutor  = errors.New("task belongs to another executor")
	ErrTaskResultLeaseExpired = errors.New("task lease expired")
	ErrTaskLeaseInactive      = errors.New("task lease is no longer active")
	ErrTaskNotInProgress      = errors.New("task is not in progress")
	ErrTaskNotFound           = errors.New("task not found")
	ErrTaskReferencesNotFound = errors.New("task references not found")
	ErrTaskNotCancellable     = errors.New("task is already in progress or does not exist")
	ErrInvalidCategoryName    = errors.New("category name is required")
	ErrInvalidItemName        = errors.New("item name is required")
	ErrInvalidTaskPlatform    = errors.New("platform name is required")
	ErrInvalidTaskAction      = errors.New("action must be buy or sell")
	ErrInvalidTaskPrice       = errors.New("price must be a positive number")
)

func (s *Service) ClaimTask(platformID int, executorToken string) (domain.Task, bool, error) {
	if platformID < 1 {
		return domain.Task{}, false, ErrInvalidPlatformID
	}

	executor, err := s.executorByToken(executorToken)
	if err != nil {
		return domain.Task{}, false, err
	}

	return s.repo.ClaimTask(platformID, executor.ID, s.taskLeaseMaxSeconds)
}

func (s *Service) GetTasks() ([]domain.TaskInfo, error) {
	return s.repo.GetTasks()
}

func (s *Service) CreateTask(categoryName, itemName, platformName, actionType, priceRaw string) (domain.TaskInfo, error) {
	categoryName = strings.TrimSpace(categoryName)
	itemName = strings.TrimSpace(itemName)
	platformName = strings.TrimSpace(platformName)
	if categoryName == "" {
		return domain.TaskInfo{}, ErrInvalidCategoryName
	}
	if itemName == "" {
		return domain.TaskInfo{}, ErrInvalidItemName
	}
	if platformName == "" {
		return domain.TaskInfo{}, ErrInvalidTaskPlatform
	}
	if actionType != "buy" && actionType != "sell" {
		return domain.TaskInfo{}, ErrInvalidTaskAction
	}

	price, err := decimal.NewFromString(strings.TrimSpace(priceRaw))
	if err != nil || price.LessThanOrEqual(decimal.Zero) {
		return domain.TaskInfo{}, ErrInvalidTaskPrice
	}

	task, err := s.repo.CreateTask(categoryName, itemName, platformName, actionType, price)
	if errors.Is(err, repository.ErrTaskReferencesNotFound) {
		return domain.TaskInfo{}, ErrTaskReferencesNotFound
	}

	return task, err
}

func (s *Service) CreateTaskByCategoryID(categoryID int, itemName, platformName, actionType, priceRaw string) (domain.TaskInfo, error) {
	if categoryID < 1 {
		return domain.TaskInfo{}, ErrInvalidCategoryID
	}

	itemName = strings.TrimSpace(itemName)
	platformName = strings.TrimSpace(platformName)
	if itemName == "" {
		return domain.TaskInfo{}, ErrInvalidItemName
	}
	if platformName == "" {
		return domain.TaskInfo{}, ErrInvalidTaskPlatform
	}
	if actionType != "buy" && actionType != "sell" {
		return domain.TaskInfo{}, ErrInvalidTaskAction
	}

	price, err := decimal.NewFromString(strings.TrimSpace(priceRaw))
	if err != nil || price.LessThanOrEqual(decimal.Zero) {
		return domain.TaskInfo{}, ErrInvalidTaskPrice
	}

	task, err := s.repo.CreateTaskByCategoryID(categoryID, itemName, platformName, actionType, price)
	if errors.Is(err, repository.ErrTaskReferencesNotFound) {
		return domain.TaskInfo{}, ErrTaskReferencesNotFound
	}

	return task, err
}

func (s *Service) DeleteTask(taskID int) error {
	if taskID < 1 {
		return ErrInvalidTaskID
	}

	deleted, err := s.repo.DeleteTask(taskID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrTaskNotCancellable
	}

	return nil
}

func (s *Service) ExtendTaskLease(taskID int, leaseToken string, leaseSeconds int, executorToken string) (time.Time, error) {
	if taskID < 1 {
		return time.Time{}, ErrInvalidTaskID
	}
	if strings.TrimSpace(leaseToken) == "" {
		return time.Time{}, ErrInvalidLeaseToken
	}
	if leaseSeconds < 1 || leaseSeconds > s.taskLeaseMaxSeconds {
		return time.Time{}, ErrInvalidLeaseDuration
	}

	executor, err := s.executorByToken(executorToken)
	if err != nil {
		return time.Time{}, err
	}

	leaseUntil, extended, err := s.repo.ExtendTaskLease(taskID, executor.ID, leaseToken, leaseSeconds)
	if err != nil {
		return time.Time{}, err
	}
	if !extended {
		return time.Time{}, ErrTaskLeaseExpired
	}

	return leaseUntil, nil
}

func (s *Service) ReportTaskResult(taskID int, leaseToken string, status domain.TaskStatus, errorText *string, executorToken string) (domain.TaskStatus, error) {
	if taskID < 1 {
		return "", ErrInvalidTaskID
	}
	if strings.TrimSpace(leaseToken) == "" {
		return "", ErrInvalidLeaseToken
	}

	if errorText != nil {
		trimmed := strings.TrimSpace(*errorText)
		errorText = &trimmed
	}
	if status == domain.TaskStatusFailed && (errorText == nil || *errorText == "") {
		return "", ErrInvalidTaskResult
	}
	if status == domain.TaskStatusCompleted && errorText != nil {
		return "", ErrInvalidTaskResult
	}
	if status != domain.TaskStatusCompleted && status != domain.TaskStatusFailed {
		return "", ErrInvalidTaskResult
	}

	executor, err := s.executorByToken(executorToken)
	if err != nil {
		return "", err
	}

	task, reported, err := s.repo.ReportTaskResult(taskID, executor.ID, leaseToken, status, errorText)
	if err != nil {
		return "", err
	}
	if reported {
		return "", nil
	}
	if task == nil {
		return "", ErrTaskNotFound
	}
	if task.Status == domain.TaskStatusCompleted || task.Status == domain.TaskStatusFailed {
		return task.Status, ErrTaskResultReported
	}
	if task.Status != domain.TaskStatusInProgress {
		return "", ErrTaskNotInProgress
	}
	if task.ExecutorID == nil || *task.ExecutorID != executor.ID {
		return "", ErrTaskBelongsToExecutor
	}
	if task.LeaseExpired {
		return "", ErrTaskResultLeaseExpired
	}
	if task.LeaseToken == nil || *task.LeaseToken != leaseToken {
		return "", ErrTaskLeaseInactive
	}

	return "", ErrTaskNotInProgress
}

func (s *Service) executorByToken(token string) (domain.Executor, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.Executor{}, ErrInvalidExecutorToken
	}

	executor, err := s.repo.GetExecutorByToken(token)
	if errors.Is(err, repository.ErrExecutorNotFound) {
		return domain.Executor{}, ErrInvalidExecutorToken
	}

	return executor, err
}
