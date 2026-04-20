package handlers

import (
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
}

// GenerateToken will generate mock token as set in MockJWT struct
func (m *MockJWT) GenerateToken(u db.User) (string, time.Time, error) {
	return m.token, time.Now().Add(time.Minute * 15), m.errorVar
}

// ValidateToken will generate mock token as set in MockJWT struct
func (m *MockJWT) ValidateToken(token string) (string, error) {
	return m.user.ID, m.errorVar
}

// GenerateRefreshToken will generate mock refresh token as set in MockJWT struct
func (m *MockJWT) GenerateRefreshToken(userID string) (string, error) {
	return m.refreshToken, m.errorVar
}

// ValidateRefreshToken will validate mock refresh token as set in MockJWT struct
func (m *MockJWT) ValidateRefreshToken(token string) (string, error) {
	return m.user.ID, m.errorVar
}

// MockRefreshTokenError is a mock where GenerateToken succeeds but GenerateRefreshToken fails
type MockRefreshTokenError struct {
	token           string
	refreshTokenErr error
}

func (m *MockRefreshTokenError) GenerateToken(u db.User) (string, time.Time, error) {
	return m.token, time.Now().Add(time.Minute * 15), nil
}

func (m *MockRefreshTokenError) ValidateToken(token string) (string, error) {
	return "", nil
}

func (m *MockRefreshTokenError) GenerateRefreshToken(userID string) (string, error) {
	return "", m.refreshTokenErr
}

func (m *MockRefreshTokenError) ValidateRefreshToken(token string) (string, error) {
	return "", nil
}

func TestLogin_Success(t *testing.T) {
	// Mock the JWT generation
	auth.JWTServiceInstance = &MockJWT{
		token:        "mock-token",
		refreshToken: "mock-refresh-token",
		errorVar:     nil,
	}

	req, err := http.NewRequest("POST", "/login", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Login)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(t, body, "mock-token")
	assert.Contains(t, body, "mock-refresh-token")
	assert.Contains(t, body, "expires_at")
}

func TestLogin_TokenGenerationError(t *testing.T) {
	// Mock the JWT generation to return an error
	auth.JWTServiceInstance = &MockJWT{
		errorVar: assert.AnError,
	}

	req, err := http.NewRequest("POST", "/login", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Login)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Internal Server Error")
}

func TestLogin_RefreshTokenGenerationError(t *testing.T) {
	// Use a mock that succeeds for GenerateToken but fails for GenerateRefreshToken
	auth.JWTServiceInstance = &MockRefreshTokenError{
		token:           "mock-token",
		refreshTokenErr: assert.AnError,
	}

	req, err := http.NewRequest("POST", "/login", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Login)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Internal Server Error")
}
