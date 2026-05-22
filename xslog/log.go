package xslog

import (
	"context"
	"log/slog"
	"os"
)

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Error(msg string, err error, args ...any) {
	slog.Error(msg, append([]any{"error", err}, args...)...)
}

func Fatal(msg string, err error, args ...any) {
	slog.Error(msg, append([]any{"error", err}, args...)...)
	os.Exit(1)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	slog.DebugContext(ctx, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, err error, args ...any) {
	slog.ErrorContext(ctx, msg, append([]any{"error", err}, args...)...)
}

func FatalContext(ctx context.Context, msg string, err error, args ...any) {
	slog.ErrorContext(ctx, msg, append([]any{"error", err}, args...)...)
	os.Exit(1)
}
