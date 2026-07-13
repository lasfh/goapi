package logger

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

var LoggerKey ctxKey

// WithLogger adiciona um logger no contexto.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, l)
}

// FromContext retorna o logger do contexto ou slog.Default().
func From(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return l
	}

	return slog.Default()
}
