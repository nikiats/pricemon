package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	app "gopricemon/internal/dealmanager/service"
)

type Service interface {
	ReceiveEvent(id, eventType string, payload json.RawMessage) error
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(router *gin.Engine) {
	router.POST("/event/item-summary", h.receiveItemSummaryEvent)
}

func (h *Handler) receiveItemSummaryEvent(c *gin.Context) {
	var request struct {
		ID      string          `json:"id"`
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.ReceiveEvent(request.ID, request.Type, request.Payload); err != nil {
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
