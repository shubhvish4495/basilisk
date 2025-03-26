package handlers

import (
	"net/http"

	"github.com/shubhvish4495/basilisk/pkg/helper"
	"github.com/shubhvish4495/basilisk/pkg/jwt"
	"github.com/shubhvish4495/basilisk/pkg/user"
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
	user := user.User{
		Username: "test-user",
		Roles:    []string{"user:self:get"},
	}
	t, err := jwt.Instance.GenerateToken(user)
	if err != nil {
		helper.GetLogger().Error("Error generating token", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	helper.GetLogger().Info("User logged in")
	helper.SendSuccess(w, http.StatusOK, t)
}
