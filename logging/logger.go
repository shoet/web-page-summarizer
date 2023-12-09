package logging

import (
	"context"
	"io"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger *zerolog.Logger
}

func NewLogger(w io.Writer) *Logger {
	l := zerolog.
		New(w).
		With().
		Timestamp().
		Logger()
	return &Logger{logger: &l}
}

func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

func (l *Logger) Error(msg string, err error) {
	l.logger.Error().Err(err).Msg(msg)
}

func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

func (l *Logger) Fatal(msg string, err error) {
	l.logger.Fatal().Err(err).Msg(msg)
}

func (l *Logger) NewTraceIdLogger(traceId string) *Logger {
	nl := l.logger.With().Str("traceId", traceId).Logger()
	return &Logger{logger: &nl}
}

type contextKey struct{}

var LoggerContextKey = contextKey{}

func SetLogger(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, LoggerContextKey, l)
}

func GetLogger(ctx context.Context) *Logger {
	return ctx.Value(LoggerContextKey).(*Logger)
}
