package services

import "errors"

var (
	ErrAccessDenied          = errors.New("access denied")
	ErrCityNotAllowed        = errors.New("not allowed city")
	ErrProductTypeNotAllowed = errors.New("not allowed product type")
	ErrStartLaterThenEnd     = errors.New("start date_time is later then end date_time")
	ErrPageParamIsInvalid    = errors.New("page parametr is invalid")
	ErrLimitParamIsInvalid   = errors.New("limit parametr is invalid")
)
