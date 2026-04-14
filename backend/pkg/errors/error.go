// Package errors provides structured error handling using samber/oops.
// Use oops directly with string codes, and map to HTTP/gRPC at the boundaries.
package errors

import (
	"errors"
	"net/http"

	"github.com/samber/oops"
	"google.golang.org/grpc/codes"
)

// Error codes as strings - use these with oops.Code()
const (
	CodeOK              = "OK"
	CodeBadRequest      = "BAD_REQUEST"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeNotFound        = "NOT_FOUND"
	CodeConflict        = "CONFLICT"
	CodePrecondition    = "PRECONDITION_FAILED"
	CodeTooManyRequests = "TOO_MANY_REQUESTS"
	CodeUnprocessable   = "UNPROCESSABLE_ENTITY"
	CodeInternal        = "INTERNAL"
	CodeUnavailable     = "UNAVAILABLE"
	CodeTimeout         = "TIMEOUT"
)

// HTTPStatus returns the HTTP status code for an error code.
func HTTPStatus(code string) int {
	switch code {
	case CodeOK:
		return http.StatusOK
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodePrecondition:
		return http.StatusPreconditionFailed
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeUnprocessable:
		return http.StatusUnprocessableEntity
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	case CodeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// GRPCStatus returns the gRPC status code for an error code.
func GRPCStatus(code string) codes.Code {
	switch code {
	case CodeOK:
		return codes.OK
	case CodeBadRequest, CodeUnprocessable:
		return codes.InvalidArgument
	case CodeUnauthorized:
		return codes.Unauthenticated
	case CodeForbidden:
		return codes.PermissionDenied
	case CodeNotFound:
		return codes.NotFound
	case CodeConflict:
		return codes.AlreadyExists
	case CodePrecondition:
		return codes.FailedPrecondition
	case CodeTooManyRequests:
		return codes.ResourceExhausted
	case CodeUnavailable:
		return codes.Unavailable
	case CodeTimeout:
		return codes.DeadlineExceeded
	default:
		return codes.Internal
	}
}

// ShouldLog returns true if errors with this code should be logged at error level.
// Generally 5xx errors should be logged, 4xx should not.
func ShouldLog(code string) bool {
	switch code {
	case CodeInternal, CodeUnavailable, CodeTimeout:
		return true
	default:
		return false
	}
}

// DefaultMessage returns a safe, generic message for each error code.
func DefaultMessage(code string) string {
	switch code {
	case CodeOK:
		return "ok"
	case CodeBadRequest:
		return "bad request"
	case CodeUnauthorized:
		return "unauthorized"
	case CodeForbidden:
		return "forbidden"
	case CodeNotFound:
		return "not found"
	case CodeConflict:
		return "conflict"
	case CodePrecondition:
		return "precondition failed"
	case CodeTooManyRequests:
		return "too many requests"
	case CodeUnprocessable:
		return "unprocessable entity"
	case CodeUnavailable:
		return "service unavailable"
	case CodeTimeout:
		return "request timeout"
	default:
		return "internal server error"
	}
}

// GetCode extracts the error code from an oops error.
// Returns CodeInternal if the error is not an oops error or has no code.
func GetCode(err error) string {
	if err == nil {
		return CodeInternal
	}

	var oopsErr oops.OopsError
	if errors.As(err, &oopsErr) {
		if code, ok := oopsErr.Code().(string); ok && code != "" {
			return code
		}
	}

	return CodeInternal
}

// GetPublicMessage extracts the public message from an oops error.
// Returns the default message for the code if no public message is set.
func GetPublicMessage(err error) string {
	if err == nil {
		return ""
	}

	var oopsErr oops.OopsError
	if errors.As(err, &oopsErr) {
		if msg := oopsErr.Public(); msg != "" {
			return msg
		}
	}

	return DefaultMessage(GetCode(err))
}

// Re-export standard library functions for convenience
var (
	Is = errors.Is
	As = errors.As
)

// New creates a new oops error builder with the given code.
// This is a convenience wrapper around oops.Code().
func New(code string) oops.OopsErrorBuilder {
	return oops.Code(code)
}

// Convenience constructors per error-handling rules. Each sets the passed
// message as both the internal error text and the Public() user-facing text,
// so ToConnectError surfaces a useful message instead of the generic default.

func NotFound(msg string) error {
	return oops.Code(CodeNotFound).Public(msg).Errorf("%s", msg)
}

func BadRequest(msg string) error {
	return oops.Code(CodeBadRequest).Public(msg).Errorf("%s", msg)
}

// Internal is used for server-side failures. It does NOT set a public message —
// the generic "internal server error" is shown instead to avoid leaking details.
func Internal(msg string) error {
	return oops.Code(CodeInternal).Errorf("%s", msg)
}

func Unauthorized(msg string) error {
	return oops.Code(CodeUnauthorized).Public(msg).Errorf("%s", msg)
}

func Forbidden(msg string) error {
	return oops.Code(CodeForbidden).Public(msg).Errorf("%s", msg)
}

func Conflict(msg string) error {
	return oops.Code(CodeConflict).Public(msg).Errorf("%s", msg)
}

// Wrap wraps an error with a short domain message, preserving the original cause.
func Wrap(err error, msg string) error {
	return oops.Wrapf(err, "%s", msg)
}
