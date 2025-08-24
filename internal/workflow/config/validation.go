package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ValidationResult represents the result of a configuration validation
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// ConfigValidator provides comprehensive configuration validation
type ConfigValidator struct {
	config *WorkflowConfig
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator(config *WorkflowConfig) *ConfigValidator {
	return &ConfigValidator{
		config: config,
	}
}

// ValidateAll performs comprehensive validation of the configuration
func (cv *ConfigValidator) ValidateAll() *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Validate basic fields
	cv.validateBasicFields(result)

	// Validate spec configuration
	cv.validateSpecConfig(result)

	// Validate content configuration
	cv.validateContentConfig(result)

	// Validate balance configuration
	cv.validateBalanceConfig(result)

	// Validate testing configuration
	cv.validateTestingConfig(result)

	// Validate performance configuration
	cv.validatePerformanceConfig(result)

	// Validate plugin configuration
	cv.validatePluginConfig(result)

	// Validate documentation configuration
	cv.validateDocumentationConfig(result)

	// Validate CI/CD configuration
	cv.validateCICDConfig(result)

	// Validate debug configuration
	cv.validateDebugConfig(result)

	// Validate directory structure
	cv.validateDirectoryStructure(result)

	// Set overall validity
	result.Valid = len(result.Errors) == 0

	return result
}

// validateBasicFields validates basic configuration fields
func (cv *ConfigValidator) validateBasicFields(result *ValidationResult) {
	if cv.config.Version == "" {
		result.Errors = append(result.Errors, "version is required")
	} else if !cv.isValidVersion(cv.config.Version) {
		result.Errors = append(result.Errors, "version must follow semantic versioning (e.g., 1.0.0)")
	}

	if cv.config.ProjectName == "" {
		result.Errors = append(result.Errors, "project name is required")
	} else if !cv.isValidProjectName(cv.config.ProjectName) {
		result.Errors = append(result.Errors, "project name must contain only alphanumeric characters, hyphens, and underscores")
	}
}

// validateSpecConfig validates spec-related configuration
func (cv *ConfigValidator) validateSpecConfig(result *ValidationResult) {
	if cv.config.Specs.SpecDir == "" {
		result.Errors = append(result.Errors, "spec directory is required")
	} else if !cv.isValidPath(cv.config.Specs.SpecDir) {
		result.Errors = append(result.Errors, "spec directory path is invalid")
	}

	// Validate validation rules
	validRules := map[string]bool{
		"requirements": true,
		"design":       true,
		"tasks":        true,
		"acceptance":   true,
	}

	for _, rule := range cv.config.Specs.ValidationRules {
		if !validRules[rule] {
			result.Warnings = append(result.Warnings, fmt.Sprintf("unknown validation rule: %s", rule))
		}
	}
}

// validateContentConfig validates content pipeline configuration
func (cv *ConfigValidator) validateContentConfig(result *ValidationResult) {
	if cv.config.Content.ContentDir == "" {
		result.Errors = append(result.Errors, "content directory is required")
	}

	if cv.config.Content.OutputDir == "" {
		result.Errors = append(result.Errors, "content output directory is required")
	}

	// Validate Luban configuration if enabled
	if cv.config.Content.LubanEnabled {
		if cv.config.Content.LubanServerURL == "" {
			result.Errors = append(result.Errors, "Luban server URL is required when Luban is enabled")
		} else if !cv.isValidURL(cv.config.Content.LubanServerURL) {
			result.Errors = append(result.Errors, "Luban server URL is invalid")
		}
	}

	// Validate validation level
	validLevels := map[string]bool{"strict": true, "normal": true, "lenient": true}
	if !validLevels[cv.config.Content.ValidationLevel] {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid content validation level: %s", cv.config.Content.ValidationLevel))
	}
}

// validateBalanceConfig validates balance framework configuration
func (cv *ConfigValidator) validateBalanceConfig(result *ValidationResult) {
	if cv.config.Balance.AnalysisThreshold < 0 || cv.config.Balance.AnalysisThreshold > 1 {
		result.Errors = append(result.Errors, "balance analysis threshold must be between 0 and 1")
	}

	validModes := map[string]bool{"auto": true, "manual": true, "hybrid": true}
	if !validModes[cv.config.Balance.RecommendationMode] {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid balance recommendation mode: %s", cv.config.Balance.RecommendationMode))
	}

	// Validate metrics to collect
	validMetrics := map[string]bool{
		"combat":      true,
		"progression": true,
		"economy":     true,
		"player":      true,
		"ai":          true,
	}

	for _, metric := range cv.config.Balance.MetricsToCollect {
		if !validMetrics[metric] {
			result.Warnings = append(result.Warnings, fmt.Sprintf("unknown balance metric: %s", metric))
		}
	}
}

// validateTestingConfig validates testing configuration
func (cv *ConfigValidator) validateTestingConfig(result *ValidationResult) {
	if cv.config.Testing.CoverageThreshold < 0 || cv.config.Testing.CoverageThreshold > 1 {
		result.Errors = append(result.Errors, "testing coverage threshold must be between 0 and 1")
	}

	// Warn if no test types are enabled
	if !cv.config.Testing.UnitTestsEnabled &&
	   !cv.config.Testing.IntegrationTestsEnabled &&
	   !cv.config.Testing.GameplayTestsEnabled &&
	   !cv.config.Testing.PerformanceTestsEnabled {
		result.Warnings = append(result.Warnings, "no test types are enabled")
	}
}

