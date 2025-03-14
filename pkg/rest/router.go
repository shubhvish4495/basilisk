package rest

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {

	// Create a subrouter for the /api/v1 path
	s := r.PathPrefix("/api/v1").Subrouter()

	// Register routes here
	s.Path("/health").
		Methods(http.MethodGet).
		HandlerFunc(healthCheck).
		Name("healthCheck")
}
