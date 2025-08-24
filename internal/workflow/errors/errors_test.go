package errors

import (
	"testing"
	"time"
)

func TestNewWorkflowError(t *testing.T) {
	err := NewWorkflowError(ErrorCategorySystem, ErrCodeSystemInitialization, "System failed to initialize")
	
	if err.Category != ErrorCategorySystem {
		t.Errorf("Expected category %v, got %v", ErrorCategorySystem, err.Category)
	}
	
	if err.Code != ErrCodeSystemInitialization {
		t.Errorf("Expected code %s, got %s", ErrCodeSystemInitialization, err.Code)
	}
	
	if err.Message != "System failed to initialize" {
		t.Errorf("Expected message 'System failed to initialize', got '%s'", err.Message)
	}
	
	if err.Severity != SeverityMedium {
		t.Errorf("Expected default severity %v, got %v", SeverityMedium, err.Severity)
	}
	
	if !err.Recoverable {
		t.Error("Expected error to be recoverable by default")
	}
	
	if err.Context == nil {
		t.Error("Expected context to be initialized")
	}
	
	if err.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestWorkflowError_Error(t *testing.T) {
	err := NewWorkflowError(ErrorCategoryContent, ErrCodeContentValidation, "Content validation failed")
	
	expected := "[CONTENT] CNT001: Content validation failed"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestWorkflowError_CategoryString(t *testing.T) {
	testCases := []struct {
		category ErrorCategory
		expected string
	}{
		{ErrorCategorySystem, "SYSTEM"},
		{ErrorCategoryContent, "CONTENT"},
		{ErrorCategoryBalance, "BALANCE"},
		{ErrorCategoryPerformance, "PERFORMANCE"},
		{ErrorCategoryPlugin, "PLUGIN"},
		{ErrorCategoryUser, "USER"},
		{ErrorCategoryValidation, "VALIDATION"},
		{ErrorCategoryConfiguration, "CONFIG"},
		{ErrorCategoryNetwork, "NETWORK"},
		{ErrorCategoryIO, "IO"},
	}
	
	for _, tc := range testCases {
		err := NewWorkflowError(tc.category, "TEST001", "Test message")
		if err.CategoryString() != tc.expected {
			t.Errorf("Expected category string '%s', got '%s'", tc.expected, err.CategoryString())
		}
	}
}

func TestWorkflowError_SeverityString(t *testing.T) {
	testCases := []struct {
		severity ErrorSeverity
		expected string
	}{
		{SeverityLow, "LOW"},
		{SeverityMedium, "MEDIUM"},
		{SeverityHigh, "HIGH"},
		{SeverityCritical, "CRITICAL"},
	}
	
	for _, tc := range testCases {
		err := NewWorkflowError(ErrorCategorySystem, "TEST001", "Test message").WithSeverity(tc.severity)
		if err.SeverityString() != tc.expected {
			t.Errorf("Expected severity string '%s', got '%s'", tc.expected, err.SeverityString())
		}
	}
}

func TestWorkflowError_WithContext(t *testing.T) {
	err := NewWorkflowError(ErrorCategorySystem, "TEST001", "Test message")
	err.WithContext("key1", "value1").WithContext("key2", 42)
	
	if err.Context["key1"] != "value1" {
		t.Errorf("Expected context key1 to be 'value1', got %v", err.Context["key1"])
	}
	
	if err.Context["key2"] != 42 {
		t.Errorf("Expected context key2 to be 42, got %v", err.Context["key2"])
	}
}

func TestWorkflowError_WithSeverity(t *testing.T) {
	err := NewWorkflowError(ErrorCategorySystem, "TEST001", "Test message")
	err.WithSeverity(SeverityHigh)
	
	if err.Severity != SeverityHigh {
		t.Errorf("Expected severity %v, got %v", SeverityHigh, err.Severity)
	}
}

func TestWorkflowError_WithRecoverable(t *testing.T) {
	err := NewWorkflowError(ErrorCategorySystem, "TEST001", "Test message")
	err.WithRecoverable(false)
	
	if err.Recoverable {
		t.Error("Expected error to be non-recoverable")
	}
}

func TestNewWorkflowErrorWithCause(t *testing.T) {
	cause := NewWorkflowError(ErrorCategoryIO, ErrCodeIORead, "Failed to read file")
	err := NewWorkflowErrorWithCause(ErrorCategorySystem, ErrCodeSystemInitialization, "System initialization failed", cause)
	
	if err.Cause != cause {
		t.Error("Expected cause to be set")
	}
	
	if err.Unwrap() != cause {
		t.Error("Expected Unwrap to return the cause")
	}
}

func TestErrorConstructors(t *testing.T) {
	testCases := []struct {
		constructor func(string, string) *WorkflowError
		category    ErrorCategory
		severity    ErrorSeverity
	}{
		{NewSystemError, ErrorCategorySystem, SeverityHigh},
		{NewContentError, ErrorCategoryContent, SeverityMedium},
		{NewBalanceError, ErrorCategoryBalance, SeverityMedium},
		{NewPerformanceError, ErrorCategoryPerformance, SeverityHigh},
		{NewPluginError, ErrorCategoryPlugin, SeverityMedium},
		{NewUserError, ErrorCategoryUser, SeverityLow},
		{NewValidationError, ErrorCategoryValidation, SeverityMedium},
		{NewConfigurationError, ErrorCategoryConfiguration, SeverityHigh},
		{NewNetworkError, ErrorCategoryNetwork, SeverityMedium},
		{NewIOError, ErrorCategoryIO, SeverityMedium},
	}
	
	for _, tc := range testCases {
		err := tc.constructor("TEST001", "Test message")
		
		if err.Category != tc.category {
			t.Errorf("Expected category %v, got %v", tc.category, err.Category)
		}
		
		if err.Severity != tc.severity {
			t.Errorf("Expected severity %v, got %v", tc.severity, err.Severity)
		}
		
		if err.Code != "TEST001" {
			t.Errorf("Expected code 'TEST001', got '%s'", err.Code)
		}
		
		if err.Message != "Test message" {
			t.Errorf("Expected message 'Test message', got '%s'", err.Message)
		}
	}
}

func TestWorkflowError_StackTrace(t *testing.T) {
	err := NewWorkflowError(ErrorCategorySystem, "TEST001", "Test message")
	
	if err.StackTrace == "" {
		t.Error("Expected stack trace to be captured")
	}
}

func TestWorkflowError_Timestamp(t *testing.T) {
	before := time.Now()
	err := NewWorkflowError(ErrorCategorySystem, "TEST001", "Test message")
	after := time.Now()
	
	if err.Timestamp.Before(before) || err.Timestamp.After(after) {
		t.Error("Expected timestamp to be set to current time")
	}
}