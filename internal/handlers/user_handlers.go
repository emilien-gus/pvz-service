package handlers

import (
	"net/http"
	"pvz/internal/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserServiceInterface
}

func NewUserHandler(userService services.UserServiceInterface) *UserHandler {
	return &UserHandler{userService: userService}
}

func (u *UserHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required,oneof=employee moderator"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	user, err := u.userService.RegisterUser(c.Request.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		if err == ErrUserExists {
			c.JSON(http.StatusBadRequest, gin.H{"error: ": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error: ": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (u *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	token, err := u.userService.LoginUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == ErrWrongPassword || err == ErrUserDoesntExist {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func DummyLoginHandler(c *gin.Context) {
	var req struct {
		Role string `json:"role" binding:"required,oneof=employee moderator"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	token, err := services.DummyLogin(req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
