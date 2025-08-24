package errors

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// RecoveryStrategy defines how to recover from different types of errors
type RecoveryStrategy interface {
	CanRecover(err *WorkflowError) bool
	Recover(ctx context.Context, err *WorkflowError) error
	GetName() string
}

// FallbackProvider provides fallback mechanisms for different error scenarios
type FallbackProvider interface {
	ProvideFallback(ctx context.Context, err *WorkflowError) (interface{}, error)
	GetFallbackType() string
}

// RecoveryManager manages error recovery strategies and fallback mechanisms
type RecoveryManager struct {
	strategies []RecoveryStrategy
	fallbacks  map[ErrorCategory][]FallbackProvider
	logger     *slog.Logger
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(logger *slog.Logger) *RecoveryManager {
	return &RecoveryManager{
		strategies: []RecoveryStrategy{},
		fallbacks:  make(map[ErrorCategory][]FallbackProvider),
		logger:     logger,
	}
}

// RegisterStrategy registers a recovery strategy
func (rm *RecoveryManager) RegisterStrategy(strategy RecoveryStrategy) {
	rm.strategies = append(rm.strategies, strategy)
	rm.logger.Debug("Registered recovery strategy", "strategy", strategy.GetName())
}

// RegisterFallback registers a fallback provider for a specific error category
func (rm *RecoveryManager) RegisterFallback(category ErrorCategory, fallback FallbackProvider) {
	if rm.fallbacks[category] == nil {
		rm.fallbacks[category] = []FallbackProvider{}
	}
	rm.fallbacks[category] = append(rm.fallbacks[category], fallback)
	rm.logger.Debug("Registered fallback provider", 
		"category", category, 
		"type", fallback.GetFallbackType())
}

// Recover attempts to recover from an error using registered strategies
func (rm *RecoveryManager) Recover(ctx context.Context, err *WorkflowError) error {
	rm.logger.Info("Attempting error recovery", 
		"error_code", err.Code, 
		"category", err.CategoryString(),
		"severity", err.SeverityString())

	// Try each recovery strategy
	for _, strategy := range rm.strategies {
		if strategy.CanRecover(err) {
			rm.logger.Debug("Attempting recovery with strategy", "strategy", strategy.GetName())
			
			if recoveryErr := strategy.Recover(ctx, err); recoveryErr == nil {
				rm.logger.Info("Successfully recovered from error", 
					"error_code", err.Code, 
					"strategy", strategy.GetName())
				return nil
			} else {
				rm.logger.Warn("Recovery strategy failed", 
					"strategy", strategy.GetName(), 
					"recovery_error", recoveryErr)
			}
		}
	}

	rm.logger.Error("Failed to recover from error", "error_code", err.Code)
	return fmt.Errorf("no recovery strategy available for error: %w", err)
}

// GetFallback attempts to get a fallback value for an error
func (rm *RecoveryManager) GetFallback(ctx context.Context, err *WorkflowError) (interface{}, error) {
	fallbacks, exists := rm.fallbacks[err.Category]
	if !exists || len(fallbacks) == 0 {
		return nil, fmt.Errorf("no fallback available for category: %s", err.CategoryString())
	}

	// Try each fallback provider
	for _, fallback := range fallbacks {
		rm.logger.Debug("Attempting fallback", "type", fallback.GetFallbackType())
		
		if value, fallbackErr := fallback.ProvideFallback(ctx, err); fallbackErr == nil {
			rm.logger.Info("Successfully provided fallback", 
				"error_code", err.Code, 
				"fallback_type", fallback.GetFallbackType())
			return value, nil
		} else {
			rm.logger.Warn("Fallback provider failed", 
				"type", fallback.GetFallbackType(), 
				"fallback_error", fallbackErr)
		}
	}

	return nil, fmt.Errorf("all fallback providers failed for error: %w", err)
}

// Built-in recovery strategies

// RetryStrategy implements a simple retry mechanism
type RetryStrategy struct {
	maxRetries int
	delay      time.Duration
	retryFunc  func(ctx context.Context, err *WorkflowError) error
}

// NewRetryStrategy creates a new retry strategy
func NewRetryStrategy(maxRetries int, delay time.Duration, retryFunc func(ctx context.Context, err *WorkflowError) error) *RetryStrategy {
	return &RetryStrategy{
		maxRetries: maxRetries,
		delay:      delay,
		retryFunc:  retryFunc,
	}
}

// CanRecover checks if the error is recoverable
func (rs *RetryStrategy) CanRecover(err *WorkflowError) bool {
	return err.Recoverable && (err.Category == ErrorCategoryNetwork || err.Category == ErrorCategoryIO)
}

// Recover attempts to recover by retrying the operation
func (rs *RetryStrategy) Recover(ctx context.Context, err *WorkflowError) error {
	for i := 0; i < rs.maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(rs.delay):
			if retryErr := rs.retryFunc(ctx, err); retryErr == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("retry strategy exhausted after %d attempts", rs.maxRetries)
}

// GetName returns the strategy name
func (rs *RetryStrategy) GetName() string {
	return "RetryStrategy"
}

// ResetStrategy implements a reset mechanism for system errors
type ResetStrategy struct {
	resetFunc func(ctx context.Context, err *WorkflowError) error
}

// NewResetStrategy creates a new reset strategy
func NewResetStrategy(resetFunc func(ctx context.Context, err *WorkflowError) error) *ResetStrategy {
	return &ResetStrategy{
		resetFunc: resetFunc,
	}
}

// CanRecover checks if the error can be recovered by resetting
func (rs *ResetStrategy) CanRecover(err *WorkflowError) bool {
	return err.Category == ErrorCategorySystem && err.Recoverable
}

// Recover attempts to recover by resetting the system component
func (rs *ResetStrategy) Recover(ctx context.Context, err *WorkflowError) error {
	return rs.resetFunc(ctx, err)
}

// GetName returns the strategy name
func (rs *ResetStrategy) GetName() string {
	return "ResetStrategy"
}

// Built-in fallback providers

// DefaultValueFallback provides default values for different error scenarios
type DefaultValueFallback struct {
	defaultValues map[string]interface{}
}

// NewDefaultValueFallback creates a new default value fallback
func NewDefaultValueFallback(defaults map[string]interface{}) *DefaultValueFallback {
	return &DefaultValueFallback{
		defaultValues: defaults,
	}
}

// ProvideFallback provides a default value based on the error context
func (dvf *DefaultValueFallback) ProvideFallback(ctx context.Context, err *WorkflowError) (interface{}, error) {
	if contextType, exists := err.Context["type"]; exists {
		if defaultValue, hasDefault := dvf.defaultValues[contextType.(string)]; hasDefault {
			return defaultValue, nil
		}
	}
	
	// Provide generic defaults based on error code
	switch err.Code {
	case ErrCodeContentLoading:
		return map[string]interface{}{}, nil
	case ErrCodeConfigurationLoad:
		return struct{}{}, nil
	default:
		return nil, fmt.Errorf("no default value available for error code: %s", err.Code)
	}
}

// GetFallbackType returns the fallback type
func (dvf *DefaultValueFallback) GetFallbackType() string {
	return "DefaultValue"
}

// CacheFallback provides cached values when fresh data is unavailable
type CacheFallback struct {
	cache map[string]interface{}
}

// NewCacheFallback creates a new cache fallback
func NewCacheFallback() *CacheFallback {
	return &CacheFallback{
		cache: make(map[string]interface{}),
	}
}

// SetCacheValue sets a cached value
func (cf *CacheFallback) SetCacheValue(key string, value interface{}) {
	cf.cache[key] = value
}

// ProvideFallback provides a cached value if available
func (cf *CacheFallback) ProvideFallback(ctx context.Context, err *WorkflowError) (interface{}, error) {
	if cacheKey, exists := err.Context["cache_key"]; exists {
		if cachedValue, hasCached := cf.cache[cacheKey.(string)]; hasCached {
			return cachedValue, nil
		}
	}
	
	return nil, fmt.Errorf("no cached value available for error: %s", err.Code)
}

// GetFallbackType returns the fallback type
func (cf *CacheFallback) GetFallbackType() string {
	return "Cache"
}

// ErrorRecoveryConfig holds configuration for error recovery
type ErrorRecoveryConfig struct {
	EnableRetry       bool          `json:"enable_retry"`
	MaxRetries        int           `json:"max_retries"`
	RetryDelay        time.Duration `json:"retry_delay"`
	EnableReset       bool          `json:"enable_reset"`
	EnableFallback    bool          `json:"enable_fallback"`
	CacheEnabled      bool          `json:"cache_enabled"`
	LogRecoveryEvents bool          `json:"log_recovery_events"`
}

// DefaultErrorRecoveryConfig returns default error recovery configuration
func DefaultErrorRecoveryConfig() ErrorRecoveryConfig {
	return ErrorRecoveryConfig{
		EnableRetry:       true,
		MaxRetries:        3,
		RetryDelay:        time.Second * 2,
		EnableReset:       true,
		EnableFallback:    true,
		CacheEnabled:      true,
		LogRecoveryEvents: true,
	}
}