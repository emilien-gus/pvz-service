package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pvz/internal/models"
	"pvz/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockReceptionService struct {
	mock.Mock
}

func (m *MockReceptionService) CreateReception(ctx context.Context, pvzID uuid.UUID, role string) (models.Reception, error) {
	args := m.Called(ctx, pvzID, role)
	return args.Get(0).(models.Reception), args.Error(1)
}

func (m *MockReceptionService) CloseReception(ctx context.Context, pvzID uuid.UUID, role string) (models.Reception, error) {
	args := m.Called(ctx, pvzID, role)
	return args.Get(0).(models.Reception), args.Error(1)
}

func TestReceptionHandler_Create(t *testing.T) {
	mockService := new(MockReceptionService)
	handler := NewReceptionHandler(mockService)

	router := gin.Default()
	router.POST("/receptions", jwtAuthMock(), handler.Create)

	pvzID := uuid.New()
	validBody := map[string]string{"pvzId": pvzID.String()}

	t.Run("successful reception create", func(t *testing.T) {
		expectedReception := models.Reception{
			ID:       uuid.New(),
			DateTime: time.Now(),
			PVZID:    pvzID,
			Status:   models.ReceptionStatusInProgress,
		}

		mockService.On("CreateReception", mock.Anything, pvzID, "employee").Return(expectedReception, nil)

		jsonBody, _ := json.Marshal(validBody)
		req := httptest.NewRequest("POST", "/receptions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.Reception
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response.DateTime)
		assert.Equal(t, expectedReception.PVZID, response.PVZID)
		assert.Equal(t, expectedReception.Status, response.Status)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		invalidBody := map[string]string{"pvzId": "invalid-uuid"}
		jsonBody, _ := json.Marshal(invalidBody)

		req := httptest.NewRequest("POST", "/receptions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request format")
	})

	t.Run("active reception already exists", func(t *testing.T) {
		mockService.ExpectedCalls = []*mock.Call{}
		mockService.On("CreateReception", mock.Anything, pvzID, "employee").Return(models.Reception{}, repository.ErrActiveReceptionExists)

		jsonBody, _ := json.Marshal(validBody)
		req := httptest.NewRequest("POST", "/receptions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), repository.ErrActiveReceptionExists.Error())
		mockService.AssertExpectations(t)
	})
}

func TestReceptionHandler_Close(t *testing.T) {
	mockService := new(MockReceptionService)
	handler := NewReceptionHandler(mockService)

	router := gin.Default()
	router.PUT("/receptions/:pvzId/close_last_reception", jwtAuthMock(), handler.Close)

	pvzID := uuid.New()
	validPath := "/receptions/" + pvzID.String() + "/close_last_reception"

	t.Run("successful reception close", func(t *testing.T) {
		expectedReception := models.Reception{
			ID:       uuid.New(),
			DateTime: time.Now(),
			PVZID:    pvzID,
			Status:   models.ReceptionStatusInProgress,
		}

		mockService.On("CloseReception", mock.Anything, pvzID, "employee").Return(expectedReception, nil)

		req := httptest.NewRequest("PUT", validPath, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.Reception
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response.DateTime)
		assert.Equal(t, expectedReception.PVZID, response.PVZID)
		assert.Equal(t, expectedReception.Status, response.Status)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		invalidPath := "/receptions/" + "invalid_uuid" + "/close_last_reception"

		req := httptest.NewRequest("PUT", invalidPath, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("no active reception", func(t *testing.T) {
		mockService.ExpectedCalls = []*mock.Call{}
		mockService.On("CloseReception", mock.Anything, pvzID, "employee").Return(models.Reception{}, repository.ErrNoActiveReception)

		req := httptest.NewRequest("PUT", validPath, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), repository.ErrNoActiveReception.Error())
		mockService.AssertExpectations(t)
	})
}

func jwtAuthMock() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("role", "employee")
		c.Next()
	}
}
