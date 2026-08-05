package http

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"gopricemon/internal/domain"
)

type Service interface {
	GetSummary(categoryID, offset, limit int, maxAge *int) (domain.SummaryPage, error)
	GetInventory() ([]domain.InventorySummary, error)
	GetTradeSettings() (domain.TradeSettings, error)
	SetTradeSettings(minimumProfit decimal.Decimal, maximumSummaryAgeSecs, maximumConcurrentTrades int) error
	SetOffer(offer domain.Offer) error
	SetOffers(offers []domain.Offer) error
	SetZeroCount(categoryID int, itemName string, platformID int, platformToken string) (bool, error)
	ReplaceInventory(items []domain.InventoryItem) error
	ChangeInventory(items []domain.InventoryDelta) error
	GetPlatforms() ([]domain.Platform, error)
	CreatePlatform(name string) (domain.Platform, error)
	DeletePlatform(platformID int) error
	RegeneratePlatformToken(platformID int) (string, error)
	GetExecutors() ([]domain.Executor, error)
	CreateExecutor(name string) (domain.Executor, error)
	DeleteExecutor(executorID int) error
	RegenerateExecutorToken(executorID int) (string, error)
	ClaimTask(platformID int, executorToken string) (domain.Task, bool, error)
	ExtendTaskLease(taskID int, leaseToken string, leaseSeconds int, executorToken string) (time.Time, error)
	ReportTaskResult(taskID int, leaseToken string, status domain.TaskStatus, errorText *string, executorToken string) (domain.TaskStatus, error)
	GetTasks() ([]domain.TaskInfo, error)
	GetTask(taskID int) (domain.TaskInfo, error)
	CreateTask(itemID int, platformName, actionType, price, taskKey string) (domain.TaskInfo, error)
	DeleteTask(taskID int) error
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(router *gin.Engine) {
	router.Use(cors, serverErrorLogger)

	router.GET("/categories/:categoryID/summary", h.getSummary)
	router.GET("/", h.indexPage)
	router.GET("/summary", h.summaryPage)
	router.GET("/trade-settings", h.tradeSettingsPageOrJSON)
	router.PUT("/trade-settings", h.setTradeSettings)
	router.PUT("/offers", h.setOffer)
	router.PUT("/offers/bulk", h.setOffers)
	router.PUT("/offers/zero-count", h.setZeroCount)
	router.GET("/inventory", h.inventoryPageOrList)
	router.PUT("/inventory", h.replaceInventory)
	router.PATCH("/inventory", h.changeInventory)

	platforms := router.Group("/platforms")
	platforms.GET("", h.platformsPageOrList)
	platforms.POST("", h.createPlatform)
	platforms.DELETE("/:platformID", h.deletePlatform)
	platforms.PUT("/:platformID/token", h.regeneratePlatformToken)

	executors := router.Group("/executors")
	executors.GET("", h.executorsPageOrList)
	executors.POST("", h.createExecutor)
	executors.DELETE("/:executorID", h.deleteExecutor)
	executors.PUT("/:executorID/token", h.regenerateExecutorToken)

	tasks := router.Group("/tasks")
	tasks.GET("", h.tasksPageOrList)
	tasks.POST("", h.createTask)
	tasks.GET("/:taskID", h.getTask)
	tasks.DELETE("/:taskID", h.deleteTask)
	tasks.POST("/claim", h.claimTask)
	tasks.PUT("/:taskID/lease", h.extendTaskLease)
	tasks.POST("/:taskID/result", h.reportTaskResult)
}

func serverErrorLogger(c *gin.Context) {
	startedAt := time.Now()
	c.Next()

	if c.Writer.Status() < http.StatusInternalServerError {
		return
	}

	log.Printf(
		"server error: status=%d method=%s path=%s client_ip=%s duration=%s errors=%q",
		c.Writer.Status(),
		c.Request.Method,
		c.Request.URL.RequestURI(),
		c.ClientIP(),
		time.Since(startedAt).Round(time.Millisecond),
		c.Errors.String(),
	)
}

func cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}
