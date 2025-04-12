package repository

import (
	"context"
	"database/sql"
	"fmt"
	"pvz/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserRepositoryInterface interface {
	InsertUser(ctx context.Context, email, password, role string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) InsertUser(ctx context.Context, email, password, role string) (*models.User, error) {
	id := uuid.New()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	query, args, err := sq.Insert("users").Columns("id", "email", "password", "role").Values(id, email, hashedPassword, role).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	_, err = ur.db.ExecContext(ctx, query, args...)
	if err != nil {
		// check for 23505 error (unique_violation) in PostgreSQL
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	user := &models.User{
		ID:    id,
		Email: email,
		Role:  role,
	}
	return user, nil
}

func (ur *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query, args, err := sq.Select("*").From("users").Where(sq.Eq{"email": email}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var user models.User
	err = ur.db.QueryRowContext(ctx, query, args).Scan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
