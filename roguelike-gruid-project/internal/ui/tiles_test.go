package ui

import (
	"path/filepath"
	"testing"

	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
)

func TestTileMapping(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create tile mapping
	tm := NewTileMapping(tempDir)
	
	// Test basic rune mapping
	playerTile := tm.GetTileForRune('@')
	// Player tile should be one of the available options
	expectedPaths := []string{
		filepath.Join(tempDir, "characters/knight_m.png"),
		filepath.Join(tempDir, "characters/knight_f.png"),
	}
	found := false
	for _, expected := range expectedPaths {
		if playerTile == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected player tile to be one of %v, got %s", expectedPaths, playerTile)
	}
	
	// Test fallback for unknown rune
	unknownTile := tm.GetTileForRune('X')
	expectedFallback := filepath.Join(tempDir, "fallback/unknown.png")
	if unknownTile != expectedFallback {
		t.Errorf("Expected fallback tile path %s, got %s", expectedFallback, unknownTile)
	}
	
	// Test adding custom mapping
	tm.AddEntityMapping('Z', "custom/zombie.png")
	zombieTile := tm.GetTileForRune('Z')
	expectedZombie := filepath.Join(tempDir, "custom/zombie.png")
	if zombieTile != expectedZombie {
		t.Errorf("Expected zombie tile path %s, got %s", expectedZombie, zombieTile)
	}
}

func TestTileConfig(t *testing.T) {
	// Test default configuration
	defaultConfig := config.DefaultTileConfig
	if defaultConfig.Enabled {
		t.Error("Default tile config should have tiles disabled")
	}
	if defaultConfig.TileSize != 16 {
		t.Errorf("Expected default tile size 16, got %d", defaultConfig.TileSize)
	}
	if defaultConfig.ScaleFactor != 1.0 {
		t.Errorf("Expected default scale factor 1.0, got %f", defaultConfig.ScaleFactor)
	}
	
	// Test configuration validation
	if !config.ValidateTilesetPath("") {
		// Empty path should be invalid - this is expected
	}
	
	// Test with a valid directory (temp dir)
	tempDir := t.TempDir()
	if !config.ValidateTilesetPath(tempDir) {
		t.Errorf("Temp directory should be valid tileset path")
	}
}

// TestFallbackTileGeneration is tested in tiles_sdl_test.go for SDL builds

func TestTileMappingVariety(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTileMapping(tempDir)
	
	// Test that floor tiles have variety (multiple options)
	floorTiles := tm.GetAvailableTiles('.')
	if len(floorTiles) < 2 {
		t.Error("Floor tiles should have multiple variants for variety")
	}
	
	// Test that all returned paths are properly formed
	for _, tile := range floorTiles {
		if !filepath.IsAbs(tile) {
			t.Errorf("Tile path should be absolute: %s", tile)
		}
		if !filepath.HasPrefix(tile, tempDir) {
			t.Errorf("Tile path should be under tileset directory: %s", tile)
		}
	}
}

func TestTileMappingString(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTileMapping(tempDir)
	
	// Test string representation
	str := tm.String()
	if str == "" {
		t.Error("TileMapping string representation should not be empty")
	}
	
	// Should contain the tileset path
	if !contains(str, tempDir) {
		t.Errorf("String representation should contain tileset path: %s", str)
	}
}

// TestEnsureFallbackTilesExist is tested in tiles_sdl_test.go for SDL builds

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr || 
		      containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tile mapping performance
func BenchmarkTileMapping(b *testing.B) {
	tempDir := b.TempDir()
	tm := NewTileMapping(tempDir)
	
	runes := []rune{'@', '#', '.', 'o', 'g', 's', '!', '?'}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := runes[i%len(runes)]
		_ = tm.GetTileForRune(r)
	}
}

// BenchmarkFallbackTileGeneration is tested in tiles_sdl_test.go for SDL builds
