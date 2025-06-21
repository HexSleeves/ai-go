//go:build !js
// +build !js

package ui

import (
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"os"
	"sync"

	"codeberg.org/anaseto/gruid"
)

// SpriteAtlas manages a spritesheet and extracts individual sprites
type SpriteAtlas struct {
	spritesheet image.Image
	tileSize    int
	tilesPerRow int
	tilesPerCol int
	margin      int
	cache       map[gruid.Point]image.Image
	mutex       sync.RWMutex
}

// NewSpriteAtlas creates a new sprite atlas from a spritesheet file
// Parameters:
//   - spritesheetPath: path to the spritesheet image file
//   - tileSize: size of each individual sprite in pixels (assumes square sprites)
//   - margin: margin/padding around the spritesheet edges in pixels
func NewSpriteAtlas(spritesheetPath string, tileSize int, margin int) (*SpriteAtlas, error) {
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
	tilesPerRow := (bounds.Dx() - margin) / tileSize
	tilesPerCol := (bounds.Dy() - margin) / tileSize

	slog.Info("Loaded spritesheet: %dx%d pixels, %dx%d tiles (%d total)",
		bounds.Dx(), bounds.Dy(), tilesPerRow, tilesPerCol, tilesPerRow*tilesPerCol)

	return &SpriteAtlas{
		spritesheet: img,
		tileSize:    tileSize,
		tilesPerRow: tilesPerRow,
		tilesPerCol: tilesPerCol,
		margin:      margin,
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
		slog.Warn("Sprite coordinates out of bounds: (%d,%d), max: (%d,%d)",
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
	// Calculate pixel coordinates with margin offset
	startX := sa.margin + x*sa.tileSize
	startY := sa.margin + y*sa.tileSize
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
		slog.Warn("Sprite index out of bounds: %d, max: %d",
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
// The returned dimensions account for any margin specified during atlas creation
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
	slog.Debug("Sprite atlas cache cleared")
}

// PreloadSprites preloads commonly used sprites into cache
func (sa *SpriteAtlas) PreloadSprites(coords []gruid.Point) {
	slog.Info("Preloading %d sprites into cache", len(coords))
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
// Uses 16x16 pixel tiles with a 1-pixel margin around the spritesheet edges
func NewKenneyRoguelikeAtlas(spritesheetPath string) (*KenneyRoguelikeAtlas, error) {
	atlas, err := NewSpriteAtlas(spritesheetPath, 16, 1) // Kenney's tiles are 16x16 with 1px margin
	if err != nil {
		return nil, err
	}

	return &KenneyRoguelikeAtlas{atlas: atlas}, nil
}

// Common sprite coordinates for Kenney's Roguelike pack (colored-transparent_packed.png)
// Coordinates are based on the 57x31 grid layout with 16x16 pixel tiles
var (
	// Characters - Top rows contain various character types
	KenneyPlayer    = gruid.Point{X: 28, Y: 0} // Knight character (armored)
	KenneyPlayerAlt = gruid.Point{X: 29, Y: 0} // Alternative knight
	KenneyRogue     = gruid.Point{X: 30, Y: 0} // Rogue character
	KenneyMage      = gruid.Point{X: 31, Y: 0} // Mage character

	// Monsters - Various creatures throughout the sheet
	KenneyOrc      = gruid.Point{X: 26, Y: 2}  // Orc warrior
	KenneyGoblin   = gruid.Point{X: 13, Y: 19} // Goblin
	KenneySkeleton = gruid.Point{X: 15, Y: 19} // Skeleton
	KenneyDragon   = gruid.Point{X: 28, Y: 8}  // Dragon (large creature)
	KenneyBat      = gruid.Point{X: 46, Y: 18} // Bat
	KenneySpider   = gruid.Point{X: 47, Y: 18} // Spider
	KenneyRat      = gruid.Point{X: 45, Y: 18} // Rat
	KenneySnake    = gruid.Point{X: 44, Y: 18} // Snake

	// Environment tiles - Walls, floors, doors scattered throughout
	KenneyWall       = gruid.Point{X: 0, Y: 13} // Stone wall
	KenneyWallAlt    = gruid.Point{X: 2, Y: 2}  // Alternative wall
	KenneyFloor      = gruid.Point{X: 2, Y: 0}  // Stone floor
	KenneyFloorAlt   = gruid.Point{X: 3, Y: 0}  // Alternative floor
	KenneyDoor       = gruid.Point{X: 4, Y: 2}  // Closed door
	KenneyDoorOpen   = gruid.Point{X: 5, Y: 2}  // Open door
	KenneyStairsDown = gruid.Point{X: 6, Y: 2}  // Stairs going down
	KenneyStairsUp   = gruid.Point{X: 7, Y: 2}  // Stairs going up

	// Items - Potions, weapons, armor, etc.
	KenneyPotionRed   = gruid.Point{X: 39, Y: 11} // Red health potion
	KenneyPotionBlue  = gruid.Point{X: 42, Y: 11} // Blue mana potion
	KenneyPotionGreen = gruid.Point{X: 43, Y: 11} // Green poison potion
	KenneyScroll      = gruid.Point{X: 33, Y: 15} // Magic scroll
	KenneyBook        = gruid.Point{X: 34, Y: 15} // Spellbook
	KenneySword       = gruid.Point{X: 32, Y: 8}  // Iron sword
	KenneySwordGold   = gruid.Point{X: 0, Y: 19}  // Golden sword
	KenneyDagger      = gruid.Point{X: 1, Y: 19}  // Dagger
	KenneyBow         = gruid.Point{X: 2, Y: 19}  // Bow
	KenneyShield      = gruid.Point{X: 3, Y: 19}  // Wooden shield
	KenneyShieldMetal = gruid.Point{X: 4, Y: 19}  // Metal shield
	KenneyArmor       = gruid.Point{X: 5, Y: 19}  // Chest armor
	KenneyHelmet      = gruid.Point{X: 6, Y: 19}  // Helmet
	KenneyCoin        = gruid.Point{X: 8, Y: 17}  // Gold coin
	KenneyGem         = gruid.Point{X: 9, Y: 17}  // Precious gem
	KenneyKey         = gruid.Point{X: 10, Y: 17} // Key
	KenneyChest       = gruid.Point{X: 11, Y: 17} // Treasure chest
	KenneyFood        = gruid.Point{X: 12, Y: 17} // Food/bread
	KenneyRing        = gruid.Point{X: 13, Y: 17} // Magic ring
	KenneyAmulet      = gruid.Point{X: 14, Y: 17} // Amulet
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

	// Monsters - More variety with specific sprites
	'o': KenneyOrc,
	'g': KenneyGoblin,
	's': KenneySkeleton,
	'D': KenneyDragon,
	'b': KenneyBat,    // Bat
	'S': KenneySpider, // Spider (capital S for larger creature)
	'r': KenneyRat,    // Rat
	'~': KenneySnake,  // Snake (using ~ for serpentine movement)

	// Items - Using appropriate sprites for each type
	'!':  KenneyPotionRed, // Health potions (red)
	'?':  KenneyScroll,    // Scrolls and books
	'/':  KenneySword,     // Sword
	'\\': KenneyDagger,    // Dagger (now using proper dagger sprite)
	')':  KenneyBow,       // Bow (now using proper bow sprite)
	'[':  KenneyArmor,     // Armor (now using proper armor sprite)
	']':  KenneyShield,    // Shield
	'$':  KenneyCoin,      // Gold coins
	'*':  KenneyGem,       // Gems (now using proper gem sprite)
	'%':  KenneyFood,      // Food (now using proper food sprite)
	'=':  KenneyRing,      // Ring (now using proper ring sprite)
	'"':  KenneyAmulet,    // Amulet (now using proper amulet sprite)
	'&':  KenneyChest,     // Treasure chest
	'-':  KenneyKey,       // Key

	// Additional item variations
	'¡': KenneyPotionBlue,  // Mana potions (inverted !)
	'¿': KenneyBook,        // Spellbooks (inverted ?)
	'†': KenneySwordGold,   // Special/magical sword
	'‡': KenneyShieldMetal, // Metal/magical shield
	'°': KenneyHelmet,      // Helmet
}

// GetPlayerSprite returns the player character sprite
func (kra *KenneyRoguelikeAtlas) GetPlayerSprite() image.Image {
	return kra.atlas.GetSprite(KenneyPlayer.X, KenneyPlayer.Y)
}

// GetPlayerSpriteByClass returns a player sprite based on character class
func (kra *KenneyRoguelikeAtlas) GetPlayerSpriteByClass(class string) image.Image {
	var coord gruid.Point
	switch class {
	case "knight", "warrior", "fighter":
		coord = KenneyPlayer
	case "knight_alt", "paladin":
		coord = KenneyPlayerAlt
	case "rogue", "thief", "assassin":
		coord = KenneyRogue
	case "mage", "wizard", "sorcerer":
		coord = KenneyMage
	default:
		coord = KenneyPlayer // Default to knight
	}
	return kra.atlas.GetSprite(coord.X, coord.Y)
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
	case "bat":
		coord = KenneyBat
	case "spider":
		coord = KenneySpider
	case "rat":
		coord = KenneyRat
	case "snake":
		coord = KenneySnake
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
	case "wall_alt":
		coord = KenneyWallAlt
	case "floor":
		coord = KenneyFloor
	case "floor_alt":
		coord = KenneyFloorAlt
	case "door":
		coord = KenneyDoor
	case "door_open":
		coord = KenneyDoorOpen
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
	case "potion", "health_potion":
		coord = KenneyPotionRed
	case "mana_potion":
		coord = KenneyPotionBlue
	case "poison_potion":
		coord = KenneyPotionGreen
	case "scroll":
		coord = KenneyScroll
	case "book", "spellbook":
		coord = KenneyBook
	case "sword":
		coord = KenneySword
	case "golden_sword", "magic_sword":
		coord = KenneySwordGold
	case "dagger":
		coord = KenneyDagger
	case "bow":
		coord = KenneyBow
	case "shield":
		coord = KenneyShield
	case "metal_shield", "magic_shield":
		coord = KenneyShieldMetal
	case "armor":
		coord = KenneyArmor
	case "helmet":
		coord = KenneyHelmet
	case "coin", "gold":
		coord = KenneyCoin
	case "gem":
		coord = KenneyGem
	case "key":
		coord = KenneyKey
	case "chest":
		coord = KenneyChest
	case "food":
		coord = KenneyFood
	case "ring":
		coord = KenneyRing
	case "amulet":
		coord = KenneyAmulet
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
