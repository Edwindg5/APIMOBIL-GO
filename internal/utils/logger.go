package utils

import (
	"fmt"
	"log"
	"os"
)

// Logger es un logger personalizado
type Logger struct {
	logger *log.Logger
}

// NewLogger crea una nueva instancia del logger
func NewLogger(prefix string) *Logger {
	return &Logger{
		logger: log.New(os.Stdout, fmt.Sprintf("[%s] ", prefix), log.LstdFlags),
	}
}

// Info registra un mensaje informativo
func (l *Logger) Info(msg string) {
	l.logger.Println("[INFO]", msg)
}

// Error registra un error
func (l *Logger) Error(msg string, err error) {
	if err != nil {
		l.logger.Println("[ERROR]", msg, "-", err.Error())
	} else {
		l.logger.Println("[ERROR]", msg)
	}
}

// Debug registra un mensaje de debug
func (l *Logger) Debug(msg string) {
	l.logger.Println("[DEBUG]", msg)
}

// Warn registra una advertencia
func (l *Logger) Warn(msg string) {
	l.logger.Println("[WARN]", msg)
}

// Fatal registra un error fatal y sale
func (l *Logger) Fatal(msg string, err error) {
	if err != nil {
		l.logger.Fatalln("[FATAL]", msg, "-", err.Error())
	} else {
		l.logger.Fatalln("[FATAL]", msg)
	}
}
