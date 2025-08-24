package errors

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/io"
)

// WorkflowLogLevel represents workflow-specific log levels
type WorkflowLogLevel int

const (
	WorkflowLogLevelDebug WorkflowLogLevel = iota
	WorkflowLogLevelInfo
	WorkflowLogLevelWarn
	WorkflowLogLevelError
	WorkflowLogLevelCritical
)

// WorkflowLogContext represents workflow-specific logging context
type WorkflowLogContext struct {
	Component   string                 `json:"component"`
	Operation   string                 `json:"operation"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	StartTime   time.Time              `json:"start_time"`
	Duration    time.Duration          `json:"duration,omitempty"`
}

// WorkflowLogger extends the standard slog.Logger with workflow-specific functionality
type WorkflowLogger struct {
	logger  *slog.Logger
	context WorkflowLogContext
	config  WorkflowLogConfig
}

// WorkflowLogConfig holds configuration for workflow logging
type WorkflowLogConfig struct {
	Level           WorkflowLogLevel `json:"level"`
	EnableStructured bool            `json:"enable_structured"`
	EnableFile      bool            `json:"enable_file"`
	FilePath        string          `json:"file_path"`
	EnableConsole   bool            `json:"enable_console"`
	EnableMetrics   bool            `json:"enable_metrics"`
	BufferSize      int             `json:"buffer_size"`
	FlushInterval   time.Duration   `json:"flush_interval"`
}

// NewWorkflowLogger creates a new workflow logger
func NewWorkflowLogger(component string, config WorkflowLogConfig) (*WorkflowLogger, error) {
	// Create base logger
	var handler slog.Handler

	if config.EnableFile {
		// Ensure log directory exists
		logDir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, NewIOError(ErrCodeIOAccess, "failed to create log directory").
				WithContext("directory", logDir).
				WithCause(err)
		}

		// Open log file
		logFile, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, NewIOError(ErrCodeIOWrite, "failed to open log file").
				WithContext("file", config.FilePath).
				WithCause(err)
		}

		if config.EnableStructured {
			handler = slog.NewJSONHandler(logFile, &slog.HandlerOptions{
				Level: convertWorkflowLogLevel(config.Level),
			})
		} else {
			handler = slog.NewTextHandler(logFile, &slog.HandlerOptions{
				Level: convertWorkflowLogLevel(config.Level),
			})
		}
	} else {
		// Use console output
		if config.EnableStructured {
			handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
				Level: convertWorkflowLogLevel(config.Level),
			})
		} else {
			handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: convertWorkflowLogLevel(config.Level),
			})
		}
	}

	logger := slog.New(handler)

	return &WorkflowLogger{
		logger: logger,
		context: WorkflowLogContext{
			Component: component,
			StartTime: time.Now(),
			Metadata:  make(map[string]interface{}),
		},
		config: config,
	}, nil
}

// WithContext creates a new logger with additional context
func (wl *WorkflowLogger) WithContext(ctx WorkflowLogContext) *WorkflowLogger {
	newContext := wl.context
	if ctx.Component != "" {
		newContext.Component = ctx.Component
	}
	if ctx.Operation != "" {
		newContext.Operation = ctx.Operation
	}
	if ctx.RequestID != "" {
		newContext.RequestID = ctx.RequestID
	}
	if ctx.UserID != "" {
		newContext.UserID = ctx.UserID
	}
	if ctx.SessionID != "" {
		newContext.SessionID = ctx.SessionID
	}
	if ctx.Metadata != nil {
		for k, v := range ctx.Metadata {
			newContext.Metadata[k] = v
		}
	}

	return &WorkflowLogger{
		logger:  wl.logger,
		context: newContext,
		config:  wl.config,
	}
}

// WithOperation creates a new logger with operation context
func (wl *WorkflowLogger) WithOperation(operation string) *WorkflowLogger {
	return wl.WithContext(WorkflowLogContext{Operation: operation})
}

// WithRequestID creates a new logger with request ID context
func (wl *WorkflowLogger) WithRequestID(requestID string) *WorkflowLogger {
	return wl.WithContext(WorkflowLogContext{RequestID: requestID})
}

// WithMetadata creates a new logger with additional metadata
func (wl *WorkflowLogger) WithMetadata(key string, value interface{}) *WorkflowLogger {
	metadata := make(map[string]interface{})
	for k, v := range wl.context.Metadata {
		metadata[k] = v
	}
	metadata[key] = value

	return wl.WithContext(WorkflowLogContext{Metadata: metadata})
}

// Debug logs a debug message with workflow context
func (wl *WorkflowLogger) Debug(msg string, args ...interface{}) {
	wl.logWithContext(slog.LevelDebug, msg, args...)
}

// Info logs an info message with workflow context
func (wl *WorkflowLogger) Info(msg string, args ...interface{}) {
	wl.logWithContext(slog.LevelInfo, msg, args...)
}

// Warn logs a warning message with workflow context
func (wl *WorkflowLogger) Warn(msg string, args ...interface{}) {
	wl.logWithContext(slog.LevelWarn, msg, args...)
}

// Error logs an error message with workflow context
func (wl *WorkflowLogger) Error(msg string, args ...interface{}) {
	wl.logWithContext(slog.LevelError, msg, args...)
}

// Critical logs a critical message with workflow context
func (wl *WorkflowLogger) Critical(msg string, args ...interface{}) {
	wl.logWithContext(slog.LevelError, msg, args...)
}

// LogError logs a WorkflowError with full context
func (wl *WorkflowLogger) LogError(err *WorkflowError) {
	attrs := []slog.Attr{
		slog.String("error_code", err.Code),
		slog.String("error_category", err.CategoryString()),
		slog.String("error_severity", err.SeverityString()),
		slog.Bool("recoverable", err.Recoverable),
		slog.Time("error_timestamp", err.Timestamp),
	}

	// Add error context
	for key, value := range err.Context {
		attrs = append(attrs, slog.Any("error_"+key, value))
	}

	// Add workflow context
	attrs = append(attrs, wl.getContextAttrs()...)

	// Add stack trace if available
	if err.StackTrace != "" {
		attrs = append(attrs, slog.String("stack_trace", err.StackTrace))
	}

	// Add cause if available
	if err.Cause != nil {
		attrs = append(attrs, slog.String("cause", err.Cause.Error()))
	}

	level := convertWorkflowErrorSeverity(err.Severity)
	wl.logger.LogAttrs(context.Background(), level, err.Message, attrs...)
}

// StartOperation starts timing an operation
func (wl *WorkflowLogger) StartOperation(operation string) *WorkflowLogger {
	return wl.WithContext(WorkflowLogContext{
		Operation: operation,
		StartTime: time.Now(),
	})
}

// EndOperation ends timing an operation and logs the duration
func (wl *WorkflowLogger) EndOperation(msg string, args ...interface{}) {
	duration := time.Since(wl.context.StartTime)

	attrs := []slog.Attr{
		slog.Duration("duration", duration),
	}
	attrs = append(attrs, wl.getContextAttrs()...)

	// Add additional args
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if ok {
				attrs = append(attrs, slog.Any(key, args[i+1]))
			}
		}
	}

	wl.logger.LogAttrs(context.Background(), slog.LevelInfo, msg, attrs...)
}

// logWithContext logs a message with workflow context
func (wl *WorkflowLogger) logWithContext(level slog.Level, msg string, args ...interface{}) {
	attrs := wl.getContextAttrs()

	// Add additional args
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if ok {
				attrs = append(attrs, slog.Any(key, args[i+1]))
			}
		}
	}

	wl.logger.LogAttrs(context.Background(), level, msg, attrs...)
}

// getContextAttrs returns workflow context as slog attributes
func (wl *WorkflowLogger) getContextAttrs() []slog.Attr {
	attrs := []slog.Attr{
		slog.String("component", wl.context.Component),
	}

	if wl.context.Operation != "" {
		attrs = append(attrs, slog.String("operation", wl.context.Operation))
	}

	if wl.context.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", wl.context.RequestID))
	}

	if wl.context.UserID != "" {
		attrs = append(attrs, slog.String("user_id", wl.context.UserID))
	}

	if wl.context.SessionID != "" {
		attrs = append(attrs, slog.String("session_id", wl.context.SessionID))
	}

	// Add metadata
	for key, value := range wl.context.Metadata {
		attrs = append(attrs, slog.Any(key, value))
	}

	return attrs
}

// Helper functions

// convertWorkflowLogLevel converts WorkflowLogLevel to slog.Level
func convertWorkflowLogLevel(level WorkflowLogLevel) slog.Level {
	switch level {
	case WorkflowLogLevelDebug:
		return slog.LevelDebug
	case WorkflowLogLevelInfo:
		return slog.LevelInfo
	case WorkflowLogLevelWarn:
		return slog.LevelWarn
	case WorkflowLogLevelError, WorkflowLogLevelCritical:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// convertWorkflowErrorSeverity converts ErrorSeverity to slog.Level
func convertWorkflowErrorSeverity(severity ErrorSeverity) slog.Level {
	switch severity {
	case SeverityLow:
		return slog.LevelInfo
	case SeverityMedium:
		return slog.LevelWarn
	case SeverityHigh, SeverityCritical:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// DefaultWorkflowLogConfig returns default workflow logging configuration
func DefaultWorkflowLogConfig() WorkflowLogConfig {
	// Get log directory
	logDir, err := io.GetLogDir(true)
	if err != nil {
		logDir = "logs" // fallback
	}

	return WorkflowLogConfig{
		Level:           WorkflowLogLevelInfo,
		EnableStructured: true,
		EnableFile:      true,
		FilePath:        filepath.Join(logDir, "workflow.log"),
		EnableConsole:   true,
		EnableMetrics:   true,
		BufferSize:      1024,
		FlushInterval:   time.Second * 5,
	}
}

// WorkflowLoggerManager manages multiple workflow loggers
type WorkflowLoggerManager struct {
	loggers map[string]*WorkflowLogger
	config  WorkflowLogConfig
}

// NewWorkflowLoggerManager creates a new workflow logger manager
func NewWorkflowLoggerManager(config WorkflowLogConfig) *WorkflowLoggerManager {
	return &WorkflowLoggerManager{
		loggers: make(map[string]*WorkflowLogger),
		config:  config,
	}
}

// GetLogger gets or creates a logger for a component
func (wlm *WorkflowLoggerManager) GetLogger(component string) (*WorkflowLogger, error) {
	if logger, exists := wlm.loggers[component]; exists {
		return logger, nil
	}

	logger, err := NewWorkflowLogger(component, wlm.config)
	if err != nil {
		return nil, err
	}

	wlm.loggers[component] = logger
	return logger, nil
}

// SetLogLevel sets the log level for all loggers
func (wlm *WorkflowLoggerManager) SetLogLevel(level WorkflowLogLevel) {
	wlm.config.Level = level
	// Note: In a real implementation, you'd want to update existing loggers
}

// Close closes all loggers and flushes any buffered logs
func (wlm *WorkflowLoggerManager) Close() error {
	// In a real implementation, you'd close file handles and flush buffers
	wlm.loggers = make(map[string]*WorkflowLogger)
	return nil
}
