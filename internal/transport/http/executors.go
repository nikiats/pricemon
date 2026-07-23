package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gopricemon/internal/domain"
	"gopricemon/internal/service"
)

type executorResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

func (h *Handler) getExecutors(c *gin.Context) {
	executors, err := h.service.GetExecutors()
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, executorResponses(executors))
}

func (h *Handler) createExecutor(c *gin.Context) {
	var request struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	executor, err := h.service.CreateExecutor(request.Name)
	if err != nil {
		if errors.Is(err, service.ErrInvalidExecutorName) || errors.Is(err, service.ErrExecutorAlreadyExists) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, executorResponses([]domain.Executor{executor})[0])
}

func (h *Handler) deleteExecutor(c *gin.Context) {
	executorID, err := strconv.Atoi(c.Param("executorID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid executor ID"})
		return
	}

	err = h.service.DeleteExecutor(executorID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidExecutorID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrExecutorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) regenerateExecutorToken(c *gin.Context) {
	executorID, err := strconv.Atoi(c.Param("executorID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid executor ID"})
		return
	}

	token, err := h.service.RegenerateExecutorToken(executorID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidExecutorID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrExecutorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func executorResponses(executors []domain.Executor) []executorResponse {
	response := make([]executorResponse, len(executors))
	for i, executor := range executors {
		response[i] = executorResponse{
			ID:    executor.ID,
			Name:  executor.Name,
			Token: executor.Token,
		}
	}

	return response
}
