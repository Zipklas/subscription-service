package logger

import (
	"context"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func New(level slog.Level) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

// Методы с контекстом
func (l *Logger) Debug(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.DebugContext(ctx, msg, args...)
}

func (l *Logger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.InfoContext(ctx, msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.WarnContext(ctx, msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.ErrorContext(ctx, msg, args...)
}

// Методы без контекста (для обратной совместимости)
func (l *Logger) Debugf(msg string, args ...interface{}) {
	l.Logger.Debug(msg, args...)
}

func (l *Logger) Infof(msg string, args ...interface{}) {
	l.Logger.Info(msg, args...)
}

func (l *Logger) Warnf(msg string, args ...interface{}) {
	l.Logger.Warn(msg, args...)
}

func (l *Logger) Errorf(msg string, args ...interface{}) {
	l.Logger.Error(msg, args...)
}
