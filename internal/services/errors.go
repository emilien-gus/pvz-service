package services

import "errors"

var (
	ErrAccessDenied   = errors.New("access denied")
	ErrCityNotAllowed = errors.New("not allowed city")
)
