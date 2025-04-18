package test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

const secret = "3q2+7wX9zY8RtKpLmN1v5xQjZ0oWcVbGyA6sFhJdU"

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
