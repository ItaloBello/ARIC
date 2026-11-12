package apperror

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	Message    string
	StatusCode int
	// RetryAfter (segundos) vira o cabeçalho Retry-After, quando maior que zero.
	RetryAfter int
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

// BadGateway indica falha em um serviço externo (ex.: ViaCEP fora do ar).
func BadGateway(message string) *AppError {
	return &AppError{Message: message, StatusCode: http.StatusBadGateway}
}

// TooManyRequests: limite de tentativas atingido. retryAfter em segundos.
func TooManyRequests(message string, retryAfter int) *AppError {
	return &AppError{Message: message, StatusCode: http.StatusTooManyRequests, RetryAfter: retryAfter}
}

func Created(message string) *AppError {
	return &AppError{Message: message, StatusCode: http.StatusCreated}
}

// Responder envia o erro como JSON. AppError usa o próprio status;
// qualquer outro erro vira 500 genérico (o detalhe vai só para o log).
func Responder(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.RetryAfter > 0 {
			c.Header("Retry-After", strconv.Itoa(appErr.RetryAfter))
		}
		c.JSON(appErr.StatusCode, gin.H{"error": appErr.Message})
		return
	}
	log.Println("[erro não tratado]", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "erro interno"})
}
