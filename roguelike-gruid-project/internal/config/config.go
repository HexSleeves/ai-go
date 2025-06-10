package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// Global configuration instance
var Config *FullConfig

// Init initializes the configuration
func Init() {
	var err error
	Config, err = LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load game config: %v", err)
	}

	// Set log level based on config
	switch Config.Advanced.LogLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Set log output to file if enabled
	if Config.Advanced.LogToFile {
		// Ensure the logs directory exists
		logDir := "logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			logrus.Fatalf("Failed to create log directory: %v", err)
		}

		// Create a log file with a timestamp
		logFilePath := filepath.Join(logDir, fmt.Sprintf("game_%s.log", time.Now().Format("20060102_150405")))
var logFile *os.File

 // Set log output to file if enabled
 if Config.Advanced.LogToFile {
   // ... existing code ...

  logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
   if err == nil {
    logrus.SetOutput(logFile)
   } else {
     logrus.Infof("Failed to log to file, using default stderr: %v", err)
   }
	logrus.Infof("Game config loaded - Debug mode: %v, Log level: %s", Config.Advanced.DebugMode, Config.Advanced.LogLevel)
}
