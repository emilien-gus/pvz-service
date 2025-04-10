package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"pvz/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

var secretKey []byte

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		if string(secretKey) == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: JWT secret is missing"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(
			tokenString,
			&services.CustomClaims{},
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(os.Getenv("JWT_SECRET")), nil
			},
		)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*services.CustomClaims); ok && token.Valid {
			c.Set("user_id", claims.UserID)
			c.Set("role", claims.Role)
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
	}
}

func InitSecretKey() {
	if err := godotenv.Load(".env.secret"); err != nil {
		log.Fatal("Ошибка загрузки .env.secret файла")
	}
	secretKey = []byte(os.Getenv("JWT_SECRET"))
}

func SetSecretKey(secret string) {
	secretKey = []byte(secret)
}
