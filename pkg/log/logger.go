package log

import "github.com/sirupsen/logrus"

var defaultLoggerInstance ILogger

type ILogger interface {
	logrus.StdLogger
}

func NewLogger() ILogger {
	return logrus.New()
}

func Get() ILogger {
	if defaultLoggerInstance == nil {
		defaultLoggerInstance = NewLogger()
	}
	return defaultLoggerInstance
}
