package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sql.DB

// Init database connection
func InitDB() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
		return err
	}
	psqlInfo := getDBConnectionString()

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("Database connection error: %v", err)
	}

	err = DB.Ping() // making sure db works and responds
	if err != nil {
		return fmt.Errorf("Failed to connect database: %v", err)
	}
	return nil
}

// Closing database
func CloseDB() error {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("DB is nil at closing")
}

func getDBConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
}
