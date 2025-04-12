package handlers

import (
	"errors"
	"net/http"
	"pvz/internal/repository"
	"pvz/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReceptionHandler struct {
	receptionService *services.ReceptionService
}

func NewReceptionHandler(receptionService *services.ReceptionService) *ReceptionHandler {
	return &ReceptionHandler{receptionService: receptionService}
}

func (h *ReceptionHandler) Create(c *gin.Context) {
	var req struct {
		PVZID string `json:"pvz_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	id, err := uuid.Parse(req.PVZID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	role, err := getUserRole(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reception, err := h.receptionService.CreateReception(c.Request.Context(), id, role)

	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		} else if errors.Is(err, repository.ErrPVZNotFound) || errors.Is(err, repository.ErrActiveReceptionExists) {
			c.JSON(http.StatusBadRequest, err.Error())
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, reception)
}
