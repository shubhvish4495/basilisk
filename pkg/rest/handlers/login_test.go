package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shubhvish4495/basilisk/pkg/jwt"
	"github.com/shubhvish4495/basilisk/pkg/user"
	"github.com/stretchr/testify/assert"
)

// MockJWT is a struct to mock jwt service
type MockJWT struct {
	token    string
	errorVar error
	user     *user.User
}

// GenerateToken will generate mock token as set in MockJWT struct
func (m *MockJWT) GenerateToken(u user.User) (string, error) {
	return m.token, m.errorVar
}

// ValidateToken will generate mock token as set in MockJWT struct
func (m *MockJWT) ValidateToken(token string) (*user.User, error) {
	return m.user, m.errorVar
}

func TestLogin_Success(t *testing.T) {
	// Mock the JWT generation
	jwt.Instance = &MockJWT{
		token:    "mock-token",
		errorVar: nil,
	}

	req, err := http.NewRequest("POST", "/login", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Login)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "mock-token")
}

func TestLogin_TokenGenerationError(t *testing.T) {
	// Mock the JWT generation to return an error
	jwt.Instance = &MockJWT{
		errorVar: assert.AnError,
	}

	req, err := http.NewRequest("POST", "/login", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Login)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Internal server error")
}
