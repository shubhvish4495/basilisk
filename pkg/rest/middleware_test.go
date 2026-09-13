package rest

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"basilisk/pkg/auth"
	"basilisk/pkg/db"

	"github.com/stretchr/testify/assert"
)

// MockJWT is a struct to mock jwt service
type MockJWT struct {
	token        string
	refreshToken string
	errorVar     error
	user         *db.User
	sessionID    string
}

// GenerateToken will generate mock token as set in MockJWT struct
func (m *MockJWT) GenerateToken(ctx context.Context, logger *slog.Logger, userID, sessionID string) (string, time.Time, error) {
	return m.token, time.Now().Add(time.Minute * 15), m.errorVar
}

// ValidateToken will generate mock token as set in MockJWT struct
func (m *MockJWT) ValidateToken(ctx context.Context, logger *slog.Logger, token string) (string, string, error) {
	return m.user.ID, m.sessionID, m.errorVar
}

// GenerateRefreshToken will generate mock refresh token as set in MockJWT struct
func (m *MockJWT) GenerateRefreshToken(ctx context.Context, logger *slog.Logger, userID, sessionID string) (string, error) {
	return m.refreshToken, m.errorVar
}

// ValidateRefreshToken will validate mock refresh token as set in MockJWT struct
func (m *MockJWT) ValidateRefreshToken(ctx context.Context, logger *slog.Logger, token string) (string, error) {
	return m.user.ID, m.errorVar
}

func (m *MockJWT) AddSesssionToDenyList(ctx context.Context, logger *slog.Logger, sessionID string) error {
	return nil
}

func TestLoggingMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := LoggingMiddleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRecoveryMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	middleware := RecoveryMiddleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Internal Server Error")
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	validToken := "valid-token"
	auth.JWTServiceInstance = &MockJWT{
		token: validToken,
		user: &db.User{
			ID:   "mock-id",
			Name: "test-user",
		},
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	invalidToken := "invalid-token"
	auth.JWTServiceInstance = &MockJWT{token: "valid-token", errorVar: assert.AnError, user: &db.User{ID: "mock-id"}}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Set("Authorization", "Bearer "+invalidToken)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unauthorized")
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unauthorized")
}
