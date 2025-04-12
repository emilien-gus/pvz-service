package handlers

import (
	"errors"
	"net/http"
	"pvz/internal/services"

	"github.com/gin-gonic/gin"
)

type PVZHandler struct {
	pvzService services.PVZService
}

func NewPVZService(pvzService services.PVZService) *PVZHandler {
	return &PVZHandler{pvzService: pvzService}
}

func (h *PVZHandler) CreatePVZ(c *gin.Context) {
	var req struct {
		City string `json:"city" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	role, err := getUserRole(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pvz, err := h.pvzService.CreatePVZ(c.Request.Context(), req.City, role)
	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		} else if errors.Is(err, services.ErrCityNotAllowed) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, pvz)
}
