package logger

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
)

type LoggerStruct struct {
	Logger  *slog.Logger
	LogFile *os.File
}

// convierte un string como "debug" o "error" en un slog.Level.
func ParseLevel(levelStr string) (slog.Level, error) {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, errors.New("nivel de log desconocido: " + levelStr)
	}
}

// crea un logger que escribe a consola y a archivo.
func NewLogger(filename string, level slog.Level) (*LoggerStruct, error) {
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	handler := slog.NewTextHandler(multiWriter, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)

	return &LoggerStruct{
		Logger:  logger,
		LogFile: logFile,
	}, nil
}

// cierra el archivo de log para liberar recursos.
func (l *LoggerStruct) Close() {
	l.LogFile.Close()
}

// Info registra un mensaje con nivel INFO.
func (l *LoggerStruct) Info(msg string) {
	l.Logger.Info(msg)
}

// Debug registra un mensaje con nivel DEBUG.
func (l *LoggerStruct) Debug(msg string) {
	l.Logger.Debug(msg)
}

// Warn registra un mensaje con nivel WARNING.
func (l *LoggerStruct) Warn(msg string) {
	l.Logger.Warn(msg)
}

// Error registra un mensaje con nivel ERROR.
func (l *LoggerStruct) Error(msg string) {
	l.Logger.Error(msg)
}
