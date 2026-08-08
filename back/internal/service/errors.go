package service

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource conflict")
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("operation forbidden")
	ErrInternal     = errors.New("internal error")

	// ErrUnauthorized отличается от ErrForbidden: первое означает «непонятно,
	// кто ты» (401), второе — «понятно, но нельзя» (403). Раньше неверный
	// пароль отдавался как ErrInvalidInput, то есть 400.
	ErrUnauthorized = errors.New("unauthorized")
)
