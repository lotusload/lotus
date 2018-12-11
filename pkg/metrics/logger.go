package metrics

import "log"

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type logger struct{}

func (l logger) Infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}
