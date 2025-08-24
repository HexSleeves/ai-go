package errors

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestNewErrorReporter(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	reporter := NewErrorReporter(logger)
	
	if reporter == nil {
		t.Fatal("Expected error reporter to be created")
	}
	
	if reporter.logger != logger {
		t.Error("Expected logger to be set")
	}
	
	if reporter.metrics == nil {
		t.Error("Expected metrics to be initialized")
	}
	
	if reporter.subscribers == nil {
		t.Error("Expected subscribers slice to be initialized")
	}
}

func TestErrorReporter_Subscribe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	reporter := NewErrorReporter(logger)
	
	subscriber := NewFileErrorSubscriber("test.log", logger)
	reporter.Subscribe(subscriber)
	
	if len(reporter.subscribers) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", len(reporter.subscribers))
	}
	
	if reporter.subscribers[0] != subscriber {
		t.Error("Expected subscriber to be registered")
	}
}

func TestErrorReporter_Unsubscribe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	reporter := NewErrorReporter(logger)
	
	subscriber := NewFileErrorSubscriber("test.log", logger)
	reporter.Subscribe(subscriber)
	
	if len(reporter.subscribers) != 1 {
		t.Error("Expected subscriber to be added")
	}
	
	reporter.Unsubscribe(subscriber.GetName())
	
	if len(reporter.subscribers) != 0 {
		t.Errorf("Expected 0 subscribers after unsubscribe, got %d", len(reporter.subscribers))
	}
}

func TestErrorReporter_Report(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	reporter := NewErrorReporter(logger)
	
	// Track if subscriber was called
	subscriberCalled := false
	mockSubscriber := &MockErrorSubscriber{
		name: "mock",
		onError: func(ctx context.Context, err *WorkflowError) {
			subscriberCalled = true
		},
	}
	
	reporter.Subscribe(mockSubscriber)
	
	err := NewSystemError(ErrCodeSystemInitialization, "System failed")
	ctx := context.Background()
	
	reporter.Report(ctx, err)
	
	// Give some time for goroutine to execute
	time.Sleep(time.Millisecond * 10)
	
	if !subscriberCalled {
		t.Error("Expected subscriber to be called")
	}
	
	// Check metrics
	metrics := reporter.GetMetrics()
	if metrics.TotalErrors != 1 {
		t.Errorf("Expected 1 total error, got %d", metrics.TotalErrors)
	}
	
	if metrics.ErrorsByCategory[ErrorCategorySystem] != 1 {
		t.Errorf("Expected 1 system error, got %d", metrics.ErrorsByCategory[ErrorCategorySystem])
	}
	
	if metrics.ErrorsBySeverity[SeverityHigh] != 1 {
		t.Errorf("Expected 1 high severity error, got %d", metrics.ErrorsBySeverity[SeverityHigh])
	}
}

func TestErrorReporter_ReportRecovery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	reporter := NewErrorReporter(logger)
	
	err := NewNetworkError(ErrCodeNetworkConnection, "Connection failed")
	ctx := context.Background()
	
	reporter.ReportRecovery(ctx, err, "RetryStrategy")
	
	metrics := reporter.GetMetrics()
	if metrics.RecoveredErrors != 1 {
		t.Errorf("Expected 1 recovered error, got %d", metrics.RecoveredErrors)
	}
}

func TestErrorReporter_ReportRecoveryFailure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	reporter := NewErrorReporter(logger)
	
	err := NewSystemError(ErrCodeSystemInitialization, "System failed")
	ctx := context.Background()
	
	reporter.ReportRecoveryFailure(ctx, err)
	
	metrics := reporter.GetMetrics()
	if metrics.UnrecoveredErrors != 1 {
		t.Errorf("Expected 1 unrecovered error, got %d", metrics.UnrecoveredErrors)
	}
}

func TestErrorReporter_GetMetricsSnapshot(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	reporter := NewErrorReporter(logger)
	
	err1 := NewSystemError(ErrCodeSystemInitialization, "System failed")
	err2 := NewContentError(ErrCodeContentValidation, "Validation failed")
	
	ctx := context.Background()
	reporter.Report(ctx, err1)
	reporter.Report(ctx, err2)
	
	snapshot := reporter.GetMetricsSnapshot()
	
	if snapshot.TotalErrors != 2 {
		t.Errorf("Expected 2 total errors in snapshot, got %d", snapshot.TotalErrors)
	}
	
	if snapshot.ErrorsByCategory[ErrorCategorySystem] != 1 {
		t.Errorf("Expected 1 system error in snapshot, got %d", snapshot.ErrorsByCategory[ErrorCategorySystem])
	}
	
	if snapshot.ErrorsByCategory[ErrorCategoryContent] != 1 {
		t.Errorf("Expected 1 content error in snapshot, got %d", snapshot.ErrorsByCategory[ErrorCategoryContent])
	}
}

