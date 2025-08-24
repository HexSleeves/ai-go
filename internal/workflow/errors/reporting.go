package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ErrorReporter handles error reporting and metrics collection
type ErrorReporter struct {
	logger      *slog.Logger
	metrics     *ErrorMetrics
	subscribers []ErrorSubscriber
	mu          sync.RWMutex
}

// ErrorSubscriber defines an interface for error event subscribers
type ErrorSubscriber interface {
	OnError(ctx context.Context, err *WorkflowError)
	GetName() string
}

// ErrorMetrics tracks error statistics
type ErrorMetrics struct {
	TotalErrors       int64                    `json:"total_errors"`
	ErrorsByCategory  map[ErrorCategory]int64  `json:"errors_by_category"`
	ErrorsBySeverity  map[ErrorSeverity]int64  `json:"errors_by_severity"`
	ErrorsByCode      map[string]int64         `json:"errors_by_code"`
	RecoveredErrors   int64                    `json:"recovered_errors"`
	UnrecoveredErrors int64                    `json:"unrecovered_errors"`
	LastError         *WorkflowError           `json:"last_error,omitempty"`
	LastErrorTime     time.Time                `json:"last_error_time"`
	mu                sync.RWMutex
}

// NewErrorReporter creates a new error reporter
func NewErrorReporter(logger *slog.Logger) *ErrorReporter {
	return &ErrorReporter{
		logger:      logger,
		metrics:     NewErrorMetrics(),
		subscribers: []ErrorSubscriber{},
	}
}

// NewErrorMetrics creates a new error metrics instance
func NewErrorMetrics() *ErrorMetrics {
	return &ErrorMetrics{
		ErrorsByCategory: make(map[ErrorCategory]int64),
		ErrorsBySeverity: make(map[ErrorSeverity]int64),
		ErrorsByCode:     make(map[string]int64),
	}
}

// Subscribe adds an error subscriber
func (er *ErrorReporter) Subscribe(subscriber ErrorSubscriber) {
	er.mu.Lock()
	defer er.mu.Unlock()
	
	er.subscribers = append(er.subscribers, subscriber)
	er.logger.Debug("Added error subscriber", "subscriber", subscriber.GetName())
}

// Unsubscribe removes an error subscriber
func (er *ErrorReporter) Unsubscribe(subscriberName string) {
	er.mu.Lock()
	defer er.mu.Unlock()
	
	for i, subscriber := range er.subscribers {
		if subscriber.GetName() == subscriberName {
			er.subscribers = append(er.subscribers[:i], er.subscribers[i+1:]...)
			er.logger.Debug("Removed error subscriber", "subscriber", subscriberName)
			break
		}
	}
}

// Report reports an error and updates metrics
func (er *ErrorReporter) Report(ctx context.Context, err *WorkflowError) {
	// Update metrics
	er.metrics.RecordError(err)
	
	// Log the error
	er.logError(err)
	
	// Notify subscribers
	er.notifySubscribers(ctx, err)
}

// ReportRecovery reports a successful error recovery
func (er *ErrorReporter) ReportRecovery(ctx context.Context, err *WorkflowError, strategy string) {
	er.metrics.RecordRecovery()
	
	er.logger.Info("Error recovery successful",
		"error_code", err.Code,
		"category", err.CategoryString(),
		"strategy", strategy,
		"recovery_time", time.Since(err.Timestamp))
}

// ReportRecoveryFailure reports a failed error recovery
func (er *ErrorReporter) ReportRecoveryFailure(ctx context.Context, err *WorkflowError) {
	er.metrics.RecordRecoveryFailure()
	
	er.logger.Error("Error recovery failed",
		"error_code", err.Code,
		"category", err.CategoryString(),
		"severity", err.SeverityString())
}

// GetMetrics returns current error metrics
func (er *ErrorReporter) GetMetrics() *ErrorMetrics {
	return er.metrics
}

// GetMetricsSnapshot returns a snapshot of current metrics
func (er *ErrorReporter) GetMetricsSnapshot() ErrorMetrics {
	er.metrics.mu.RLock()
	defer er.metrics.mu.RUnlock()
	
	snapshot := ErrorMetrics{
		TotalErrors:       er.metrics.TotalErrors,
		ErrorsByCategory:  make(map[ErrorCategory]int64),
		ErrorsBySeverity:  make(map[ErrorSeverity]int64),
		ErrorsByCode:      make(map[string]int64),
		RecoveredErrors:   er.metrics.RecoveredErrors,
		UnrecoveredErrors: er.metrics.UnrecoveredErrors,
		LastError:         er.metrics.LastError,
		LastErrorTime:     er.metrics.LastErrorTime,
	}
	
	// Deep copy maps
	for k, v := range er.metrics.ErrorsByCategory {
		snapshot.ErrorsByCategory[k] = v
	}
	for k, v := range er.metrics.ErrorsBySeverity {
		snapshot.ErrorsBySeverity[k] = v
	}
	for k, v := range er.metrics.ErrorsByCode {
		snapshot.ErrorsByCode[k] = v
	}
	
	return snapshot
}

