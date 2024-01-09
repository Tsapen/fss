package fss

import (
	"context"

	"github.com/rs/zerolog"
)

type cxtKey int

const (
	reqIDKey cxtKey = iota
	loggerKey
)

// WithReqID adds request id into context.
func WithReqID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, reqIDKey, reqID)
}

// ReqIDFromCtx gets request id from context.
func ReqIDFromCtx(ctx context.Context) string {
	return ctx.Value(reqIDKey).(string)
}

// WithLogger adds logger into context.
func WithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromCtx gets logger from context.
func LoggerFromCtx(ctx context.Context) zerolog.Logger {
	return ctx.Value(loggerKey).(zerolog.Logger)
}
