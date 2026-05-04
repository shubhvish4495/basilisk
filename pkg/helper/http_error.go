package helper

import "net/http"

var (
	// InternalServerErrorMessage is the default message for internal server errors
	InternalServerError = httpError{
		Code: http.StatusInternalServerError,
		Err:  "Internal Server Error",
	}

	// BadRequestErrorMessage is the default message for bad request errors
	BadRequestError = httpError{
		Code: http.StatusBadRequest,
		Err:  "Bad Request",
	}
	// NotFoundErrorMessage is the default message for not found errors
	NotFoundError = httpError{
		Code: http.StatusNotFound,
		Err:  "Not Found",
	}
	// UnauthorizedErrorMessage is the default message for unauthorized errors
	UnauthorizedError = httpError{
		Code: http.StatusUnauthorized,
		Err:  "Unauthorized",
	}
	// ForbiddenErrorMessage is the default message for forbidden errors
	ForbiddenError = httpError{
		Code: http.StatusForbidden,
		Err:  "Forbidden",
	}
	// ConflictErrorMessage is the default message for conflict errors
	ConflictError = httpError{
		Code: http.StatusConflict,
		Err:  "Conflict",
	}
	// TooManyRequestsErrorMessage is the default message for too many requests errors
	TooManyRequestsError = httpError{
		Code: http.StatusTooManyRequests,
		Err:  "Too Many Requests",
	}
)

type HttpError interface {
	StatusCode() int
	Error() string
}

type httpError struct {
	Code int
	Err  string
}

func (e httpError) Error() string {
	return e.Err
}

func (e httpError) StatusCode() int {
	return e.Code
}

func NewHttpError(code int, err string) HttpError {
	return httpError{
		Code: code,
		Err:  err,
	}
}
