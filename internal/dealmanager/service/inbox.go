package service

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

type InboxWorker struct {
	repo        repository.Repository
	maxTradeAge int
}

func NewEventsWorker(repository repository.Repository, maxTradeAge int) *InboxWorker {
	return &InboxWorker{
		repo:        repository,
		maxTradeAge: maxTradeAge,
	}
}

func (s *InboxWorker) ReceiveEvent(id string, payload json.RawMessage) error {
	if strings.TrimSpace(id) == "" || !json.Valid(payload) {
		return ErrInvalidEvent
	}

	return s.repo.CreateInboxEvent(domain.InboxEvent{
		ID:      id,
		Payload: payload,
	})
}

func (s *InboxWorker) ProcessPendingEvents() error {
	activeTask, err := s.repo.GetActiveSequentialTask()
	if err != nil {
		return err
	}
	if activeTask != nil {
		return nil
	}

	for {
		events, err := s.repo.GetPendingEvents(eventBatchSize)
		if err != nil {
			return err
		}

		ids := make([]string, 0, len(events))
		for _, event := range events {
			var payload domain.ItemSummaryPayload
			if err = json.Unmarshal(event.Payload, &payload); err != nil {
				return err
			}

			if time.Since(payload.ActualAt) > time.Second*time.Duration(s.maxTradeAge) {
				ids = append(ids, event.ID)
				continue
			}

			if err = s.repo.CreateSequentialTask(domain.SequentialTask{
				InboxEventID:   event.ID,
				CategoryID:     payload.CategoryID,
				ItemID:         payload.ItemID,
				PlatformSellID: payload.PlatformSellID,
				PlatformBuyID:  payload.PlatformBuyID,
				SellPrice:      payload.SellPrice,
				BuyPrice:       payload.BuyPrice,
				Status:         domain.SequentialTaskStatusNotStarted,
			}); err != nil {
				return err
			}

			ids = append(ids, event.ID)
			return s.repo.MarkEventsProcessed(ids)
		}

		if err = s.repo.MarkEventsProcessed(ids); err != nil {
			return err
		}
		if len(events) < eventBatchSize {
			return nil
		}
	}
}

func (s *InboxWorker) RunEventWorker(ctx context.Context, interval time.Duration) {
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
