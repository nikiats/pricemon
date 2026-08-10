package repository

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"gopricemon/internal/domain"
)

var (
	ErrPlatformAlreadyExists  = errors.New("platform already exists")
	ErrPlatformNotFound       = errors.New("platform not found")
	ErrExecutorAlreadyExists  = errors.New("executor already exists")
	ErrExecutorNotFound       = errors.New("executor not found")
	ErrTaskReferencesNotFound = errors.New("task references not found")
)

type Repository interface {
	SetOffers(offers []domain.Offer) error
	SetZeroCount(itemID, platformID int) (bool, error)
	GetPlatform(platformID int) (domain.Platform, error)
	GetOrCreateItem(categoryID int, name string) (int, error)
	ReplaceInventory(items []domain.InventoryItem) error
	ChangeInventory(items []domain.InventoryDelta) (bool, error)
	GetInventory() ([]domain.InventorySummary, error)
	GetSummary(categoryID, offset, limit int, maxAge *int) ([]domain.ItemSummary, error)
	GetPlatforms() ([]domain.Platform, error)
	CreatePlatform(name string) (domain.Platform, error)
	DeletePlatform(platformID int) error
	RegeneratePlatformToken(platformID int) (string, error)
	GetExecutors() ([]domain.Executor, error)
	CreateExecutor(name string) (domain.Executor, error)
	DeleteExecutor(executorID int) error
	RegenerateExecutorToken(executorID int) (string, error)
	GetExecutorByToken(token string) (domain.Executor, error)
	ClaimTask(platformID, executorID, leaseSeconds int) (domain.Task, bool, error)
	ExtendTaskLease(taskID, executorID int, leaseToken string, leaseSeconds int) (time.Time, bool, error)
	ReportTaskResult(taskID, executorID int, leaseToken string, status domain.TaskStatus, errorText *string, completedAt time.Time) (*domain.TaskReportState, bool, error)
	GetTasks() ([]domain.TaskInfo, error)
	GetTask(taskID int) (*domain.TaskInfo, error)
	CreateTask(itemID int, platformName, actionType string, price decimal.Decimal, taskKey string) (domain.TaskInfo, error)
	DeleteTask(taskID int) (bool, error)
	ResetActiveTasks(message string) error
	GetOutboxMessages(limit int) ([]domain.OutboxMessage, error)
	MarkOutboxMessageProcessed(id string) error
	MarkOutboxMessageFailed(id, message string) error
	GetTradeSettings() (domain.TradeSettings, error)
	InitializeTradeSettings(minimumProfit, maximumBuyPrice decimal.Decimal, maximumSummaryAgeSecs, maximumConcurrentTrades int) error
	SetTradeSettings(minimumProfit, maximumBuyPrice decimal.Decimal, maximumSummaryAgeSecs, maximumConcurrentTrades int) error
}
