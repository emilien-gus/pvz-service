package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func getUserRole(c *gin.Context) (string, error) {
	const roleKey = "role"

	roleValue, exists := c.Get(roleKey)
	if !exists {
		return "", fmt.Errorf("user role not found in context")
	}

	role, ok := roleValue.(string)
	if !ok {
		return "", fmt.Errorf("invalid role type: expected string, got %T", roleValue)
	}

	return role, nil
}
