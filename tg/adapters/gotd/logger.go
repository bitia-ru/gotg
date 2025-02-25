package gotd

import (
	"fmt"
	"log"
	"os"
)

type LogLevelType int

const (
	DEBUG LogLevelType = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	logger   *log.Logger
	logLevel LogLevelType
}

func NewLogger(out *os.File, prefix string, flag int, level LogLevelType) *Logger {
	return &Logger{
		logger:   log.New(out, prefix, flag),
		logLevel: level,
	}
}

func (l *Logger) log(level LogLevelType, v ...interface{}) {
	if level >= l.logLevel {
		l.logger.Output(3, fmt.Sprintln(v...))
	}
}

func (l *Logger) Debug(v ...interface{}) {
	l.log(DEBUG, v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.log(INFO, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.log(WARN, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.log(ERROR, v...)
}
