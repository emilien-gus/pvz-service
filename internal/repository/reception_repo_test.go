package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"pvz/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var (
	receptionSelectInInsertQuery = regexp.QuoteMeta(`SELECT 1 FROM receptions WHERE (pvz_id = $1 AND status = $2) LIMIT 1`)
	receptionInsertQuery         = regexp.QuoteMeta(`INSERT INTO receptions (id, pvz_id) VALUES ($1,$2) RETURNING id, date_time, pvz_id, status`)
	receptionUpdateQuery         = regexp.QuoteMeta(`UPDATE receptions SET status = $1 WHERE id = ( SELECT id FROM receptions WHERE pvz_id = $2 AND status = $3 ORDER BY date_time DESC LIMIT 1 ) RETURNING id, date_time, pvz_id, status`)
)

func TestReceptionRepository_InsertReception_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewReceptionRepository(db)
	pvzID := uuid.New()
	id := uuid.New()
	now := time.Now()

	mock.ExpectBegin()

	mock.ExpectQuery(receptionSelectInInsertQuery).
		WithArgs(pvzID, models.ReceptionStatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(receptionInsertQuery).
		WithArgs(sqlmock.AnyArg(), pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(id, now, pvzID, models.ReceptionStatusInProgress))

	mock.ExpectCommit()

	result, err := repo.InsertReception(context.Background(), pvzID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, pvzID, result.PVZID)
	assert.Equal(t, models.ReceptionStatusInProgress, result.Status)
}

func TestReceptionRepository_InsertReception_ActiveReceptionExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewReceptionRepository(db)
	pvzID := uuid.New()

	mock.ExpectBegin()

	mock.ExpectQuery(receptionSelectInInsertQuery).
		WithArgs(pvzID, models.ReceptionStatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectRollback()

	_, err = repo.InsertReception(context.Background(), pvzID)
	assert.ErrorIs(t, err, ErrActiveReceptionExists)
}

func TestReceptionRepository_InsertReception_PVZNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewReceptionRepository(db)
	pvzID := uuid.New()

	mock.ExpectBegin()

	mock.ExpectQuery(receptionSelectInInsertQuery).
		WithArgs(pvzID, models.ReceptionStatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(receptionInsertQuery).
		WithArgs(sqlmock.AnyArg(), pvzID).
		WillReturnError(&pq.Error{Code: "23503"})

	mock.ExpectRollback()

	_, err = repo.InsertReception(context.Background(), pvzID)
	assert.ErrorIs(t, err, ErrPVZNotFound)
}

func TestReceptionRepository_InsertReception_DBerrorInSelect(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewReceptionRepository(db)
	pvzID := uuid.New()

	mock.ExpectBegin()

	mock.ExpectQuery(receptionSelectInInsertQuery).
		WithArgs(pvzID, models.ReceptionStatusInProgress).
		WillReturnError(errors.New("some db error"))

	mock.ExpectRollback()

	_, err = repo.InsertReception(context.Background(), pvzID)
	assert.ErrorContains(t, err, "failed to check active reception:")
}

func TestReceptionRepository_InsertReception_DBerrorInInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewReceptionRepository(db)
	pvzID := uuid.New()

	mock.ExpectBegin()

	mock.ExpectQuery(receptionSelectInInsertQuery).
		WithArgs(pvzID, models.ReceptionStatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(receptionInsertQuery).
		WithArgs(sqlmock.AnyArg(), pvzID).
		WillReturnError(errors.New("some db error"))

	mock.ExpectRollback()

	_, err = repo.InsertReception(context.Background(), pvzID)
	assert.ErrorContains(t, err, "database error:")
}

func TestReceptionRepository_UpdateLastReceptionStatus_NoActiveReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewReceptionRepository(db)
	pvzID := uuid.New()

	mock.ExpectQuery(receptionUpdateQuery).
		WithArgs(models.ReceptionStatusClosed, pvzID, models.ReceptionStatusInProgress).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.UpdateLastReceptionStatus(context.Background(), pvzID)
	assert.ErrorIs(t, err, ErrNoActiveReception)
}

func TestReceptionRepository_UpdateLastReceptionStatus_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewReceptionRepository(db)
	pvzID := uuid.New()

	mock.ExpectQuery(receptionInsertQuery).
		WithArgs(models.ReceptionStatusClosed, pvzID, models.ReceptionStatusInProgress).
		WillReturnError(errors.New("some db error"))

	_, err = repo.UpdateLastReceptionStatus(context.Background(), pvzID)
	assert.ErrorContains(t, err, "execute update")
}

func TestReceptionRepository_UpdateLastReceptionStatus_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewReceptionRepository(db)
	pvzID := uuid.New()
	id := uuid.New()
	now := time.Now()

	mock.ExpectQuery(receptionUpdateQuery).
		WithArgs(models.ReceptionStatusClosed, pvzID, models.ReceptionStatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(id, now, pvzID, models.ReceptionStatusClosed))

	result, err := repo.UpdateLastReceptionStatus(context.Background(), pvzID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.ReceptionStatusClosed, result.Status)
	assert.Equal(t, pvzID, result.PVZID)
}
