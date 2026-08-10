package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"gopricemon/internal/service"
)

type tradeSettingsResponse struct {
	MinimumProfit           decimal.Decimal `json:"minimumProfit"`
	MaximumBuyPrice         decimal.Decimal `json:"maximumBuyPrice"`
	MaximumSummaryAgeSecs   int             `json:"maximumSummaryAgeSecs"`
	MaximumConcurrentTrades int             `json:"maximumConcurrentTrades"`
}

func (h *Handler) getTradeSettings(c *gin.Context) {
	settings, err := h.service.GetTradeSettings()
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, tradeSettingsResponse{
		MinimumProfit:           settings.MinimumProfit,
		MaximumBuyPrice:         settings.MaximumBuyPrice,
		MaximumSummaryAgeSecs:   settings.MaximumSummaryAgeSecs,
		MaximumConcurrentTrades: settings.MaximumConcurrentTrades,
	})
}

func (h *Handler) setTradeSettings(c *gin.Context) {
	var request struct {
		MinimumProfit           string `json:"minimumProfit"`
		MaximumBuyPrice         string `json:"maximumBuyPrice"`
		MaximumSummaryAgeSecs   int    `json:"maximumSummaryAgeSecs"`
		MaximumConcurrentTrades int    `json:"maximumConcurrentTrades"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	minimumProfit, err := decimal.NewFromString(strings.TrimSpace(request.MinimumProfit))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid minimum profit"})
		return
	}
	maximumBuyPrice, err := decimal.NewFromString(strings.TrimSpace(request.MaximumBuyPrice))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid maximum buy price"})
		return
	}
	if err = h.service.SetTradeSettings(minimumProfit, maximumBuyPrice, request.MaximumSummaryAgeSecs, request.MaximumConcurrentTrades); err != nil {
		if errors.Is(err, service.ErrInvalidMinimumProfit) || errors.Is(err, service.ErrInvalidMaximumBuyPrice) || errors.Is(err, service.ErrInvalidSummaryAge) || errors.Is(err, service.ErrInvalidConcurrentTrades) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
