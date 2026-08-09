package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"gopricemon/internal/dealmanager/domain"
	app "gopricemon/internal/dealmanager/service"
)

type InboxService interface {
	ReceiveEvent(id string, payload json.RawMessage) error
}

type SequentialTasksService interface {
	GetSequentialTasks(offset, limit int) (domain.SequentialTaskPage, error)
	ResetActiveSequentialTasks() error
}

type Handler struct {
	inboxService           InboxService
	sequentialTasksService SequentialTasksService
}

func NewHandler(inboxService InboxService, sequentialTasksService SequentialTasksService) *Handler {
	return &Handler{inboxService: inboxService, sequentialTasksService: sequentialTasksService}
}

func (h *Handler) Register(router *gin.Engine) {
	router.POST("/event/item-summary", h.receiveItemSummaryEvent)
	router.GET("/sequential-tasks", h.getSequentialTasks)
	router.POST("/sequential-tasks/reset", h.resetSequentialTasks)
}

func (h *Handler) receiveItemSummaryEvent(c *gin.Context) {
	var request struct {
		ID      string          `json:"id"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.inboxService.ReceiveEvent(request.ID, request.Payload); err != nil {
		if errors.Is(err, app.ErrInvalidEvent) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusAccepted)
}
