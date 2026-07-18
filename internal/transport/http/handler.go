package http

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"

	"gopricemon/internal/domain"
)

type Service interface {
	GetOffers(categoryID, afterID, limit int) (domain.OfferPage, error)
	SetOffer(offer domain.Offer) error
	GetPlatforms() ([]domain.Platform, error)
	CreatePlatform(name string) (domain.Platform, error)
	RegeneratePlatformToken(platformID int) (string, error)
}

type Handler struct {
	service       Service
	adminPassword string
}

func NewHandler(service Service, adminPassword string) *Handler {
	return &Handler{service: service, adminPassword: adminPassword}
}

func (h *Handler) Register(router *gin.Engine) {
	router.Use(cors)

	router.GET("/categories/:categoryID/offers", h.getOffers)
	router.PUT("/offers", h.setOffer)

	admin := router.Group("/admin", h.adminAuth)
	admin.GET("", h.adminPage)
	admin.GET("/platforms", h.getPlatforms)
	admin.POST("/platforms", h.createPlatform)
	admin.PUT("/platforms/:platformID/token", h.regeneratePlatformToken)
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

func (h *Handler) adminAuth(c *gin.Context) {
	username, password, ok := c.Request.BasicAuth()
	valid := ok && username == "admin" && subtle.ConstantTimeCompare([]byte(password), []byte(h.adminPassword)) == 1
	if !valid {
		c.Header("WWW-Authenticate", `Basic realm="admin"`)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Next()
}
