package helper

import (
	"encoding/json"
	"net/http"
	"os"
)

type SuccessResponse struct {
	Data interface{} `json:"data,omitempty"`
}

// ErrorResponse defines the structure of an error response
type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

// SendError sends a JSON response with the specified error message and status code.
// It sets the "Content-Type" header to "application/json" and writes the status code to the response.
// If an error is provided, its message is included in the response.
//
// Parameters:
//   - w: The http.ResponseWriter to write the response to.
//   - statusCode: The HTTP status code to set in the response.
//   - err: An optional error whose message will be included in the response if not nil.
//
//nolint:errcheck
func SendError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{}

	if err != nil {
		response.Error = err.Error()
	} else {
		response.Error = "Internal Server Error"
	}
	json.NewEncoder(w).Encode(response)
}

// SendSuccess sends a JSON response with a success message and data.
// It sets the Content-Type header to "application/json" and writes the provided status code.
//
// Parameters:
//   - w: The http.ResponseWriter to write the response to.
//   - statusCode: The HTTP status code to set in the response.
//   - data: The data to include in the response body.
//
// The response is encoded as a JSON object with the following structure:
//
//	{
//	  "Code": statusCode,
//	  "Data": data
//	}
//
//nolint:errcheck
func SendSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := SuccessResponse{
		Data: data,
	}

	json.NewEncoder(w).Encode(response)
}

// CheckFileExist checks if a file exists at the given file path.
// It returns true if the file exists, and false otherwise.
//
// Parameters:
//   - filePath: A string representing the path to the file.
//
// Returns:
//   - bool: true if the file exists, false if it does not.
func CheckFileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
