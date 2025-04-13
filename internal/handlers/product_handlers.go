package handlers

import (
	"errors"
	"net/http"
	"pvz/internal/repository"
	"pvz/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	productService *services.ProductService
}

func NewProductService(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) Add(c *gin.Context) {
	var req struct {
		ProductType string `json:"type" binding:"required"`
		PVZID       string `json:"pvz_id" binding:"required"`
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

	product, err := h.productService.AddProduct(c.Request.Context(), req.ProductType, id, role)

	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		} else if errors.Is(err, repository.ErrNoActiveReception) {
			c.JSON(http.StatusBadRequest, err.Error())
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
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

	err = h.productService.DeleteProduct(c.Request.Context(), pvzId, role)
	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		} else if errors.Is(err, repository.ErrPVZNotFound) {
			c.JSON(http.StatusBadRequest, err.Error())
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message:": "product deleted successfully"})
}
