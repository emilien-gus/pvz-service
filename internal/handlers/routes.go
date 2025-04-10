package handlers

import (
	"database/sql"
	"pvz/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(db *sql.DB, r *gin.Engine) {
	middleware.InitSecretKey()

	r.Use(middleware.JWTMiddleware())
}
