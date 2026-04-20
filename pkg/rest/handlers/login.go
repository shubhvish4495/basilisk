package handlers

import (
	"log/slog"
	"net/http"

	"basilisk/pkg/auth"
	"basilisk/pkg/db"
	"basilisk/pkg/helper"
)

// Login handles the user login process, generates a JWT token for the user,
// and writes it to the response. If there is an error during token generation,
// it responds with an internal server error status.
//
// Parameters:
//   - w: http.ResponseWriter to write the response.
//   - r: *http.Request containing the login request.
//
// Response:
//   - 200 OK: If the token is successfully generated, the token is written to the response.
//   - 500 Internal Server Error: If there is an error during token generation.
//
// nolint:errcheck
func Login(w http.ResponseWriter, r *http.Request) {
	user := db.User{
		ID: "random-user-uuid",
	}

	// generate token
	t, expiresAt, err := auth.JWTServiceInstance.GenerateToken(user)
	if err != nil {
		slog.Error("Error generating token", "error", err)
		helper.SendError(w, helper.InternalServerError)
		return
	}

	//generate refresh token
	rfTkn, err := auth.JWTServiceInstance.GenerateRefreshToken(user.ID)
	if err != nil {
		slog.Error("error while generating refresh token")
		helper.SendError(w, helper.InternalServerError)
		return
	}

	slog.Info("User logged in")
	helper.SendSuccessResponse(w, http.StatusOK, map[string]any{
		"token":         t,
		"expires_at":    expiresAt,
		"refresh_token": rfTkn,
	})
}
