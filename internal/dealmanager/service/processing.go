package app

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"gopricemon/internal/dealmanager/domain"
	"gopricemon/internal/dealmanager/repository"
)

const eventBatchSize = 100

var ErrInvalidEvent = errors.New("invalid event")

type EventsWorker struct {
	repository repository.Repository
}

func NewEventsWorker(repository repository.Repository) *EventsWorker {
	return &EventsWorker{repository: repository}
}

func (s *EventsWorker) ReceiveEvent(id string, payload json.RawMessage) error {
	if strings.TrimSpace(id) == "" || !json.Valid(payload) {
		return ErrInvalidEvent
	}

	return s.repository.CreateInboxEvent(domain.InboxEvent{
		ID:      id,
		Payload: payload,
	})
}

func (s *EventsWorker) ProcessPendingEvents() error {
	for {
		events, err := s.repository.GetPendingEvents(eventBatchSize)
		if err != nil {
			return err
		}

		ids := make([]string, 0, len(events))
		for _, event := range events {
			var summary ItemSummaryEvent
			if err = json.Unmarshal(event.Payload, &summary); err != nil {
				if err = s.repository.MarkEventFailed(event.ID, err.Error()); err != nil {
					return err
				}
				continue
			}

			// TODO: обработать событие.
			// в результате создается sequential_task или событие отбраковывается.
			ids = append(ids, event.ID)
		}

		if err = s.repository.MarkEventsProcessed(ids); err != nil {
			return err
		}
		if len(events) < eventBatchSize {
			return nil
		}
	}
}

type ItemSummaryEvent struct {
	CategoryID     int    `json:"categoryID"`
	ItemID         int    `json:"itemID"`
	PlatformSellID int    `json:"platformSellID"`
	PlatformBuyID  int    `json:"platformBuyID"`
	SellPrice      string `json:"sellPrice"`
	BuyPrice       string `json:"buyPrice"`
}

func (s *EventsWorker) RunEventWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.ProcessPendingEvents(); err != nil {
				log.Printf("dealmanager event worker: %v", err)
			}
		}
	}
}
