package config

import (
	"testing"
)

func TestConfigValidator_ValidateAll(t *testing.T) {
	// Test valid configuration
	validConfig := DefaultWorkflowConfig()
	validator := NewConfigValidator(&validConfig)
	result := validator.ValidateAll()

	if !result.Valid {
		t.Errorf("Valid config should pass validation. Errors: %v", result.Errors)
	}

	if len(result.Errors) > 0 {
		t.Errorf("Valid config should have no errors. Got: %v", result.Errors)
	}
}

func TestConfigValidator_ValidateBasicFields(t *testing.T) {
	// Test empty version
	config := DefaultWorkflowConfig()
	config.Version = ""
	validator := NewConfigValidator(&config)
	result := validator.ValidateAll()

	if result.Valid {
		t.Error("Config with empty version should be invalid")
	}

	if len(result.Errors) == 0 {
		t.Error("Config with empty version should have errors")
	}

	// Test invalid version format
	config = DefaultWorkflowConfig()
	config.Version = "invalid-version"
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with invalid version format should be invalid")
	}

	// Test empty project name
	config = DefaultWorkflowConfig()
	config.ProjectName = ""
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with empty project name should be invalid")
	}

	// Test invalid project name
	config = DefaultWorkflowConfig()
	config.ProjectName = "invalid project name!"
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with invalid project name should be invalid")
	}
}

func TestConfigValidator_ValidateSpecConfig(t *testing.T) {
	// Test empty spec directory
	config := DefaultWorkflowConfig()
	config.Specs.SpecDir = ""
	validator := NewConfigValidator(&config)
	result := validator.ValidateAll()

	if result.Valid {
		t.Error("Config with empty spec directory should be invalid")
	}

	// Test unknown validation rule
	config = DefaultWorkflowConfig()
	config.Specs.ValidationRules = []string{"unknown-rule"}
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	// Should be valid but have warnings
	if !result.Valid {
		t.Error("Config with unknown validation rule should be valid but have warnings")
	}

	if len(result.Warnings) == 0 {
		t.Error("Config with unknown validation rule should have warnings")
	}
}

func TestConfigValidator_ValidateContentConfig(t *testing.T) {
	// Test empty content directory
	config := DefaultWorkflowConfig()
	config.Content.ContentDir = ""
	validator := NewConfigValidator(&config)
	result := validator.ValidateAll()

	if result.Valid {
		t.Error("Config with empty content directory should be invalid")
	}

	// Test empty output directory
	config = DefaultWorkflowConfig()
	config.Content.OutputDir = ""
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with empty output directory should be invalid")
	}

	// Test Luban enabled without URL
	config = DefaultWorkflowConfig()
	config.Content.LubanEnabled = true
	config.Content.LubanServerURL = ""
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with Luban enabled but no URL should be invalid")
	}

	// Test invalid validation level
	config = DefaultWorkflowConfig()
	config.Content.ValidationLevel = "invalid"
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with invalid validation level should be invalid")
	}
}

func TestConfigValidator_ValidateBalanceConfig(t *testing.T) {
	// Test invalid analysis threshold (too low)
	config := DefaultWorkflowConfig()
	config.Balance.AnalysisThreshold = -0.1
	validator := NewConfigValidator(&config)
	result := validator.ValidateAll()

	if result.Valid {
		t.Error("Config with negative analysis threshold should be invalid")
	}

	// Test invalid analysis threshold (too high)
	config = DefaultWorkflowConfig()
	config.Balance.AnalysisThreshold = 1.1
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with analysis threshold > 1 should be invalid")
	}

	// Test invalid recommendation mode
	config = DefaultWorkflowConfig()
	config.Balance.RecommendationMode = "invalid"
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with invalid recommendation mode should be invalid")
	}

	// Test unknown metric
	config = DefaultWorkflowConfig()
	config.Balance.MetricsToCollect = []string{"unknown-metric"}
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	// Should be valid but have warnings
	if !result.Valid {
		t.Error("Config with unknown metric should be valid but have warnings")
	}

	if len(result.Warnings) == 0 {
		t.Error("Config with unknown metric should have warnings")
	}
}

func TestConfigValidator_ValidateTestingConfig(t *testing.T) {
	// Test invalid coverage threshold (too low)
	config := DefaultWorkflowConfig()
	config.Testing.CoverageThreshold = -0.1
	validator := NewConfigValidator(&config)
	result := validator.ValidateAll()

	if result.Valid {
		t.Error("Config with negative coverage threshold should be invalid")
	}

	// Test invalid coverage threshold (too high)
	config = DefaultWorkflowConfig()
	config.Testing.CoverageThreshold = 1.1
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with coverage threshold > 1 should be invalid")
	}

	// Test no test types enabled
	config = DefaultWorkflowConfig()
	config.Testing.UnitTestsEnabled = false
	config.Testing.IntegrationTestsEnabled = false
	config.Testing.GameplayTestsEnabled = false
	config.Testing.PerformanceTestsEnabled = false
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	// Should be valid but have warnings
	if !result.Valid {
		t.Error("Config with no test types should be valid but have warnings")
	}

	if len(result.Warnings) == 0 {
		t.Error("Config with no test types should have warnings")
	}
}

