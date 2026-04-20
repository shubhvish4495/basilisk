package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"basilisk/pkg/db"
)

const (
	ownServiceName = "basilisk-auth-service"
)

var (
	JWTServiceInstance JWTInterface
)

type OwnClaims struct {
	jwt.RegisteredClaims
	db.User `json:"user"`
}

type JWTInterface interface {
	ValidateToken(token string) (string, error)
	GenerateToken(user db.User) (string, error)
}

type jwtService struct {
	secret string
}

// LoadJWTService initializes the JWT service with the provided secret.
// The secret is expected to be a base64 encoded string, which will be decoded
// and used to configure the JWT service.
//
// Parameters:
//   - context: context
//   - secret: A base64 encoded string representing the JWT secret.
//
// Returns:
//   - error: An error if the secret cannot be decoded, otherwise nil.
func LoadJWTService(ctx context.Context, secret string) error {
	// jwt secret is base64 encoded. We will decode it first and then set it in config
	decStr, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return fmt.Errorf("error while decoding jwt secret %v", err)
	}
	JWTServiceInstance = &jwtService{
		secret: string(decStr),
	}
	return nil
}

// ValidateToken validates a JWT token string.
// It parses the token with custom claims and checks its validity.
// If the token is invalid or there is an error during parsing, it returns an error.
//
// Parameters:
//   - token: A string representing the JWT token to be validated.
//
// Returns:
//   - error: An error if the token is invalid or if there is an error during parsing.
func (j *jwtService) ValidateToken(token string) (string, error) {
	claimsData := OwnClaims{}
	t, err := jwt.ParseWithClaims(token, &claimsData, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})
	if err != nil {
		return "", err
	}

	// if token is not valid return invalid token error
	if !t.Valid {
		return "", fmt.Errorf("invalid token")
	}

	// get audience from claims
	aud, err := t.Claims.GetAudience()
	if err != nil {
		return "", err
	}

	// check audience for token
	if !checkTokenAudience(aud) {
		return "", fmt.Errorf("invalid audience")
	}

	return claimsData.User.ID, nil
}

// checkTokenAudience checks if the provided audience contains the service's own name.
// It iterates through the audience claims and returns true if a match is found.
// Otherwise, it returns false.
//
// Parameters:
//
//	audience (jwt.ClaimStrings): A list of audience claims.
//
// Returns:
//
//	bool: True if the audience contains the service's own name, false otherwise.
func checkTokenAudience(audience jwt.ClaimStrings) bool {
	return slices.Contains(audience, ownServiceName)
}

// GenerateToken generates a JWT token for the given user.
// The token includes custom claims such as user details, expiration time,
// issued at time, issuer, audience, subject, and not before time.
// The token is signed using the HS256 signing method and a secret key.
//
// Parameters:
//   - user: The user for whom the token is being generated.
//
// Returns:
//   - string: The signed JWT token as a string.
//   - error: An error if the token generation fails.
func (j *jwtService) GenerateToken(user db.User) (string, error) {
	claims := OwnClaims{
		User: db.User{
			ID: user.ID,
		},
	}

	// populate claims field
	claims.ExpiresAt = &jwt.NumericDate{Time: time.Now().Add(time.Minute * 15)}
	claims.IssuedAt = &jwt.NumericDate{Time: time.Now()}
	claims.Issuer = ownServiceName
	claims.Audience = jwt.ClaimStrings{ownServiceName}
	claims.Subject = "auth"
	claims.NotBefore = &jwt.NumericDate{Time: time.Now()}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}
