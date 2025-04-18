package test

import (
	"avito-shop/internal/handlers"
	"avito-shop/internal/middleware"
	"avito-shop/internal/repository"
	"avito-shop/internal/services"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"pvz/internal/data"
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
	err := data.InitDB()
	assert.NoError(t, err)
	defer func() {
		data.CloseDB()
		assert.NoError(t, err)
	}()

	router := setupRoutesForBasicTest(data.DB)
	server := httptest.NewServer(router)
	defer server.Close()

	employeeToken := getToken(server.URL, "employee")
	moderatorToken := getToken(server.URL, "moderator")

	req, _ := http.NewRequest("POST", server.URL+"/pvz", nil)
	req.Header.Set("Authorization", "Bearer "+moderatorToken)

	client := &http.Client{}
	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// err = deleteUserByID(db, user.ID)
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }

}
