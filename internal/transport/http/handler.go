package http

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"gopricemon/internal/domain"
)

type Service interface {
	GetSummary(categoryID, offset, limit int, maxAge *int) (domain.SummaryPage, error)
	SetOffer(offer domain.Offer) error
	SetZeroCount(categoryID int, itemName string, platformID int, platformToken string) (bool, error)
	GetPlatforms() ([]domain.Platform, error)
	CreatePlatform(name string) (domain.Platform, error)
	DeletePlatform(platformID int) error
	RegeneratePlatformToken(platformID int) (string, error)
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
	router.PUT("/offers", h.setOffer)
	router.PUT("/offers/zero-count", h.setZeroCount)

	platforms := router.Group("/platforms")
	platforms.GET("", h.platformsPageOrList)
	platforms.POST("", h.createPlatform)
	platforms.DELETE("/:platformID", h.deletePlatform)
	platforms.PUT("/:platformID/token", h.regeneratePlatformToken)
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
