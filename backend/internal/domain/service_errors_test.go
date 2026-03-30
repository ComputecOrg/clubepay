package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/clubepay/backend/internal/domain"
)

func TestServiceError_Error(t *testing.T) {
	err := domain.NewServiceError(400, "bad input", nil)
	assert.Equal(t, "bad input", err.Error())
}

func TestServiceError_Unwrap(t *testing.T) {
	cause := errors.New("db connection failed")
	err := domain.NewServiceError(500, "internal error", cause)
	assert.ErrorIs(t, err, cause)
}

func TestServiceError_ErrorsAs(t *testing.T) {
	err := domain.NewErrNotFound("nao encontrado")
	var svcErr *domain.ServiceError
	assert.True(t, errors.As(err, &svcErr))
	assert.Equal(t, 404, svcErr.Code)
}

func TestServiceError_Constructors(t *testing.T) {
	tests := []struct {
		name string
		err  *domain.ServiceError
		code int
	}{
		{"not found", domain.NewErrNotFound("x"), 404},
		{"conflict", domain.NewErrConflict("x"), 409},
		{"forbidden", domain.NewErrForbidden("x"), 403},
		{"bad request", domain.NewErrBadRequest("x"), 400},
		{"internal", domain.NewErrInternal("x", nil), 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.err.Code)
			assert.Equal(t, "x", tt.err.Message)
		})
	}
}
