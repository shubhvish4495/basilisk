package helper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

var (
	defaultLimit  = 10
	defaultOffset = 0
)

type SuccessResponse struct {
	Data any `json:"data,omitempty"`
}

// ErrorResponse defines the structure of an error response
type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

// SendError sends an error response to the client in JSON format.
// It sets the "Content-Type" header to "application/json" and writes the
// appropriate HTTP status code based on the provided HttpError.
//
// Parameters:
//   - w: The HTTP response writer used to send the response.
//   - err: An instance of HttpError containing the error details.
//
// The function encodes an ErrorResponse struct into JSON and writes it to the response.
// If the provided error is nil, it defaults to sending an internal server error message.
// nolint:errcheck
func SendError(w http.ResponseWriter, err HttpError) {
	if err == nil {
		err = InternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode())

	response := ErrorResponse{
		Error: err.Error(),
	}

	json.NewEncoder(w).Encode(response)
}

// SendSuccessResponse sends a JSON response with a success message and data.
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

func SendSuccessResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := SuccessResponse{
		Data: data,
	}

	_ = json.NewEncoder(w).Encode(response)
}

// PaginatedSuccessResponse defines the structure for paginated API responses.
// It includes the data payload along with pagination metadata.
type PaginatedSuccessResponse struct {
	Data       any `json:"data,omitempty"`
	TotalCount int `json:"total_count"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
}

// SendPaginatedSuccessResponse sends a JSON response with paginated data.
// It sets the Content-Type header to "application/json" and writes the provided status code.
// The response includes the data, limit, and offset. Note: TotalCount is set to 0 by default.
// For responses with proper total count, use SendPaginatedSuccessResponseWithCount instead.
//
// Parameters:
//   - w: The http.ResponseWriter to write the response to
//   - statusCode: The HTTP status code to set in the response
//   - data: The data to include in the response body
//   - limit: The maximum number of items per page
//   - offset: The number of items to skip (for pagination)
func SendPaginatedSuccessResponse(w http.ResponseWriter, statusCode int, data any, limit, offset int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := PaginatedSuccessResponse{
		Data:       data,
		TotalCount: 0, // Default value, caller should use SendPaginatedSuccessResponseWithCount for proper total
		Limit:      limit,
		Offset:     offset,
	}

	_ = json.NewEncoder(w).Encode(response)
}

// SendPaginatedSuccessResponseWithCount sends a JSON response with paginated data and total count.
// It sets the Content-Type header to "application/json" and writes the provided status code.
// The response includes the data, total count, limit, and offset.
func SendPaginatedSuccessResponseWithCount(w http.ResponseWriter, statusCode int, data any, totalCount, limit, offset int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := PaginatedSuccessResponse{
		Data:       data,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}

	_ = json.NewEncoder(w).Encode(response)
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

// GetLimitAndOffset extracts the "limit" and "offset" query parameters from the given HTTP request.
// If the parameters are not present or cannot be converted to integers, it falls back to DefaultLimit and DefaultOffset.
// Any conversion errors are logged using the provided logger.
// Returns the limit and offset values as integers.
func GetLimitAndOffset(r *http.Request) (int, int) {
	limit := defaultLimit
	offset := defaultOffset

	queryParams := r.URL.Query()
	limitStr := queryParams.Get("limit")
	offsetStr := queryParams.Get("offset")

	if limitStr != "" {
		cVar, err := strconv.Atoi(limitStr)
		if err == nil {
			limit = cVar
		} else {
			slog.Error("error converting limit to int, setting default limit", "error", err)
		}
	}

	if offsetStr != "" {
		cVar, err := strconv.Atoi(offsetStr)
		if err == nil {

			offset = cVar
		} else {
			slog.Error("error converting offset to int, setting default offset", "error", err)
		}
	}

	return limit, offset
}

// Base64DecodeString converts a string to a JSON-encoded byte slice.
// Note: Despite the name suggesting base64 decoding, this function actually
// marshals the input string as a JSON RawMessage.
//
// Parameters:
//   - s: The input string to be JSON-encoded
//
// Returns:
//   - []byte: The JSON-encoded byte representation of the input string
//   - error: An error if JSON marshaling fails, nil otherwise
func Base64DecodeString(s string) ([]byte, error) {
	return json.RawMessage(s).MarshalJSON()
}

func HashString(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}

func GetLogger(ctx context.Context) *slog.Logger {
	requestId, _ := GetRequestIdFromContext(ctx)
	return slog.Default().With("request_id", requestId)
}
