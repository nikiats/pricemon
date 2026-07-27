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

type taskInfoResponse struct {
	ID           int               `json:"id"`
	ItemName     *string           `json:"itemName"`
	CategoryName *string           `json:"categoryName"`
	PlatformName string            `json:"platformName"`
	ExecutorName *string           `json:"executorName"`
	ActionType   string            `json:"actionType"`
	Price        decimal.Decimal   `json:"price"`
	Status       domain.TaskStatus `json:"status"`
	Error        *string           `json:"error"`
	LeaseUntil   *time.Time        `json:"leaseUntil"`
	CompletedAt  *time.Time        `json:"completedAt"`
}

func (h *Handler) getTasks(c *gin.Context) {
	tasks, err := h.service.GetTasks()
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if len(tasks) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	response := make([]taskInfoResponse, len(tasks))
	for i, task := range tasks {
		response[i] = taskInfoResponse{
			ID:           task.ID,
			ItemName:     task.ItemName,
			CategoryName: task.CategoryName,
			PlatformName: task.PlatformName,
			ExecutorName: task.ExecutorName,
			ActionType:   task.ActionType,
			Price:        task.Price,
			Status:       task.Status,
			Error:        task.Error,
			LeaseUntil:   task.LeaseUntil,
			CompletedAt:  task.CompletedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) getTask(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("taskID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	task, err := h.service.GetTask(taskID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTaskID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, taskInfoResponse{
		ID:           task.ID,
		ItemName:     task.ItemName,
		CategoryName: task.CategoryName,
		PlatformName: task.PlatformName,
		ExecutorName: task.ExecutorName,
		ActionType:   task.ActionType,
		Price:        task.Price,
		Status:       task.Status,
		Error:        task.Error,
		LeaseUntil:   task.LeaseUntil,
		CompletedAt:  task.CompletedAt,
	})
}

func (h *Handler) createTask(c *gin.Context) {
	var request struct {
		ItemID       int    `json:"itemID"`
		PlatformName string `json:"platformName"`
		ActionType   string `json:"actionType"`
		Price        string `json:"price"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	task, err := h.service.CreateTask(request.ItemID, request.PlatformName, request.ActionType, request.Price)
	if err != nil {
		if errors.Is(err, service.ErrInvalidItemID) ||
			errors.Is(err, service.ErrInvalidTaskPlatform) ||
			errors.Is(err, service.ErrInvalidTaskAction) ||
			errors.Is(err, service.ErrInvalidTaskPrice) ||
			errors.Is(err, service.ErrTaskReferencesNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, taskInfoResponse{
		ID:           task.ID,
		ItemName:     task.ItemName,
		CategoryName: task.CategoryName,
		PlatformName: task.PlatformName,
		ExecutorName: task.ExecutorName,
		ActionType:   task.ActionType,
		Price:        task.Price,
		Status:       task.Status,
		Error:        task.Error,
		LeaseUntil:   task.LeaseUntil,
		CompletedAt:  task.CompletedAt,
	})
}

func (h *Handler) deleteTask(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("taskID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	if err := h.service.DeleteTask(taskID); err != nil {
		if errors.Is(err, service.ErrInvalidTaskID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrTaskNotCancellable) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
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

	reportedStatus, err := h.service.ReportTaskResult(taskID, request.LeaseToken, request.Status, request.Error, token)
	if err != nil {
		handleTaskResultError(c, reportedStatus, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func handleTaskResultError(c *gin.Context, status domain.TaskStatus, err error) {
	if errors.Is(err, service.ErrTaskResultReported) {
		message := err.Error()
		if status == domain.TaskStatusCompleted {
			message = "task is already completed"
		}
		c.JSON(http.StatusConflict, gin.H{"error": message, "status": status})
		return
	}
	if errors.Is(err, service.ErrTaskBelongsToExecutor) ||
		errors.Is(err, service.ErrTaskResultLeaseExpired) ||
		errors.Is(err, service.ErrTaskLeaseInactive) ||
		errors.Is(err, service.ErrTaskNotInProgress) ||
		errors.Is(err, service.ErrTaskNotFound) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	handleTaskError(c, err)
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
