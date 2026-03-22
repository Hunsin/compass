package middleware

import (
	"context"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const internalMsg = "internal server error"

// serverStream wraps grpc.ServerStream to inject a logger-enriched context,
// ensuring the logger is available to stream handlers.
type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss serverStream) Context() context.Context {
	return ss.ctx
}

// sanitize ensures that Internal status errors never expose raw error details
// to the client. It replaces the message with a generic one and logs the
// original as a safety net for errors not caught by service-level handlers.
func sanitize(log *zerolog.Logger, method string, err error) error {
	if err == nil {
		return nil
	}

	// For an unhandled error, log it and return a generic Internal error.
	s, ok := status.FromError(err)
	if !ok {
		log.Error().Str("method", method).Err(err).Msg("unhandled error")
		return status.Error(codes.Internal, internalMsg)
	}

	// For an internal error, log the message and return a generic one.
	if s.Code() == codes.Internal {
		log.Error().Str("method", method).Str("error", s.Message()).Msg("internal error")
		return status.Error(codes.Internal, internalMsg)
	}
	return err
}

// UnaryInterceptor returns a gRPC unary server interceptor that sanitizes
// Internal errors as a safety net against accidental detail disclosure.
func UnaryInterceptor(log *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(log.WithContext(ctx), req)
		return resp, sanitize(log, info.FullMethod, err)
	}
}

// StreamInterceptor returns a gRPC stream server interceptor that sanitizes
// Internal errors as a safety net against accidental detail disclosure.
func StreamInterceptor(log *zerolog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ss = serverStream{ss, log.WithContext(ss.Context())}
		return sanitize(log, info.FullMethod, handler(srv, ss))
	}
}
