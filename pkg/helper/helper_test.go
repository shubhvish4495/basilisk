package helper

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		wantBody   ErrorResponse
	}{
		{
			name:       "With error message",
			statusCode: http.StatusBadRequest,
			err:        errors.New("invalid request"),
			wantBody: ErrorResponse{
				Error: "invalid request",
			},
		},
		{
			name:       "Without error message",
			statusCode: http.StatusInternalServerError,
			err:        nil,
			wantBody: ErrorResponse{
				Error: "Internal Server Error",
			},
		},
		{
			name:       "Custom error message",
			statusCode: http.StatusUnauthorized,
			err:        errors.New("unauthorized access"),
			wantBody: ErrorResponse{
				Error: "unauthorized access",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a response recorder
			w := httptest.NewRecorder()

			// Call SendError
			SendError(w, tt.statusCode, tt.err)

			// Check status code
			assert.Equal(t, tt.statusCode, w.Code)

			// Check Content-Type header
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			// Parse response body
			var gotBody ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&gotBody)
			assert.NoError(t, err)

			// Check response body
			assert.Equal(t, tt.wantBody, gotBody)
		})
	}
}

func TestSendSuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		data       interface{}
	}{
		{
			name:       "String data",
			statusCode: http.StatusOK,
			data:       "success",
		},
		{
			name:       "Struct data",
			statusCode: http.StatusCreated,
			data: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{
				ID:   1,
				Name: "test",
			},
		},
		{
			name:       "Nil data",
			statusCode: http.StatusNoContent,
			data:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a response recorder
			w := httptest.NewRecorder()

			// Call SendSuccess
			SendSuccess(w, tt.statusCode, tt.data)

			// Check status code
			assert.Equal(t, tt.statusCode, w.Code)

			// Check Content-Type header
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			// Parse response body
			var gotBody SuccessResponse
			err := json.NewDecoder(w.Body).Decode(&gotBody)
			assert.NoError(t, err)

			// Check response body
			expectedResponse := SuccessResponse{
				Data: tt.data,
			}

			if tt.data != nil && reflect.TypeOf(tt.data).String() != "string" {
				assert.Equal(t, expectedResponse.Data.(struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				}).ID, int(gotBody.Data.(map[string]interface{})["id"].(float64)))

				assert.Equal(t, expectedResponse.Data.(struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				}).Name, gotBody.Data.(map[string]interface{})["name"])
			} else {
				assert.Equal(t, expectedResponse, gotBody)
			}

		})
	}
}

func TestSendSuccess_LargeData(t *testing.T) {
	// Create large data structure
	largeData := make([]string, 1000)
	for i := range largeData {
		largeData[i] = "test data"
	}

	w := httptest.NewRecorder()
	SendSuccess(w, http.StatusOK, largeData)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Verify response can be read completely
	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	var response SuccessResponse
	err = json.Unmarshal(body, &response)
	assert.NoError(t, err)
	assert.Len(t, response.Data.([]interface{}), 1000)
}

func TestCheckFileExist(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		setup    func() string
		expected bool
	}{
		{
			name: "Existing file",
			setup: func() string {
				path := filepath.Join(tempDir, "existing.txt")
				err := os.WriteFile(path, []byte("test"), 0644)
				assert.NoError(t, err)
				return path
			},
			expected: true,
		},
		{
			name: "Non-existing file",
			setup: func() string {
				return filepath.Join(tempDir, "nonexisting.txt")
			},
			expected: false,
		},
		{
			name: "Directory instead of file",
			setup: func() string {
				path := filepath.Join(tempDir, "testdir")
				err := os.Mkdir(path, 0755)
				assert.NoError(t, err)
				return path
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			result := CheckFileExist(path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponseStructures(t *testing.T) {
	t.Run("SuccessResponse", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		response := SuccessResponse{Data: data}

		// Marshal to JSON
		jsonData, err := json.Marshal(response)
		assert.NoError(t, err)

		// Unmarshal back
		var decoded SuccessResponse
		err = json.Unmarshal(jsonData, &decoded)
		assert.NoError(t, err)

		// Check if data is preserved
		decodedData, ok := decoded.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "value", decodedData["key"])
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		response := ErrorResponse{Error: "test error"}

		// Marshal to JSON
		jsonData, err := json.Marshal(response)
		assert.NoError(t, err)

		// Unmarshal back
		var decoded ErrorResponse
		err = json.Unmarshal(jsonData, &decoded)
		assert.NoError(t, err)

		// Check if error message is preserved
		assert.Equal(t, "test error", decoded.Error)
	})
}

func TestContentTypeHeader(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(w http.ResponseWriter)
	}{
		{
			name: "SendError sets content type",
			testFunc: func(w http.ResponseWriter) {
				SendError(w, http.StatusBadRequest, errors.New("test error"))
			},
		},
		{
			name: "SendSuccess sets content type",
			testFunc: func(w http.ResponseWriter) {
				SendSuccess(w, http.StatusOK, "test data")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.testFunc(w)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}
