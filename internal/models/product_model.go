package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `json:"-" db:"id"`
	DateTime    time.Time `json:"-" db:"date_time"`
	ProductType string    `json:"type" db:"type"`
	ReceptionID uuid.UUID `json:"receptionId" db:"reception_id"`
}
