package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pvz/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type ProductRepositoryInterface interface {
	InsertProduct(ctx context.Context, productType string, pvzID uuid.UUID) (*models.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
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
		&product.DateTime,
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

func (r *ProductRepository) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// find id of active reception
	checkReceptionQuery, checkReceptionArgs, err := sq.
		Select("id").
		From("receptions").
		Where(sq.And{
			sq.Eq{"pvz_id": pvzID},
			sq.Eq{"status": models.ReceptionStatusInProgress},
		}).Limit(1).
		ToSql()
	if err != nil {
		return fmt.Errorf("build check reception query: %w", err)
	}

	var receptionID uuid.UUID
	err = tx.QueryRowContext(ctx, checkReceptionQuery, checkReceptionArgs...).Scan(&receptionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoActiveReception
		}
		return fmt.Errorf("get active reception: %w", err)
	}

	subQuery, subArgs, err := sq.
		Select("id").
		From("products").
		Where(sq.Eq{"reception_id": receptionID}).
		OrderBy("created_at DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return fmt.Errorf("build subquery: %w", err)
	}

	deleteSQL := fmt.Sprintf("DELETE FROM products WHERE id = (%s)", subQuery)

	result, err := tx.ExecContext(ctx, deleteSQL, subArgs...)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	// check affected rows count
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error in rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrEmptyReception
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
