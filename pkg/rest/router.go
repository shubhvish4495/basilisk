package rest

import (
	"net/http"

	"basilisk/pkg/rest/handlers"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {

	// Create a subrouter for the /api/v1 path
	s := r.PathPrefix("/api/v1").Subrouter()

	// Register routes here
	s.Path("/health").
		Methods(http.MethodGet).
		HandlerFunc(handlers.HealthCheck).
		Name("healthCheck")

	s.Path("/login/google").
		Methods(http.MethodPost).
		HandlerFunc(handlers.GoogleLogin).
		Name("Google Login")

	s.Path("/login").
		Methods(http.MethodPost).
		HandlerFunc(handlers.LoginWithPwd).
		Name("login with pass")

	// protected routes
	p := s.PathPrefix("").Subrouter()
	p.Use(AuthMiddleware)

	p.Path("/loggedInReq").
		Methods(http.MethodGet).
		HandlerFunc(handlers.LoggedInRequest).
		Name("Logged In Request")

	p.Path("/logout").
		Methods(http.MethodPost).
		HandlerFunc(handlers.LogOutHandler).
		Name("logout")

	p.Path("/refresh").
		Methods(http.MethodPost).
		HandlerFunc(handlers.RefreshTokenHandler).
		Name("refresh token handler")

}
