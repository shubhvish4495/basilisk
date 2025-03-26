package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTestEnv(t *testing.T) func() {
	// Create config directory
	err := os.MkdirAll("config", 0755)
	assert.NoError(t, err)

	// Return cleanup function
	return func() {
		err := os.RemoveAll("config")
		assert.NoError(t, err)
		// Reset config variable
		config = nil
	}
}

func TestLoad_Success(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Test data with environment variable
	os.Setenv("TEST_PASSWORD", "secret123")
	defer os.Unsetenv("TEST_PASSWORD")

	configData := `
		tlsConfig:
		certFile: "/path/to/cert.pem"
		keyFile: "/path/to/key.pem"
		database:
		host: "localhost"
		port: 5432
		user: "testuser"
		password: "${TEST_PASSWORD}"
		name: "testdb"
		jwt:
		secret: "test-secret"
		`
	err := os.WriteFile(filepath.Join("config", "config.yml"), []byte(configData), 0644)
	assert.NoError(t, err)

	// Test Load function
	err = Load()
	assert.NoError(t, err)

	// Verify all fields are correctly loaded
	assert.NotNil(t, config)
	assert.Equal(t, "/path/to/cert.pem", config.TlsConfig.CertFile)
	assert.Equal(t, "/path/to/key.pem", config.TlsConfig.KeyFile)
	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
	assert.Equal(t, "testuser", config.Database.User)
	assert.Equal(t, "secret123", config.Database.Password) // Environment variable should be expanded
	assert.Equal(t, "testdb", config.Database.Name)
	assert.Equal(t, "test-secret", config.JWT.Secret)
}

func TestLoad_FileNotFound(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestLoad_InvalidYAML(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	invalidYAML := `
		invalid:
		- yaml:
			format:
			missing:
			closing
			bracket
		`
	err := os.WriteFile(filepath.Join("config", "config.yml"), []byte(invalidYAML), 0644)
	assert.NoError(t, err)

	err = Load()
	assert.Error(t, err)
}

func TestGetConfig_Success(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	configData := `
		tlsConfig:
		certFile: "/path/to/cert.pem"
		keyFile: "/path/to/key.pem"
		database:
		host: "localhost"
		port: 5432
		user: "testuser"
		password: "testpass"
		name: "testdb"
		jwt:
		secret: "test-secret"
		`
	err := os.WriteFile(filepath.Join("config", "config.yml"), []byte(configData), 0644)
	assert.NoError(t, err)

	// First call should load config
	cfg := GetConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost", cfg.Database.Host)

	// Second call should return same instance
	cfg2 := GetConfig()
	assert.Equal(t, cfg, cfg2)
}

func TestGetConfig_PanicsOnError(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Test that GetConfig panics when config file doesn't exist
	assert.Panics(t, func() {
		GetConfig()
	})
}

func TestEnvironmentVariableExpansion(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Set multiple environment variables
	os.Setenv("DB_HOST", "test-host")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_PASSWORD", "env-password")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_PASSWORD")
	}()

	configData := `
		tlsConfig:
		certFile: "/path/to/cert.pem"
		keyFile: "/path/to/key.pem"
		database:
		host: "${DB_HOST}"
		port: 5432
		user: "testuser"
		password: "${DB_PASSWORD}"
		name: "testdb"
		jwt:
		secret: "test-secret"
		`
	err := os.WriteFile(filepath.Join("config", "config.yml"), []byte(configData), 0644)
	assert.NoError(t, err)

	err = Load()
	assert.NoError(t, err)

	assert.Equal(t, "test-host", config.Database.Host)
	assert.Equal(t, "env-password", config.Database.Password)
}
