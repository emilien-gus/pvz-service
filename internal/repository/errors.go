package repository

import "errors"

var (
	ErrUserExists            = errors.New("user already exists")
	ErrActiveReceptionExists = errors.New("active reception already exists")
	ErrPVZNotFound           = errors.New("pvz not found")
	ErrReceptionConflict     = errors.New("reception conflict")
	ErrNoActiveReception     = errors.New("no active reception")
)
