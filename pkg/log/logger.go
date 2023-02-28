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

func BigPrintf(a string, b ...interface{}) {
	l := Get()
	l.Printf("=========")
	if b != nil {
		l.Printf(a, b...)
	} else {
		l.Printf(a)
	}
	l.Printf("=========")
}
