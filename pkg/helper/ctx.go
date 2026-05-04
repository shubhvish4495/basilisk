package helper

import (
	"context"
	"errors"
)

type ctxKey int

const (
	uuidKey ctxKey = iota
	userKey
	userRoleKey
)

// SetUserToContext returns a new context with the provided user.User value associated with the userKey.
// This allows the user information to be retrieved from the context in downstream handlers or functions.
// Parameters:
//   - ctx: The original context to which the user information will be added.
//   - userId: A pointer to the user.User struct representing the user to be stored in the context.
//
// Returns:
//   - A new context.Context containing the user information.
func SetUserToContext(ctx context.Context, userId string) context.Context {
	return context.WithValue(ctx, userKey, userId)
}

// GetUserFromContext retrieves the user information stored in the provided context.
// It returns a pointer to a user.User object if present, or an error if the user is not found in the context.
// The function expects the user to be stored in the context using the userKey.
func GetUserIDFromContext(ctx context.Context) string {
	userId := ctx.Value(userKey)
	if userId == nil {
		return ""
	}
	return userId.(string)
}

func SetRequestIdToContext(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, uuidKey, requestId)
}

func GetRequestIdFromContext(ctx context.Context) (string, error) {
	requestId := ctx.Value(uuidKey)
	if requestId == nil {
		return "", errors.New("request ID not found in context")
	}
	return requestId.(string), nil
}
