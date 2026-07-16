package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"gopricemon/internal/domain"
)

type Service interface {
	GetItems(categoryID int) ([]domain.Item, error)
	SetOffer(offer domain.Offer) error
}

type Handler struct {
	service Service
}

type itemResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CategoryID int    `json:"categoryID"`
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(router *gin.Engine) {
	router.GET("/categories/:categoryID/items", h.getItems)
	router.PUT("/items/:itemID/offers", h.setOffer)
}

func (h *Handler) getItems(c *gin.Context) {
	categoryID, err := strconv.Atoi(c.Param("categoryID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	items, err := h.service.GetItems(categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	response := make([]itemResponse, len(items))
	for i, item := range items {
		response[i] = itemResponse{item.ID, item.Name, item.CategoryID}
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) setOffer(c *gin.Context) {
	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	var request struct {
		PlatformID int              `json:"platformID"`
		Side       domain.OfferSide `json:"side"`
		Price      string           `json:"price"`
	}
	if err = c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if request.Side != domain.SideSell && request.Side != domain.SideBuy {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offer side"})
		return
	}
	if request.PlatformID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid platform ID"})
		return
	}

	price, err := decimal.NewFromString(request.Price)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price"})
		return
	}

	offer := domain.Offer{
		ItemID:     itemID,
		PlatformID: request.PlatformID,
		Side:       request.Side,
		Price:      price,
	}
	if err = h.service.SetOffer(offer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
