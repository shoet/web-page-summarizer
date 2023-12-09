package logging

import (
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

func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

type LoggerKey struct{}
