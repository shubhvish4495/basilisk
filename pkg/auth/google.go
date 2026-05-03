package auth

import (
	"context"
	"errors"
	"log/slog"

	"cloud.google.com/go/auth/credentials/idtoken"
)

var (
	GoogleAuthInstance *GoogleAuth
)

type GoogleConfig struct {
	Secret   string `yaml:"auth_secret"`
	ClientID string `yaml:"auth_client_id"`
}

type GoogleAuth struct {
	Secret   string
	ClientID string
}

type GoogleUser struct {
	Email   string
	Name    string
	Picture string
	ID      string
}

// sets up google authentication
func GoogleAuthInit(c *GoogleConfig) error {
	if c.Secret == "" || c.ClientID == "" {
		return errors.New("missing google auth secret")
	}

	GoogleAuthInstance = new(GoogleAuth)
	GoogleAuthInstance.Secret = c.Secret
	GoogleAuthInstance.ClientID = c.ClientID

	return nil
}

// VaildateIDToken validates the idToken with google using the clientID
// idToken provided to us as a client
func (g *GoogleAuth) ValidateIDToken(ctx context.Context, logger *slog.Logger, idToken string) (*GoogleUser, error) {
	payload, err := idtoken.Validate(ctx, idToken, g.ClientID)
	if err != nil {
		logger.Error("error while validating google token", "error", err)
		return nil, err
	}

	user := GoogleUser{
		Email:   payload.Claims["email"].(string),
		Name:    payload.Claims["name"].(string),
		Picture: payload.Claims["picture"].(string),
		ID:      payload.Subject,
	}

	return &user, nil
}
