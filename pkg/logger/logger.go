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
	infoLogger    *defaultLogger.Logger
	warningLogger *defaultLogger.Logger
	errorLogger   *defaultLogger.Logger
	debugLogger   *defaultLogger.Logger
}

type Level int

const (
	InfoLevel Level = iota
	WarningLevel
	ErrorLevel
	DebugLevel
)

var currentLevel Level = DebugLevel

// formatContext formatta il contesto in una stringa per essere utilizzata nei log
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

// log è un metodo privato che scrive il messaggio di log utilizzando il logger specificato
func (l *Logger) log(logger *defaultLogger.Logger, v ...interface{}) {
	if l == nil {
		return // Se il logger non è inizializzato, non fare nulla
	}
	logger.Output(3, fmt.Sprint(v...))
}

// Info scrive un messaggio di log di livello informativo
func (l *Logger) Info(v ...interface{}) {
	if l == nil {
		return // Se il logger non è inizializzato, non fare nulla
	}
	if currentLevel > InfoLevel {
		l.log(l.infoLogger, v...)
	}
}

// Warn scrive un messaggio di log di livello di avviso
func (l *Logger) Warn(v ...interface{}) {
	if l == nil {
		return // Se il logger non è inizializzato, non fare nulla
	}
	if currentLevel > WarningLevel {
		l.log(l.warningLogger, v...)
	}
}

// Error scrive un messaggio di log di livello di errore
func (l *Logger) Error(v ...interface{}) {
	if l == nil {
		return // Se il logger non è inizializzato, non fare nulla
	}
	if currentLevel > ErrorLevel {
		l.log(l.errorLogger, v...)
	}
}

// Debug scrive un messaggio di log di livello di debug
func (l *Logger) Debug(v ...interface{}) {
	if l == nil {
		return // Se il logger non è inizializzato, non fare nulla
	}
	if currentLevel > DebugLevel {
		l.log(l.debugLogger, v...)
	}
}

// Log è il logger globale che può essere utilizzato in tutto il pacchetto
var Log *Logger = nil

// CreateLogger inizializza il logger globale con il contesto fornito
func CreateLogger(ctx Context) {
	prefix := formatContext(ctx)
	Log = &Logger{
		infoLogger:    defaultLogger.New(os.Stdout, "[INFO]    "+prefix, defaultLogger.Ldate|defaultLogger.Ltime|defaultLogger.Lshortfile),
		warningLogger: defaultLogger.New(os.Stdout, "[WARNING] "+prefix, defaultLogger.Ldate|defaultLogger.Ltime|defaultLogger.Lshortfile),
		errorLogger:   defaultLogger.New(os.Stderr, "[ERROR]   "+prefix, defaultLogger.Ldate|defaultLogger.Ltime|defaultLogger.Lshortfile),
		debugLogger:   defaultLogger.New(os.Stdout, "[DEBUG]   "+prefix, defaultLogger.Ldate|defaultLogger.Ltime|defaultLogger.Lshortfile),
	}
}

func SetLoggerLevel(level Level) {
	currentLevel = level
}

func GetContext(service, BuildingID, FloorID, HubID string) Context {
	return Context{
		"service":  service,
		"building": BuildingID,
		"floor":    FloorID,
		"hub":      HubID,
	}
}
