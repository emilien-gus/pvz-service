package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `json:"-" db:"id"`
	DatedTime   time.Time `json:"-" db:"date_time"`
	ProductType string    `json:"type" db:"type"`
	ReceptionID uuid.UUID `json:"reception_id" db:"reception_id"`
}
