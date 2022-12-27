package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

type Logger = zerolog.Logger

const timeFormat = "15:04:05"

// New creates a new logger instance
func New() Logger {

	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: timeFormat})

	if os.Getenv("DURABLE_DEBUG") != "" {
		logger = logger.Level(zerolog.DebugLevel)
	}
	return logger
}

func NewNoopLogger() Logger {
	return zerolog.Nop()
}