// logError logs the error with appropriate level
func (er *ErrorReporter) logError(err *WorkflowError) {
	attrs := []slog.Attr{
		slog.String("error_code", err.Code),
		slog.String("category", err.CategoryString()),
		slog.String("severity", err.SeverityString()),
		slog.Bool("recoverable", err.Recoverable),
		slog.Time("timestamp", err.Timestamp),
	}
	
	// Add context attributes
	for key, value := range err.Context {
		attrs = append(attrs, slog.Any(key, value))
	}
	
	// Log with appropriate level based on severity
	switch err.Severity {
	case SeverityLow:
		er.logger.LogAttrs(context.Background(), slog.LevelInfo, err.Message, attrs...)
	case SeverityMedium:
		er.logger.LogAttrs(context.Background(), slog.LevelWarn, err.Message, attrs...)
	case SeverityHigh, SeverityCritical:
		er.logger.LogAttrs(context.Background(), slog.LevelError, err.Message, attrs...)
	}
}

// notifySubscribers notifies all error subscribers
func (er *ErrorReporter) notifySubscribers(ctx context.Context, err *WorkflowError) {
	er.mu.RLock()
	subscribers := make([]ErrorSubscriber, len(er.subscribers))
	copy(subscribers, er.subscribers)
	er.mu.RUnlock()
	
	for _, subscriber := range subscribers {
		go func(sub ErrorSubscriber) {
			defer func() {
				if r := recover(); r != nil {
					er.logger.Error("Error subscriber panicked",
						"subscriber", sub.GetName(),
						"panic", r)
				}
			}()
			sub.OnError(ctx, err)
		}(subscriber)
	}
}

// ErrorMetrics methods

// RecordError records an error in the metrics
func (em *ErrorMetrics) RecordError(err *WorkflowError) {
	em.mu.Lock()
	defer em.mu.Unlock()
	
	em.TotalErrors++
	em.ErrorsByCategory[err.Category]++
	em.ErrorsBySeverity[err.Severity]++
	em.ErrorsByCode[err.Code]++
	em.LastError = err
	em.LastErrorTime = err.Timestamp
}

// RecordRecovery records a successful error recovery
func (em *ErrorMetrics) RecordRecovery() {
	em.mu.Lock()
	defer em.mu.Unlock()
	
	em.RecoveredErrors++
}

// RecordRecoveryFailure records a failed error recovery
func (em *ErrorMetrics) RecordRecoveryFailure() {
	em.mu.Lock()
	defer em.mu.Unlock()
	
	em.UnrecoveredErrors++
}

// GetErrorRate returns the error rate for a specific category
func (em *ErrorMetrics) GetErrorRate(category ErrorCategory) float64 {
	em.mu.RLock()
	defer em.mu.RUnlock()
	
	if em.TotalErrors == 0 {
		return 0
	}
	
	return float64(em.ErrorsByCategory[category]) / float64(em.TotalErrors)
}

// GetRecoveryRate returns the overall recovery rate
func (em *ErrorMetrics) GetRecoveryRate() float64 {
	em.mu.RLock()
	defer em.mu.RUnlock()
	
	totalRecoveryAttempts := em.RecoveredErrors + em.UnrecoveredErrors
	if totalRecoveryAttempts == 0 {
		return 0
	}
	
	return float64(em.RecoveredErrors) / float64(totalRecoveryAttempts)
}

// ToJSON returns the metrics as JSON
func (em *ErrorMetrics) ToJSON() ([]byte, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()
	
	return json.MarshalIndent(em, "", "  ")
}

// Built-in error subscribers

// FileErrorSubscriber writes errors to a file
type FileErrorSubscriber struct {
	filename string
	logger   *slog.Logger
}

// NewFileErrorSubscriber creates a new file error subscriber
func NewFileErrorSubscriber(filename string, logger *slog.Logger) *FileErrorSubscriber {
	return &FileErrorSubscriber{
		filename: filename,
		logger:   logger,
	}
}

// OnError handles error events by writing to file
func (fes *FileErrorSubscriber) OnError(ctx context.Context, err *WorkflowError) {
	// This is a simplified implementation
	// In a real implementation, you'd want proper file handling with rotation
	errorData, jsonErr := json.MarshalIndent(err, "", "  ")
	if jsonErr != nil {
		fes.logger.Error("Failed to marshal error for file logging", "error", jsonErr)
		return
	}
	
	// Log to structured logger instead of direct file writing for now
	fes.logger.Info("Error logged to file subscriber", 
		"filename", fes.filename,
		"error_data", string(errorData))
}

// GetName returns the subscriber name
func (fes *FileErrorSubscriber) GetName() string {
	return fmt.Sprintf("FileErrorSubscriber(%s)", fes.filename)
}

// MetricsErrorSubscriber updates metrics dashboards
type MetricsErrorSubscriber struct {
	metricsEndpoint string
	logger          *slog.Logger
}

// NewMetricsErrorSubscriber creates a new metrics error subscriber
func NewMetricsErrorSubscriber(endpoint string, logger *slog.Logger) *MetricsErrorSubscriber {
	return &MetricsErrorSubscriber{
		metricsEndpoint: endpoint,
		logger:          logger,
	}
}

// OnError handles error events by sending to metrics endpoint
func (mes *MetricsErrorSubscriber) OnError(ctx context.Context, err *WorkflowError) {
	// This is a placeholder implementation
	// In a real implementation, you'd send metrics to your monitoring system
	mes.logger.Debug("Error sent to metrics endpoint",
		"endpoint", mes.metricsEndpoint,
		"error_code", err.Code,
		"category", err.CategoryString())
}

// GetName returns the subscriber name
func (mes *MetricsErrorSubscriber) GetName() string {
	return fmt.Sprintf("MetricsErrorSubscriber(%s)", mes.metricsEndpoint)
}