package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Context rappresenta i dati di contesto da includere in ogni log
type Context map[string]string

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

func New(ctx Context) *Logger {
	prefix := formatContext(ctx)
	return &Logger{
		infoLogger:  log.New(os.Stdout, "[INFO]  "+prefix, log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "[ERROR] "+prefix, log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "[DEBUG] "+prefix, log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// formatta il contesto in stringa chiave=valore separati da spazi
func formatContext(ctx Context) string {
	if len(ctx) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ctx))
	for k, v := range ctx {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, " ") + " "
}

// metodi di log che includono il contesto nei messaggi

func (l *Logger) log(logger *log.Logger, v ...interface{}) {
	logger.Output(3, fmt.Sprint(v...))
}

func (l *Logger) Info(v ...interface{}) {
	l.log(l.infoLogger, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.log(l.errorLogger, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	l.log(l.debugLogger, v...)
}
