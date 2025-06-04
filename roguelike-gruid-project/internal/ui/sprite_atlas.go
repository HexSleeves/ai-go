//go:build !js
// +build !js

package ui

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"sync"

	"codeberg.org/anaseto/gruid"
	"github.com/sirupsen/logrus"
)

// SpriteAtlas manages a spritesheet and extracts individual sprites
type SpriteAtlas struct {
	spritesheet image.Image
	tileSize    int
	tilesPerRow int
	tilesPerCol int
	cache       map[gruid.Point]image.Image
	mutex       sync.RWMutex
}

// NewSpriteAtlas creates a new sprite atlas from a spritesheet file
func NewSpriteAtlas(spritesheetPath string, tileSize int) (*SpriteAtlas, error) {
	// Load the spritesheet image
	file, err := os.Open(spritesheetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open spritesheet %s: %w", spritesheetPath, err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode spritesheet PNG %s: %w", spritesheetPath, err)
	}

	bounds := img.Bounds()
	tilesPerRow := bounds.Dx() / tileSize
	tilesPerCol := bounds.Dy() / tileSize

	logrus.Infof("Loaded spritesheet: %dx%d pixels, %dx%d tiles (%d total)",
		bounds.Dx(), bounds.Dy(), tilesPerRow, tilesPerCol, tilesPerRow*tilesPerCol)

	return &SpriteAtlas{
		spritesheet: img,
		tileSize:    tileSize,
		tilesPerRow: tilesPerRow,
		tilesPerCol: tilesPerCol,
		cache:       make(map[gruid.Point]image.Image),
	}, nil
}

// GetSprite extracts a sprite at the given grid coordinates (x, y)
// Coordinates are 0-based, where (0,0) is the top-left sprite
func (sa *SpriteAtlas) GetSprite(x, y int) image.Image {
	coord := gruid.Point{X: x, Y: y}

	// Check cache first
	sa.mutex.RLock()
	if sprite, exists := sa.cache[coord]; exists {
		sa.mutex.RUnlock()
		return sprite
	}
	sa.mutex.RUnlock()

	// Validate coordinates
	if x < 0 || x >= sa.tilesPerRow || y < 0 || y >= sa.tilesPerCol {
		logrus.Warnf("Sprite coordinates out of bounds: (%d,%d), max: (%d,%d)",
			x, y, sa.tilesPerRow-1, sa.tilesPerCol-1)
		return nil
	}

	// Extract sprite from spritesheet
	sprite := sa.extractSprite(x, y)

	// Cache the extracted sprite
	sa.mutex.Lock()
	sa.cache[coord] = sprite
	sa.mutex.Unlock()

	return sprite
}

// extractSprite extracts a single sprite from the spritesheet
func (sa *SpriteAtlas) extractSprite(x, y int) image.Image {
	// Calculate pixel coordinates
	startX := x * sa.tileSize
	startY := y * sa.tileSize
	endX := startX + sa.tileSize
	endY := startY + sa.tileSize

	// Create a new image for the sprite
	sprite := image.NewRGBA(image.Rect(0, 0, sa.tileSize, sa.tileSize))

	// Copy pixels from spritesheet to sprite
	for py := startY; py < endY; py++ {
		for px := startX; px < endX; px++ {
			color := sa.spritesheet.At(px, py)
			sprite.Set(px-startX, py-startY, color)
		}
	}

	return sprite
}

// GetSpriteByIndex extracts a sprite by linear index (0-based)
// Index 0 is top-left, increases left-to-right, then top-to-bottom
func (sa *SpriteAtlas) GetSpriteByIndex(index int) image.Image {
	if index < 0 || index >= sa.tilesPerRow*sa.tilesPerCol {
		logrus.Warnf("Sprite index out of bounds: %d, max: %d",
			index, sa.tilesPerRow*sa.tilesPerCol-1)
		return nil
	}

	x := index % sa.tilesPerRow
	y := index / sa.tilesPerRow
	return sa.GetSprite(x, y)
}

