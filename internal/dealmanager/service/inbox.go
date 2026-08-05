package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"gopricemon/internal/dealmanager/domain"
	"gopricemon/internal/dealmanager/gopricemon"
	"gopricemon/internal/dealmanager/repository"
)

const eventBatchSize = 100

var ErrInvalidEvent = errors.New("invalid event")

type TradeSettingsClient interface {
	GetTradeSettings() (gopricemon.TradeSettings, error)
}

type InboxWorker struct {
	repo   repository.Repository
	client TradeSettingsClient
}

func NewEventsWorker(repository repository.Repository, client TradeSettingsClient) *InboxWorker {
	return &InboxWorker{
		repo:   repository,
		client: client,
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
	settings, err := s.client.GetTradeSettings()
	if err != nil {
		return err
	}
	activeTasks, err := s.repo.GetActiveSequentialTasks()
	if err != nil {
		return err
	}
	availableTasks := settings.MaximumConcurrentTrades - len(activeTasks)
	if availableTasks < 1 {
		return nil
	}

	for availableTasks > 0 {
		events, err := s.repo.GetPendingEvents(eventBatchSize)
		if err != nil {
			return err
		}
		if len(events) == 0 {
			return nil
		}

		ids := make([]string, 0, len(events))
		for _, event := range events {
			var payload domain.ItemSummaryPayload
			if err = json.Unmarshal(event.Payload, &payload); err != nil {
				return err
			}

			actual, err := s.isActualTrade(payload, time.Now())
			if err != nil {
				return err
			}
			if !actual {
				ids = append(ids, event.ID)
				continue
			}

			if err = s.createTrade(event.ID, payload); err != nil {
				return err
			}

			ids = append(ids, event.ID)
			availableTasks--
			if availableTasks == 0 {
				break
			}
		}

		if err = s.repo.MarkEventsProcessed(ids); err != nil {
			return err
		}
		if len(events) < eventBatchSize {
			return nil
		}
	}

	return nil
}

func (s *InboxWorker) isActualTrade(payload domain.ItemSummaryPayload, now time.Time) (bool, error) {
	if !payload.ExpiresAt.After(payload.ActualAt) || !payload.ExpiresAt.After(now) {
		return false, nil
	}

	lastFinishedAt, err := s.repo.GetLastFinishedAt(payload.ItemID)
	if err != nil {
		return false, err
	}
	if lastFinishedAt == nil {
		return true, nil
	}

	maximumAge := payload.ExpiresAt.Sub(payload.ActualAt)
	finishedAgo := now.Sub(*lastFinishedAt)
	return payload.ActualAt.After(*lastFinishedAt) && finishedAgo >= maximumAge, nil
}

func (s *InboxWorker) createTrade(inboxEventID string, payload domain.ItemSummaryPayload) error {
	task := domain.SequentialTask{
		InboxEventID:   inboxEventID,
		CategoryID:     payload.CategoryID,
		ItemID:         payload.ItemID,
		PlatformSellID: payload.PlatformSellID,
		PlatformBuyID:  payload.PlatformBuyID,
		SellPrice:      payload.SellPrice,
		BuyPrice:       payload.BuyPrice,
		Status:         domain.SequentialTaskStatusNotStarted,
	}
	item, err := s.repo.GetAvailableUnboundItem(payload.ItemID, payload.BuyPrice)
	if err != nil {
		return err
	}
	if item == nil {
		return s.repo.CreateSequentialTask(task)
	}

	task.Status = domain.SequentialTaskStatusSelling
	return s.repo.CreateSellingSequentialTask(task, item.ID)
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
