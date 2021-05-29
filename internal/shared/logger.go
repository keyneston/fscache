package shared

import (
	"os"

	"github.com/sirupsen/logrus"
)

var debug bool
var level string = "error"
var logger *logrus.Logger

func Logger() *logrus.Logger {
	if logger != nil {
		return logger
	}

	if debug {
		level = "debug"
	}

	logger = logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	logger.Out = os.Stderr
	logger.SetFormatter(&logrus.JSONFormatter{PrettyPrint: false})

	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logger.Errorf("Unable to parse level %q; defaulting to %q", level, logrus.ErrorLevel)
		parsedLevel = logrus.ErrorLevel
	}
	logger.SetLevel(parsedLevel)

	return logger
}

func SetLevel(lvl logrus.Level) {
	level = lvl.String()
	logger.SetLevel(lvl)
}
