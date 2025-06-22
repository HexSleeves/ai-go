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

// InitSlog initializes slog with the configured log level and output
func InitSlog() *slog.Logger {
	var handler slog.Handler

	logLevel := os.Getenv("LOG_LEVEL")
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
		// Use text handler for stderr output (better for development)
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})
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