// GetTileSize returns the size of individual tiles
func (sa *SpriteAtlas) GetTileSize() int {
	return sa.tileSize
}

// GetDimensions returns the dimensions of the spritesheet in tiles
func (sa *SpriteAtlas) GetDimensions() (tilesPerRow, tilesPerCol int) {
	return sa.tilesPerRow, sa.tilesPerCol
}

// GetTotalSprites returns the total number of sprites in the atlas
func (sa *SpriteAtlas) GetTotalSprites() int {
	return sa.tilesPerRow * sa.tilesPerCol
}

// ClearCache clears the sprite cache to free memory
func (sa *SpriteAtlas) ClearCache() {
	sa.mutex.Lock()
	defer sa.mutex.Unlock()
	sa.cache = make(map[gruid.Point]image.Image)
	logrus.Debug("Sprite atlas cache cleared")
}

// PreloadSprites preloads commonly used sprites into cache
func (sa *SpriteAtlas) PreloadSprites(coords []gruid.Point) {
	logrus.Infof("Preloading %d sprites into cache", len(coords))
	for _, coord := range coords {
		sa.GetSprite(coord.X, coord.Y)
	}
}

// SaveSprite saves a specific sprite to a file (useful for debugging)
func (sa *SpriteAtlas) SaveSprite(x, y int, filename string) error {
	sprite := sa.GetSprite(x, y)
	if sprite == nil {
		return fmt.Errorf("sprite at (%d,%d) not found", x, y)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, sprite)
}

// KenneyRoguelikeAtlas provides predefined coordinates for Kenney's Roguelike pack
type KenneyRoguelikeAtlas struct {
	atlas *SpriteAtlas
}

// NewKenneyRoguelikeAtlas creates an atlas specifically for Kenney's Roguelike pack
func NewKenneyRoguelikeAtlas(spritesheetPath string) (*KenneyRoguelikeAtlas, error) {
	atlas, err := NewSpriteAtlas(spritesheetPath, 16) // Kenney's tiles are 16x16
	if err != nil {
		return nil, err
	}

	return &KenneyRoguelikeAtlas{atlas: atlas}, nil
}

// Common sprite coordinates for Kenney's Roguelike pack
// These are approximate - you'll need to adjust based on the actual spritesheet layout
var (
	// Characters (approximate positions - adjust based on actual spritesheet)
	KenneyPlayer    = gruid.Point{X: 28, Y: 0} // Knight character
	KenneyPlayerAlt = gruid.Point{X: 29, Y: 0} // Alternative player

	// Monsters (approximate positions)
	KenneyOrc      = gruid.Point{X: 14, Y: 19} // Orc
	KenneyGoblin   = gruid.Point{X: 13, Y: 19} // Goblin
	KenneySkeleton = gruid.Point{X: 15, Y: 19} // Skeleton
	KenneyDragon   = gruid.Point{X: 24, Y: 18} // Dragon

	// Environment (approximate positions)
	KenneyWall       = gruid.Point{X: 1, Y: 2} // Wall
	KenneyFloor      = gruid.Point{X: 0, Y: 2} // Floor
	KenneyDoor       = gruid.Point{X: 2, Y: 2} // Door
	KenneyStairsDown = gruid.Point{X: 3, Y: 2} // Stairs down
	KenneyStairsUp   = gruid.Point{X: 4, Y: 2} // Stairs up

	// Items (approximate positions)
	KenneyPotionRed = gruid.Point{X: 9, Y: 23} // Red potion
	KenneyScroll    = gruid.Point{X: 6, Y: 23} // Scroll
	KenneySword     = gruid.Point{X: 0, Y: 29} // Sword
	KenneyShield    = gruid.Point{X: 6, Y: 29} // Shield
	KenneyCoin      = gruid.Point{X: 9, Y: 26} // Gold coin
)

