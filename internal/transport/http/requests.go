package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"gopricemon/internal/domain"
	"gopricemon/internal/service"
)

const defaultPageSize = 50

type itemResponse struct {
	ID                   int              `json:"id"`
	Name                 string           `json:"name"`
	CategoryID           int              `json:"categoryID"`
	SellPrice            *decimal.Decimal `json:"sellPrice"`
	SellCount            *int             `json:"sellCount"`
	SellPlatformName     *string          `json:"sellPlatformName"`
	SellURL              *string          `json:"sellUrl"`
	BuyPrice             *decimal.Decimal `json:"buyPrice"`
	BuyCount             *int             `json:"buyCount"`
	BuyPlatformName      *string          `json:"buyPlatformName"`
	BuyURL               *string          `json:"buyUrl"`
	OldestOfferUpdatedAt *time.Time       `json:"oldestOfferUpdatedAt"`
}

type summaryResponse struct {
	Items      []itemResponse `json:"items"`
	NextOffset *int           `json:"nextOffset,omitempty"`
}

func (h *Handler) getSummary(c *gin.Context) {
	categoryID, err := strconv.Atoi(c.Param("categoryID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	offset, limit, err := pagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
		return
	}
	maxAge, err := queryOptionalInt(c, "max_age")
	if err != nil || (maxAge != nil && *maxAge < 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid max_age"})
		return
	}

	page, err := h.service.GetSummary(categoryID, offset, limit, maxAge)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCategoryID) || errors.Is(err, service.ErrInvalidPagination) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	items := make([]itemResponse, len(page.Items))
	for i, item := range page.Items {
		items[i] = itemResponse{
			ID:                   item.ID,
			Name:                 item.Name,
			CategoryID:           item.CategoryID,
			SellPrice:            item.SellPrice,
			SellCount:            item.SellCount,
			SellPlatformName:     item.SellPlatformName,
			SellURL:              item.SellURL,
			BuyPrice:             item.BuyPrice,
			BuyCount:             item.BuyCount,
			BuyPlatformName:      item.BuyPlatformName,
			BuyURL:               item.BuyURL,
			OldestOfferUpdatedAt: item.OldestOfferUpdatedAt,
		}
	}

	c.JSON(http.StatusOK, summaryResponse{Items: items, NextOffset: page.NextOffset})
}

func pagination(c *gin.Context) (int, int, error) {
	offset, err := queryInt(c, "offset", 0)
	if err != nil {
		return 0, 0, err
	}

	limit, err := queryInt(c, "limit", defaultPageSize)
	if err != nil {
		return 0, 0, err
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

func queryOptionalInt(c *gin.Context, name string) (*int, error) {
	value := c.Query(name)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (h *Handler) setOffer(c *gin.Context) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
		return
	}

	var request struct {
		CategoryID int              `json:"categoryID"`
		ItemName   string           `json:"itemName"`
		PlatformID int              `json:"platformID"`
		Side       domain.OfferSide `json:"side"`
		Price      string           `json:"price"`
		Count      int              `json:"count"`
		URL        *string          `json:"url"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	price, err := decimal.NewFromString(request.Price)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price"})
		return
	}

	offer := domain.Offer{
		CategoryID:    request.CategoryID,
		ItemName:      request.ItemName,
		PlatformID:    request.PlatformID,
		PlatformToken: token,
		Side:          request.Side,
		Price:         price,
		Count:         request.Count,
		URL:           request.URL,
	}
	if err := h.service.SetOffer(offer); err != nil {
		if errors.Is(err, service.ErrInvalidPlatformToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrInvalidOffer) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) setZeroCount(c *gin.Context) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
		return
	}

	var request struct {
		CategoryID int    `json:"categoryID"`
		ItemName   string `json:"itemName"`
		PlatformID int    `json:"platformID"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	updated, err := h.service.SetZeroCount(request.CategoryID, request.ItemName, request.PlatformID, token)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPlatformToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrInvalidOffer) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !updated {
		c.JSON(http.StatusOK, gin.H{"result": "no offer modified"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "offer count set to zero"})
}

func bearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(header, " ")
	return token, ok && strings.EqualFold(scheme, "Bearer") && token != ""
}
