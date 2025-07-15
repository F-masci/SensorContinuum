package logger

import (
	"fmt"
	defaultLogger "log"
	"os"
	"strings"
)

// Context rappresenta i dati di contesto da includere in ogni log
type Context map[string]string

type Logger struct {
	infoLogger  *defaultLogger.Logger
	errorLogger *defaultLogger.Logger
	debugLogger *defaultLogger.Logger
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

func (l *Logger) log(logger *defaultLogger.Logger, v ...interface{}) {
	if l == nil {
		return // Se il logger non è inizializzato, non fare nulla
	}
	logger.Output(3, fmt.Sprint(v...))
}

func (l *Logger) Info(v ...interface{}) {
	if l == nil {
		return // Se il logger non è inizializzato, non fare nulla
	}
	l.log(l.infoLogger, v...)
}

func (l *Logger) Error(v ...interface{}) {
	if l == nil {
		return // Se il logger non è inizializzato, non fare nulla
	}
	l.log(l.errorLogger, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	if l == nil {
		return // Se il logger non è inizializzato, non fare nulla
	}
	l.log(l.debugLogger, v...)
}

var Log *Logger = nil

func CreateLogger(ctx Context) {
	prefix := formatContext(ctx)
	Log = &Logger{
		infoLogger:  defaultLogger.New(os.Stdout, "[INFO]  "+prefix, defaultLogger.Ldate|defaultLogger.Ltime|defaultLogger.Lshortfile),
		errorLogger: defaultLogger.New(os.Stderr, "[ERROR] "+prefix, defaultLogger.Ldate|defaultLogger.Ltime|defaultLogger.Lshortfile),
		debugLogger: defaultLogger.New(os.Stdout, "[DEBUG] "+prefix, defaultLogger.Ldate|defaultLogger.Ltime|defaultLogger.Lshortfile),
	}
}
