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
		Name("googleLogin")

	s.Path("/login").
		Methods(http.MethodPost).
		HandlerFunc(handlers.UserLoginUsingPassword).
		Name("passwordLogin")

	s.Path("/user").
		Methods(http.MethodPost).
		HandlerFunc(handlers.CreateUserFromSignUpFormHandler).
		Name("signUpForm")

	s.Path("/user/google").
		Methods(http.MethodPost).
		HandlerFunc(handlers.CreateUserFromGoogleSignUpHandler).
		Name("signUpGoogle")

	// protected routes
	p := s.PathPrefix("").Subrouter()
	p.Use(AuthMiddleware)

	// now create protected user routes
	pu := p.PathPrefix("/user").Subrouter()

	pu.Path("/self").
		Methods(http.MethodGet).
		HandlerFunc(handlers.SelfUserGetHandler).
		Name("getSelfUserDetails")

}
