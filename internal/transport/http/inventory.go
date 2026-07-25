package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"gopricemon/internal/domain"
	"gopricemon/internal/service"
)

type inventoryResponse struct {
	ID               int              `json:"id"`
	Name             string           `json:"name"`
	CategoryID       int              `json:"categoryID"`
	Quantity         int              `json:"quantity"`
	SellPrice        *decimal.Decimal `json:"sellPrice"`
	SellPlatformName *string          `json:"sellPlatformName"`
	SellURL          *string          `json:"sellUrl"`
	BuyPrice         *decimal.Decimal `json:"buyPrice"`
	BuyPlatformName  *string          `json:"buyPlatformName"`
	BuyURL           *string          `json:"buyUrl"`
}

func (h *Handler) getInventory(c *gin.Context) {
	items, err := h.service.GetInventory()
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	response := make([]inventoryResponse, len(items))
	for i, item := range items {
		response[i] = inventoryResponse{
			ID:               item.ID,
			Name:             item.Name,
			CategoryID:       item.CategoryID,
			Quantity:         item.Quantity,
			SellPrice:        item.SellPrice,
			SellPlatformName: item.SellPlatformName,
			SellURL:          item.SellURL,
			BuyPrice:         item.BuyPrice,
			BuyPlatformName:  item.BuyPlatformName,
			BuyURL:           item.BuyURL,
		}
	}

	c.JSON(http.StatusOK, response)
}

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
