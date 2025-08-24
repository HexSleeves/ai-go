package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// WorkflowConfig represents the main workflow configuration
type WorkflowConfig struct {
	// General settings
	Version     string `json:"version" yaml:"version"`
	ProjectName string `json:"project_name" yaml:"project_name"`

	// Spec system configuration
	Specs SpecConfig `json:"specs" yaml:"specs"`

	// Content pipeline configuration
	Content ContentConfig `json:"content" yaml:"content"`

	// Balance framework configuration
	Balance BalanceConfig `json:"balance" yaml:"balance"`

	// Testing configuration
	Testing TestingConfig `json:"testing" yaml:"testing"`

	// Performance configuration
	Performance PerformanceConfig `json:"performance" yaml:"performance"`

	// Plugin system configuration
	Plugins PluginConfig `json:"plugins" yaml:"plugins"`

	// Documentation configuration
	Documentation DocumentationConfig `json:"documentation" yaml:"documentation"`

	// CI/CD configuration
	CICD CICDConfig `json:"cicd" yaml:"cicd"`

	// Debug configuration
	Debug DebugConfig `json:"debug" yaml:"debug"`
}

// SpecConfig holds spec-driven development configuration
type SpecConfig struct {
	SpecDir         string   `json:"spec_dir" yaml:"spec_dir"`
	TaskTracker     bool     `json:"task_tracker" yaml:"task_tracker"`
	AutoGeneration  bool     `json:"auto_generation" yaml:"auto_generation"`
	ValidationRules []string `json:"validation_rules" yaml:"validation_rules"`
}

// ContentConfig holds content pipeline configuration
type ContentConfig struct {
	LubanEnabled    bool   `json:"luban_enabled" yaml:"luban_enabled"`
	LubanServerURL  string `json:"luban_server_url" yaml:"luban_server_url"`
	ContentDir      string `json:"content_dir" yaml:"content_dir"`
	OutputDir       string `json:"output_dir" yaml:"output_dir"`
	HotReload       bool   `json:"hot_reload" yaml:"hot_reload"`
	ValidationLevel string `json:"validation_level" yaml:"validation_level"` // strict, normal, lenient
}

// BalanceConfig holds game balance framework configuration
type BalanceConfig struct {
	MetricsEnabled     bool     `json:"metrics_enabled" yaml:"metrics_enabled"`
	SimulationEnabled  bool     `json:"simulation_enabled" yaml:"simulation_enabled"`
	AnalysisThreshold  float64  `json:"analysis_threshold" yaml:"analysis_threshold"`
	RecommendationMode string   `json:"recommendation_mode" yaml:"recommendation_mode"` // auto, manual, hybrid
	MetricsToCollect   []string `json:"metrics_to_collect" yaml:"metrics_to_collect"`
}

// TestingConfig holds testing framework configuration
type TestingConfig struct {
	UnitTestsEnabled        bool    `json:"unit_tests_enabled" yaml:"unit_tests_enabled"`
	IntegrationTestsEnabled bool    `json:"integration_tests_enabled" yaml:"integration_tests_enabled"`
	GameplayTestsEnabled    bool    `json:"gameplay_tests_enabled" yaml:"gameplay_tests_enabled"`
	PerformanceTestsEnabled bool    `json:"performance_tests_enabled" yaml:"performance_tests_enabled"`
	CoverageThreshold       float64 `json:"coverage_threshold" yaml:"coverage_threshold"`
	AutoGeneration          bool    `json:"auto_generation" yaml:"auto_generation"`
}

// PerformanceConfig holds performance optimization configuration
type PerformanceConfig struct {
	ProfilingEnabled    bool    `json:"profiling_enabled" yaml:"profiling_enabled"`
	MonitoringEnabled   bool    `json:"monitoring_enabled" yaml:"monitoring_enabled"`
	OptimizationEnabled bool    `json:"optimization_enabled" yaml:"optimization_enabled"`
	ThresholdCPU        float64 `json:"threshold_cpu" yaml:"threshold_cpu"`
	ThresholdMemory     float64 `json:"threshold_memory" yaml:"threshold_memory"`
	ThresholdFrameTime  float64 `json:"threshold_frame_time" yaml:"threshold_frame_time"`
}

