package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"basilisk/pkg/auth"
	"basilisk/pkg/db"
	"basilisk/pkg/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockGoogleAuth struct {
	user *auth.GoogleUser
	err  error
}

func (m *mockGoogleAuth) ValidateIDToken(_ context.Context, _ *slog.Logger, _ string) (*auth.GoogleUser, error) {
	return m.user, m.err
}

type mockJWT struct {
	token           string
	tokenErr        error
	refreshToken    string
	refreshTokenErr error
}

func (m *mockJWT) GenerateToken(_ string) (string, time.Time, error) {
	return m.token, time.Now().Add(15 * time.Minute), m.tokenErr
}

func (m *mockJWT) ValidateToken(_ string) (string, error) {
	return "", nil
}

func (m *mockJWT) GenerateRefreshToken(_ string) (string, error) {
	return m.refreshToken, m.refreshTokenErr
}

func (m *mockJWT) ValidateRefreshToken(_ string) (string, error) {
	return "", nil
}

type mockDB struct {
	user    *db.User
	httpErr helper.HttpError
	userID  uuid.UUID
}

func (m *mockDB) Ping() error                        { return nil }
func (m *mockDB) PingContext(_ context.Context) error { return nil }
func (m *mockDB) Close() error                       { return nil }

func (m *mockDB) GetUserByEmail(_ context.Context, _ *slog.Logger, _ string) (*db.User, helper.HttpError) {
	return m.user, m.httpErr
}

func (m *mockDB) GetUserByID(_ context.Context, _ *slog.Logger, _ string) (*db.User, helper.HttpError) {
	return m.user, m.httpErr
}

func (m *mockDB) CreateUser(_ context.Context, _ *slog.Logger, _ db.User) (uuid.UUID, helper.HttpError) {
	return m.userID, m.httpErr
}

func TestGoogleLogin(t *testing.T) {
	isVerified := true
	testUserID := uuid.New()
	validUser := &db.User{
		Name:       "Test User",
		Email:      "test@example.com",
		ProfilePic: "pic.png",
		SignUpType:  db.GoogleAuthType,
		IsVerified: &isVerified,
	}
	validUser.ID = testUserID

	tests := []struct {
		name           string
		body           string
		googleAuth     auth.GoogleAuthInterface
		jwt            auth.JWTInterface
		mockDB         *mockDB
		wantStatusCode int
		wantErrMsg     string
	}{
		{
			name: "success",
			body: `{"id_token":"valid-token"}`,
			googleAuth: &mockGoogleAuth{
				user: &auth.GoogleUser{ID: "user-123", Email: "test@example.com", Name: "Test User", Picture: "pic.png"},
			},
			jwt: &mockJWT{
				token:        "access-token",
				refreshToken: "refresh-token",
			},
			mockDB:         &mockDB{user: validUser},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid request body",
			body:           `{invalid-json`,
			googleAuth:     &mockGoogleAuth{},
			jwt:            &mockJWT{},
			mockDB:         &mockDB{},
			wantStatusCode: http.StatusInternalServerError,
			wantErrMsg:     "Internal Server Error",
		},
		{
			name: "google token validation fails",
			body: `{"id_token":"bad-token"}`,
			googleAuth: &mockGoogleAuth{
				err: errors.New("invalid token"),
			},
			jwt:            &mockJWT{},
			mockDB:         &mockDB{},
			wantStatusCode: http.StatusUnauthorized,
			wantErrMsg:     "Unauthorized",
		},
		{
			name: "jwt generate token fails",
			body: `{"id_token":"valid-token"}`,
			googleAuth: &mockGoogleAuth{
				user: &auth.GoogleUser{ID: "user-123", Email: "test@example.com"},
			},
			jwt: &mockJWT{
				tokenErr: errors.New("token generation failed"),
			},
			mockDB:         &mockDB{user: validUser},
			wantStatusCode: http.StatusInternalServerError,
			wantErrMsg:     "Internal Server Error",
		},
		{
			name: "jwt generate refresh token fails",
			body: `{"id_token":"valid-token"}`,
			googleAuth: &mockGoogleAuth{
				user: &auth.GoogleUser{ID: "user-123", Email: "test@example.com"},
			},
			jwt: &mockJWT{
				token:           "access-token",
				refreshTokenErr: errors.New("refresh token generation failed"),
			},
			mockDB:         &mockDB{user: validUser},
			wantStatusCode: http.StatusInternalServerError,
			wantErrMsg:     "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth.GoogleAuthInstance = tt.googleAuth
			auth.JWTServiceInstance = tt.jwt
			db.SetInstance(tt.mockDB)

			req := httptest.NewRequest(http.MethodPost, "/login/google", strings.NewReader(tt.body))
			ctx := helper.SetRequestIdToContext(req.Context(), "test-request-id")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			GoogleLogin(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantErrMsg != "" {
				var errResp helper.ErrorResponse
				err := json.NewDecoder(rr.Body).Decode(&errResp)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantErrMsg, errResp.Error)
			}

			if tt.wantStatusCode == http.StatusOK {
				var resp helper.SuccessResponse
				err := json.NewDecoder(rr.Body).Decode(&resp)
				assert.NoError(t, err)

				data, ok := resp.Data.(map[string]any)
				assert.True(t, ok)
				assert.Equal(t, "access-token", data["token"])
				assert.Equal(t, "refresh-token", data["refresh_token"])
				assert.NotNil(t, data["expiry"])
			}
		})
	}
}
