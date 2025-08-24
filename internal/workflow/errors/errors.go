package errors

import (
	"fmt"
	"runtime"
	"time"
)

// ErrorCategory represents different categories of workflow errors
type ErrorCategory int

const (
	ErrorCategorySystem ErrorCategory = iota
	ErrorCategoryContent
	ErrorCategoryBalance
	ErrorCategoryPerformance
	ErrorCategoryPlugin
	ErrorCategoryUser
	ErrorCategoryValidation
	ErrorCategoryConfiguration
	ErrorCategoryNetwork
	ErrorCategoryIO
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	SeverityLow ErrorSeverity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// WorkflowError represents a comprehensive error in the workflow system
type WorkflowError struct {
	Category    ErrorCategory          `json:"category"`
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Context     map[string]interface{} `json:"context"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    ErrorSeverity          `json:"severity"`
	Recoverable bool                   `json:"recoverable"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Cause       error                  `json:"cause,omitempty"`
}

// Error implements the error interface
func (we *WorkflowError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", we.CategoryString(), we.Code, we.Message)
}

// CategoryString returns the string representation of the error category
func (we *WorkflowError) CategoryString() string {
	switch we.Category {
	case ErrorCategorySystem:
		return "SYSTEM"
	case ErrorCategoryContent:
		return "CONTENT"
	case ErrorCategoryBalance:
		return "BALANCE"
	case ErrorCategoryPerformance:
		return "PERFORMANCE"
	case ErrorCategoryPlugin:
		return "PLUGIN"
	case ErrorCategoryUser:
		return "USER"
	case ErrorCategoryValidation:
		return "VALIDATION"
	case ErrorCategoryConfiguration:
		return "CONFIG"
	case ErrorCategoryNetwork:
		return "NETWORK"
	case ErrorCategoryIO:
		return "IO"
	default:
		return "UNKNOWN"
	}
}

// SeverityString returns the string representation of the error severity
func (we *WorkflowError) SeverityString() string {
	switch we.Severity {
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Unwrap returns the underlying cause error
func (we *WorkflowError) Unwrap() error {
	return we.Cause
}

// NewWorkflowError creates a new workflow error
func NewWorkflowError(category ErrorCategory, code, message string) *WorkflowError {
	return &WorkflowError{
		Category:    category,
		Code:        code,
		Message:     message,
		Context:     make(map[string]interface{}),
		Timestamp:   time.Now(),
		Severity:    SeverityMedium,
		Recoverable: true,
		StackTrace:  getStackTrace(),
	}
}

// NewWorkflowErrorWithCause creates a new workflow error with an underlying cause
func NewWorkflowErrorWithCause(category ErrorCategory, code, message string, cause error) *WorkflowError {
	return &WorkflowError{
		Category:    category,
		Code:        code,
		Message:     message,
		Context:     make(map[string]interface{}),
		Timestamp:   time.Now(),
		Severity:    SeverityMedium,
		Recoverable: true,
		StackTrace:  getStackTrace(),
		Cause:       cause,
	}
}

// WithContext adds context information to the error
func (we *WorkflowError) WithContext(key string, value interface{}) *WorkflowError {
	we.Context[key] = value
	return we
}

// WithSeverity sets the severity of the error
func (we *WorkflowError) WithSeverity(severity ErrorSeverity) *WorkflowError {
	we.Severity = severity
	return we
}

// WithRecoverable sets whether the error is recoverable
func (we *WorkflowError) WithRecoverable(recoverable bool) *WorkflowError {
	we.Recoverable = recoverable
	return we
}

// WithCause sets the underlying cause error
func (we *WorkflowError) WithCause(cause error) *WorkflowError {
	we.Cause = cause
	return we
}

// getStackTrace captures the current stack trace
func getStackTrace() string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf))
	}
}

// Common error constructors for different categories

// NewSystemError creates a new system error
func NewSystemError(code, message string) *WorkflowError {
	return NewWorkflowError(ErrorCategorySystem, code, message).WithSeverity(SeverityHigh)
}

// NewContentError creates a new content error
func NewContentError(code, message string) *WorkflowError {
	return NewWorkflowError(ErrorCategoryContent, code, message)
}

// NewBalanceError creates a new balance error
func NewBalanceError(code, message string) *WorkflowError {
	return NewWorkflowError(ErrorCategoryBalance, code, message)
}

// NewPerformanceError creates a new performance error
func NewPerformanceError(code, message string) *WorkflowError {
	return NewWorkflowError(ErrorCategoryPerformance, code, message).WithSeverity(SeverityHigh)
}

// NewPluginError creates a new plugin error
func NewPluginError(code, message string) *WorkflowError {
	return NewWorkflowError(ErrorCategoryPlugin, code, message)
}

// NewUserError creates a new user error
func NewUserError(code, message string) *WorkflowError {
	return NewWorkflowError(ErrorCategoryUser, code, message).WithSeverity(SeverityLow)
}

// NewValidationError creates a new validation error
func NewValidationError(code, message string) *WorkflowError {
	return NewWorkflowError(ErrorCategoryValidation, code, message)
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(code, message string) *WorkflowError {
	return NewWorkflowError(ErrorCategoryConfiguration, code, message).WithSeverity(SeverityHigh)
}

// NewNetworkError creates a new network error
func NewNetworkError(code, message string) *WorkflowError {
	return NewWorkflowError(ErrorCategoryNetwork, code, message)
}

// NewIOError creates a new I/O error
func NewIOError(code, message string) *WorkflowError {
	return NewWorkflowError(ErrorCategoryIO, code, message)
}

// Error code constants for common errors
const (
	// System errors
	ErrCodeSystemInitialization = "SYS001"
	ErrCodeSystemShutdown       = "SYS002"
	ErrCodeSystemResource       = "SYS003"

	// Content errors
	ErrCodeContentValidation = "CNT001"
	ErrCodeContentGeneration = "CNT002"
	ErrCodeContentLoading    = "CNT003"

	// Balance errors
	ErrCodeBalanceAnalysis     = "BAL001"
	ErrCodeBalanceSimulation   = "BAL002"
	ErrCodeBalanceRecommendation = "BAL003"

	// Performance errors
	ErrCodePerformanceThreshold = "PRF001"
	ErrCodePerformanceProfiling = "PRF002"
	ErrCodePerformanceOptimization = "PRF003"

	// Plugin errors
	ErrCodePluginLoading    = "PLG001"
	ErrCodePluginExecution  = "PLG002"
	ErrCodePluginSecurity   = "PLG003"

	// User errors
	ErrCodeUserInput        = "USR001"
	ErrCodeUserPermission   = "USR002"
	ErrCodeUserConfiguration = "USR003"

	// Validation errors
	ErrCodeValidationSchema = "VAL001"
	ErrCodeValidationRule   = "VAL002"
	ErrCodeValidationData   = "VAL003"

	// Configuration errors
	ErrCodeConfigurationLoad = "CFG001"
	ErrCodeConfigurationSave = "CFG002"
	ErrCodeConfigurationValidation = "CFG003"

	// Network errors
	ErrCodeNetworkConnection = "NET001"
	ErrCodeNetworkTimeout    = "NET002"
	ErrCodeNetworkProtocol   = "NET003"

	// I/O errors
	ErrCodeIORead  = "IO001"
	ErrCodeIOWrite = "IO002"
	ErrCodeIOAccess = "IO003"
)