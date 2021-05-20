package shared

import (
	"os"

	"github.com/sirupsen/logrus"
)

var debug bool
var logger *logrus.Logger

func Logger() *logrus.Logger {
	if logger != nil {
		return logger
	}

	logger = logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	if debug {
		logger.SetLevel(logrus.DebugLevel)
	}
	logger.Out = os.Stderr
	logger.SetFormatter(&logrus.JSONFormatter{PrettyPrint: false})
	return logger
}
