package log

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor that:
// 1. Injects the logger into the request context
// 2. Logs request/response details
func UnaryServerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Add logger to context
		ctx = WithLogger(ctx, logger)

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		code := codes.OK
		if err != nil {
			if s, ok := status.FromError(err); ok {
				code = s.Code()
			} else {
				code = codes.Unknown
			}
		}

		// Log the request
		logLevel := slog.LevelInfo
		if code != codes.OK {
			logLevel = slog.LevelError
		}

		logger.Log(ctx, logLevel, "grpc request",
			slog.String("method", info.FullMethod),
			slog.String("code", code.String()),
			slog.Duration("duration", duration),
			slog.Bool("error", err != nil),
		)

		return resp, err
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor that:
// 1. Injects the logger into the stream context
// 2. Logs stream details
func StreamServerInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		// Create wrapped stream with logger in context
		ctx := WithLogger(ss.Context(), logger)
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Call the handler
		err := handler(srv, wrapped)

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		code := codes.OK
		if err != nil {
			if s, ok := status.FromError(err); ok {
				code = s.Code()
			} else {
				code = codes.Unknown
			}
		}

		// Log the request
		logLevel := slog.LevelInfo
		if code != codes.OK {
			logLevel = slog.LevelError
		}

		logger.Log(ctx, logLevel, "grpc stream",
			slog.String("method", info.FullMethod),
			slog.String("code", code.String()),
			slog.Duration("duration", duration),
			slog.Bool("error", err != nil),
		)

		return err
	}
}

// wrappedServerStream wraps grpc.ServerStream with custom context.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
