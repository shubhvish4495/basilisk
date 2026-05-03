package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	ownServiceName       = "basilisk-auth-service"
	TokenTypeAccess      = "access"
	TokenTypeRefresh     = "refresh"
	refreshTokenDuration = time.Hour * 24 * 30 // 30 days
	accessTokenDuration  = time.Hour * 24 * 7  // 1 week
)

var (
	JWTServiceInstance JWTInterface
)

type OwnClaims struct {
	jwt.RegisteredClaims
	UserID    string `json:"user_id"`
	TokenType string `json:"type"`
}

type JWTInterface interface {
	ValidateToken(token string) (string, error)
	GenerateToken(userID string) (string, time.Time, error)
	ValidateRefreshToken(token string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
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
	t, err := jwt.ParseWithClaims(token, &claimsData, func(token *jwt.Token) (any, error) {
		return []byte(j.secret), nil
	})
	if err != nil {
		return "", err
	}

	// if token is not valid return invalid token error
	if !t.Valid {
		return "", fmt.Errorf("invalid token")
	}

	// validate token type is correct
	if claimsData.TokenType != TokenTypeAccess {
		return "", fmt.Errorf("invalid token type")
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

	return claimsData.UserID, nil
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

// GenerateToken generates a JWT token for the given user ID.
// The token includes custom claims such as user ID, expiration time,
// issued at time, issuer, audience, subject, and not before time.
// The token is signed using the HS256 signing method and a secret key.
//
// Parameters:
//   - userID: The unique identifier of the user.
//
// Returns:
//   - string: The signed JWT token as a string.
//   - error: An error if the token generation fails.
func (j *jwtService) GenerateToken(userID string) (string, time.Time, error) {
	claims := OwnClaims{
		UserID:    userID,
		TokenType: TokenTypeAccess,
	}

	// populate claims field
	expiresAt := time.Now().Add(accessTokenDuration)
	claims.ExpiresAt = &jwt.NumericDate{Time: expiresAt}
	claims.IssuedAt = &jwt.NumericDate{Time: time.Now()}
	claims.Issuer = ownServiceName
	claims.Audience = jwt.ClaimStrings{ownServiceName}
	claims.Subject = "auth"
	claims.NotBefore = &jwt.NumericDate{Time: time.Now()}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return signedToken, expiresAt, nil
}

// GenerateRefreshToken generates a long-lived JWT refresh token for the given user.
// The token includes the user ID as the subject and is valid for 30 days.
// It is signed using the HS256 signing method.
//
// Parameters:
//   - userID: The unique identifier of the user.
//
// Returns:
//   - string: The signed JWT refresh token.
//   - error: An error if the token signing fails.
func (j *jwtService) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(refreshTokenDuration)

	var claims OwnClaims

	// populate claims field
	claims.ExpiresAt = &jwt.NumericDate{Time: expiresAt}
	claims.IssuedAt = &jwt.NumericDate{Time: now}
	claims.Issuer = ownServiceName
	claims.Audience = jwt.ClaimStrings{ownServiceName}
	claims.Subject = userID
	claims.NotBefore = &jwt.NumericDate{Time: now}
	claims.TokenType = TokenTypeRefresh

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

// ValidateRefreshToken validates a JWT refresh token string.
// It parses and verifies the token's signature, expiration, issuer, audience,
// and token type. If valid, it returns the user UUID stored in the subject claim.
//
// Parameters:
//   - token: A string representing the JWT refresh token to be validated.
//
// Returns:
//   - string: The user UUID extracted from the token's subject claim.
//   - error: An error if the token is invalid, expired, or missing required claims.
func (j *jwtService) ValidateRefreshToken(token string) (string, error) {
	// Parse the token with standard claims
	claims := OwnClaims{}
	t, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (any, error) {
		return []byte(j.secret), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return "", fmt.Errorf("refresh token has expired")
		}
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check token validity
	if !t.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}

	// Verify issuer
	if claims.Issuer != ownServiceName {
		return "", fmt.Errorf("invalid issuer")
	}

	// if token not of refresh type throw error
	if claims.TokenType != TokenTypeRefresh {
		return "", fmt.Errorf("invalid token type")
	}

	// Verify audience
	aud, err := claims.GetAudience()
	if err != nil {
		return "", fmt.Errorf("error getting audience: %w", err)
	}
	if !checkTokenAudience(aud) {
		return "", fmt.Errorf("invalid audience")
	}

	// Get and verify subject (user UUID)
	userUUID := claims.Subject
	if userUUID == "" {
		return "", fmt.Errorf("missing user UUID in refresh token")
	}

	return userUUID, nil
}
