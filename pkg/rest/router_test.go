package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"basilisk/pkg/auth"
	"basilisk/pkg/rest/handlers"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRoutes(t *testing.T) {
	r := mux.NewRouter()
	RegisterRoutes(r)

	tests := []struct {
		name       string
		method     string
		path       string
		handler    http.HandlerFunc
		middleware bool
		jwtDetails auth.JWTInterface
	}{
		{
			name:    "HealthCheck",
			method:  http.MethodGet,
			path:    "/api/v1/health",
			handler: handlers.HealthCheck,
		},
		{
			name:    "Login - No Error",
			method:  http.MethodGet,
			path:    "/api/v1/login",
			handler: handlers.Login,
			jwtDetails: &MockJWT{
				token:    "correct-token",
				errorVar: nil,
			},
		},
		{
			name:       "GetUsers",
			method:     http.MethodGet,
			path:       "/api/v1/users",
			handler:    handlers.GetUsers,
			middleware: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			assert.NoError(t, err)

			if tt.jwtDetails != nil {
				auth.JWTServiceInstance = tt.jwtDetails
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if tt.middleware {
				// Assuming AuthMiddleware sets a specific header or status code
				assert.Equal(t, http.StatusUnauthorized, rr.Code)
			} else {
				assert.Equal(t, http.StatusOK, rr.Code)
			}
		})
	}
}
