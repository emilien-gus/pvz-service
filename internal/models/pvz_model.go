package models

import (
	"time"

	"github.com/google/uuid"
)

type PVZ struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Registration time.Time `json:"registration" db:"registration_date"`
	City         string    `json:"city" db:"city"`
}
