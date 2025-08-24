package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/workflow/errors"
)

func TestCLIApp_Run_Help(t *testing.T) {
	app := createTestApp(t)
	
	args := []string{"workflow", "help"}
	err := app.Run(context.Background(), args)
	
	if err != nil {
		t.Errorf("Expected help command to succeed, got error: %v", err)
	}
}

func TestCLIApp_Run_Version(t *testing.T) {
	app := createTestApp(t)
	
	args := []string{"workflow", "version"}
	err := app.Run(context.Background(), args)
	
	if err != nil {
		t.Errorf("Expected version command to succeed, got error: %v", err)
	}
}

func TestCLIApp_Run_Status(t *testing.T) {
	app := createTestApp(t)
	
	// Create a temporary config file
	tempDir := t.TempDir()
	
	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	args := []string{"workflow", "status"}
	err := app.Run(context.Background(), args)
	
	if err != nil {
		t.Errorf("Expected status command to succeed, got error: %v", err)
	}
}

func TestCLIApp_Run_ConfigInit(t *testing.T) {
	app := createTestApp(t)
	
	// Create a temporary directory
	tempDir := t.TempDir()
	
	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	args := []string{"workflow", "config", "init"}
	err := app.Run(context.Background(), args)
	
	if err != nil {
		t.Errorf("Expected config init command to succeed, got error: %v", err)
	}
	
	// Verify config file was created
	configPath := filepath.Join(tempDir, "workflow.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}
}

func TestCLIApp_Run_ConfigShow(t *testing.T) {
	app := createTestApp(t)
	
	// Create a temporary directory with config
	tempDir := t.TempDir()
	
	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	// First initialize config
	initArgs := []string{"workflow", "config", "init"}
	err := app.Run(context.Background(), initArgs)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}
	
	// Then show config
	showArgs := []string{"workflow", "config", "show"}
	err = app.Run(context.Background(), showArgs)
	
	if err != nil {
		t.Errorf("Expected config show command to succeed, got error: %v", err)
	}
}

func TestCLIApp_Run_ConfigValidate(t *testing.T) {
	app := createTestApp(t)
	
	// Create a temporary directory with config
	tempDir := t.TempDir()
	
	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	// First initialize config
	initArgs := []string{"workflow", "config", "init"}
	err := app.Run(context.Background(), initArgs)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}
	
	// Then validate config
	validateArgs := []string{"workflow", "config", "validate"}
	err = app.Run(context.Background(), validateArgs)
	
	if err != nil {
		t.Errorf("Expected config validate command to succeed, got error: %v", err)
	}
}

func TestCLIApp_Run_SpecCommands(t *testing.T) {
	app := createTestApp(t)
	
	// Create a temporary directory with config
	tempDir := t.TempDir()
	
	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	// First initialize config
	initArgs := []string{"workflow", "config", "init"}
	err := app.Run(context.Background(), initArgs)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}
	
	testCases := []struct {
		name string
		args []string
	}{
		{"spec status", []string{"workflow", "spec", "status"}},
		{"spec validate", []string{"workflow", "spec", "validate"}},
		{"spec generate", []string{"workflow", "spec", "generate"}},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := app.Run(context.Background(), tc.args)
			if err != nil {
				t.Errorf("Expected %s command to succeed, got error: %v", tc.name, err)
			}
		})
	}
}

func TestCLIApp_Run_ContentCommands(t *testing.T) {
	app := createTestApp(t)
	
	// Create a temporary directory with config
	tempDir := t.TempDir()
	
	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	// First initialize config
	initArgs := []string{"workflow", "config", "init"}
	err := app.Run(context.Background(), initArgs)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}
	
	testCases := []struct {
		name string
		args []string
	}{
		{"content generate", []string{"workflow", "content", "generate"}},
		{"content validate", []string{"workflow", "content", "validate"}},
		{"content reload", []string{"workflow", "content", "reload"}},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := app.Run(context.Background(), tc.args)
			if err != nil {
				t.Errorf("Expected %s command to succeed, got error: %v", tc.name, err)
			}
		})
	}
}

