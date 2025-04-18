package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var (
	usersInsertQuery = regexp.QuoteMeta(`INSERT INTO users (id,email,password,role) VALUES ($1,$2,$3,$4)`)
	usersSelectQuery = regexp.QuoteMeta(`SELECT * FROM users WHERE email = $1`)
)

func TestUserRepository_InsertUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	email := "test@example.com"
	password := "securepassword"
	role := "user"

	mock.ExpectExec(usersInsertQuery).
		WithArgs(sqlmock.AnyArg(), email, sqlmock.AnyArg(), role).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user, err := repo.InsertUser(context.Background(), email, password, role)
	assert.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, role, user.Role)
	assert.NotEqual(t, uuid.Nil, user.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_InsertUser_Conflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	email := "duplicate@example.com"
	password := "password"
	role := "user"

	mock.ExpectExec(usersInsertQuery).
		WithArgs(sqlmock.AnyArg(), email, sqlmock.AnyArg(), role).
		WillReturnError(&pq.Error{Code: "23505"}) // Unique constraint

	user, err := repo.InsertUser(context.Background(), email, password, role)
	assert.Nil(t, user)
	assert.Equal(t, ErrUserExists, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_InsertUser_HashError(t *testing.T) {
	repo := NewUserRepository(nil)

	longPassword := string(make([]byte, 1<<20)) // too long password

	user, err := repo.InsertUser(context.Background(), "email@example.com", longPassword, "user")
	assert.Nil(t, user)
	assert.Error(t, err)
}

func TestUserRepository_GetUserByEmail_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	expectedID := uuid.New()
	email := "user@example.com"
	role := "admin"
	password := "hashed-password"

	rows := sqlmock.NewRows([]string{"id", "email", "password", "role"}).
		AddRow(expectedID, email, password, role)

	mock.ExpectQuery(usersSelectQuery).
		WithArgs(email).
		WillReturnRows(rows)

	user, err := repo.GetUserByEmail(context.Background(), email)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedID, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, password, user.Password)
	assert.Equal(t, role, user.Role)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	email := "unknown@example.com"

	mock.ExpectQuery(usersSelectQuery).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByEmail(context.Background(), email)
	assert.NoError(t, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByEmail_DBerror(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	email := "error@example.com"

	mock.ExpectQuery(usersSelectQuery).
		WithArgs(email).
		WillReturnError(errors.New("db error"))

	user, err := repo.GetUserByEmail(context.Background(), email)
	assert.Nil(t, user)
	assert.EqualError(t, err, "db error")

	assert.NoError(t, mock.ExpectationsWereMet())
}
