package http

import "github.com/gin-gonic/gin"

func (h *Handler) adminPage(c *gin.Context) {
	c.File("public/admin.html")
}

func (h *Handler) summaryPage(c *gin.Context) {
	c.File("public/summary.html")
}
