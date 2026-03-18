package logutil

import (
	"context"
	"io"

	"github.com/rs/zerolog"
)

type ctxKey string

const keyLogger ctxKey = "logger"

// DefaultLogger creates a zerolog logger that writes to w with timestamps enabled.
func DefaultLogger(w io.Writer) zerolog.Logger {
	return zerolog.New(w).With().Timestamp().Logger()
}

// WithLogger returns a new context with the given logger attached.
func WithLogger(ctx context.Context, log zerolog.Logger) context.Context {
	return context.WithValue(ctx, keyLogger, log)
}

// FromContext retrieves the logger stored in ctx. The ok reports whether one was found.
func FromContext(ctx context.Context) (log zerolog.Logger, ok bool) {
	log, ok = ctx.Value(keyLogger).(zerolog.Logger)
	return
}
