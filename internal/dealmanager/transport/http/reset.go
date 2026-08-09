package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) resetSequentialTasks(c *gin.Context) {
	if err := h.sequentialTasksService.ResetActiveSequentialTasks(); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
