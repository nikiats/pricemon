package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"gopricemon/internal/domain"
	"gopricemon/internal/service"
)

type taskResponse struct {
	ID         int             `json:"id"`
	PlatformID int             `json:"platformID"`
	ActionType string          `json:"actionType"`
	Price      decimal.Decimal `json:"price"`
	LeaseToken string          `json:"leaseToken"`
	LeaseUntil time.Time       `json:"leaseUntil"`
}

func (h *Handler) claimTask(c *gin.Context) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
		return
	}

	var request struct {
		PlatformID int `json:"platformID"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	task, claimed, err := h.service.ClaimTask(request.PlatformID, token)
	if err != nil {
		handleTaskError(c, err)
		return
	}
	if !claimed {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, taskResponse{
		ID:         task.ID,
		PlatformID: task.PlatformID,
		ActionType: task.ActionType,
		Price:      task.Price,
		LeaseToken: task.LeaseToken,
		LeaseUntil: task.LeaseUntil,
	})
}

func (h *Handler) extendTaskLease(c *gin.Context) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
		return
	}

	taskID, err := strconv.Atoi(c.Param("taskID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var request struct {
		LeaseToken string `json:"leaseToken"`
		Seconds    int    `json:"seconds"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	leaseUntil, err := h.service.ExtendTaskLease(taskID, request.LeaseToken, request.Seconds, token)
	if err != nil {
		handleTaskError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"leaseUntil": leaseUntil})
}

func (h *Handler) reportTaskResult(c *gin.Context) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
		return
	}

	taskID, err := strconv.Atoi(c.Param("taskID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var request struct {
		LeaseToken string            `json:"leaseToken"`
		Status     domain.TaskStatus `json:"status"`
		Error      *string           `json:"error"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.ReportTaskResult(taskID, request.LeaseToken, request.Status, request.Error, token); err != nil {
		handleTaskError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func handleTaskError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidExecutorToken) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, service.ErrInvalidPlatformID) ||
		errors.Is(err, service.ErrInvalidTaskID) ||
		errors.Is(err, service.ErrInvalidLeaseToken) ||
		errors.Is(err, service.ErrInvalidLeaseDuration) ||
		errors.Is(err, service.ErrInvalidTaskResult) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, service.ErrTaskLeaseExpired) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	_ = c.Error(err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
