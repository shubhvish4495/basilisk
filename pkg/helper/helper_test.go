package helper

import (
	"encoding/json"
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
		name     string
		err      HttpError
		wantCode int
		wantBody ErrorResponse
	}{
		{
			name:     "With bad request error",
			err:      NewHttpError(http.StatusBadRequest, "invalid request"),
			wantCode: http.StatusBadRequest,
			wantBody: ErrorResponse{
				Error: "invalid request",
			},
		},
		{
			name:     "With nil error defaults to internal server error",
			err:      nil,
			wantCode: http.StatusInternalServerError,
			wantBody: ErrorResponse{
				Error: "Internal Server Error",
			},
		},
		{
			name:     "With unauthorized error",
			err:      NewHttpError(http.StatusUnauthorized, "unauthorized access"),
			wantCode: http.StatusUnauthorized,
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
			SendError(w, tt.err)

			// Check status code
			assert.Equal(t, tt.wantCode, w.Code)

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

func TestSendSuccessResponse(t *testing.T) {
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

			// Call SendSuccessResponse
			SendSuccessResponse(w, tt.statusCode, tt.data)

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

func TestSendSuccessResponse_LargeData(t *testing.T) {
	// Create large data structure
	largeData := make([]string, 1000)
	for i := range largeData {
		largeData[i] = "test data"
	}

	w := httptest.NewRecorder()
	SendSuccessResponse(w, http.StatusOK, largeData)

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

func TestSendPaginatedSuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := []string{"a", "b", "c"}
	SendPaginatedSuccessResponse(w, http.StatusOK, data, 10, 5)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response PaginatedSuccessResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, 5, response.Offset)
	assert.Equal(t, 0, response.TotalCount)
	assert.Len(t, response.Data.([]interface{}), 3)
}

func TestSendPaginatedSuccessResponseWithCount(t *testing.T) {
	w := httptest.NewRecorder()
	data := []string{"a", "b"}
	SendPaginatedSuccessResponseWithCount(w, http.StatusOK, data, 42, 10, 20)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response PaginatedSuccessResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, 42, response.TotalCount)
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, 20, response.Offset)
	assert.Len(t, response.Data.([]interface{}), 2)
}

func TestGetLimitAndOffset(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantLimit  int
		wantOffset int
	}{
		{
			name:       "No params uses defaults",
			query:      "",
			wantLimit:  10,
			wantOffset: 0,
		},
		{
			name:       "Custom limit and offset",
			query:      "limit=25&offset=50",
			wantLimit:  25,
			wantOffset: 50,
		},
		{
			name:       "Only limit provided",
			query:      "limit=5",
			wantLimit:  5,
			wantOffset: 0,
		},
		{
			name:       "Only offset provided",
			query:      "offset=15",
			wantLimit:  10,
			wantOffset: 15,
		},
		{
			name:       "Invalid limit falls back to default",
			query:      "limit=abc&offset=10",
			wantLimit:  10,
			wantOffset: 10,
		},
		{
			name:       "Invalid offset falls back to default",
			query:      "limit=5&offset=xyz",
			wantLimit:  5,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)
			limit, offset := GetLimitAndOffset(r)
			assert.Equal(t, tt.wantLimit, limit)
			assert.Equal(t, tt.wantOffset, offset)
		})
	}
}

func TestBase64DecodeString(t *testing.T) {
	t.Run("Valid JSON string", func(t *testing.T) {
		input := `{"key":"value"}`
		result, err := Base64DecodeString(input)
		assert.NoError(t, err)
		assert.JSONEq(t, input, string(result))
	})

	t.Run("Simple string", func(t *testing.T) {
		input := `"hello"`
		result, err := Base64DecodeString(input)
		assert.NoError(t, err)
		assert.Equal(t, `"hello"`, string(result))
	})
}

func TestHashString(t *testing.T) {
	t.Run("Consistent hashing", func(t *testing.T) {
		hash1 := HashString("test")
		hash2 := HashString("test")
		assert.Equal(t, hash1, hash2)
	})

	t.Run("Different inputs produce different hashes", func(t *testing.T) {
		hash1 := HashString("hello")
		hash2 := HashString("world")
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("Known SHA256 value", func(t *testing.T) {
		// SHA256 of "test" is well-known
		hash := HashString("test")
		assert.Equal(t, "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", hash)
	})

	t.Run("Empty string", func(t *testing.T) {
		hash := HashString("")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64) // SHA256 hex is 64 chars
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
				SendError(w, NewHttpError(http.StatusBadRequest, "test error"))
			},
		},
		{
			name: "SendSuccessResponse sets content type",
			testFunc: func(w http.ResponseWriter) {
				SendSuccessResponse(w, http.StatusOK, "test data")
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