func TestConfigValidator_ValidatePerformanceConfig(t *testing.T) {
	// Test invalid CPU threshold (too low)
	config := DefaultWorkflowConfig()
	config.Performance.ThresholdCPU = -1
	validator := NewConfigValidator(&config)
	result := validator.ValidateAll()

	if result.Valid {
		t.Error("Config with negative CPU threshold should be invalid")
	}

	// Test invalid CPU threshold (too high)
	config = DefaultWorkflowConfig()
	config.Performance.ThresholdCPU = 101
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with CPU threshold > 100 should be invalid")
	}

	// Test invalid memory threshold
	config = DefaultWorkflowConfig()
	config.Performance.ThresholdMemory = -1
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with negative memory threshold should be invalid")
	}

	// Test invalid frame time threshold
	config = DefaultWorkflowConfig()
	config.Performance.ThresholdFrameTime = -1
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with negative frame time threshold should be invalid")
	}

	// Test aggressive optimization warning
	config = DefaultWorkflowConfig()
	config.Performance.OptimizationEnabled = true
	config.Performance.ThresholdCPU = 30
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	// Should be valid but have warnings
	if !result.Valid {
		t.Error("Config with aggressive optimization should be valid but have warnings")
	}

	if len(result.Warnings) == 0 {
		t.Error("Config with aggressive optimization should have warnings")
	}
}

func TestConfigValidator_ValidatePluginConfig(t *testing.T) {
	// Test invalid security level
	config := DefaultWorkflowConfig()
	config.Plugins.SecurityLevel = "invalid"
	validator := NewConfigValidator(&config)
	result := validator.ValidateAll()

	if result.Valid {
		t.Error("Config with invalid plugin security level should be invalid")
	}

	// Test plugins enabled without directory
	config = DefaultWorkflowConfig()
	config.Plugins.Enabled = true
	config.Plugins.PluginDir = ""
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with plugins enabled but no directory should be invalid")
	}

	// Test plugins enabled without sandbox
	config = DefaultWorkflowConfig()
	config.Plugins.Enabled = true
	config.Plugins.SandboxEnabled = false
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	// Should be valid but have warnings
	if !result.Valid {
		t.Error("Config with plugins enabled without sandbox should be valid but have warnings")
	}

	if len(result.Warnings) == 0 {
		t.Error("Config with plugins enabled without sandbox should have warnings")
	}
}

func TestConfigValidator_ValidateDocumentationConfig(t *testing.T) {
	// Test invalid output format
	config := DefaultWorkflowConfig()
	config.Documentation.OutputFormat = "invalid"
	validator := NewConfigValidator(&config)
	result := validator.ValidateAll()

	if result.Valid {
		t.Error("Config with invalid documentation output format should be invalid")
	}

	// Test invalid update frequency
	config = DefaultWorkflowConfig()
	config.Documentation.UpdateFrequency = "invalid"
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	if result.Valid {
		t.Error("Config with invalid documentation update frequency should be invalid")
	}
}

func TestConfigValidator_ValidateDebugConfig(t *testing.T) {
	// Test invalid log level
	config := DefaultWorkflowConfig()
	config.Debug.LogLevel = "invalid"
	validator := NewConfigValidator(&config)
	result := validator.ValidateAll()

	if result.Valid {
		t.Error("Config with invalid debug log level should be invalid")
	}

	// Test time travel enabled warning
	config = DefaultWorkflowConfig()
	config.Debug.TimeTravelEnabled = true
	validator = NewConfigValidator(&config)
	result = validator.ValidateAll()

	// Should be valid but have warnings
	if !result.Valid {
		t.Error("Config with time travel enabled should be valid but have warnings")
	}

	if len(result.Warnings) == 0 {
		t.Error("Config with time travel enabled should have warnings")
	}
}

func TestConfigValidator_HelperMethods(t *testing.T) {
	validator := NewConfigValidator(&WorkflowConfig{})

	// Test version validation
	if !validator.isValidVersion("1.0.0") {
		t.Error("1.0.0 should be a valid version")
	}

	if !validator.isValidVersion("1.0.0-beta") {
		t.Error("1.0.0-beta should be a valid version")
	}

	if validator.isValidVersion("invalid") {
		t.Error("'invalid' should not be a valid version")
	}

	// Test project name validation
	if !validator.isValidProjectName("valid-project_name") {
		t.Error("'valid-project_name' should be a valid project name")
	}

	if validator.isValidProjectName("invalid project name!") {
		t.Error("'invalid project name!' should not be a valid project name")
	}

	// Test URL validation
	if !validator.isValidURL("http://localhost:8080") {
		t.Error("'http://localhost:8080' should be a valid URL")
	}

	if !validator.isValidURL("https://example.com") {
		t.Error("'https://example.com' should be a valid URL")
	}

	if validator.isValidURL("invalid-url") {
		t.Error("'invalid-url' should not be a valid URL")
	}
}