package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"pvz/internal/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) RegisterUser(ctx context.Context, email, password, role string) (models.User, error) {
	args := m.Called(ctx, email, password, role)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserService) LoginUser(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func TestRegisterHandler(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := gin.Default()
	router.POST("/register", handler.Register)

	t.Run("successful registration", func(t *testing.T) {
		expectedUser := models.User{
			Email: "test@example.com",
			Role:  "employee",
		}
		mockService.On("RegisterUser", mock.Anything, "test@example.com", "password123", "employee").
			Return(expectedUser, nil)

		body := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
			"role":     "employee",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.User
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, expectedUser, response)
	})

	t.Run("Invalid email", func(t *testing.T) {
		body := map[string]string{
			"email":    "invalid-email",
			"password": "password123",
			"role":     "employee",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request format")
	})

	t.Run("user already exists", func(t *testing.T) {
		mockService.ExpectedCalls = []*mock.Call{}
		mockService.On("RegisterUser", mock.Anything, "exists@example.com", "password123", "employee").
			Return(models.User{}, ErrUserExists)

		body := map[string]string{
			"email":    "exists@example.com",
			"password": "password123",
			"role":     "employee",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), ErrUserExists.Error())
	})
}

func TestLoginHandler(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := gin.Default()
	router.POST("/login", handler.Login)

	t.Run("successful login", func(t *testing.T) {
		mockService.On("LoginUser", mock.Anything, "test@example.com", "password123").
			Return("fake-jwt-token", nil)

		body := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "fake-jwt-token")
		mockService.AssertExpectations(t)
	})

	t.Run("wrong password", func(t *testing.T) {
		mockService.On("LoginUser", mock.Anything, "test@example.com", "wrong").
			Return("", ErrWrongPassword)

		body := map[string]string{
			"email":    "test@example.com",
			"password": "wrong",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid email or password")
		mockService.AssertExpectations(t)
	})

	t.Run("User not found", func(t *testing.T) {
		mockService.On("LoginUser", mock.Anything, "nonexistent@example.com", "password123").
			Return("", ErrUserDoesntExist)

		body := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid email or password")
		mockService.AssertExpectations(t)
	})
}

func TestDummyLoginHandler(t *testing.T) {
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Setenv("JWT_SECRET", originalSecret)
	}()

	router := gin.Default()
	router.POST("/dummyLogin", DummyLoginHandler)

	t.Run("Успешный Dummy-логин для employee", func(t *testing.T) {
		body := map[string]string{
			"role": "employee",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["token"])
	})

	t.Run("Успешный Dummy-логин для moderator", func(t *testing.T) {
		body := map[string]string{
			"role": "moderator",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["token"])
	})

	t.Run("Невалидная роль", func(t *testing.T) {
		body := map[string]string{
			"role": "invalid-role",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request format")
	})
}
