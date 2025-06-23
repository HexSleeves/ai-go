package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// Global configuration instance
var Config *FullConfig

// logFile holds the current log file for reuse by slog
var logFile *os.File

// getSlogLevel converts string log level to slog.Level
func getSlogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// RemoveTime removes the top-level time attribute.
// It is intended to be used as a ReplaceAttr function,
// to make example output deterministic.
func RemoveTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		return slog.Attr{}
	}
	return a
}

// InitSlog initializes slog with the configured log level and output
func InitSlog() *slog.Logger {
	var handler slog.Handler

	logLevel := os.Getenv("LOG")
	if logLevel == "" {
		logLevel = Config.Advanced.LogLevel
	}

	// Use LogLevel or environment variable
	level := getSlogLevel(logLevel)

	if Config.Advanced.LogToFile && logFile != nil {
		// Use JSON handler for file output (better for production)
		handler = slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			Level: level,
		})
	} else {

		if level == slog.LevelDebug {
			handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: level,
			})
		} else {
			handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level:       level,
				ReplaceAttr: RemoveTime,
			})
		}

	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// Init initializes the configuration
func Init() {
	var err error
	Config, err = LoadConfig()
	if err != nil {
		slog.Error("Failed to load game config", "error", err)
		os.Exit(1)
	}

	// Initialize slog with the configuration
	InitSlog()

	// Set log output to file if enabled
	if Config.Advanced.LogToFile {
		// Ensure the logs directory exists
		logDir := "logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			slog.Error("Failed to create log directory", "error", err)
			os.Exit(1)
		}

		// Create a log file with a timestamp
		logFilePath := filepath.Join(logDir, fmt.Sprintf("game_%s.log", time.Now().Format("20060102_150405")))

		// Set log output to file if enabled
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			slog.Info("Failed to log to file, using default stderr", "error", err)
		}
	}

	slog.Info("Game config loaded", "debugMode", Config.Advanced.DebugMode, "logLevel", Config.Advanced.LogLevel)
}
