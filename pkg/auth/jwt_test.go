package auth

import (
	"basilisk/pkg/cache"
	"basilisk/pkg/helper"
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const (
	testSecret    = "dGVzdC1zZWNyZXQ=" // base64 encoded "test-secret"
	mockSessionID = "test-session-id"
)

type MockCache struct {
	AddError    error
	GetError    error
	DeleteError error
	Data        []byte
}

func (m *MockCache) Add(ctx context.Context, logger *slog.Logger, key string, data any, ttl ...time.Duration) error {
	return m.AddError
}

func (m *MockCache) Get(ctx context.Context, logger *slog.Logger, key string, target any) error {
	switch ptr := target.(type) {
	case *string:
		*ptr = string(m.Data)
	case *[]byte:
		*ptr = m.Data
	default:
		if err := json.Unmarshal(m.Data, target); err != nil {
			logger.Error("error while unmarshalling data", "error", err)
			return helper.InternalServerError
		}
	}
	return m.GetError
}

func (m *MockCache) Delete(ctx context.Context, logger *slog.Logger, key string) error {
	return m.DeleteError
}

func setupTest(t *testing.T) {
	err := LoadJWTService(context.Background(), testSecret)
	assert.NoError(t, err)

	m := MockCache{}
	cache.SetInstance(&m)
}

func TestInit(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{
			name:    "Valid base64 secret",
			secret:  testSecret,
			wantErr: false,
		},
		{
			name:    "Invalid base64 secret",
			secret:  "invalid-base64",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LoadJWTService(context.Background(), tt.secret)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, JWTServiceInstance)
			}
		})
	}
}

func TestJWTService_GenerateToken(t *testing.T) {
	setupTest(t)

	t.Run("Generate valid token", func(t *testing.T) {
		token, _, err := JWTServiceInstance.GenerateToken(context.TODO(), slog.Default(), "user-123", mockSessionID)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate the generated token
		userID, _, err := JWTServiceInstance.ValidateToken(context.TODO(), slog.Default(), token)
		assert.NoError(t, err)
		assert.Equal(t, "user-123", userID)
	})
}

func TestJWTService_ValidateToken(t *testing.T) {
	setupTest(t)

	testUserID := "user-123"
	testSessionID := "test-session-id"

	tests := []struct {
		name      string
		setupFunc func() string
		wantErr   bool
	}{
		{
			name: "Valid token",
			setupFunc: func() string {
				token, _, _ := JWTServiceInstance.GenerateToken(context.TODO(), slog.Default(), testUserID, testSessionID)
				return token
			},
			wantErr: false,
		},
		{
			name: "Expired token",
			setupFunc: func() string {
				claims := OwnClaims{
					UserID: testUserID,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(-time.Hour)},
						IssuedAt:  &jwt.NumericDate{Time: time.Now().Add(-time.Hour * 2)},
						NotBefore: &jwt.NumericDate{Time: time.Now().Add(-time.Hour * 2)},
						Issuer:    ownServiceName,
						Subject:   "auth",
						Audience:  jwt.ClaimStrings{ownServiceName},
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("test-secret"))
				return tokenString
			},
			wantErr: true,
		},
		{
			name: "Invalid audience",
			setupFunc: func() string {
				claims := OwnClaims{
					UserID: testUserID,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Hour)},
						IssuedAt:  &jwt.NumericDate{Time: time.Now()},
						NotBefore: &jwt.NumericDate{Time: time.Now()},
						Issuer:    ownServiceName,
						Subject:   "auth",
						Audience:  jwt.ClaimStrings{"wrong-audience"},
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("test-secret"))
				return tokenString
			},
			wantErr: true,
		},
		{
			name: "Invalid token format",
			setupFunc: func() string {
				return "invalid.token.format"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupFunc()
			userID, _, err := JWTServiceInstance.ValidateToken(context.TODO(), slog.Default(), token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testUserID, userID)
			}
		})
	}
}

