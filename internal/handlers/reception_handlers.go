package handlers

import (
	"net/http"
	"pvz/internal/repository"
	"pvz/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReceptionHandler struct {
	receptionService services.ReceptionServiceInterface
}

func NewReceptionHandler(receptionService services.ReceptionServiceInterface) *ReceptionHandler {
	return &ReceptionHandler{receptionService: receptionService}
}

func (h *ReceptionHandler) Create(c *gin.Context) {
	var req struct {
		PVZID string `json:"pvzId" binding:"required"`
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
		if err == services.ErrAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else if err == repository.ErrActiveReceptionExists || err == repository.ErrPVZNotFound {
			c.JSON(http.StatusBadRequest, err.Error())
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, reception)
}

func (h *ReceptionHandler) Close(c *gin.Context) {
	pvzIdRaw := c.Param("pvzId")

	pvzId, err := uuid.Parse(pvzIdRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	role, err := getUserRole(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reception, err := h.receptionService.CloseReception(c.Request.Context(), pvzId, role)

	if err != nil {
		if err == services.ErrAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else if err == repository.ErrNoActiveReception {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, reception)
}
