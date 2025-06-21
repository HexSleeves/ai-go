package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// GameplayConfig holds gameplay-related settings
type GameplayConfig struct {
	// Difficulty settings
	MonsterSpawnRate        float64 `json:"monster_spawn_rate"`
	MonsterDamageMultiplier float64 `json:"monster_damage_multiplier"`
	PlayerHealthMultiplier  float64 `json:"player_health_multiplier"`
	XPMultiplier            float64 `json:"xp_multiplier"`

	// Game mechanics
	AutoSave          bool `json:"auto_save"`
	AutoSaveInterval  int  `json:"auto_save_interval"` // minutes
	PermaDeath        bool `json:"perma_death"`
	ShowDamageNumbers bool `json:"show_damage_numbers"`
	ShowHealthBars    bool `json:"show_health_bars"`

	// Turn system
	TurnTimeLimit  int `json:"turn_time_limit"` // seconds, 0 = no limit
	AnimationSpeed int `json:"animation_speed"` // 1-10 scale

	// FOV settings
	FOVRadius    int    `json:"fov_radius"`
	FOVAlgorithm string `json:"fov_algorithm"` // "shadowcast", "raycasting"

	// Map generation
	RoomMinSize        int `json:"room_min_size"`
	RoomMaxSize        int `json:"room_max_size"`
	MaxRooms           int `json:"max_rooms"`
	MaxMonstersPerRoom int `json:"max_monsters_per_room"`
	DungeonWidth       int `json:"dungeon_width"`
	DungeonHeight      int `json:"dungeon_height"`
}

// DisplayConfig holds display-related settings
type DisplayConfig struct {
	// Window settings
	WindowWidth  int  `json:"window_width"`
	WindowHeight int  `json:"window_height"`
	Fullscreen   bool `json:"fullscreen"`
	VSync        bool `json:"vsync"`

	// Graphics settings
	TileSize     int     `json:"tile_size"`
	FontSize     int     `json:"font_size"`
	FontPath     string  `json:"font_path"`
	ColorScheme  string  `json:"color_scheme"` // "classic", "modern", "high_contrast"
	TilesEnabled bool    `json:"tiles_enabled"`
	TilesetPath  string  `json:"tileset_path"`
	ScaleFactorX float32 `json:"scale_factor_x"`
	ScaleFactorY float32 `json:"scale_factor_y"`
	UseSmoothing bool    `json:"use_smoothing"`
	CacheSize    int     `json:"cache_size"`

	// UI settings
	ShowFPS        bool `json:"show_fps"`
	ShowMinimap    bool `json:"show_minimap"`
	MessageLogSize int  `json:"message_log_size"`

	// Accessibility
	HighContrast   bool   `json:"high_contrast"`
	LargeText      bool   `json:"large_text"`
	ColorBlindMode string `json:"color_blind_mode"` // "none", "protanopia", "deuteranopia", "tritanopia"
}

// InputConfig holds input-related settings
type InputConfig struct {
	// Key bindings
	KeyBindings map[string]string `json:"key_bindings"`

	// Mouse settings
	MouseEnabled     bool    `json:"mouse_enabled"`
	MouseSensitivity float64 `json:"mouse_sensitivity"`

	// Gamepad settings
	GamepadEnabled  bool    `json:"gamepad_enabled"`
	GamepadDeadzone float64 `json:"gamepad_deadzone"`

	// Input behavior
	RepeatDelay     int `json:"repeat_delay"`      // milliseconds
	RepeatRate      int `json:"repeat_rate"`       // keys per second
	DoubleClickTime int `json:"double_click_time"` // milliseconds
}

// AudioConfig holds audio-related settings
type AudioConfig struct {
	// Volume settings
	MasterVolume float64 `json:"master_volume"` // 0.0 - 1.0
	SFXVolume    float64 `json:"sfx_volume"`
	MusicVolume  float64 `json:"music_volume"`

	// Audio behavior
	AudioEnabled    bool `json:"audio_enabled"`
	MuteOnFocusLoss bool `json:"mute_on_focus_loss"`

	// Audio quality
	SampleRate int `json:"sample_rate"`
	BufferSize int `json:"buffer_size"`
}

// AdvancedConfig holds advanced/debug settings
type AdvancedConfig struct {
	// Debug settings
	DebugMode     bool   `json:"debug_mode"`
	ShowDebugInfo bool   `json:"show_debug_info"`
	LogLevel      string `json:"log_level"` // "debug", "info", "warn", "error"
	LogToFile     bool   `json:"log_to_file"`

	// Performance settings
	MaxFPS          int `json:"max_fps"`
	VMemoryLimit    int `json:"vmemory_limit"` // MB
	GCTargetPercent int `json:"gc_target_percent"`

	// Development settings
	EnableProfiling bool `json:"enable_profiling"`
	ProfilePort     int  `json:"profile_port"`
	EnableMetrics   bool `json:"enable_metrics"`
	MetricsPort     int  `json:"metrics_port"`
}

