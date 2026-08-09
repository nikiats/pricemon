package repository

import (
	"time"

	"github.com/shopspring/decimal"

	"gopricemon/internal/dealmanager/domain"
)

type Repository interface {
	CreateInboxEvent(event domain.InboxEvent) error
	GetPendingEvents(limit int) ([]domain.InboxEvent, error)
	GetActiveSequentialTasks() ([]domain.SequentialTask, error)
	GetSequentialTasks(offset, limit int) ([]domain.SequentialTask, error)
	GetLastFinishedAt(itemID int) (*time.Time, error)
	CreateSequentialTask(task domain.SequentialTask) error
	CreateSellingSequentialTask(task domain.SequentialTask, unboundItemID int) error
	UpdateSequentialTask(task domain.SequentialTask) error
	GetAvailableUnboundItem(itemID int, maximumPrice decimal.Decimal) (*domain.UnboundItem, error)
	CompleteSellingSequentialTask(taskID int) error
	FailSellingSequentialTask(task domain.SequentialTask, purchasedAt *time.Time) error
	MarkEventsProcessed(ids []string) error
	MarkEventFailed(id, message string) error
}
