package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type TaskStatus string

const (
	TaskStatusNotStarted TaskStatus = "not started"
	TaskStatusInProgress TaskStatus = "in progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

type Task struct {
	ID         int
	PlatformID int
	ActionType string
	Price      decimal.Decimal
	LeaseToken string
	LeaseUntil time.Time
}

type TaskReportState struct {
	Status       TaskStatus
	ExecutorID   *int
	LeaseToken   *string
	LeaseExpired bool
}

type TaskInfo struct {
	ID           int
	ItemName     *string
	CategoryName *string
	PlatformName string
	ExecutorName *string
	ActionType   string
	Price        decimal.Decimal
	Status       TaskStatus
	Error        *string
	LeaseUntil   *time.Time
	CompletedAt  *time.Time
}
