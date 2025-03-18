package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/shubhvish4495/basilisk/pkg/user"
)

const (
	ownServiceName = "basilisk-auth-service"
)

var (
	tokenSecret string
)

type OwnClaims struct {
	jwt.RegisteredClaims
	UserDetails user.User `json:"user"`
}

// SetTokenSecret will set secret for JWT token generation
func SetTokenSecret(secret string) {
	tokenSecret = secret
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
func ValidateToken(token string) (*user.User, error) {
	claimsData := OwnClaims{}
	t, err := jwt.ParseWithClaims(token, &claimsData, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return nil, err
	}

	// if token is not valid return invalid token error
	if !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// get audience from claims
	aud, err := t.Claims.GetAudience()
	if err != nil {
		return nil, err
	}

	// check audience for token
	if !checkTokenAudience(aud) {
		return nil, fmt.Errorf("invalid audience")
	}

	return &claimsData.UserDetails, nil
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
	for _, a := range audience {
		if a == ownServiceName {
			return true
		}
	}
	return false
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
func GenerateToken(user user.User) (string, error) {
	claims := OwnClaims{
		UserDetails: user,
	}

	// populate claims field
	claims.ExpiresAt = &jwt.NumericDate{Time: time.Now().Add(time.Minute * 15)}
	claims.IssuedAt = &jwt.NumericDate{Time: time.Now()}
	claims.Issuer = ownServiceName
	claims.Audience = jwt.ClaimStrings{ownServiceName}
	claims.Subject = "auth"
	claims.NotBefore = &jwt.NumericDate{Time: time.Now()}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))
}
