package repository

import (
	"context"
	"database/sql"
	"fmt"
	"pvz/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type ProductRepositoryInterface interface {
	InsertProduct(ctx context.Context, productType string, pvzID uuid.UUID) (*models.Product, error)
}

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) InsertProduct(ctx context.Context, productType string, pvzID uuid.UUID) (*models.Product, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	checkQuery, checkArgs, err := sq.Select("id").
		From("reception").
		Where(sq.And{
			sq.Eq{"pvz_id": pvzID},
			sq.Eq{"status": models.ReceptionStatusInProgress},
		}).
		Limit(1).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build check query: %w", err)
	}

	var receptionID uuid.UUID
	err = tx.QueryRowContext(ctx, checkQuery, checkArgs...).Scan(&receptionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoActiveReception
		}
		return nil, fmt.Errorf("failed to check active reception: %w", err)
	}

	id := uuid.New()
	insertQuery, insertArgs, err := sq.Insert("products").
		Columns("id, reception_id").
		Values(id, receptionID).
		Suffix("RETURNING id, date_time, type, reception_id").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	var product models.Product
	err = tx.QueryRowContext(ctx, insertQuery, insertArgs...).Scan(
		&product.ID,
		&product.DatedTime,
		&product.ProductType,
		&product.ReceptionID,
	)

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &product, nil
}
