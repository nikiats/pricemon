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
