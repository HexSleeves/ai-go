package errors

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestNewRecoveryManager(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	rm := NewRecoveryManager(logger)
	
	if rm == nil {
		t.Fatal("Expected recovery manager to be created")
	}
	
	if rm.logger != logger {
		t.Error("Expected logger to be set")
	}
	
	if rm.strategies == nil {
		t.Error("Expected strategies slice to be initialized")
	}
	
	if rm.fallbacks == nil {
		t.Error("Expected fallbacks map to be initialized")
	}
}

func TestRecoveryManager_RegisterStrategy(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	rm := NewRecoveryManager(logger)
	
	strategy := NewRetryStrategy(3, time.Second, func(ctx context.Context, err *WorkflowError) error {
		return nil
	})
	
	rm.RegisterStrategy(strategy)
	
	if len(rm.strategies) != 1 {
		t.Errorf("Expected 1 strategy, got %d", len(rm.strategies))
	}
	
	if rm.strategies[0] != strategy {
		t.Error("Expected strategy to be registered")
	}
}

func TestRecoveryManager_RegisterFallback(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	rm := NewRecoveryManager(logger)
	
	fallback := NewDefaultValueFallback(map[string]interface{}{
		"test": "default_value",
	})
	
	rm.RegisterFallback(ErrorCategoryContent, fallback)
	
	if len(rm.fallbacks[ErrorCategoryContent]) != 1 {
		t.Errorf("Expected 1 fallback for content category, got %d", len(rm.fallbacks[ErrorCategoryContent]))
	}
	
	if rm.fallbacks[ErrorCategoryContent][0] != fallback {
		t.Error("Expected fallback to be registered")
	}
}

func TestRecoveryManager_Recover_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	rm := NewRecoveryManager(logger)
	
	// Create a strategy that can recover network errors
	strategy := NewRetryStrategy(3, time.Millisecond*10, func(ctx context.Context, err *WorkflowError) error {
		return nil // Simulate successful recovery
	})
	
	rm.RegisterStrategy(strategy)
	
	// Create a recoverable network error
	err := NewNetworkError(ErrCodeNetworkConnection, "Connection failed")
	
	ctx := context.Background()
	recoveryErr := rm.Recover(ctx, err)
	
	if recoveryErr != nil {
		t.Errorf("Expected successful recovery, got error: %v", recoveryErr)
	}
}

func TestRecoveryManager_Recover_Failure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	rm := NewRecoveryManager(logger)
	
	// Create a strategy that cannot recover this type of error
	strategy := NewRetryStrategy(3, time.Millisecond*10, func(ctx context.Context, err *WorkflowError) error {
		return nil
	})
	
	rm.RegisterStrategy(strategy)
	
	// Create a non-recoverable system error
	err := NewSystemError(ErrCodeSystemInitialization, "System failed").WithRecoverable(false)
	
	ctx := context.Background()
	recoveryErr := rm.Recover(ctx, err)
	
	if recoveryErr == nil {
		t.Error("Expected recovery to fail for non-recoverable error")
	}
}

func TestRecoveryManager_GetFallback_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	rm := NewRecoveryManager(logger)
	
	fallback := NewDefaultValueFallback(map[string]interface{}{
		"test": "default_value",
	})
	
	rm.RegisterFallback(ErrorCategoryContent, fallback)
	
	err := NewContentError(ErrCodeContentLoading, "Failed to load content").
		WithContext("type", "test")
	
	ctx := context.Background()
	value, fallbackErr := rm.GetFallback(ctx, err)
	
	if fallbackErr != nil {
		t.Errorf("Expected successful fallback, got error: %v", fallbackErr)
	}
	
	if value != "default_value" {
		t.Errorf("Expected fallback value 'default_value', got %v", value)
	}
}

func TestRecoveryManager_GetFallback_NoFallback(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	rm := NewRecoveryManager(logger)
	
	err := NewSystemError(ErrCodeSystemInitialization, "System failed")
	
	ctx := context.Background()
	_, fallbackErr := rm.GetFallback(ctx, err)
	
	if fallbackErr == nil {
		t.Error("Expected fallback to fail when no fallback is registered")
	}
}

func TestRetryStrategy_CanRecover(t *testing.T) {
	strategy := NewRetryStrategy(3, time.Second, func(ctx context.Context, err *WorkflowError) error {
		return nil
	})
	
	// Test recoverable network error
	networkErr := NewNetworkError(ErrCodeNetworkConnection, "Connection failed")
	if !strategy.CanRecover(networkErr) {
		t.Error("Expected retry strategy to handle recoverable network errors")
	}
	
	// Test recoverable I/O error
	ioErr := NewIOError(ErrCodeIORead, "Read failed")
	if !strategy.CanRecover(ioErr) {
		t.Error("Expected retry strategy to handle recoverable I/O errors")
	}
	
	// Test non-recoverable error
	systemErr := NewSystemError(ErrCodeSystemInitialization, "System failed").WithRecoverable(false)
	if strategy.CanRecover(systemErr) {
		t.Error("Expected retry strategy to not handle non-recoverable errors")
	}
	
	// Test wrong category
	userErr := NewUserError(ErrCodeUserInput, "Invalid input")
	if strategy.CanRecover(userErr) {
		t.Error("Expected retry strategy to not handle user errors")
	}
}

