package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

type Logger struct {
	level  Level
	logger *log.Logger
}

var GlobalLogger *Logger

func Init(levelStr string) {
	var level Level
	switch levelStr {
	case "DEBUG":
		level = DEBUG
	case "INFO":
		level = INFO
	case "WARN":
		level = WARN
	case "ERROR":
		level = ERROR
	case "FATAL":
		level = FATAL
	default:
		level = INFO
	}

	GlobalLogger = &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) log(level Level, prefix string, v ...interface{}) {
	if level >= l.level {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		_, file, line, ok := runtime.Caller(3)
		if !ok {
			file = "unknown"
			line = 0
		}

		shortFile := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				shortFile = file[i+1:]
				break
			}
		}

		message := fmt.Sprintf("[%s] %s %s:%d %s", timestamp, prefix, shortFile, line, fmt.Sprint(v...))
		l.logger.Println(message)
	}
}

func (l *Logger) Debug(v ...interface{}) {
	l.log(DEBUG, "[DEBUG]", v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.log(INFO, "[INFO]", v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.log(WARN, "[WARN]", v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.log(ERROR, "[ERROR]", v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.log(FATAL, "[FATAL]", v...)
	os.Exit(1)
}

func Debug(v ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Debug(v...)
	}
}

func Info(v ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Info(v...)
	}
}

func Warn(v ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Warn(v...)
	}
}

func Error(v ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Error(v...)
	}
}

func Fatal(v ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Fatal(v...)
	}
}