func TestErrorMetrics_RecordError(t *testing.T) {
	metrics := NewErrorMetrics()
	
	err := NewSystemError(ErrCodeSystemInitialization, "System failed")
	metrics.RecordError(err)
	
	if metrics.TotalErrors != 1 {
		t.Errorf("Expected 1 total error, got %d", metrics.TotalErrors)
	}
	
	if metrics.ErrorsByCategory[ErrorCategorySystem] != 1 {
		t.Errorf("Expected 1 system error, got %d", metrics.ErrorsByCategory[ErrorCategorySystem])
	}
	
	if metrics.ErrorsBySeverity[SeverityHigh] != 1 {
		t.Errorf("Expected 1 high severity error, got %d", metrics.ErrorsBySeverity[SeverityHigh])
	}
	
	if metrics.ErrorsByCode[ErrCodeSystemInitialization] != 1 {
		t.Errorf("Expected 1 error with code %s, got %d", ErrCodeSystemInitialization, metrics.ErrorsByCode[ErrCodeSystemInitialization])
	}
	
	if metrics.LastError != err {
		t.Error("Expected last error to be set")
	}
}

func TestErrorMetrics_GetErrorRate(t *testing.T) {
	metrics := NewErrorMetrics()
	
	// Record some errors
	systemErr := NewSystemError(ErrCodeSystemInitialization, "System failed")
	contentErr := NewContentError(ErrCodeContentValidation, "Validation failed")
	
	metrics.RecordError(systemErr)
	metrics.RecordError(contentErr)
	metrics.RecordError(contentErr) // Another content error
	
	systemRate := metrics.GetErrorRate(ErrorCategorySystem)
	expectedSystemRate := 1.0 / 3.0 // 1 out of 3 total errors
	
	if systemRate != expectedSystemRate {
		t.Errorf("Expected system error rate %f, got %f", expectedSystemRate, systemRate)
	}
	
	contentRate := metrics.GetErrorRate(ErrorCategoryContent)
	expectedContentRate := 2.0 / 3.0 // 2 out of 3 total errors
	
	if contentRate != expectedContentRate {
		t.Errorf("Expected content error rate %f, got %f", expectedContentRate, contentRate)
	}
}

func TestErrorMetrics_GetRecoveryRate(t *testing.T) {
	metrics := NewErrorMetrics()
	
	// No recovery attempts yet
	rate := metrics.GetRecoveryRate()
	if rate != 0 {
		t.Errorf("Expected recovery rate 0 with no attempts, got %f", rate)
	}
	
	// Record some recovery attempts
	metrics.RecordRecovery()
	metrics.RecordRecovery()
	metrics.RecordRecoveryFailure()
	
	rate = metrics.GetRecoveryRate()
	expectedRate := 2.0 / 3.0 // 2 successful out of 3 total attempts
	
	if rate != expectedRate {
		t.Errorf("Expected recovery rate %f, got %f", expectedRate, rate)
	}
}

func TestErrorMetrics_ToJSON(t *testing.T) {
	metrics := NewErrorMetrics()
	
	err := NewSystemError(ErrCodeSystemInitialization, "System failed")
	metrics.RecordError(err)
	
	jsonData, jsonErr := metrics.ToJSON()
	if jsonErr != nil {
		t.Errorf("Expected successful JSON marshaling, got error: %v", jsonErr)
	}
	
	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON data")
	}
}

func TestFileErrorSubscriber(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	subscriber := NewFileErrorSubscriber("test.log", logger)
	
	if subscriber.GetName() != "FileErrorSubscriber(test.log)" {
		t.Errorf("Expected name 'FileErrorSubscriber(test.log)', got '%s'", subscriber.GetName())
	}
	
	// Test OnError doesn't panic
	err := NewSystemError(ErrCodeSystemInitialization, "System failed")
	ctx := context.Background()
	
	// This should not panic
	subscriber.OnError(ctx, err)
}

func TestMetricsErrorSubscriber(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	subscriber := NewMetricsErrorSubscriber("http://localhost:8080/metrics", logger)
	
	expectedName := "MetricsErrorSubscriber(http://localhost:8080/metrics)"
	if subscriber.GetName() != expectedName {
		t.Errorf("Expected name '%s', got '%s'", expectedName, subscriber.GetName())
	}
	
	// Test OnError doesn't panic
	err := NewSystemError(ErrCodeSystemInitialization, "System failed")
	ctx := context.Background()
	
	// This should not panic
	subscriber.OnError(ctx, err)
}

// MockErrorSubscriber for testing
type MockErrorSubscriber struct {
	name    string
	onError func(ctx context.Context, err *WorkflowError)
}

func (m *MockErrorSubscriber) OnError(ctx context.Context, err *WorkflowError) {
	if m.onError != nil {
		m.onError(ctx, err)
	}
}

func (m *MockErrorSubscriber) GetName() string {
	return m.name
}