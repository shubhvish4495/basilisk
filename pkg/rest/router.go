package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"basilisk/pkg/rest/handlers"
)

func RegisterRoutes(r *mux.Router) {

	// Create a subrouter for the /api/v1 path
	s := r.PathPrefix("/api/v1").Subrouter()

	// Register routes here
	s.Path("/health").
		Methods(http.MethodGet).
		HandlerFunc(handlers.HealthCheck).
		Name("healthCheck")

	s.Path("/login").
		Methods(http.MethodGet).
		HandlerFunc(handlers.Login).
		Name("login")

	// protected routes
	p := s.PathPrefix("").Subrouter()
	p.Use(AuthMiddleware)

	p.Path("/users").
		Methods(http.MethodGet).
		HandlerFunc(handlers.GetUsers).
		Name("getUsers")
}
