package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pvz/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ReceptionRepositoryInterface interface {
	InsertReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
}

type ReceptionRepository struct {
	db *sql.DB
}

func NewReceptionRepository(db *sql.DB) *ReceptionRepository {
	return &ReceptionRepository{db: db}
}

func (r *ReceptionRepository) InsertReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	checkQuery, checkArgs, err := sq.Select("Count(*)").
		From("reception").
		Where(sq.And{
			sq.Eq{"pvz_id": pvzID},
			sq.Eq{"status": models.ReceptionStatusInProgress},
		}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build check query: %w", err)
	}

	var InProgressCount int
	err = tx.QueryRowContext(ctx, checkQuery, checkArgs...).Scan(&InProgressCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check active reception: %w", err)
	}

	if InProgressCount > 0 {
		return nil, ErrActiveReceptionExists
	}

	insertQuery, insertArgs, err := sq.Insert("reception").
		Columns("pvz_id").
		Values(pvzID).
		Suffix("RETURNING id, date_time, pvz_id, status").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	var reception models.Reception
	err = tx.QueryRowContext(ctx, insertQuery, insertArgs...).Scan(
		&reception.ID,
		&reception.DateTime,
		&reception.PVZID,
		&reception.Status,
	)

	if err != nil {
		//foreign_key_violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrPVZNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &reception, nil
}