// PluginConfig holds plugin system configuration
type PluginConfig struct {
	Enabled         bool     `json:"enabled" yaml:"enabled"`
	PluginDir       string   `json:"plugin_dir" yaml:"plugin_dir"`
	SandboxEnabled  bool     `json:"sandbox_enabled" yaml:"sandbox_enabled"`
	SecurityLevel   string   `json:"security_level" yaml:"security_level"` // strict, normal, permissive
	AllowedPlugins  []string `json:"allowed_plugins" yaml:"allowed_plugins"`
	BlockedPlugins  []string `json:"blocked_plugins" yaml:"blocked_plugins"`
}

// DocumentationConfig holds documentation system configuration
type DocumentationConfig struct {
	AutoGeneration    bool   `json:"auto_generation" yaml:"auto_generation"`
	OutputFormat      string `json:"output_format" yaml:"output_format"` // markdown, html, pdf
	IncludeExamples   bool   `json:"include_examples" yaml:"include_examples"`
	IncludeDiagrams   bool   `json:"include_diagrams" yaml:"include_diagrams"`
	UpdateFrequency   string `json:"update_frequency" yaml:"update_frequency"` // on_change, daily, weekly
}

// CICDConfig holds CI/CD pipeline configuration
type CICDConfig struct {
	Enabled           bool     `json:"enabled" yaml:"enabled"`
	Provider          string   `json:"provider" yaml:"provider"` // github, gitlab, jenkins
	AutoDeploy        bool     `json:"auto_deploy" yaml:"auto_deploy"`
	QualityGates      []string `json:"quality_gates" yaml:"quality_gates"`
	NotificationLevel string   `json:"notification_level" yaml:"notification_level"` // all, errors, none
}

// DebugConfig holds debug and development tools configuration
type DebugConfig struct {
	VisualizationEnabled bool   `json:"visualization_enabled" yaml:"visualization_enabled"`
	TimeTravelEnabled    bool   `json:"time_travel_enabled" yaml:"time_travel_enabled"`
	HotReloadEnabled     bool   `json:"hot_reload_enabled" yaml:"hot_reload_enabled"`
	LogLevel             string `json:"log_level" yaml:"log_level"` // debug, info, warn, error
	MetricsEnabled       bool   `json:"metrics_enabled" yaml:"metrics_enabled"`
}

