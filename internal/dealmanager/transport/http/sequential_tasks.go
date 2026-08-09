package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

const (
	defaultPageSize = 50
	maximumPageSize = 100
)

var errInvalidPagination = errors.New("invalid pagination")

type sequentialTaskResponse struct {
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

type sequentialTasksPageResponse struct {
	Items      []sequentialTaskResponse `json:"items"`
	NextOffset *int                     `json:"nextOffset,omitempty"`
}

func (h *Handler) getSequentialTasks(c *gin.Context) {
	offset, limit, err := pagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
		return
	}

	page, err := h.sequentialTasksService.GetSequentialTasks(offset, limit)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	items := make([]sequentialTaskResponse, len(page.Items))
	for i, task := range page.Items {
		item := sequentialTaskResponse{
			ID:                 task.ID,
			ItemID:             task.ItemID,
			PurchasePlatformID: task.PurchasePlatformID,
			PurchasePrice:      task.PurchasePrice,
			SalePlatformID:     task.SalePlatformID,
			SalePrice:          task.SalePrice,
			Status:             string(task.Status),
			Error:              task.Error,
			CreatedAt:          task.CreatedAt,
			FinishedAt:         task.FinishedAt,
		}
		items[i] = item
	}

	c.JSON(http.StatusOK, sequentialTasksPageResponse{Items: items, NextOffset: page.NextOffset})
}

func pagination(c *gin.Context) (int, int, error) {
	offset, err := queryInt(c, "offset", 0)
	if err != nil {
		return 0, 0, err
	}
	if offset < 0 {
		return 0, 0, errInvalidPagination
	}
	limit, err := queryInt(c, "limit", defaultPageSize)
	if err != nil {
		return 0, 0, err
	}
	if limit < 1 || limit > maximumPageSize {
		return 0, 0, errInvalidPagination
	}

	return offset, limit, nil
}

func queryInt(c *gin.Context, name string, defaultValue int) (int, error) {
	value := c.Query(name)
	if value == "" {
		return defaultValue, nil
	}

	return strconv.Atoi(value)
}
