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
	}

	psqlInfo := getDBConnectionString()
	log.Printf("Connecting to DB with: %s", psqlInfo) // Логирование строки подключения

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("Database connection error: %v", err)
	}

	// Проверка подключения
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("Failed to connect database: %v", err)
	}

	log.Println("Successfully connected to database")
	return nil
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func getDBConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
}
