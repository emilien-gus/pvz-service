package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ReceptionStatusInProgress = "in_progress"
	ReceptionStatusClosed     = "close"
)

type Reception struct {
	ID       uuid.UUID `json:"-" db:"id"`
	DateTime time.Time `json:"dateTime" db:"date_time"`
	PVZID    uuid.UUID `json:"pvzId" db:"pvz_id"`
	Status   string    `json:"status" db:"status"` // "in_progress" or "closed"
}

type ReceptionWithProducts struct {
	Reception `json:"reception"`
	Products  []Product `json:"products"`
}
