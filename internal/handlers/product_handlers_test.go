package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pvz/internal/models"
	"pvz/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) AddProduct(ctx context.Context, productType string, pvzID uuid.UUID, role string) (models.Product, error) {
	args := m.Called(ctx, productType, pvzID, role)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *MockProductService) DeleteProduct(ctx context.Context, pvzID uuid.UUID, role string) error {
	args := m.Called(ctx, pvzID, role)
	return args.Error(0)
}

func TestProductHandler_Add(t *testing.T) {
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)

	router := gin.Default()
	router.POST("/products", jwtAuthMock(), handler.Add)

	pvzID := uuid.New()
	validBody := map[string]string{
		"type":  "электроника",
		"pvzId": pvzID.String(),
	}

	t.Run("successful product addition", func(t *testing.T) {
		expectedProduct := models.Product{
			ProductType: validBody["type"],
			ReceptionID: uuid.New(),
		}

		mockService.On("AddProduct", mock.Anything, "электроника", pvzID, "employee").Return(expectedProduct, nil)

		jsonBody, _ := json.Marshal(validBody)
		req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.Product
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, expectedProduct.ProductType, response.ProductType)
		assert.Equal(t, expectedProduct.ReceptionID, response.ReceptionID)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid PVZ ID format", func(t *testing.T) {
		invalidBody := map[string]string{
			"type":  "electronics",
			"pvzId": "invalid-uuid",
		}

		jsonBody, _ := json.Marshal(invalidBody)
		req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request format")
	})

	t.Run("no active reception for PVZ", func(t *testing.T) {
		mockService.ExpectedCalls = []*mock.Call{}
		mockService.On("AddProduct", mock.Anything, "электроника", pvzID, "employee").Return(models.Product{}, repository.ErrNoActiveReception)

		jsonBody, _ := json.Marshal(validBody)
		req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), repository.ErrNoActiveReception.Error())
		mockService.AssertExpectations(t)
	})
}

func TestProductHandler_Delete(t *testing.T) {
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)

	router := gin.Default()
	router.DELETE("/products/:pvzId/delete_last_product", jwtAuthMock(), handler.Delete)

	pvzID := uuid.New()
	validPath := "/products/" + pvzID.String() + "/delete_last_product"
	t.Run("successful product deletion", func(t *testing.T) {
		mockService.On("DeleteProduct", mock.Anything, pvzID, "employee").Return(nil)

		req := httptest.NewRequest("DELETE", validPath, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "product deleted successfully")
		mockService.AssertExpectations(t)
	})

	t.Run("invalid PVZ ID format", func(t *testing.T) {
		invalidPath := "/products/invalid_uuid/delete_last_product"
		req := httptest.NewRequest("DELETE", invalidPath, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request format")
	})

	t.Run("PVZ not found", func(t *testing.T) {
		nonExistentPVZ := uuid.New()
		mockService.ExpectedCalls = []*mock.Call{}
		mockService.On("DeleteProduct", mock.Anything, nonExistentPVZ, "employee").Return(repository.ErrPVZNotFound)

		req := httptest.NewRequest("DELETE", "/products/"+nonExistentPVZ.String()+"/delete_last_product", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), repository.ErrPVZNotFound.Error())
		mockService.AssertExpectations(t)
	})
}
