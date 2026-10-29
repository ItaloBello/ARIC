package apperror

import "net/http"

type AppError struct {
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

// Construtores para erros comuns
func NotFound(message string) *AppError {
	return &AppError{Message: message, StatusCode: http.StatusNotFound}
}

func BadRequest(message string) *AppError {
	return &AppError{Message: message, StatusCode: http.StatusBadRequest}
}

func Conflict(message string) *AppError {
	return &AppError{Message: message, StatusCode: http.StatusConflict}
}

func Unauthorized(message string) *AppError {
	return &AppError{Message: message, StatusCode: http.StatusUnauthorized}
}

func Forbidden(message string) *AppError {
	return &AppError{Message: message, StatusCode: http.StatusForbidden}
}

func Internal(message string) *AppError {
	return &AppError{Message: message, StatusCode: http.StatusInternalServerError}
}

func Created(message string) *AppError {
	return &AppError{Message: message, StatusCode: http.StatusCreated}
}