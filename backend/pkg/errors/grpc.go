package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPCError converts an error to a gRPC status error.
// Only exposes public messages to prevent information leakage.
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	code := GetCode(err)
	message := GetPublicMessage(err)

	return status.Error(GRPCStatus(code), message)
}

// FromGRPCError converts a gRPC status error to an oops error.
func FromGRPCError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return New(CodeInternal).Wrapf(err, "unknown error")
	}

	code := grpcCodeToAppCode(st.Code())
	return New(code).Errorf("%s", st.Message())
}

// grpcCodeToAppCode maps gRPC codes to application codes.
func grpcCodeToAppCode(c codes.Code) string {
	switch c {
	case codes.OK:
		return CodeOK
	case codes.InvalidArgument:
		return CodeBadRequest
	case codes.Unauthenticated:
		return CodeUnauthorized
	case codes.PermissionDenied:
		return CodeForbidden
	case codes.NotFound:
		return CodeNotFound
	case codes.AlreadyExists:
		return CodeConflict
	case codes.FailedPrecondition:
		return CodePrecondition
	case codes.ResourceExhausted:
		return CodeTooManyRequests
	case codes.Unavailable:
		return CodeUnavailable
	case codes.DeadlineExceeded:
		return CodeTimeout
	default:
		return CodeInternal
	}
}
