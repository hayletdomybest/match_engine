package log

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ZeroLogger struct {
}

func NewZeroLogger(level LogLevel) *ZeroLogger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.SetGlobalLevel(zerolog.Level(level))
	return &ZeroLogger{}
}

func (logger *ZeroLogger) Info(format string, v ...any) {
	log.Info().Msgf(format, v...)
}

func (logger *ZeroLogger) InfoJson(obj interface{}) {
	log.Info().Any("data", obj).Send()
}
