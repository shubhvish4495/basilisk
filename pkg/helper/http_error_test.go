package helper

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHttpError(t *testing.T) {
	err := NewHttpError(http.StatusBadRequest, "bad input")
	assert.Equal(t, http.StatusBadRequest, err.StatusCode())
	assert.Equal(t, "bad input", err.Error())
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      HttpError
		wantCode int
		wantMsg  string
	}{
		{"InternalServerError", InternalServerError, http.StatusInternalServerError, "Internal Server Error"},
		{"BadRequestError", BadRequestError, http.StatusBadRequest, "Bad Request"},
		{"NotFoundError", NotFoundError, http.StatusNotFound, "Not Found"},
		{"UnauthorizedError", UnauthorizedError, http.StatusUnauthorized, "Unauthorized"},
		{"ForbiddenError", ForbiddenError, http.StatusForbidden, "Forbidden"},
		{"ConflictError", ConflictError, http.StatusConflict, "Conflict"},
		{"TooManyRequestsError", TooManyRequestsError, http.StatusTooManyRequests, "Too Many Requests"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, tt.err.StatusCode())
			assert.Equal(t, tt.wantMsg, tt.err.Error())
		})
	}
}