// DefaultWorkflowConfig returns a configuration with sensible defaults
func DefaultWorkflowConfig() WorkflowConfig {
	return WorkflowConfig{
		Version:     "1.0.0",
		ProjectName: "roguelike-workflow",
		Specs: SpecConfig{
			SpecDir:         ".kiro/specs",
			TaskTracker:     true,
			AutoGeneration:  true,
			ValidationRules: []string{"requirements", "design", "tasks"},
		},
		Content: ContentConfig{
			LubanEnabled:    false, // Start disabled, enable when Luban is integrated
			LubanServerURL:  "http://localhost:8080",
			ContentDir:      "assets/content",
			OutputDir:       "internal/content/generated",
			HotReload:       true,
			ValidationLevel: "normal",
		},
		Balance: BalanceConfig{
			MetricsEnabled:     true,
			SimulationEnabled:  true,
			AnalysisThreshold:  0.8,
			RecommendationMode: "hybrid",
			MetricsToCollect:   []string{"combat", "progression", "economy"},
		},
		Testing: TestingConfig{
			UnitTestsEnabled:        true,
			IntegrationTestsEnabled: true,
			GameplayTestsEnabled:    true,
			PerformanceTestsEnabled: true,
			CoverageThreshold:       0.8,
			AutoGeneration:          true,
		},
		Performance: PerformanceConfig{
			ProfilingEnabled:    true,
			MonitoringEnabled:   true,
			OptimizationEnabled: false, // Start disabled for safety
			ThresholdCPU:        80.0,
			ThresholdMemory:     512.0, // MB
			ThresholdFrameTime:  16.67, // 60 FPS target
		},
		Plugins: PluginConfig{
			Enabled:        false, // Start disabled
			PluginDir:      "plugins",
			SandboxEnabled: true,
			SecurityLevel:  "strict",
			AllowedPlugins: []string{},
			BlockedPlugins: []string{},
		},
		Documentation: DocumentationConfig{
			AutoGeneration:  true,
			OutputFormat:    "markdown",
			IncludeExamples: true,
			IncludeDiagrams: true,
			UpdateFrequency: "on_change",
		},
		CICD: CICDConfig{
			Enabled:           false, // Start disabled
			Provider:          "github",
			AutoDeploy:        false,
			QualityGates:      []string{"tests", "coverage", "security"},
			NotificationLevel: "errors",
		},
		Debug: DebugConfig{
			VisualizationEnabled: true,
			TimeTravelEnabled:    false, // Start disabled for performance
			HotReloadEnabled:     true,
			LogLevel:             "info",
			MetricsEnabled:       true,
		},
	}
}
// ConfigLoader handles loading and saving workflow configuration
type ConfigLoader struct {
	configPath string
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(configPath string) *ConfigLoader {
	return &ConfigLoader{
		configPath: configPath,
	}
}

// LoadConfig loads configuration from file, creating default if not exists
func (cl *ConfigLoader) LoadConfig() (*WorkflowConfig, error) {
	// Check if config file exists
	if _, err := os.Stat(cl.configPath); os.IsNotExist(err) {
		// Create default config
		config := DefaultWorkflowConfig()
		if err := cl.SaveConfig(&config); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return &config, nil
	}

	// Determine file format based on extension
	ext := filepath.Ext(cl.configPath)

	data, err := os.ReadFile(cl.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config WorkflowConfig

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	// Validate configuration
	if err := cl.ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func (cl *ConfigLoader) SaveConfig(config *WorkflowConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(cl.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Determine file format based on extension
	ext := filepath.Ext(cl.configPath)

	var data []byte
	var err error

	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML config: %w", err)
		}
	case ".json":
		data, err = json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON config: %w", err)
		}
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err := os.WriteFile(cl.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateConfig validates the workflow configuration
func (cl *ConfigLoader) ValidateConfig(config *WorkflowConfig) error {
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	if config.ProjectName == "" {
		return fmt.Errorf("project name is required")
	}

	// Validate spec configuration
	if config.Specs.SpecDir == "" {
		return fmt.Errorf("spec directory is required")
	}

	// Validate content configuration
	if config.Content.ContentDir == "" {
		return fmt.Errorf("content directory is required")
	}

	if config.Content.OutputDir == "" {
		return fmt.Errorf("content output directory is required")
	}

	// Validate balance configuration
	if config.Balance.AnalysisThreshold < 0 || config.Balance.AnalysisThreshold > 1 {
		return fmt.Errorf("balance analysis threshold must be between 0 and 1")
	}

	// Validate testing configuration
	if config.Testing.CoverageThreshold < 0 || config.Testing.CoverageThreshold > 1 {
		return fmt.Errorf("testing coverage threshold must be between 0 and 1")
	}

	// Validate performance configuration
	if config.Performance.ThresholdCPU < 0 || config.Performance.ThresholdCPU > 100 {
		return fmt.Errorf("CPU threshold must be between 0 and 100")
	}

	if config.Performance.ThresholdMemory <= 0 {
		return fmt.Errorf("memory threshold must be positive")
	}

	if config.Performance.ThresholdFrameTime <= 0 {
		return fmt.Errorf("frame time threshold must be positive")
	}

	// Validate plugin configuration
	validSecurityLevels := map[string]bool{"strict": true, "normal": true, "permissive": true}
	if !validSecurityLevels[config.Plugins.SecurityLevel] {
		return fmt.Errorf("invalid plugin security level: %s", config.Plugins.SecurityLevel)
	}

	// Validate documentation configuration
	validFormats := map[string]bool{"markdown": true, "html": true, "pdf": true}
	if !validFormats[config.Documentation.OutputFormat] {
		return fmt.Errorf("invalid documentation output format: %s", config.Documentation.OutputFormat)
	}

	// Validate debug configuration
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[config.Debug.LogLevel] {
		return fmt.Errorf("invalid debug log level: %s", config.Debug.LogLevel)
	}

	return nil
}

// MergeWithDefaults fills missing fields with default values
func (cl *ConfigLoader) MergeWithDefaults(config *WorkflowConfig) *WorkflowConfig {
	defaults := DefaultWorkflowConfig()

	// Merge missing fields - this is a simplified version
	// In a real implementation, you'd want to do deep merging
	if config.Version == "" {
		config.Version = defaults.Version
	}

	if config.ProjectName == "" {
		config.ProjectName = defaults.ProjectName
	}

	// Add more field merging as needed

	return config
}
