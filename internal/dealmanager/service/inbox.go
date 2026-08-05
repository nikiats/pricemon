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

			now := time.Now()
			if !payload.ExpiresAt.After(payload.ActualAt) || !now.Before(payload.ExpiresAt) {
				ids = append(ids, event.ID)
				continue
			}
			lastFinishedAt, err := s.repo.GetLastFinishedAt(payload.ItemID)
			if err != nil {
				return err
			}
			if lastFinishedAt != nil && (now.Before(lastFinishedAt.Add(payload.ExpiresAt.Sub(payload.ActualAt))) || !payload.ActualAt.After(*lastFinishedAt)) {
				ids = append(ids, event.ID)
				continue
			}

			task := domain.SequentialTask{
				InboxEventID:   event.ID,
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
			if item != nil {
				task.Status = domain.SequentialTaskStatusSelling
				err = s.repo.CreateSellingSequentialTask(task, item.ID)
			} else {
				err = s.repo.CreateSequentialTask(task)
			}
			if err != nil {
				return err
			}

			ids = append(ids, event.ID)
			availableTasks--
			if availableTasks == 0 {
				return s.repo.MarkEventsProcessed(ids)
			}
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
