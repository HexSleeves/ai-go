package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultWorkflowConfig(t *testing.T) {
	config := DefaultWorkflowConfig()

	// Test basic fields
	if config.Version == "" {
		t.Error("Default config should have a version")
	}

	if config.ProjectName == "" {
		t.Error("Default config should have a project name")
	}

	// Test spec configuration
	if config.Specs.SpecDir == "" {
		t.Error("Default config should have a spec directory")
	}

	if !config.Specs.TaskTracker {
		t.Error("Default config should enable task tracker")
	}

	// Test content configuration
	if config.Content.ContentDir == "" {
		t.Error("Default config should have a content directory")
	}

	if config.Content.OutputDir == "" {
		t.Error("Default config should have an output directory")
	}

	// Test balance configuration
	if config.Balance.AnalysisThreshold < 0 || config.Balance.AnalysisThreshold > 1 {
		t.Error("Default balance analysis threshold should be between 0 and 1")
	}

	// Test testing configuration
	if config.Testing.CoverageThreshold < 0 || config.Testing.CoverageThreshold > 1 {
		t.Error("Default testing coverage threshold should be between 0 and 1")
	}

	// Test performance configuration
	if config.Performance.ThresholdCPU < 0 || config.Performance.ThresholdCPU > 100 {
		t.Error("Default CPU threshold should be between 0 and 100")
	}
}

func TestConfigLoader_LoadConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "workflow_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test loading non-existent config (should create default)
	configPath := filepath.Join(tempDir, "workflow.json")
	loader := NewConfigLoader(configPath)

	config, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config == nil {
		t.Fatal("Config should not be nil")
	}

	// Verify default values
	if config.Version == "" {
		t.Error("Loaded config should have a version")
	}

	// Verify config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should have been created")
	}
}

func TestConfigLoader_SaveConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "workflow_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "workflow.json")
	loader := NewConfigLoader(configPath)

	// Create a test config
	config := DefaultWorkflowConfig()
	config.ProjectName = "test-project"

	// Save the config
	err = loader.SaveConfig(&config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should exist after saving")
	}

	// Load the config back and verify
	loadedConfig, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.ProjectName != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", loadedConfig.ProjectName)
	}
}

func TestConfigLoader_ValidateConfig(t *testing.T) {
	loader := NewConfigLoader("")

	// Test valid config
	validConfig := DefaultWorkflowConfig()
	err := loader.ValidateConfig(&validConfig)
	if err != nil {
		t.Errorf("Valid config should not produce validation error: %v", err)
	}

	// Test invalid config - empty version
	invalidConfig := DefaultWorkflowConfig()
	invalidConfig.Version = ""
	err = loader.ValidateConfig(&invalidConfig)
	if err == nil {
		t.Error("Config with empty version should produce validation error")
	}

	// Test invalid config - empty project name
	invalidConfig = DefaultWorkflowConfig()
	invalidConfig.ProjectName = ""
	err = loader.ValidateConfig(&invalidConfig)
	if err == nil {
		t.Error("Config with empty project name should produce validation error")
	}

	// Test invalid config - invalid threshold
	invalidConfig = DefaultWorkflowConfig()
	invalidConfig.Balance.AnalysisThreshold = 1.5
	err = loader.ValidateConfig(&invalidConfig)
	if err == nil {
		t.Error("Config with invalid balance threshold should produce validation error")
	}
}

func TestConfigLoader_YAMLSupport(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "workflow_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "workflow.yaml")
	loader := NewConfigLoader(configPath)

	// Create and save a test config
	config := DefaultWorkflowConfig()
	config.ProjectName = "yaml-test-project"

	err = loader.SaveConfig(&config)
	if err != nil {
		t.Fatalf("Failed to save YAML config: %v", err)
	}

	// Load the config back and verify
	loadedConfig, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load YAML config: %v", err)
	}

	if loadedConfig.ProjectName != "yaml-test-project" {
		t.Errorf("Expected project name 'yaml-test-project', got '%s'", loadedConfig.ProjectName)
	}
}

func TestConfigLoader_UnsupportedFormat(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "workflow_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "workflow.txt")
	loader := NewConfigLoader(configPath)

	config := DefaultWorkflowConfig()

	// Should fail to save unsupported format
	err = loader.SaveConfig(&config)
	if err == nil {
		t.Error("Should fail to save unsupported file format")
	}
}