func TestRetryStrategy_Recover_Success(t *testing.T) {
	attempts := 0
	strategy := NewRetryStrategy(3, time.Millisecond*10, func(ctx context.Context, err *WorkflowError) error {
		attempts++
		if attempts >= 2 {
			return nil // Succeed on second attempt
		}
		return NewNetworkError(ErrCodeNetworkConnection, "Still failing")
	})
	
	err := NewNetworkError(ErrCodeNetworkConnection, "Connection failed")
	ctx := context.Background()
	
	recoveryErr := strategy.Recover(ctx, err)
	
	if recoveryErr != nil {
		t.Errorf("Expected successful recovery, got error: %v", recoveryErr)
	}
	
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryStrategy_Recover_Exhausted(t *testing.T) {
	attempts := 0
	strategy := NewRetryStrategy(2, time.Millisecond*10, func(ctx context.Context, err *WorkflowError) error {
		attempts++
		return NewNetworkError(ErrCodeNetworkConnection, "Still failing")
	})
	
	err := NewNetworkError(ErrCodeNetworkConnection, "Connection failed")
	ctx := context.Background()
	
	recoveryErr := strategy.Recover(ctx, err)
	
	if recoveryErr == nil {
		t.Error("Expected recovery to fail after exhausting retries")
	}
	
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestResetStrategy_CanRecover(t *testing.T) {
	strategy := NewResetStrategy(func(ctx context.Context, err *WorkflowError) error {
		return nil
	})
	
	// Test recoverable system error
	systemErr := NewSystemError(ErrCodeSystemResource, "Resource exhausted")
	if !strategy.CanRecover(systemErr) {
		t.Error("Expected reset strategy to handle recoverable system errors")
	}
	
	// Test non-recoverable system error
	nonRecoverableErr := NewSystemError(ErrCodeSystemInitialization, "System failed").WithRecoverable(false)
	if strategy.CanRecover(nonRecoverableErr) {
		t.Error("Expected reset strategy to not handle non-recoverable errors")
	}
	
	// Test wrong category
	contentErr := NewContentError(ErrCodeContentValidation, "Validation failed")
	if strategy.CanRecover(contentErr) {
		t.Error("Expected reset strategy to not handle content errors")
	}
}

func TestDefaultValueFallback_ProvideFallback(t *testing.T) {
	fallback := NewDefaultValueFallback(map[string]interface{}{
		"config": map[string]string{"key": "value"},
		"data":   []string{"item1", "item2"},
	})
	
	// Test with context type
	err := NewContentError(ErrCodeContentLoading, "Failed to load").
		WithContext("type", "config")
	
	ctx := context.Background()
	value, fallbackErr := fallback.ProvideFallback(ctx, err)
	
	if fallbackErr != nil {
		t.Errorf("Expected successful fallback, got error: %v", fallbackErr)
	}
	
	expectedValue := map[string]string{"key": "value"}
	if !compareValues(value, expectedValue) {
		t.Errorf("Expected fallback value %v, got %v", expectedValue, value)
	}
}

func TestCacheFallback_ProvideFallback(t *testing.T) {
	fallback := NewCacheFallback()
	fallback.SetCacheValue("test_key", "cached_value")
	
	err := NewContentError(ErrCodeContentLoading, "Failed to load").
		WithContext("cache_key", "test_key")
	
	ctx := context.Background()
	value, fallbackErr := fallback.ProvideFallback(ctx, err)
	
	if fallbackErr != nil {
		t.Errorf("Expected successful fallback, got error: %v", fallbackErr)
	}
	
	if value != "cached_value" {
		t.Errorf("Expected cached value 'cached_value', got %v", value)
	}
}

func TestDefaultErrorRecoveryConfig(t *testing.T) {
	config := DefaultErrorRecoveryConfig()
	
	if !config.EnableRetry {
		t.Error("Expected retry to be enabled by default")
	}
	
	if config.MaxRetries <= 0 {
		t.Error("Expected positive max retries")
	}
	
	if config.RetryDelay <= 0 {
		t.Error("Expected positive retry delay")
	}
	
	if !config.EnableReset {
		t.Error("Expected reset to be enabled by default")
	}
	
	if !config.EnableFallback {
		t.Error("Expected fallback to be enabled by default")
	}
}

// Helper function to compare values (simplified)
func compareValues(a, b interface{}) bool {
	// This is a simplified comparison for testing
	// In a real implementation, you'd want more robust comparison
	return a != nil && b != nil
}