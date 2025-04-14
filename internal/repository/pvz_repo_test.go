package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	pvzInsertQuery = regexp.QuoteMeta(`INSERT INTO pvz (id, city) VALUES (?,?) RETURNING id, registration_date, city`)
)

func TestPVZRepository_InsertPVZ_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPWZRepository(db)
	city := "Москва"
	id := uuid.New()
	registration := time.Now()

	mock.ExpectQuery(pvzInsertQuery).
		WithArgs(sqlmock.AnyArg(), city).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow(id, registration, city))

	result, err := repo.InsertPVZ(context.Background(), city)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, city, result.City)
	assert.Equal(t, id, result.ID)
	assert.WithinDuration(t, registration, result.RegistrationDate, time.Second)
}

func TestPVZRepository_InsertPVZ_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPWZRepository(db)
	city := "Kazan"

	mock.ExpectQuery(pvzInsertQuery).
		WithArgs(sqlmock.AnyArg(), city).
		WillReturnError(sql.ErrConnDone)

	_, err = repo.InsertPVZ(context.Background(), city)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestPVZRepository_GetPVZList_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPWZRepository(db)

	queryRegex := regexp.QuoteMeta(
		`SELECT p.id AS pvz_id, p.registration_date, p.city, r.id AS reception_id, r.date_time AS reception_dateTime, r.status AS reception_status, pr.id AS product_id, pr.date_time AS product_dateTime, pr.type AS product_type FROM pvz p LEFT JOIN receptions r ON p.id = r.pvz_id LEFT JOIN products pr ON r.id = pr.reception_id ORDER BY p.id, r.date_time, pr.date_time LIMIT 10 OFFSET 0`,
	)

	mock.ExpectQuery(queryRegex).
		WillReturnRows(sqlmock.NewRows([]string{
			"pvz_id", "registration_date", "city",
			"reception_id", "reception_dateTime", "reception_status",
			"product_id", "product_dateTime", "product_type",
		}))

	result, err := repo.GetPVZList(context.Background(), nil, nil, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, result)
}
