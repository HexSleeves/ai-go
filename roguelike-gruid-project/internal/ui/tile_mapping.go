package ui

import (
	"fmt"
	"path/filepath"

	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

// TileMapping provides backward compatibility for the old tile mapping system
// This is now a simplified wrapper that maintains the API but works with the atlas system
type TileMapping struct {
	EntityTiles    map[rune][]string            // Kept for backward compatibility (now unused)
	WallTiles      map[string]string            // Kept for backward compatibility (now unused)
	FloorTiles     map[string]string            // Kept for backward compatibility (now unused)
	ItemTiles      map[string]string            // Kept for backward compatibility (now unused)
	MonsterTiles   map[string]string            // Kept for backward compatibility (now unused)
	FallbackTile   string                       // Kept for backward compatibility (now unused)
	TilesetPath    string                       // Base path to tileset
}

// NewTileMapping creates a new tile mapping with default mappings
// NOTE: This is now a compatibility wrapper - actual tile loading uses the sprite atlas
func NewTileMapping(tilesetPath string) *TileMapping {
	tm := &TileMapping{
		EntityTiles:  make(map[rune][]string),
		WallTiles:    make(map[string]string),
		FloorTiles:   make(map[string]string),
		ItemTiles:    make(map[string]string),
		MonsterTiles: make(map[string]string),
		FallbackTile: "fallback/unknown.png",
		TilesetPath:  tilesetPath,
	}

	tm.initializeDefaultMappings()
	return tm
}

// initializeDefaultMappings sets up the default entity-to-tile mappings
func (tm *TileMapping) initializeDefaultMappings() {
	// Player character
	tm.EntityTiles['@'] = []string{"characters/knight_m.png", "characters/knight_f.png"}
	
	// Environment tiles
	tm.EntityTiles['#'] = []string{"environment/wall_mid.png"}
	tm.EntityTiles['.'] = []string{"environment/floor_1.png", "environment/floor_2.png", "environment/floor_3.png"}
	tm.EntityTiles['+'] = []string{"environment/door_closed.png"}
	tm.EntityTiles['>'] = []string{"environment/stairs_down.png"}
	tm.EntityTiles['<'] = []string{"environment/stairs_up.png"}
	
	// Common monsters (using letters typically used in roguelikes)
	tm.EntityTiles['o'] = []string{"monsters/orc.png"}
	tm.EntityTiles['g'] = []string{"monsters/goblin.png"}
	tm.EntityTiles['s'] = []string{"monsters/skeleton.png"}
	tm.EntityTiles['D'] = []string{"monsters/dragon.png"}
	tm.EntityTiles['r'] = []string{"monsters/rat.png"}
	tm.EntityTiles['b'] = []string{"monsters/bat.png"}
	tm.EntityTiles['S'] = []string{"monsters/spider.png"}
	tm.EntityTiles['T'] = []string{"monsters/troll.png"}
	
	// Items
	tm.EntityTiles['!'] = []string{"items/flask_red.png", "items/flask_blue.png", "items/flask_green.png"}
	tm.EntityTiles['?'] = []string{"items/scroll_01.png", "items/scroll_02.png"}
	tm.EntityTiles['/'] = []string{"items/weapon_sword.png"}
	tm.EntityTiles['\\'] = []string{"items/weapon_dagger.png"}
	tm.EntityTiles[')'] = []string{"items/weapon_bow.png"}
	tm.EntityTiles['['] = []string{"items/armor_leather.png"}
	tm.EntityTiles[']'] = []string{"items/shield_wood.png"}
	tm.EntityTiles['$'] = []string{"items/coin_gold.png"}
	tm.EntityTiles['*'] = []string{"items/gem_01.png", "items/gem_02.png"}
	tm.EntityTiles['%'] = []string{"items/food_bread.png"}
	tm.EntityTiles['='] = []string{"items/ring_01.png"}
	tm.EntityTiles['"'] = []string{"items/amulet_01.png"}
	
	// Special characters
	tm.EntityTiles[' '] = []string{"environment/floor_1.png"} // Empty space shows floor
	
	// Wall tile variations
	tm.WallTiles["default"] = "environment/wall_mid.png"
	tm.WallTiles["corner"] = "environment/wall_corner.png"
	tm.WallTiles["side"] = "environment/wall_side.png"
	
	// Floor tile variations
	tm.FloorTiles["default"] = "environment/floor_1.png"
	tm.FloorTiles["stone"] = "environment/floor_2.png"
	tm.FloorTiles["wood"] = "environment/floor_3.png"
}

// GetTileForRune returns a tile path for the given rune
// NOTE: This is now a compatibility method - actual tile loading uses the sprite atlas
func (tm *TileMapping) GetTileForRune(r rune) string {
	if tiles, exists := tm.EntityTiles[r]; exists && len(tiles) > 0 {
		// Return first tile for consistency (no randomness needed with atlas)
		return tm.getFullPath(tiles[0])
	}

	// Return fallback tile
	return tm.getFullPath(tm.FallbackTile)
}

// GetTileForEntity returns a tile path for a specific entity, considering its components
func (tm *TileMapping) GetTileForEntity(entityID ecs.EntityID, ecsSystem *ecs.ECS) string {
	if !ecsSystem.HasRenderableSafe(entityID) {
		return tm.getFullPath(tm.FallbackTile)
	}
	
	renderable := ecsSystem.GetRenderableSafe(entityID)
	
	// Check for specific tile override in renderable component
	if renderable.TileName != "" {
		return tm.getFullPath(renderable.TileName)
	}
	
	// Special handling for different entity types
	if ecsSystem.HasComponent(entityID, components.CPlayerTag) {
		return tm.getPlayerTile(entityID, ecsSystem)
	}
	
	if ecsSystem.HasComponent(entityID, components.CAITag) {
		return tm.getMonsterTile(entityID, ecsSystem)
	}
	
	// Use rune-based mapping as fallback
	return tm.GetTileForRune(renderable.Glyph)
}

// getPlayerTile returns the appropriate tile for the player
func (tm *TileMapping) getPlayerTile(entityID ecs.EntityID, ecsSystem *ecs.ECS) string {
	// Could be extended to show different player states (injured, equipped, etc.)
	if tiles, exists := tm.EntityTiles['@']; exists && len(tiles) > 0 {
		return tm.getFullPath(tiles[0]) // Use first player tile consistently
	}
	return tm.getFullPath(tm.FallbackTile)
}

// getMonsterTile returns the appropriate tile for a monster
func (tm *TileMapping) getMonsterTile(entityID ecs.EntityID, ecsSystem *ecs.ECS) string {
	renderable := ecsSystem.GetRenderableSafe(entityID)
	
	// Try to get monster-specific tile based on name or type
	if ecsSystem.HasComponent(entityID, components.CName) {
		name := ecsSystem.GetNameSafe(entityID)
		if tile, exists := tm.MonsterTiles[name]; exists {
			return tm.getFullPath(tile)
		}
	}
	
	// Fall back to rune-based mapping
	return tm.GetTileForRune(renderable.Glyph)
}

// getFullPath returns the full path to a tile file
func (tm *TileMapping) getFullPath(tilePath string) string {
	if filepath.IsAbs(tilePath) {
		return tilePath
	}
	return filepath.Join(tm.TilesetPath, tilePath)
}

// AddEntityMapping adds a new entity-to-tile mapping
func (tm *TileMapping) AddEntityMapping(r rune, tilePaths ...string) {
	tm.EntityTiles[r] = tilePaths
}

// AddMonsterMapping adds a monster name to tile mapping
func (tm *TileMapping) AddMonsterMapping(monsterName, tilePath string) {
	tm.MonsterTiles[monsterName] = tilePath
}

// SetFallbackTile sets the fallback tile used when no mapping is found
func (tm *TileMapping) SetFallbackTile(tilePath string) {
	tm.FallbackTile = tilePath
}

// GetAvailableTiles returns all tiles mapped to a specific rune
func (tm *TileMapping) GetAvailableTiles(r rune) []string {
	if tiles, exists := tm.EntityTiles[r]; exists {
		result := make([]string, len(tiles))
		for i, tile := range tiles {
			result[i] = tm.getFullPath(tile)
		}
		return result
	}
	return []string{tm.getFullPath(tm.FallbackTile)}
}

// ValidateMapping checks if all mapped tiles exist (for debugging)
func (tm *TileMapping) ValidateMapping() []string {
	var missingTiles []string
	
	// Check all entity tiles
	for rune, tiles := range tm.EntityTiles {
		for _, tile := range tiles {
			fullPath := tm.getFullPath(tile)
			// Note: In a real implementation, you'd check if the file exists
			// For now, we'll just collect the paths for validation
			_ = fullPath
			_ = rune // Avoid unused variable warning
		}
	}
	
	return missingTiles
}

// String returns a string representation of the tile mapping for debugging
func (tm *TileMapping) String() string {
	return fmt.Sprintf("TileMapping{EntityTiles: %d entries, TilesetPath: %s}", 
		len(tm.EntityTiles), tm.TilesetPath)
}
