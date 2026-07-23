package http

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) platformsPageOrList(c *gin.Context) {
	c.Header("Vary", "Accept")

	if strings.Contains(c.GetHeader("Accept"), "text/html") {
		c.File("public/platforms.html")
		return
	}

	h.getPlatforms(c)
}

func (h *Handler) indexPage(c *gin.Context) {
	c.File("public/index.html")
}

func (h *Handler) summaryPage(c *gin.Context) {
	c.File("public/summary.html")
}
