package test

import (
	"database/sql"
	"net/http/httptest"
	"pvz/internal/handlers"
	"pvz/internal/middleware"
	"pvz/internal/repository"
	"pvz/internal/services"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func setupRoutesForBasicTest(db *sql.DB) *gin.Engine {
	r := gin.Default()

	middleware.SetSecretKey(secret)

	pvzRepo := repository.NewPWZRepository(db)
	receptionRepo := repository.NewReceptionRepository(db)
	productRepo := repository.NewProductRepository(db)

	pvzService := services.NewPVZService(pvzRepo)
	receptionService := services.NewReceptionService(receptionRepo)
	productService := services.NewProductService(productRepo)

	PVZHandler := handlers.NewPVZHandler(pvzService)
	receptionHandler := handlers.NewReceptionHandler(receptionService)
	productHandler := handlers.NewProductHandler(productService)

	r.POST("/dummyLogin", handlers.DummyLoginHandler)

	r.Use(middleware.JWTMiddleware())

	r.POST("/pvz", PVZHandler.CreatePVZ)
	r.GET("/pvz", PVZHandler.GetPVZInfo)
	r.PUT("/pvz/:pvzId/close_last_reception", receptionHandler.Close)
	r.POST("/reception", receptionHandler.Create)
	r.POST("/products", productHandler.Add)

	return r
}

// test creating pvz, opening reception, adding 50 products.
func TestBasicScenario(t *testing.T) {
	db := initTestDB()
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	middleware.SetSecretKey(secret)
	router := setupRoutesForBasicTest(db)
	server := httptest.NewServer(router)
	defer server.Close()

	employeeToken := getToken(server.URL, "employee")
	moderatorToken := getToken(server.URL, "moderator")

	pvz := createPVZ(t, server.URL, moderatorToken)

	openReception(t, server.URL, employeeToken, pvz.ID)

	for range 50 {
		addProduct(t, server.URL, employeeToken, "обувь", pvz.ID)
	}

	closeReception(t, server.URL, employeeToken, pvz.ID)
	err := deletePVZById(db, pvz.ID)
	assert.NoError(t, err)
}
