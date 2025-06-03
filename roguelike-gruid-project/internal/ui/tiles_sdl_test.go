//go:build js || sdl
// +build js sdl

package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFallbackTileGeneration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Test fallback tile creation
	err := CreateFallbackTiles(tempDir, 16)
	if err != nil {
		t.Fatalf("Failed to create fallback tiles: %v", err)
	}
	
	// Check that fallback directory was created
	fallbackDir := filepath.Join(tempDir, "fallback")
	if _, err := os.Stat(fallbackDir); os.IsNotExist(err) {
		t.Error("Fallback directory was not created")
	}
	
	// Check that unknown tile was created
	unknownTile := filepath.Join(fallbackDir, "unknown.png")
	if _, err := os.Stat(unknownTile); os.IsNotExist(err) {
		t.Error("Unknown fallback tile was not created")
	}
	
	// Check that player tile was created
	playerTile := filepath.Join(fallbackDir, "player.png")
	if _, err := os.Stat(playerTile); os.IsNotExist(err) {
		t.Error("Player fallback tile was not created")
	}
}

func TestEnsureFallbackTilesExist(t *testing.T) {
	tempDir := t.TempDir()
	
	// Initially, no fallback tiles should exist
	fallbackDir := filepath.Join(tempDir, "fallback")
	unknownTile := filepath.Join(fallbackDir, "unknown.png")
	
	if _, err := os.Stat(unknownTile); !os.IsNotExist(err) {
		t.Error("Unknown tile should not exist initially")
	}
	
	// Call EnsureFallbackTilesExist
	EnsureFallbackTilesExist(tempDir, 16)
	
	// Now fallback tiles should exist
	if _, err := os.Stat(unknownTile); os.IsNotExist(err) {
		t.Error("Unknown tile should exist after calling EnsureFallbackTilesExist")
	}
	
	// Calling again should not cause errors
	EnsureFallbackTilesExist(tempDir, 16)
}

// Benchmark fallback tile generation
func BenchmarkFallbackTileGeneration(b *testing.B) {
	tempDir := b.TempDir()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clean up previous run
		os.RemoveAll(filepath.Join(tempDir, "fallback"))
		
		err := CreateFallbackTiles(tempDir, 16)
		if err != nil {
			b.Fatalf("Failed to create fallback tiles: %v", err)
		}
	}
}
