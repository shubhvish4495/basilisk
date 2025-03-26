package rest

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/google/uuid"

	"github.com/shubhvish4495/basilisk/pkg/auth"
	"github.com/shubhvish4495/basilisk/pkg/helper"
)

type ctxKey int

const (
	uuidKey ctxKey = iota
	userKey
)

type CustomResponseLogger struct {
	http.ResponseWriter
	StatusCode int
}

func (c *CustomResponseLogger) WriteHeader(code int) {
	c.StatusCode = code
	c.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware is a middleware that logs the details of each HTTP request and response.
// It generates a unique UUID for each request, which is included in the log entries for both
// the request and the response. The middleware logs the HTTP method, URL path, and the generated
// UUID for the request, and the status code, URL path, and the UUID for the response.
//
// Parameters:
//
//	next - the next http.Handler to be called after the middleware has processed the request.
//
// Returns:
//
//	An http.Handler that wraps the provided handler with logging functionality.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuid := uuid.New()
		ctx := context.WithValue(r.Context(), uuidKey, uuid)
		wr := &CustomResponseLogger{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		helper.GetLogger().Info(fmt.Sprintf("Request: method=%s url=%s, reqID=%s", r.Method, r.URL.Path, uuid))
		next.ServeHTTP(wr, r.WithContext(ctx))
		helper.GetLogger().Info(fmt.Sprintf("Response: status=%d url=%s reqID=%s", wr.StatusCode, r.URL.Path, uuid))
	})
}

// RecoveryMiddleware is a middleware that recovers from panics in the HTTP handler chain.
// If a panic occurs, it logs the panic details along with the stack trace and sends a
// 500 Internal Server Error response to the client.
//
// Usage:
//
//	http.Handle("/path", RecoveryMiddleware(yourHandler))
//
// Parameters:
//
//	next - the next http.Handler in the chain
//
// Returns:
//
//	http.Handler - a new handler that wraps the provided handler with panic recovery
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic details
				helper.GetLogger().Error("Recovered from panic: %v\nStack Trace:\n%s", err, debug.Stack())

				// Send a 500 Internal Server Error response
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware is a middleware function that checks for a valid authentication token
// in the request header. If the token is valid, it adds the token data to the request
// context and calls the next handler in the chain. If the token is invalid, it responds
// with an "Unauthorized" error and a 401 status code.
//
// The token is expected to be in the format "Bearer my-secret-token". The token is
// validated using the jwt.ValidateToken function.
//
// Parameters:
// - next: The next http.Handler to be called if the token is valid.
//
// Returns:
//   - http.Handler: A new http.Handler that wraps the original handler with authentication
//     token validation.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request has a valid authentication token
		authHeader := r.Header.Get("Authorization")
		splittedHeader := strings.Split(authHeader, " ")

		// Check if the token is in the format
		// "Bearer
		if len(splittedHeader) != 2 && splittedHeader[0] != "Bearer" {
			helper.GetLogger().Error("Invalid token format")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		data, err := auth.JWTServiceInstance.ValidateToken(splittedHeader[1])
		if err != nil {
			helper.GetLogger().Error("Invalid token", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, data)
		// Call the next handler in the chain
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