// RuneToSpriteMapping maps game runes directly to sprite atlas coordinates
var RuneToSpriteMapping = map[rune]gruid.Point{
	// Player character
	'@': KenneyPlayer,

	// Environment tiles
	'#': KenneyWall,
	'.': KenneyFloor,
	'+': KenneyDoor,
	'>': KenneyStairsDown,
	'<': KenneyStairsUp,
	' ': KenneyFloor, // Empty space shows floor

	// Monsters
	'o': KenneyOrc,
	'g': KenneyGoblin,
	's': KenneySkeleton,
	'D': KenneyDragon,

	// Items
	'!': KenneyPotionRed, // Potions
	'?': KenneyScroll,    // Scrolls
	'/': KenneySword,     // Sword
	'\\': KenneySword,    // Dagger (using sword sprite)
	')': KenneySword,     // Bow (using sword sprite for now)
	'[': KenneyShield,    // Armor (using shield sprite)
	']': KenneyShield,    // Shield
	'$': KenneyCoin,      // Gold
	'*': KenneyCoin,      // Gems (using coin sprite)
	'%': KenneyPotionRed, // Food (using potion sprite)
	'=': KenneyCoin,      // Ring (using coin sprite)
	'"': KenneyCoin,      // Amulet (using coin sprite)
}

// GetPlayerSprite returns the player character sprite
func (kra *KenneyRoguelikeAtlas) GetPlayerSprite() image.Image {
	return kra.atlas.GetSprite(KenneyPlayer.X, KenneyPlayer.Y)
}

// GetMonsterSprite returns a monster sprite by type
func (kra *KenneyRoguelikeAtlas) GetMonsterSprite(monsterType string) image.Image {
	var coord gruid.Point
	switch monsterType {
	case "orc":
		coord = KenneyOrc
	case "goblin":
		coord = KenneyGoblin
	case "skeleton":
		coord = KenneySkeleton
	case "dragon":
		coord = KenneyDragon
	default:
		coord = KenneyOrc // Default to orc
	}
	return kra.atlas.GetSprite(coord.X, coord.Y)
}

// GetEnvironmentSprite returns an environment sprite by type
func (kra *KenneyRoguelikeAtlas) GetEnvironmentSprite(envType string) image.Image {
	var coord gruid.Point
	switch envType {
	case "wall":
		coord = KenneyWall
	case "floor":
		coord = KenneyFloor
	case "door":
		coord = KenneyDoor
	case "stairs_down":
		coord = KenneyStairsDown
	case "stairs_up":
		coord = KenneyStairsUp
	default:
		coord = KenneyFloor // Default to floor
	}
	return kra.atlas.GetSprite(coord.X, coord.Y)
}

// GetItemSprite returns an item sprite by type
func (kra *KenneyRoguelikeAtlas) GetItemSprite(itemType string) image.Image {
	var coord gruid.Point
	switch itemType {
	case "potion":
		coord = KenneyPotionRed
	case "scroll":
		coord = KenneyScroll
	case "sword":
		coord = KenneySword
	case "shield":
		coord = KenneyShield
	case "coin":
		coord = KenneyCoin
	default:
		coord = KenneyPotionRed // Default to potion
	}
	return kra.atlas.GetSprite(coord.X, coord.Y)
}

// GetSpriteForRune returns a sprite for the given rune using the direct mapping
func (kra *KenneyRoguelikeAtlas) GetSpriteForRune(r rune) image.Image {
	if coord, exists := RuneToSpriteMapping[r]; exists {
		return kra.atlas.GetSprite(coord.X, coord.Y)
	}
	// Return floor sprite as fallback
	return kra.atlas.GetSprite(KenneyFloor.X, KenneyFloor.Y)
}

// GetAtlas returns the underlying sprite atlas
func (kra *KenneyRoguelikeAtlas) GetAtlas() *SpriteAtlas {
	return kra.atlas
}
