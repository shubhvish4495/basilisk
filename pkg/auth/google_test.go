package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoogleAuthInit(t *testing.T) {
	tests := []struct {
		name    string
		config  *GoogleConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &GoogleConfig{
				Secret:   "test-secret",
				ClientID: "test-client-id",
			},
			wantErr: false,
		},
		{
			name: "Missing secret",
			config: &GoogleConfig{
				Secret:   "",
				ClientID: "test-client-id",
			},
			wantErr: true,
		},
		{
			name: "Missing client ID",
			config: &GoogleConfig{
				Secret:   "test-secret",
				ClientID: "",
			},
			wantErr: true,
		},
		{
			name: "Both missing",
			config: &GoogleConfig{
				Secret:   "",
				ClientID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global instance before each test
			GoogleAuthInstance = nil

			err := GoogleAuthInit(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, GoogleAuthInstance)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, GoogleAuthInstance)
				assert.Equal(t, tt.config.Secret, GoogleAuthInstance.Secret)
				assert.Equal(t, tt.config.ClientID, GoogleAuthInstance.ClientID)
			}
		})
	}
}
