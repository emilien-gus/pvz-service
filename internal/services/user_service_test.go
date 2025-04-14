package services

import (
	"context"
	"os"
	"pvz/internal/models"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) InsertUser(ctx context.Context, email, password, role string) (*models.User, error) {
	args := m.Called(ctx, email, password, role)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

func TestUserService_RegisterUser(t *testing.T) {
	mockRepo := new(MockUserRepo)
	userService := NewUserService(mockRepo)

	email := "test@example.com"
	password := "password123"
	role := "moderator"
	user := &models.User{
		ID:    uuid.New(),
		Email: email,
		Role:  role,
	}

	mockRepo.On("InsertUser", mock.Anything, email, password, role).Return(user, nil)

	t.Run("successful user registration", func(t *testing.T) {
		createdUser, err := userService.RegisterUser(context.Background(), email, password, role)

		assert.NoError(t, err)
		assert.Equal(t, email, createdUser.Email)
		assert.Equal(t, role, createdUser.Role)
	})
}

func TestUserService_LoginUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepo)
	userService := NewUserService(mockRepo)

	email := "test@example.com"
	password := "securepass"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.User{
		ID:       uuid.New(),
		Email:    email,
		Password: string(hashedPassword),
		Role:     "staff",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(user, nil)
	os.Setenv("JWT_SECRET", "supersecret")

	t.Run("successful login", func(t *testing.T) {
		token, err := userService.LoginUser(context.Background(), email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

func TestUserService_LoginUser_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepo)
	userService := NewUserService(mockRepo)

	email := "user@example.com"
	user := &models.User{
		ID:       uuid.New(),
		Email:    email,
		Password: "$2a$10$invalidhash", // wrong hash
	}

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(user, nil)

	t.Run("invalid password", func(t *testing.T) {
		token, err := userService.LoginUser(context.Background(), email, "wrongpassword")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "wrong password", err.Error())
	})
}

func TestUserService_LoginUser_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepo)
	userService := NewUserService(mockRepo)

	email := "notfound@example.com"

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return((*models.User)(nil), nil)

	t.Run("user not found", func(t *testing.T) {
		token, err := userService.LoginUser(context.Background(), email, "any")

		assert.Error(t, err)
		assert.Equal(t, "no such user", err.Error())
		assert.Empty(t, token)
	})
}

func TestDummyLogin_Success(t *testing.T) {
	_ = os.Setenv("JWT_SECRET", "testsecret")

	t.Run("successful dummy login", func(t *testing.T) {
		token, err := DummyLogin("moderator")

		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		parsedToken, parseErr := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("testsecret"), nil
		})

		assert.NoError(t, parseErr)
		assert.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(*CustomClaims)
		assert.True(t, ok)
		assert.Equal(t, "moderator", claims.Role)
		assert.NotEmpty(t, claims.UserID)
	})
}

func TestDummyLogin_ValidRole(t *testing.T) {
	t.Run("invalid role", func(t *testing.T) {
		role := ""
		_, err := DummyLogin(role)

		assert.ErrorIs(t, err, ErrInvalidRole)
	})
}
