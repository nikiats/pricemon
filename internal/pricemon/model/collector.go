package model

import "time"

type Collector struct {
	ID         string
	PlatformID int64
	Name       string
	CreatedAt  time.Time
}

type CollectorWithKey struct {
	Collector
	APIKey string
}
