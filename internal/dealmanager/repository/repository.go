package repository

import "gopricemon/internal/dealmanager/domain"

type Repository interface {
	CreateInboxEvent(event domain.InboxEvent) error
	GetPendingEvents(limit int) ([]domain.InboxEvent, error)
	MarkEventsProcessed(ids []string) error
}