func TestCLIApp_Run_BalanceCommands(t *testing.T) {
	app := createTestApp(t)
	
	// Create a temporary directory with config
	tempDir := t.TempDir()
	
	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	// First initialize config
	initArgs := []string{"workflow", "config", "init"}
	err := app.Run(context.Background(), initArgs)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}
	
	testCases := []struct {
		name string
		args []string
	}{
		{"balance analyze", []string{"workflow", "balance", "analyze"}},
		{"balance simulate", []string{"workflow", "balance", "simulate"}},
		{"balance recommend", []string{"workflow", "balance", "recommend"}},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := app.Run(context.Background(), tc.args)
			if err != nil {
				t.Errorf("Expected %s command to succeed, got error: %v", tc.name, err)
			}
		})
	}
}

func TestCLIApp_Run_TestCommands(t *testing.T) {
	app := createTestApp(t)
	
	// Create a temporary directory with config
	tempDir := t.TempDir()
	
	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	// First initialize config
	initArgs := []string{"workflow", "config", "init"}
	err := app.Run(context.Background(), initArgs)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}
	
	testCases := []struct {
		name string
		args []string
	}{
		{"test run", []string{"workflow", "test", "run"}},
		{"test generate", []string{"workflow", "test", "generate"}},
		{"test coverage", []string{"workflow", "test", "coverage"}},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := app.Run(context.Background(), tc.args)
			if err != nil {
				t.Errorf("Expected %s command to succeed, got error: %v", tc.name, err)
			}
		})
	}
}

func TestCLIApp_Run_UnknownCommand(t *testing.T) {
	app := createTestApp(t)
	
	args := []string{"workflow", "unknown-command"}
	err := app.Run(context.Background(), args)
	
	if err == nil {
		t.Error("Expected unknown command to fail")
	}
}

func TestCLIApp_Run_NoArgs(t *testing.T) {
	app := createTestApp(t)
	
	args := []string{"workflow"}
	err := app.Run(context.Background(), args)
	
	// The app shows help when no args are provided, which is not an error
	if err != nil {
		t.Errorf("Expected no args to show help without error, got: %v", err)
	}
}

func TestCLIApp_Run_InvalidSubcommand(t *testing.T) {
	app := createTestApp(t)
	
	// Create a temporary directory with config
	tempDir := t.TempDir()
	
	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	// First initialize config
	initArgs := []string{"workflow", "config", "init"}
	err := app.Run(context.Background(), initArgs)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}
	
	args := []string{"workflow", "spec", "invalid-subcommand"}
	err = app.Run(context.Background(), args)
	
	if err == nil {
		t.Error("Expected invalid subcommand to fail")
	}
}

func TestCLIApp_findConfigFile(t *testing.T) {
	app := createTestApp(t)
	
	// Create a temporary directory
	tempDir := t.TempDir()
	
	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	// Test default case (no config file exists)
	configPath := app.findConfigFile()
	expectedDefault := "workflow.yaml"
	if configPath != expectedDefault {
		t.Errorf("Expected default config path '%s', got '%s'", expectedDefault, configPath)
	}
	
	// Create a config file and test detection
	testConfigPath := "workflow.json"
	file, err := os.Create(testConfigPath)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	file.Close()
	
	configPath = app.findConfigFile()
	if configPath != testConfigPath {
		t.Errorf("Expected config path '%s', got '%s'", testConfigPath, configPath)
	}
}

// Helper function to create a test app
func createTestApp(t *testing.T) *CLIApp {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce noise in tests
	}))
	
	workflowLogger, err := errors.NewWorkflowLogger("workflow-cli-test", errors.WorkflowLogConfig{
		Level:           errors.WorkflowLogLevelError,
		EnableStructured: false,
		EnableFile:      false,
		EnableConsole:   false,
		EnableMetrics:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create workflow logger: %v", err)
	}
	
	return NewCLIApp(logger, workflowLogger)
}