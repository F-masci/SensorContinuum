package logger

import (
	"SensorContinuum/pkg/types"
	"errors"
	"fmt"
	defaultLogger "log"
	"os"
	"sort"
	"strings"
)

// Context rappresenta i dati di contesto da includere in ogni log
type Context map[string]string

// Logger è la struttura principale del pacchetto logger
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

var currentLevel = ErrorLevel

// formatContext formatta il contesto in una stringa per essere utilizzata nei log
func formatContext(ctx Context) string {
	if len(ctx) == 0 {
		return ""
	}
	keys := make([]string, 0, len(ctx))
	for k := range ctx {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(ctx))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, ctx[k]))
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
	if currentLevel >= InfoLevel {
		l.log(l.infoLogger, v...)
	}
}

// Warn scrive un messaggio di log di livello di avviso
func (l *Logger) Warn(v ...interface{}) {
	if l == nil {
		return // Se il logger non è inizializzato, non fare nulla
	}
	if currentLevel >= WarningLevel {
		l.log(l.warningLogger, v...)
	}
}

// Error scrive un messaggio di log di livello di errore
func (l *Logger) Error(v ...interface{}) {
	if l == nil {
		return // Se il logger non è inizializzato, non fare nulla
	}
	if currentLevel >= ErrorLevel {
		l.log(l.errorLogger, v...)
	}
}

// Debug scrive un messaggio di log di livello di debug
func (l *Logger) Debug(v ...interface{}) {
	if l == nil {
		return // Se il logger non è inizializzato, non fare nulla
	}
	if currentLevel >= DebugLevel {
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

func PrintCurrentLevel() {
	var levelStr string
	switch currentLevel {
	case InfoLevel:
		levelStr = "Info"
	case WarningLevel:
		levelStr = "Warning"
	case ErrorLevel:
		levelStr = "Error"
	case DebugLevel:
		levelStr = "Debug"
	default:
		levelStr = "Unknown"
	}
	Log.Info("Current logger level: ", levelStr)
}

// LoadLoggerFromEnv carica il livello del logger dalle variabili d'ambiente
// Se la variabile LOG_LEVEL non è impostata, il livello di default è ErrorLevel
func LoadLoggerFromEnv() error {
	LoggerLevelStr, exists := os.LookupEnv("LOG_LEVEL")
	if exists {
		switch LoggerLevelStr {
		case "debug":
			SetLoggerLevel(DebugLevel)
		case "info":
			SetLoggerLevel(InfoLevel)
		case "warning":
			SetLoggerLevel(WarningLevel)
		case "error":
			SetLoggerLevel(ErrorLevel)
		default:
			return errors.New("invalid LOG_LEVEL value, must be 'debug', 'info', 'warning' or 'error'")
		}
	}
	return nil
}

func GetContext(service, macrozone, zone, hub, sensor string) (ctx Context) {
	ctx = make(Context)
	if service != "" {
		ctx["service"] = service
	}
	if macrozone != "" {
		ctx["macrozone"] = macrozone
	}
	if zone != "" {
		ctx["zone"] = zone
	}
	if hub != "" {
		ctx["hub"] = hub
	}
	if sensor != "" {
		ctx["sensor"] = sensor
	}
	return
}

func GetSensorAgentContext(macrozone, zone, sensor string) Context {
	return GetContext("sensor-agent", macrozone, zone, "", sensor)
}

func GetEdgeHubContext(service types.Service, macrozone, zone, hub string) Context {
	return GetContext(string(service), macrozone, zone, hub, "")
}

func GetProximityHubContext(macrozone, hub string) Context {
	return GetContext("proximity-fog-hub", macrozone, "", hub, "")
}

func GetIntermediateHubContext(hub string) Context {
	return GetContext("intermediate-fog-hub", "", "", hub, "")
}
