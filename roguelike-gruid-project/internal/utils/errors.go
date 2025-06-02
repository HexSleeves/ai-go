package utils

import (
	"fmt"
	"runtime"
)

// GameError represents a game-specific error with context
type GameError struct {
	Type    string
	Message string
	Cause   error
	Context map[string]interface{}
	Stack   string
}

// Error implements the error interface
func (e *GameError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *GameError) Unwrap() error {
	return e.Cause
}

// NewGameError creates a new game error with stack trace
func NewGameError(errorType, message string, cause error) *GameError {
	// Capture stack trace
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	
	return &GameError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
		Stack:   string(buf[:n]),
	}
}

// WithContext adds context information to the error
func (e *GameError) WithContext(key string, value interface{}) *GameError {
	e.Context[key] = value
	return e
}

// Common error types
const (
	ErrorTypeECS        = "ECS_ERROR"
	ErrorTypeGame       = "GAME_ERROR"
	ErrorTypeMap        = "MAP_ERROR"
	ErrorTypeAI         = "AI_ERROR"
	ErrorTypeInput      = "INPUT_ERROR"
	ErrorTypeSave       = "SAVE_ERROR"
	ErrorTypeLoad       = "LOAD_ERROR"
	ErrorTypeValidation = "VALIDATION_ERROR"
	ErrorTypeSystem     = "SYSTEM_ERROR"
)

// Predefined error constructors
func NewECSError(message string, cause error) *GameError {
	return NewGameError(ErrorTypeECS, message, cause)
}

func NewGameLogicError(message string, cause error) *GameError {
	return NewGameError(ErrorTypeGame, message, cause)
}

func NewMapError(message string, cause error) *GameError {
	return NewGameError(ErrorTypeMap, message, cause)
}

func NewAIError(message string, cause error) *GameError {
	return NewGameError(ErrorTypeAI, message, cause)
}

func NewInputError(message string, cause error) *GameError {
	return NewGameError(ErrorTypeInput, message, cause)
}

func NewSaveError(message string, cause error) *GameError {
	return NewGameError(ErrorTypeSave, message, cause)
}

func NewLoadError(message string, cause error) *GameError {
	return NewGameError(ErrorTypeLoad, message, cause)
}

func NewValidationError(message string, cause error) *GameError {
	return NewGameError(ErrorTypeValidation, message, cause)
}

func NewSystemError(message string, cause error) *GameError {
	return NewGameError(ErrorTypeSystem, message, cause)
}

// Result represents a result that can contain either a value or an error
type Result[T any] struct {
	Value T
	Error error
}

// NewResult creates a successful result
func NewResult[T any](value T) Result[T] {
	return Result[T]{Value: value, Error: nil}
}

// NewResultError creates a failed result
func NewResultError[T any](err error) Result[T] {
	var zero T
	return Result[T]{Value: zero, Error: err}
}

// IsOk returns true if the result contains a value
func (r Result[T]) IsOk() bool {
	return r.Error == nil
}

// IsErr returns true if the result contains an error
func (r Result[T]) IsErr() bool {
	return r.Error != nil
}

// Unwrap returns the value or panics if there's an error
func (r Result[T]) Unwrap() T {
	if r.Error != nil {
		panic(fmt.Sprintf("called Unwrap on error result: %v", r.Error))
	}
	return r.Value
}

// UnwrapOr returns the value or the provided default if there's an error
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.Error != nil {
		return defaultValue
	}
	return r.Value
}

// Map transforms the value if the result is ok
func (r Result[T]) Map(fn func(T) T) Result[T] {
	if r.Error != nil {
		return r
	}
	return NewResult(fn(r.Value))
}

// MapError transforms the error if the result is an error
func (r Result[T]) MapError(fn func(error) error) Result[T] {
	if r.Error == nil {
		return r
	}
	return NewResultError[T](fn(r.Error))
}

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	handlers map[string]func(*GameError)
}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		handlers: make(map[string]func(*GameError)),
	}
}

// RegisterHandler registers a handler for a specific error type
func (eh *ErrorHandler) RegisterHandler(errorType string, handler func(*GameError)) {
	eh.handlers[errorType] = handler
}

// Handle processes an error using the appropriate handler
func (eh *ErrorHandler) Handle(err error) {
	if gameErr, ok := err.(*GameError); ok {
		if handler, exists := eh.handlers[gameErr.Type]; exists {
			handler(gameErr)
			return
		}
	}
	
	// Default handler
	eh.defaultHandler(err)
}

// defaultHandler is the fallback error handler
func (eh *ErrorHandler) defaultHandler(err error) {
	fmt.Printf("Unhandled error: %v\n", err)
}

// Validation helpers
func ValidateNotNil(value interface{}, name string) error {
	if value == nil {
		return NewValidationError(fmt.Sprintf("%s cannot be nil", name), nil)
	}
	return nil
}

func ValidatePositive(value int, name string) error {
	if value <= 0 {
		return NewValidationError(fmt.Sprintf("%s must be positive, got %d", name, value), nil)
	}
	return nil
}

func ValidateRange(value, min, max int, name string) error {
	if value < min || value > max {
		return NewValidationError(
			fmt.Sprintf("%s must be between %d and %d, got %d", name, min, max, value), 
			nil,
		)
	}
	return nil
}

func ValidateNotEmpty(value string, name string) error {
	if value == "" {
		return NewValidationError(fmt.Sprintf("%s cannot be empty", name), nil)
	}
	return nil
}

// Panic recovery helper
func RecoverPanic() error {
	if r := recover(); r != nil {
		// Capture stack trace
		buf := make([]byte, 1024)
		n := runtime.Stack(buf, false)
		
		return NewSystemError(
			fmt.Sprintf("panic recovered: %v", r),
			nil,
		).WithContext("stack", string(buf[:n]))
	}
	return nil
}

// SafeExecute executes a function and recovers from panics
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if recovered := RecoverPanic(); recovered != nil {
			err = recovered
		}
	}()
	
	return fn()
}

// Retry executes a function with retry logic
func Retry(attempts int, fn func() error) error {
	var lastErr error
	
	for i := 0; i < attempts; i++ {
		if err := fn(); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	
	return NewSystemError(
		fmt.Sprintf("operation failed after %d attempts", attempts),
		lastErr,
	)
}
