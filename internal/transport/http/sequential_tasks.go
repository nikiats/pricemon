package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type dealManagerSequentialTaskResponse struct {
	ID                 int             `json:"id"`
	ItemID             int             `json:"itemID"`
	PurchasePlatformID int             `json:"purchasePlatformID"`
	PurchasePrice      decimal.Decimal `json:"purchasePrice"`
	SalePlatformID     int             `json:"salePlatformID"`
	SalePrice          decimal.Decimal `json:"salePrice"`
	Status             string          `json:"status"`
	Error              *string         `json:"error,omitempty"`
	CreatedAt          time.Time       `json:"createdAt"`
	FinishedAt         *time.Time      `json:"finishedAt,omitempty"`
}

type dealManagerSequentialTasksPageResponse struct {
	Items      []dealManagerSequentialTaskResponse `json:"items"`
	NextOffset *int                                `json:"nextOffset,omitempty"`
}

type sequentialTaskResponse struct {
	ID                   int             `json:"id"`
	ItemID               int             `json:"itemID"`
	PurchasePlatformName string          `json:"purchasePlatformName"`
	PurchasePrice        decimal.Decimal `json:"purchasePrice"`
	SalePlatformName     string          `json:"salePlatformName"`
	SalePrice            decimal.Decimal `json:"salePrice"`
	Status               string          `json:"status"`
	Error                *string         `json:"error,omitempty"`
	CreatedAt            time.Time       `json:"createdAt"`
	FinishedAt           *time.Time      `json:"finishedAt,omitempty"`
}

type sequentialTasksPageResponse struct {
	Items      []sequentialTaskResponse `json:"items"`
	NextOffset *int                     `json:"nextOffset,omitempty"`
}

func (h *Handler) getSequentialTasks(c *gin.Context) {
	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, h.dealManagerAPIBaseURL+c.Request.URL.RequestURI(), nil)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	response, err := h.client.Do(request)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "dealmanager is unavailable"})
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		c.JSON(response.StatusCode, gin.H{"error": "dealmanager request failed"})
		return
	}

	var page dealManagerSequentialTasksPageResponse
	if err = json.NewDecoder(response.Body).Decode(&page); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid dealmanager response"})
		return
	}
	platforms, err := h.service.GetPlatforms()
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	platformNames := make(map[int]string, len(platforms))
	for _, platform := range platforms {
		platformNames[platform.ID] = platform.Name
	}

	items := make([]sequentialTaskResponse, len(page.Items))
	for i, task := range page.Items {
		items[i] = sequentialTaskResponse{
			ID:                   task.ID,
			ItemID:               task.ItemID,
			PurchasePlatformName: platformNames[task.PurchasePlatformID],
			PurchasePrice:        task.PurchasePrice,
			SalePlatformName:     platformNames[task.SalePlatformID],
			SalePrice:            task.SalePrice,
			Status:               task.Status,
			Error:                task.Error,
			CreatedAt:            task.CreatedAt,
			FinishedAt:           task.FinishedAt,
		}
	}

	c.JSON(http.StatusOK, sequentialTasksPageResponse{Items: items, NextOffset: page.NextOffset})
}

func (h *Handler) resetSequentialTasks(c *gin.Context) {
	if err := h.service.ResetActiveTasks(); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, h.dealManagerAPIBaseURL+"/sequential-tasks/reset", nil)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	response, err := h.client.Do(request)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "dealmanager is unavailable"})
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		c.JSON(http.StatusBadGateway, gin.H{"error": "dealmanager reset failed"})
		return
	}

	c.Status(http.StatusNoContent)
}
