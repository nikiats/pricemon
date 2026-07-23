package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"gopricemon/internal/domain"
	"gopricemon/internal/service"
)

func (h *Handler) replaceInventory(c *gin.Context) {
	var request struct {
		Items []domain.InventoryItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.Items == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.ReplaceInventory(request.Items); err != nil {
		if errors.Is(err, service.ErrInvalidInventory) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) changeInventory(c *gin.Context) {
	var request struct {
		Items []domain.InventoryDelta `json:"items"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.Items == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.ChangeInventory(request.Items); err != nil {
		if errors.Is(err, service.ErrInvalidInventory) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrInsufficientInventory) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
