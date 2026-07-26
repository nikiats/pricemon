package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"gopricemon/internal/repository"
)

const outboxBatchSize = 100

type OutboxPublisher struct {
	repo       repository.Repository
	url        string
	interval   time.Duration
	httpClient *http.Client
}

func NewOutboxPublisher(repo repository.Repository, url string, interval time.Duration) *OutboxPublisher {
	return &OutboxPublisher{
		repo:       repo,
		url:        url + "/event/item-summary",
		interval:   interval,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (o *OutboxPublisher) Run(ctx context.Context) {
	ticker := time.NewTicker(o.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := o.publish(ctx); err != nil {
				log.Printf("outbox publisher: %v", err)
			}
		}
	}
}

func (o *OutboxPublisher) publish(ctx context.Context) error {
	messages, err := o.repo.GetOutboxMessages(outboxBatchSize)
	if err != nil {
		return err
	}

	for _, message := range messages {
		if err = o.send(ctx, message.ID, message.Payload); err != nil {
			if markErr := o.repo.MarkOutboxMessageFailed(message.ID, err.Error()); markErr != nil {
				return markErr
			}
			continue
		}

		if err = o.repo.MarkOutboxMessageProcessed(message.ID); err != nil {
			return err
		}
	}

	return nil
}

func (o *OutboxPublisher) send(ctx context.Context, id string, payload json.RawMessage) error {
	body, err := json.Marshal(struct {
		ID      string          `json:"id"`
		Payload json.RawMessage `json:"payload"`
	}{ID: id, Payload: payload})
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, o.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := o.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("dealmanager returned %s", response.Status)
	}

	return nil
}
