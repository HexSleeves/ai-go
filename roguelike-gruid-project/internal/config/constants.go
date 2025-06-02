package config

// Game settings & map generation constants
const (
	DungeonWidth  = 80
	DungeonHeight = 24
	FovRadius     = 10 // How far the player can see
)

// UI Layout constants
const (
	// Main game viewport (map display area)
	MapViewportWidth  = 60
	MapViewportHeight = 16
	MapViewportX      = 0
	MapViewportY      = 0

	// Stats panel (top-right)
	StatsPanelWidth  = 20
	StatsPanelHeight = 12
	StatsPanelX      = 60
	StatsPanelY      = 0

	// Message log panel (bottom)
	MessageLogWidth  = 80
	MessageLogHeight = 8
	MessageLogX      = 0
	MessageLogY      = 16

	// UI styling
	BorderStyle = "single" // "single", "double", "none"
)
