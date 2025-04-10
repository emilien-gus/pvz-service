package repository

import (
	"context"
	"database/sql"
	"fmt"
	"pvz/internal/models"

	sq "github.com/Masterminds/squirrel"
)

type PVZRepositoryInterface interface {
	InsertPVZ(ctx context.Context, city string) (*models.PVZ, error)
}

type PVZRepository struct {
	db *sql.DB
}

func NewPWZRepository(db *sql.DB) *PVZRepository {
	return &PVZRepository{db: db}
}

func (p *PVZRepository) InsertPVZ(ctx context.Context, city string) (*models.PVZ, error) {
	query, args, err := sq.Insert("pvz").Columns("city").Values(city).Suffix("RETURNING id, registration_date, city").ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var pvz models.PVZ
	err = p.db.QueryRowContext(ctx, query, args...).Scan(&pvz.ID, &pvz.Registration, &pvz.City)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &pvz, nil
}
