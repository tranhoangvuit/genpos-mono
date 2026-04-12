package errors

import (
	stderrors "errors"

	"connectrpc.com/connect"
)

// ToConnectError converts a pkg/errors domain error to a connect.Error.
// Only exposes public messages to prevent information leakage.
func ToConnectError(err error) *connect.Error {
	if err == nil {
		return nil
	}

	code := GetCode(err)
	message := GetPublicMessage(err)

	return connect.NewError(connectCode(code), stderrors.New(message))
}

func connectCode(code string) connect.Code {
	switch code {
	case CodeBadRequest, CodeUnprocessable:
		return connect.CodeInvalidArgument
	case CodeUnauthorized:
		return connect.CodeUnauthenticated
	case CodeForbidden:
		return connect.CodePermissionDenied
	case CodeNotFound:
		return connect.CodeNotFound
	case CodeConflict:
		return connect.CodeAlreadyExists
	case CodePrecondition:
		return connect.CodeFailedPrecondition
	case CodeTooManyRequests:
		return connect.CodeResourceExhausted
	case CodeUnavailable:
		return connect.CodeUnavailable
	case CodeTimeout:
		return connect.CodeDeadlineExceeded
	default:
		return connect.CodeInternal
	}
}
