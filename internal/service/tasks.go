package service

import (
	"errors"
	"strings"
	"time"

	"gopricemon/internal/domain"
	"gopricemon/internal/repository"
)

var (
	ErrInvalidExecutorToken = errors.New("invalid executor token")
	ErrInvalidTaskID        = errors.New("invalid task ID")
	ErrInvalidLeaseToken    = errors.New("invalid lease token")
	ErrInvalidLeaseDuration = errors.New("invalid lease duration")
	ErrInvalidTaskResult    = errors.New("invalid task result")
	ErrTaskLeaseExpired     = errors.New("task lease expired or belongs to another executor")
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

func (s *Service) ReportTaskResult(taskID int, leaseToken string, status domain.TaskStatus, errorText *string, executorToken string) error {
	if taskID < 1 {
		return ErrInvalidTaskID
	}
	if strings.TrimSpace(leaseToken) == "" {
		return ErrInvalidLeaseToken
	}

	if errorText != nil {
		trimmed := strings.TrimSpace(*errorText)
		errorText = &trimmed
	}
	if status == domain.TaskStatusFailed && (errorText == nil || *errorText == "") {
		return ErrInvalidTaskResult
	}
	if status == domain.TaskStatusCompleted && errorText != nil {
		return ErrInvalidTaskResult
	}
	if status != domain.TaskStatusCompleted && status != domain.TaskStatusFailed {
		return ErrInvalidTaskResult
	}

	executor, err := s.executorByToken(executorToken)
	if err != nil {
		return err
	}

	reported, err := s.repo.ReportTaskResult(taskID, executor.ID, leaseToken, status, errorText)
	if err != nil {
		return err
	}
	if !reported {
		return ErrTaskLeaseExpired
	}

	return nil
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
