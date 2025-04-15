package handlers

import (
	"database/sql"
	"pvz/internal/middleware"
	"pvz/internal/repository"
	"pvz/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(db *sql.DB, r *gin.Engine) {
	middleware.InitSecretKey()
	userRepo := repository.NewUserRepository(db)
	pvzRepo := repository.NewPWZRepository(db)
	receptionRepo := repository.NewReceptionRepository(db)
	productRepo := repository.NewProductRepository(db)

	userService := services.NewUserService(userRepo)
	pvzService := services.NewPVZService(pvzRepo)
	receptionService := services.NewReceptionService(receptionRepo)
	productService := services.NewProductService(productRepo)

	userHandler := NewUserHandler(userService)
	PVZHandler := NewPVZHandler(pvzService)
	receptionHandler := NewReceptionHandler(receptionService)
	productHandler := NewProductHandler(productService)

	r.POST("/dummyLogin", DummyLoginHandler)
	r.POST("/register", userHandler.Register)
	r.POST("/login", userHandler.Login)

	r.Use(middleware.JWTMiddleware())

	r.POST("/pvz", PVZHandler.CreatePVZ)
	r.GET("/pvz", PVZHandler.GetPVZInfo)
	r.PUT("/pvz/:pvzId/close_last_reception", receptionHandler.Close)
	r.DELETE("/pvz/:pvzId/delete_last_product", productHandler.Delete)
	r.POST("/reception", receptionHandler.Create)
	r.POST("/products", productHandler.Add)
}
