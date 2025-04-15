package repository

import (
	"context"
	"database/sql"
	"fmt"
	"pvz/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ReceptionRepositoryInterface interface {
	InsertReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	UpdateLastReceptionStatus(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
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

	checkQuery, checkArgs, err := sq.Select("1").
		From("receptions").
		Where(sq.And{
			sq.Eq{"pvz_id": pvzID},
			sq.Eq{"status": models.ReceptionStatusInProgress},
		}).Limit(1).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build check query: %w", err)
	}

	var exists bool
	err = tx.QueryRowContext(ctx, checkQuery, checkArgs...).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			exists = false // no in_progress rows
		} else {
			return nil, fmt.Errorf("failed to check active reception: %w", err)
		}
	}

	if exists {
		return nil, ErrActiveReceptionExists
	}

	id := uuid.New()
	insertQuery, insertArgs, err := sq.Insert("receptions").
		Columns("id, pvz_id").
		Values(id, pvzID).
		Suffix("RETURNING id, date_time, pvz_id, status").
		PlaceholderFormat(sq.Dollar).
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
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return nil, ErrPVZNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &reception, nil
}

func (r *ReceptionRepository) UpdateLastReceptionStatus(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	query, args, err := sq.
		Update("receptions").
		Set("status", models.ReceptionStatusClosed).
		Where(`
            id = (
                SELECT id FROM receptions 
                WHERE pvz_id = $2 AND status = $3
                ORDER BY date_time DESC 
                LIMIT 1
            )`,
			pvzID,
			models.ReceptionStatusInProgress,
		).
		Suffix("RETURNING id, date_time, pvz_id, status").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build update query: %w", err)
	}

	var reception models.Reception
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&reception.ID,
		&reception.DateTime,
		&reception.PVZID,
		&reception.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoActiveReception
		}
		return nil, fmt.Errorf("execute update: %w", err)
	}

	return &reception, nil
}
