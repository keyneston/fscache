package shared

import (
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
)

var debug bool
var level string = "error"
var logger *zerolog.Logger
var configLoggerOnce *sync.Once = &sync.Once{}
var pretty bool

func init() {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	l := zerolog.New(os.Stderr)
	logger = &l
}

func Logger() *zerolog.Logger {
	configLoggerOnce.Do(func() {
		if debug {
			level = "debug"
		}

		parsedLevel, err := zerolog.ParseLevel(level)
		if err != nil {
			logger.Error().Str("input", level).Msgf("Unable to parse level %q; defaulting to %q", level, logrus.ErrorLevel)
			parsedLevel = zerolog.ErrorLevel
		}
		zerolog.SetGlobalLevel(parsedLevel)

	})

	return logger
}

func SetLevel(lvl zerolog.Level) {
	level = lvl.String()
	zerolog.SetGlobalLevel(lvl)
}

func SetPrettyLogging() {
	pretty = true

	l := logger.Output(&zerolog.ConsoleWriter{
		Out: os.Stderr,
	})
	logger = &l
}
