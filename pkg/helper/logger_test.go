// pkg/helper/logger_test.go
package helper

import (
	"testing"
)

// MockLogger is a mock implementation of the Logger interface for testing
type MockLogger struct {
	infoCount  int
	errorCount int
	debugCount int
}

func (m *MockLogger) Info(format string, args ...interface{})  { m.infoCount++ }
func (m *MockLogger) Error(format string, args ...interface{}) { m.errorCount++ }
func (m *MockLogger) Debug(format string, args ...interface{}) { m.debugCount++ }

func TestInitLogger(t *testing.T) {
	// Reset the global logger before each test
	l = nil

	t.Run("Initialize with nil logger", func(t *testing.T) {
		InitLogger(nil)
		if l == nil {
			t.Error("Logger should not be nil after initialization")
		}
	})

	t.Run("Initialize with custom logger", func(t *testing.T) {
		// Reset the global logger
		l = nil

		mockLogger := &MockLogger{}
		InitLogger(mockLogger)

		if l != mockLogger {
			t.Error("Logger should be set to the provided mock logger")
		}
	})
}

func TestGetLogger(t *testing.T) {
	t.Run("Get logger when not initialized", func(t *testing.T) {
		// Reset the global logger
		l = nil

		logger := GetLogger()
		if logger == nil {
			t.Error("GetLogger should return a non-nil logger")
		}
	})

	t.Run("Get logger when already initialized", func(t *testing.T) {
		// Reset the global logger
		l = nil

		mockLogger := &MockLogger{}
		InitLogger(mockLogger)

		logger := GetLogger()
		if logger != mockLogger {
			t.Error("GetLogger should return the previously initialized logger")
		}
	})
}

func TestLoggerInterface(t *testing.T) {
	mockLogger := &MockLogger{}
	InitLogger(mockLogger)

	t.Run("Test logger methods", func(t *testing.T) {
		logger := GetLogger()

		// Test Info method
		logger.Info("test info")
		if mockLogger.infoCount != 1 {
			t.Errorf("Expected info count to be 1, got %d", mockLogger.infoCount)
		}

		// Test Error method
		logger.Error("test error")
		if mockLogger.errorCount != 1 {
			t.Errorf("Expected error count to be 1, got %d", mockLogger.errorCount)
		}

		// Test Debug method
		logger.Debug("test debug")
		if mockLogger.debugCount != 1 {
			t.Errorf("Expected debug count to be 1, got %d", mockLogger.debugCount)
		}
	})
}
