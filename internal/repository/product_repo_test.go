package repository

import (
	"context"
	"database/sql"
	"pvz/internal/models"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	productSelectInInsertQuery = regexp.QuoteMeta(`SELECT id FROM reception WHERE (pvz_id = ? AND status = ?) LIMIT 1`)
	productInsertQuery         = regexp.QuoteMeta(`INSERT INTO products (id, reception_id) VALUES (?,?) RETURNING id, date_time, type, reception_id`)
	productSelectInDeleteQuery = regexp.QuoteMeta(`SELECT id FROM receptions WHERE (pvz_id = ? AND status = ?) LIMIT 1`)
	productDeleteQuery         = regexp.QuoteMeta(`DELETE FROM products WHERE id = (SELECT id FROM products WHERE reception_id = ? ORDER BY created_at DESC LIMIT 1)`)
)

func TestProductRepository_InsertProduct_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)
	pvzID := uuid.New()
	receptionID := uuid.New()
	productID := uuid.New()
	now := time.Now()

	mock.ExpectBegin()

	mock.ExpectQuery(productSelectInInsertQuery).
		WithArgs(pvzID, models.ReceptionStatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(receptionID))

	mock.ExpectQuery(productInsertQuery).
		WithArgs(sqlmock.AnyArg(), receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
			AddRow(productID, now, "", receptionID))

	mock.ExpectCommit()

	result, err := repo.InsertProduct(context.Background(), "", pvzID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, productID, result.ID)
	assert.Equal(t, receptionID, result.ReceptionID)
}

func TestProductRepository_InsertProduct_NoActiveReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)
	pvzID := uuid.New()

	mock.ExpectBegin()

	mock.ExpectQuery(productSelectInInsertQuery).
		WithArgs(pvzID, models.ReceptionStatusInProgress).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectRollback()

	_, err = repo.InsertProduct(context.Background(), "", pvzID)
	assert.ErrorIs(t, err, ErrNoActiveReception)
}

func TestProductRepository_DeleteLastProduct_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)
	pvzID := uuid.New()
	receptionID := uuid.New()

	mock.ExpectBegin()

	mock.ExpectQuery(productSelectInDeleteQuery).
		WithArgs(pvzID, models.ReceptionStatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(receptionID))

	mock.ExpectExec(productDeleteQuery).
		WithArgs(receptionID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.DeleteLastProduct(context.Background(), pvzID)
	assert.NoError(t, err)
}

func TestProductRepository_DeleteLastProduct_NoActiveReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)
	pvzID := uuid.New()

	mock.ExpectBegin()

	mock.ExpectQuery(productSelectInDeleteQuery).
		WithArgs(pvzID, models.ReceptionStatusInProgress).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectRollback()

	err = repo.DeleteLastProduct(context.Background(), pvzID)
	assert.ErrorIs(t, err, ErrNoActiveReception)
}

func TestProductRepository_DeleteLastProduct_EmptyReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)
	pvzID := uuid.New()
	receptionID := uuid.New()

	mock.ExpectBegin()

	mock.ExpectQuery(productSelectInDeleteQuery).
		WithArgs(pvzID, models.ReceptionStatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(receptionID))

	mock.ExpectExec(productDeleteQuery).
		WithArgs(receptionID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectRollback()

	err = repo.DeleteLastProduct(context.Background(), pvzID)
	assert.ErrorIs(t, err, ErrEmptyReception)
}
