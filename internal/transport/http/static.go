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

func (h *Handler) executorsPageOrList(c *gin.Context) {
	c.Header("Vary", "Accept")

	if strings.Contains(c.GetHeader("Accept"), "text/html") {
		c.File("public/executors.html")
		return
	}

	h.getExecutors(c)
}

func (h *Handler) tasksPageOrList(c *gin.Context) {
	c.Header("Vary", "Accept")

	if strings.Contains(c.GetHeader("Accept"), "text/html") {
		c.File("public/tasks.html")
		return
	}

	h.getTasks(c)
}

func (h *Handler) inventoryPageOrList(c *gin.Context) {
	c.Header("Vary", "Accept")

	if strings.Contains(c.GetHeader("Accept"), "text/html") {
		c.File("public/inventory.html")
		return
	}

	h.getInventory(c)
}

func (h *Handler) indexPage(c *gin.Context) {
	c.File("public/index.html")
}

func (h *Handler) summaryPage(c *gin.Context) {
	c.File("public/summary.html")
}

func (h *Handler) sequentialTasksPageOrJSON(c *gin.Context) {
	c.Header("Vary", "Accept")

	if strings.Contains(c.GetHeader("Accept"), "text/html") {
		c.File("public/sequential-tasks.html")
		return
	}

	h.getSequentialTasks(c)
}

func (h *Handler) tradeSettingsPageOrJSON(c *gin.Context) {
	c.Header("Vary", "Accept")

	if strings.Contains(c.GetHeader("Accept"), "text/html") {
		c.File("public/trade-settings.html")
		return
	}

	h.getTradeSettings(c)
}
