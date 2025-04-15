package handlers

import (
	"errors"
	"net/http"
	"pvz/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type PVZHandler struct {
	pvzService services.PVZServiceInterface
}

func NewPVZHandler(pvzService services.PVZServiceInterface) *PVZHandler {
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

func (h *PVZHandler) GetPVZInfo(c *gin.Context) {
	startDateStr := c.DefaultQuery("startDate", "")
	endDateStr := c.DefaultQuery("endDate", "")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page is not int"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit is not int"})
		return
	}

	var startDate, endDate *time.Time
	if startDateStr != "" {
		parsedStartDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid startDate format"})
			return
		}
		startDate = &parsedStartDate
	}

	if endDateStr != "" {
		parsedEndDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endDate format"})
			return
		}
		endDate = &parsedEndDate
	}
	role, err := getUserRole(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pvzs, err := h.pvzService.GetPVZList(c.Request.Context(), startDate, endDate, page, limit, role)
	if err != nil {
		switch err {
		case services.ErrAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case services.ErrPageParamIsInvalid, services.ErrLimitParamIsInvalid, services.ErrStartLaterThenEnd:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, pvzs)
}
