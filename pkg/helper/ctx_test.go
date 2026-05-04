package helper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAndGetUserFromContext(t *testing.T) {
	t.Run("Set and retrieve user ID", func(t *testing.T) {
		ctx := context.Background()
		ctx = SetUserToContext(ctx, "user-123")
		got := GetUserIDFromContext(ctx)
		assert.Equal(t, "user-123", got)
	})

	t.Run("Returns empty string when user not set", func(t *testing.T) {
		ctx := context.Background()
		got := GetUserIDFromContext(ctx)
		assert.Equal(t, "", got)
	})
}

func TestSetAndGetRequestIdFromContext(t *testing.T) {
	t.Run("Set and retrieve request ID", func(t *testing.T) {
		ctx := context.Background()
		ctx = SetRequestIdToContext(ctx, "req-abc-123")
		got, err := GetRequestIdFromContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "req-abc-123", got)
	})

	t.Run("Returns error when request ID not set", func(t *testing.T) {
		ctx := context.Background()
		got, err := GetRequestIdFromContext(ctx)
		assert.Error(t, err)
		assert.Equal(t, "", got)
		assert.Equal(t, "request ID not found in context", err.Error())
	})
}
