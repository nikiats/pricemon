package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gopricemon/internal/domain"
	"gopricemon/internal/service"
)

type platformResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

func (h *Handler) getPlatforms(c *gin.Context) {
	platforms, err := h.service.GetPlatforms()
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, platformResponses(platforms))
}

func (h *Handler) createPlatform(c *gin.Context) {
	var request struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	platform, err := h.service.CreatePlatform(request.Name)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPlatformName) || errors.Is(err, service.ErrPlatformAlreadyExists) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, platformResponses([]domain.Platform{platform})[0])
}

func (h *Handler) deletePlatform(c *gin.Context) {
	platformID, err := strconv.Atoi(c.Param("platformID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid platform ID"})
		return
	}

	err = h.service.DeletePlatform(platformID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPlatformID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrPlatformNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) regeneratePlatformToken(c *gin.Context) {
	platformID, err := strconv.Atoi(c.Param("platformID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid platform ID"})
		return
	}

	token, err := h.service.RegeneratePlatformToken(platformID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPlatformID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func platformResponses(platforms []domain.Platform) []platformResponse {
	response := make([]platformResponse, len(platforms))
	for i, platform := range platforms {
		response[i] = platformResponse{
			ID:    platform.ID,
			Name:  platform.Name,
			Token: platform.Token,
		}
	}

	return response
}
