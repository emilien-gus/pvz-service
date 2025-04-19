package test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"pvz/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const secret = "3q2+7wX9zY8RtKpLmN1v5xQjZ0oWcVbGyA6sFhJdU"

func initTestDB() *sql.DB {
	connStr := "host=postgres port=5432 user=postgres password=password dbname=pvz sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
		return nil
	}

	return db
}

func getToken(baseURL, role string) string {
	authRequest := map[string]string{
		"role": role,
	}
	body, _ := json.Marshal(authRequest)
	resp, _ := http.Post(baseURL+"/dummyLogin", "application/json", bytes.NewBuffer(body))

	var authResponse map[string]string
	json.NewDecoder(resp.Body).Decode(&authResponse)
	return authResponse["token"]
}

func deletePVZById(db *sql.DB, pvzId uuid.UUID) error {
	query := "DELETE FROM pvz WHERE id = $1"
	_, err := db.Exec(query, pvzId)
	return err
}

func newAuthorizedRequest(method, url, token string, body []byte) *http.Request {
	req, _ := http.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func createPVZ(t *testing.T, baseURL, token string) models.PVZ {
	body := map[string]interface{}{
		"city": "Москва",
	}
	reqBody, err := json.Marshal(body)
	assert.NoError(t, err)

	req := newAuthorizedRequest("POST", baseURL+"/pvz", token, reqBody)
	client := &http.Client{}
	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result models.PVZ
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	return result
}

func openReception(t *testing.T, baseURL, token string, pvzID uuid.UUID) models.Reception {
	body := map[string]interface{}{
		"pvzId": pvzID,
	}

	reqBody, err := json.Marshal(body)
	assert.NoError(t, err)

	req := newAuthorizedRequest("POST", baseURL+"/reception", token, reqBody)
	client := &http.Client{}
	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result models.Reception
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	return result
}

func addProduct(t *testing.T, baseURL, token string, productType string, pvzId uuid.UUID) models.Product {
	body := map[string]interface{}{
		"type":  productType,
		"pvzId": pvzId,
	}

	reqBody, err := json.Marshal(body)
	assert.NoError(t, err)

	req := newAuthorizedRequest("POST", baseURL+"/products", token, reqBody)
	client := &http.Client{}
	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result models.Product
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	return result
}

func closeReception(t *testing.T, baseURL, token string, pvzId uuid.UUID) models.Reception {
	req := newAuthorizedRequest("PUT", baseURL+"/pvz/"+pvzId.String()+"/close_last_reception", token, nil)
	client := &http.Client{}
	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result models.Reception
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	return result
}
