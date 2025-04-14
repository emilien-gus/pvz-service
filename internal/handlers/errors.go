package handlers

import "errors"

var (
	ErrUserExists      = errors.New("user already exists")
	ErrWrongPassword   = errors.New("wrong password")
	ErrUserDoesntExist = errors.New("no such user")
)
