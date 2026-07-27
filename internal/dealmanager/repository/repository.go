package repository

import "gopricemon/internal/dealmanager/domain"

type Repository interface {
	CreateInboxEvent(event domain.InboxEvent) error
	GetPendingEvents(limit int) ([]domain.InboxEvent, error)
	GetActiveSequentialTask() (*domain.SequentialTask, error)
	CreateSequentialTask(task domain.SequentialTask) error
	UpdateSequentialTask(task domain.SequentialTask) error
	MarkEventsProcessed(ids []string) error
	MarkEventFailed(id, message string) error
}
