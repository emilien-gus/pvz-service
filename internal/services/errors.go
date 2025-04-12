package services

import "errors"

var (
	ErrAccessDenied          = errors.New("access denied")
	ErrCityNotAllowed        = errors.New("not allowed city")
	ErrProductTypeNotAllowed = errors.New("not allowed product type")
)