func TestCheckTokenAudience(t *testing.T) {
	tests := []struct {
		name     string
		audience jwt.ClaimStrings
		want     bool
	}{
		{
			name:     "Valid audience",
			audience: jwt.ClaimStrings{ownServiceName},
			want:     true,
		},
		{
			name:     "Invalid audience",
			audience: jwt.ClaimStrings{"wrong-service"},
			want:     false,
		},
		{
			name:     "Multiple audiences with valid one",
			audience: jwt.ClaimStrings{"service1", ownServiceName, "service2"},
			want:     true,
		},
		{
			name:     "Empty audience",
			audience: jwt.ClaimStrings{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkTokenAudience(tt.audience)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestJWTService_TokenClaims(t *testing.T) {
	setupTest(t)

	token, exp, err := JWTServiceInstance.GenerateToken(context.TODO(), slog.Default(), "user-123", "test-session-id")
	assert.NoError(t, err)

	// Parse the token without validation to check claims
	parser := jwt.Parser{}
	parsedToken, _ := parser.ParseWithClaims(token, &OwnClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})

	claims, ok := parsedToken.Claims.(*OwnClaims)
	assert.True(t, ok)

	// Verify all claims are set correctly
	assert.Equal(t, ownServiceName, claims.Issuer)
	assert.Equal(t, "auth", claims.Subject)
	assert.Contains(t, claims.Audience, ownServiceName)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
	assert.NotNil(t, exp)

	// Verify user details in claims
	assert.Equal(t, "user-123", claims.UserID)
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	setupTest(t)

	t.Run("Generate valid refresh token", func(t *testing.T) {
		token, err := JWTServiceInstance.GenerateRefreshToken(context.TODO(), slog.Default(), "user-123", mockSessionID)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	setupTest(t)
	sessionId := "test-session-id"

	tests := []struct {
		name       string
		setupFunc  func() string
		wantErr    bool
		wantUserID string
	}{
		{
			name: "Valid refresh token",
			setupFunc: func() string {
				token, _ := JWTServiceInstance.GenerateRefreshToken(context.TODO(), slog.Default(), "user-123", mockSessionID)
				return token
			},
			wantErr:    false,
			wantUserID: "user-123",
		},
		{
			name: "Expired refresh token",
			setupFunc: func() string {
				claims := OwnClaims{
					SessionID: sessionId,
					TokenType: TokenTypeRefresh,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(-time.Hour)},
						IssuedAt:  &jwt.NumericDate{Time: time.Now().Add(-time.Hour * 2)},
						NotBefore: &jwt.NumericDate{Time: time.Now().Add(-time.Hour * 2)},
						Issuer:    ownServiceName,
						Subject:   "user-123",
						Audience:  jwt.ClaimStrings{ownServiceName},
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("test-secret"))
				return tokenString
			},
			wantErr: true,
		},
		{
			name: "Access token used as refresh token",
			setupFunc: func() string {
				token, _, _ := JWTServiceInstance.GenerateToken(context.TODO(), slog.Default(), "user-123", mockSessionID)
				return token
			},
			wantErr: true,
		},
		{
			name: "Invalid token string",
			setupFunc: func() string {
				return "not-a-valid-token"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupFunc()
			userID, err := JWTServiceInstance.ValidateRefreshToken(context.TODO(), slog.Default(), token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUserID, userID)
			}
		})
	}
}

func TestJWTService_SecretDecoding(t *testing.T) {
	// Test that the secret is correctly decoded from base64
	originalSecret := "my-secret-key"
	encodedSecret := base64.StdEncoding.EncodeToString([]byte(originalSecret))

	err := LoadJWTService(context.Background(), encodedSecret)
	assert.NoError(t, err)

	jwtSvc, ok := JWTServiceInstance.(*jwtService)
	assert.True(t, ok)
	assert.Equal(t, originalSecret, jwtSvc.secret)
}
