package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// TileConfig holds configuration for tile-based rendering
type TileConfig struct {
	Enabled      bool    `json:"tiles_enabled"`
	TileSize     int     `json:"tile_size"`
	ScaleFactor  float32 `json:"scale_factor"`
	TilesetPath  string  `json:"tileset_path"`
	UseSmoothing bool    `json:"use_smoothing"`
	CacheSize    int     `json:"cache_size"`
}

// DefaultTileConfig provides sensible defaults for tile configuration
var DefaultTileConfig = TileConfig{
	Enabled:      false, // Start with ASCII as default for compatibility
	TileSize:     16,
	ScaleFactor:  1.0,
	TilesetPath:  "assets/tiles/",
	UseSmoothing: true,
	CacheSize:    1000, // Maximum number of cached tile images
}

// TileConfigFile represents the structure of the tile configuration file
type TileConfigFile struct {
	Tiles TileConfig `json:"tiles"`
}

// LoadTileConfig loads tile configuration from file or returns defaults
func LoadTileConfig() TileConfig {
	configPath := getTileConfigPath()

	// If config file doesn't exist, return defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultTileConfig
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultTileConfig
	}

	var configFile TileConfigFile
	if err := json.Unmarshal(data, &configFile); err != nil {
		return DefaultTileConfig
	}

	// Validate and apply defaults for missing fields
	config := configFile.Tiles
	if config.TileSize <= 0 {
		config.TileSize = DefaultTileConfig.TileSize
	}
	if config.ScaleFactor <= 0 {
		config.ScaleFactor = DefaultTileConfig.ScaleFactor
	}
	if config.TilesetPath == "" {
		config.TilesetPath = DefaultTileConfig.TilesetPath
	}
	if config.CacheSize <= 0 {
		config.CacheSize = DefaultTileConfig.CacheSize
	}

	return config
}

// SaveTileConfig saves tile configuration to file
func SaveTileConfig(config TileConfig) error {
	configPath := getTileConfigPath()

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	configFile := TileConfigFile{
		Tiles: config,
	}

	data, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// getTileConfigPath returns the path to the tile configuration file
func getTileConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "tile_config.json" // Fallback to current directory
	}

	configDir := filepath.Join(homeDir, ".config", "roguelike-gruid")
	return filepath.Join(configDir, "tile_config.json")
}

// ToggleTiles toggles tile rendering on/off and saves the configuration
func ToggleTiles(currentConfig *TileConfig) error {
	currentConfig.Enabled = !currentConfig.Enabled
	return SaveTileConfig(*currentConfig)
}

// ValidateTilesetPath checks if the tileset path exists and is accessible
func ValidateTilesetPath(path string) bool {
	if path == "" {
		return false
	}

	// Check if path exists and is a directory
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}
