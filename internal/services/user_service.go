package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"pvz/internal/models"
	"pvz/internal/repository"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CustomClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.StandardClaims
}

var allowedRoles = map[string]bool{
	"moderator": true,
	"employee":  true,
}

type UserService struct {
	userRepo repository.UserRepositoryInterface
}

func NewUserService(userRepo repository.UserRepositoryInterface) *UserService {
	return &UserService{userRepo: userRepo}
}

func (u *UserService) RegisterUser(ctx context.Context, email, password, role string) (models.User, error) {
	user := &models.User{}

	user, err := u.userRepo.InsertUser(ctx, email, password, role)
	if err != nil {
		return *user, err
	}

	return *user, nil
}

func (u *UserService) LoginUser(ctx context.Context, email, password string) (string, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", errors.New("no such user")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("wrong password")
	}

	token, err := generateJWT(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func DummyLogin(role string) (string, error) {
	if _, ok := allowedRoles[role]; !ok {
		return "", ErrInvalidRole
	}

	user := models.User{
		ID:   uuid.New(),
		Role: role,
	}

	token, err := generateJWT(&user)
	if err != nil {
		return "", fmt.Errorf("failed to generate token for DummyLogin: %w", err)
	}
	return token, nil
}

func generateJWT(user *models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET not set")
	}

	claims := CustomClaims{
		UserID: user.ID,
		Role:   user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