// validatePerformanceConfig validates performance configuration
func (cv *ConfigValidator) validatePerformanceConfig(result *ValidationResult) {
	if cv.config.Performance.ThresholdCPU < 0 || cv.config.Performance.ThresholdCPU > 100 {
		result.Errors = append(result.Errors, "CPU threshold must be between 0 and 100")
	}

	if cv.config.Performance.ThresholdMemory <= 0 {
		result.Errors = append(result.Errors, "memory threshold must be positive")
	}

	if cv.config.Performance.ThresholdFrameTime <= 0 {
		result.Errors = append(result.Errors, "frame time threshold must be positive")
	}

	// Warn about aggressive optimization settings
	if cv.config.Performance.OptimizationEnabled && cv.config.Performance.ThresholdCPU < 50 {
		result.Warnings = append(result.Warnings, "low CPU threshold with optimization enabled may cause frequent optimizations")
	}
}

// validatePluginConfig validates plugin system configuration
func (cv *ConfigValidator) validatePluginConfig(result *ValidationResult) {
	validSecurityLevels := map[string]bool{"strict": true, "normal": true, "permissive": true}
	if !validSecurityLevels[cv.config.Plugins.SecurityLevel] {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid plugin security level: %s", cv.config.Plugins.SecurityLevel))
	}

	if cv.config.Plugins.Enabled && cv.config.Plugins.PluginDir == "" {
		result.Errors = append(result.Errors, "plugin directory is required when plugins are enabled")
	}

	// Warn about security implications
	if cv.config.Plugins.Enabled && !cv.config.Plugins.SandboxEnabled {
		result.Warnings = append(result.Warnings, "plugins are enabled without sandboxing - this may pose security risks")
	}
}

// validateDocumentationConfig validates documentation configuration
func (cv *ConfigValidator) validateDocumentationConfig(result *ValidationResult) {
	validFormats := map[string]bool{"markdown": true, "html": true, "pdf": true}
	if !validFormats[cv.config.Documentation.OutputFormat] {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid documentation output format: %s", cv.config.Documentation.OutputFormat))
	}

	validFrequencies := map[string]bool{"on_change": true, "daily": true, "weekly": true, "manual": true}
	if !validFrequencies[cv.config.Documentation.UpdateFrequency] {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid documentation update frequency: %s", cv.config.Documentation.UpdateFrequency))
	}
}

// validateCICDConfig validates CI/CD configuration
func (cv *ConfigValidator) validateCICDConfig(result *ValidationResult) {
	if cv.config.CICD.Enabled {
		validProviders := map[string]bool{"github": true, "gitlab": true, "jenkins": true, "azure": true}
		if !validProviders[cv.config.CICD.Provider] {
			result.Errors = append(result.Errors, fmt.Sprintf("invalid CI/CD provider: %s", cv.config.CICD.Provider))
		}

		validNotificationLevels := map[string]bool{"all": true, "errors": true, "none": true}
		if !validNotificationLevels[cv.config.CICD.NotificationLevel] {
			result.Errors = append(result.Errors, fmt.Sprintf("invalid CI/CD notification level: %s", cv.config.CICD.NotificationLevel))
		}

		// Validate quality gates
		validGates := map[string]bool{
			"tests":     true,
			"coverage":  true,
			"security":  true,
			"lint":      true,
			"build":     true,
		}

		for _, gate := range cv.config.CICD.QualityGates {
			if !validGates[gate] {
				result.Warnings = append(result.Warnings, fmt.Sprintf("unknown quality gate: %s", gate))
			}
		}
	}
}

// validateDebugConfig validates debug configuration
func (cv *ConfigValidator) validateDebugConfig(result *ValidationResult) {
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[cv.config.Debug.LogLevel] {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid debug log level: %s", cv.config.Debug.LogLevel))
	}

	// Warn about performance implications
	if cv.config.Debug.TimeTravelEnabled {
		result.Warnings = append(result.Warnings, "time travel debugging is enabled - this may impact performance")
	}
}

// validateDirectoryStructure validates that required directories exist or can be created
func (cv *ConfigValidator) validateDirectoryStructure(result *ValidationResult) {
	directories := []string{
		cv.config.Specs.SpecDir,
		cv.config.Content.ContentDir,
		cv.config.Content.OutputDir,
	}

	if cv.config.Plugins.Enabled {
		directories = append(directories, cv.config.Plugins.PluginDir)
	}

	for _, dir := range directories {
		if dir == "" {
			continue // Already validated in specific validators
		}

		// Check if directory exists or can be created
		if err := cv.ensureDirectoryExists(dir); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("directory %s does not exist and cannot be created: %v", dir, err))
		}
	}
}

// Helper methods

func (cv *ConfigValidator) isValidVersion(version string) bool {
	// Simple semantic version validation
	matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`, version)
	return matched
}

func (cv *ConfigValidator) isValidProjectName(name string) bool {
	// Allow alphanumeric characters, hyphens, and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}

func (cv *ConfigValidator) isValidPath(path string) bool {
	// Basic path validation - check for invalid characters
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return false
		}
	}
	return true
}

func (cv *ConfigValidator) isValidURL(url string) bool {
	// Simple URL validation
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func (cv *ConfigValidator) ensureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Try to create the directory
		return os.MkdirAll(dir, 0755)
	}
	return nil
}
