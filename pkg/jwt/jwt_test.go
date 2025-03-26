package jwt

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shubhvish4495/basilisk/pkg/user"
	"github.com/stretchr/testify/assert"
)

const (
	testSecret = "dGVzdC1zZWNyZXQ=" // base64 encoded "test-secret"
)

func setupTest(t *testing.T) {
	err := Init(testSecret)
	assert.NoError(t, err)
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
			err := Init(tt.secret)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, Instance)
			}
		})
	}
}

func TestJWTService_GenerateToken(t *testing.T) {
	setupTest(t)

	testUser := user.User{
		ID:       123,
		Username: "testuser",
		Roles:    []string{"test-Roles"},
	}

	t.Run("Generate valid token", func(t *testing.T) {
		token, err := Instance.GenerateToken(testUser)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate the generated token
		parsedUser, err := Instance.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, testUser.ID, parsedUser.ID)
		assert.Equal(t, testUser.Username, parsedUser.Username)
		assert.Equal(t, testUser.Roles, parsedUser.Roles)
	})
}

func TestJWTService_ValidateToken(t *testing.T) {
	setupTest(t)

	testUser := user.User{
		ID:       123,
		Username: "testuser",
		Roles:    []string{"test-Roles"},
	}

	tests := []struct {
		name      string
		setupFunc func() string
		wantErr   bool
	}{
		{
			name: "Valid token",
			setupFunc: func() string {
				token, _ := Instance.GenerateToken(testUser)
				return token
			},
			wantErr: false,
		},
		{
			name: "Expired token",
			setupFunc: func() string {
				claims := OwnClaims{
					UserDetails: testUser,
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
					UserDetails: testUser,
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
			user, err := Instance.ValidateToken(token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, testUser.ID, user.ID)
				assert.Equal(t, testUser.Username, user.Username)
				assert.Equal(t, testUser.Roles, user.Roles)
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

	testUser := user.User{
		ID:       123,
		Username: "testuser",
		Roles:    []string{"test-Roles"},
	}

	token, err := Instance.GenerateToken(testUser)
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

	// Verify user details in claims
	assert.Equal(t, testUser.ID, claims.UserDetails.ID)
	assert.Equal(t, testUser.Username, claims.UserDetails.Username)
	assert.Equal(t, testUser.Roles, claims.UserDetails.Roles)
}

func TestJWTService_SecretDecoding(t *testing.T) {
	// Test that the secret is correctly decoded from base64
	originalSecret := "my-secret-key"
	encodedSecret := base64.StdEncoding.EncodeToString([]byte(originalSecret))

	err := Init(encodedSecret)
	assert.NoError(t, err)

	jwtService, ok := Instance.(*JWTService)
	assert.True(t, ok)
	assert.Equal(t, originalSecret, jwtService.secret)
}