// FullConfig combines all configuration sections
type FullConfig struct {
	Gameplay GameplayConfig `json:"gameplay"`
	Display  DisplayConfig  `json:"display"`
	Input    InputConfig    `json:"input"`
	Audio    AudioConfig    `json:"audio"`
	Advanced AdvancedConfig `json:"advanced"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() FullConfig {
	return FullConfig{
		Gameplay: GameplayConfig{
			MonsterSpawnRate:        1.0,
			MonsterDamageMultiplier: 1.0,
			PlayerHealthMultiplier:  1.0,
			XPMultiplier:            1.0,
			AutoSave:                true,
			AutoSaveInterval:        5,
			PermaDeath:              false,
			ShowDamageNumbers:       true,
			ShowHealthBars:          true,
			TurnTimeLimit:           0,
			AnimationSpeed:          5,
			FOVRadius:               10,
			FOVAlgorithm:            "shadowcast",
			RoomMinSize:             6,
			RoomMaxSize:             10,
			MaxRooms:                10,
			MaxMonstersPerRoom:      2,
			DungeonWidth:            80,
			DungeonHeight:           24,
		},
		Display: DisplayConfig{
			WindowWidth:    1280, // 80 chars * 16 pixels = 1280
			WindowHeight:   480,  // 24 chars * 20 pixels = 480 (adjusted for smaller font)
			Fullscreen:     false,
			VSync:          true,
			TileSize:       16,
			FontSize:       24,
			FontPath:       "",
			ColorScheme:    "classic",
			ShowFPS:        true,
			ShowMinimap:    true,
			MessageLogSize: 50,
			HighContrast:   false,
			LargeText:      false,
			ColorBlindMode: "none",
			TilesEnabled:   true,
			TilesetPath:    TilesetPath,
			ScaleFactorX:   1.0,
			ScaleFactorY:   1.0,
			UseSmoothing:   true,
			CacheSize:      1024,
		},
		Input: InputConfig{
			KeyBindings: map[string]string{
				"move_north": "w,k,up",
				"move_south": "s,j,down",
				"move_west":  "a,h,left",
				"move_east":  "d,l,right",
				"wait":       "space,period",
				"inventory":  "i",
				"pickup":     "g",
				"drop":       "D",
				"quit":       "q,escape",
				"save":       "ctrl+s",
				"load":       "ctrl+l",
			},
			MouseEnabled:     true,
			MouseSensitivity: 1.0,
			GamepadEnabled:   false,
			GamepadDeadzone:  0.2,
			RepeatDelay:      500,
			RepeatRate:       10,
			DoubleClickTime:  300,
		},
		Audio: AudioConfig{
			MasterVolume:    0.8,
			SFXVolume:       0.8,
			MusicVolume:     0.6,
			AudioEnabled:    true,
			MuteOnFocusLoss: true,
			SampleRate:      44100,
			BufferSize:      1024,
		},
		Advanced: AdvancedConfig{
			DebugMode:       false,
			ShowDebugInfo:   false,
			LogLevel:        "info",
			LogToFile:       false,
			MaxFPS:          60,
			VMemoryLimit:    512,
			GCTargetPercent: 100,
			EnableProfiling: false,
			ProfilePort:     6060,
			EnableMetrics:   false,
			MetricsPort:     8080,
		},
	}
}

const (
	ConfigDir  = "config"
	ConfigFile = "game_config.json"

	TilesetPath = "assets/tiles"
)

// LoadConfig loads configuration from file or creates default
func LoadConfig() (*FullConfig, error) {
	configPath := filepath.Join(ConfigDir, ConfigFile)

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(ConfigDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		config := DefaultConfig()
		if err := SaveConfig(&config); err != nil {
			slog.Warn("Failed to save default config", "error", err)
		}
		return &config, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config FullConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and fill missing fields with defaults
	config = mergeWithDefaults(config)

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *FullConfig) error {
	configPath := filepath.Join(ConfigDir, ConfigFile)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// mergeWithDefaults fills missing fields with default values
func mergeWithDefaults(config FullConfig) FullConfig {
	defaults := DefaultConfig()

	if config.Input.KeyBindings == nil {
		config.Input.KeyBindings = defaults.Input.KeyBindings
	}

	// Merge missing gameplay fields
	if config.Gameplay.DungeonWidth == 0 {
		config.Gameplay.DungeonWidth = defaults.Gameplay.DungeonWidth
	}
	if config.Gameplay.DungeonHeight == 0 {
		config.Gameplay.DungeonHeight = defaults.Gameplay.DungeonHeight
	}

	// Merge missing display fields
	if config.Display.TilesetPath == "" {
		config.Display.TilesetPath = defaults.Display.TilesetPath
	}

	return config
}

// ValidateConfig validates configuration values
func ValidateConfig(config *FullConfig) error {
	// Validate gameplay settings
	if config.Gameplay.DungeonWidth < 20 || config.Gameplay.DungeonWidth > 200 {
		return fmt.Errorf("dungeon width must be between 20 and 200")
	}

	if config.Gameplay.DungeonHeight < 10 || config.Gameplay.DungeonHeight > 100 {
		return fmt.Errorf("dungeon height must be between 10 and 100")
	}

	// Validate gameplay settings
	if config.Gameplay.FOVRadius < 1 || config.Gameplay.FOVRadius > 20 {
		return fmt.Errorf("FOV radius must be between 1 and 20")
	}

	// Validate display settings
	if config.Display.ScaleFactorX < 0.1 || config.Display.ScaleFactorX > 5.0 {
		return fmt.Errorf("scale factor X must be between 0.1 and 5.0")
	}

	if config.Display.ScaleFactorY < 0.1 || config.Display.ScaleFactorY > 5.0 {
		return fmt.Errorf("scale factor Y must be between 0.1 and 5.0")
	}
	if config.Gameplay.AnimationSpeed < 1 || config.Gameplay.AnimationSpeed > 10 {
		return fmt.Errorf("animation speed must be between 1 and 10")
	}

	// Validate display settings
	if config.Display.WindowWidth < 640 || config.Display.WindowHeight < 480 {
		return fmt.Errorf("window size must be at least 640x480")
	}

	// Validate audio settings
	if config.Audio.MasterVolume < 0 || config.Audio.MasterVolume > 1 {
		return fmt.Errorf("master volume must be between 0 and 1")
	}

	return nil
}
