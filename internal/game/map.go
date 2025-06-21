package game

import (
	"log/slog"
	"math/rand"
	"slices"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/rl" // Use rl package which contains FOV
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

// Game settings & map generation constants
const (
	maxRooms           = 10
	roomMinSize        = 6
	roomMaxSize        = 10
	maxMonstersPerRoom = 2 // Max monsters per room (excluding first)
)

// TileType represents the type of a map tile.

const (
	WallCell rl.Cell = iota
	FloorCell
)

// Map represents the game map's logical state and visibility.
type Map struct {
	Grid     rl.Grid // Stores the map cells (rune, style, attributes)
	Width    int
	Height   int
	Explored []uint64 // Bitset for explored tiles (Global map knowledge)
}

// NewMap creates a new map initialized with walls and visibility data.
func NewMap(width, height int) *Map {
	m := &Map{
		Grid:     rl.NewGrid(width, height),
		Explored: make([]uint64, (width*height+63)/64),
		Width:    width,
		Height:   height,
	}

	return m
}

// generateMap creates a new map layout with rooms and tunnels, and spawns monsters.
// It now takes the game struct to access ECS and TurnQueue.
func (m *Map) generateMap(g *Game, width, height int, items map[string]components.Item) gruid.Point {
	m.Grid.Fill(WallCell)

	var rooms []Rect
	var playerStart gruid.Point = gruid.Point{X: 0, Y: 0}

	for range maxRooms {
		w := rand.Intn(roomMaxSize-roomMinSize+1) + roomMinSize
		h := rand.Intn(roomMaxSize-roomMinSize+1) + roomMinSize
		x := rand.Intn(width - w - 1)  // -1 to ensure room fits
		y := rand.Intn(height - h - 1) // -1 to ensure room fits

		newRoom := NewRect(x, y, w, h)

		// Check for intersections with existing rooms
		intersects := slices.ContainsFunc(rooms, newRoom.Intersects)

		if !intersects {
			createRoom(m.Grid, newRoom)
			newCenter := newRoom.Center()

			if len(rooms) == 0 {
				// This is the first room, place player here
				playerStart = newCenter
			} else {
				// Connect to the previous room's center
				prevCenter := rooms[len(rooms)-1].Center()

				// Randomly decide tunnel order (H then V or V then H)
				if rand.Intn(2) == 0 {
					createHTunnel(m.Grid, prevCenter.X, newCenter.X, prevCenter.Y)
					createVTunnel(m.Grid, prevCenter.Y, newCenter.Y, newCenter.X)
				} else {
					createVTunnel(m.Grid, prevCenter.Y, newCenter.Y, prevCenter.X)
					createHTunnel(m.Grid, prevCenter.X, newCenter.X, newCenter.Y)
				}
				// Spawn monsters in this room (if not the first room)
				m.placeMonsters(g, newRoom)
				// Spawn items in this room
				m.placeItems(g, newRoom, items)
			}
			rooms = append(rooms, newRoom)
		}
	}

	return playerStart
}

// InBounds checks if coordinates are within map bounds.
func (m *Map) InBounds(p gruid.Point) bool {
	return p.X >= 0 && p.X < m.Width && p.Y >= 0 && p.Y < m.Height
}

// isWalkable checks if a tile is a floor tile.
func (m *Map) isWalkable(p gruid.Point) bool {
	if !m.InBounds(p) {
		return false
	}
	return m.Grid.At(p) == FloorCell
}

// IsWall checks if the tile at the given point is a wall.
func (m *Map) IsWall(p gruid.Point) bool {
	if !m.InBounds(p) {
		return true
	}
	return m.Grid.At(p) == WallCell
}

// --- Map State Methods ---

// IsOpaque checks if a tile blocks FOV. Required by FOV calculations.
func (m *Map) IsOpaque(p gruid.Point) bool {
	// Treat out-of-bounds positions as opaque walls
	if !m.InBounds(p) {
		return true
	}

	return m.Grid.At(p) == WallCell
}

// SetExplored marks a point as explored in the global map bitset.
// This should typically only be called based on the player's FOV results.
func (m *Map) SetExplored(p gruid.Point) {
	if !m.InBounds(p) {
		return
	}

	idx := p.Y*m.Width + p.X
	sliceIdx := idx / 64
	bitIdx := uint(idx % 64)
	mask := uint64(1 << bitIdx)

	// Check bounds before writing
	if sliceIdx < len(m.Explored) {
		m.Explored[sliceIdx] |= mask
	}
}

// IsExplored checks if a point has ever been explored using the global bitset.
func (m *Map) IsExplored(p gruid.Point) bool {
	if !m.InBounds(p) {
		return false
	}
	idx := p.Y*m.Width + p.X
	sliceIdx := idx / 64
	bitIdx := uint(idx % 64)

	// Check bounds before reading
	if sliceIdx >= len(m.Explored) {
		return false // Treat out-of-bounds index as not explored
	}
	return (m.Explored[sliceIdx] & (1 << bitIdx)) != 0
}

// --- End Map State Methods ---

// Rune determines the character representation for a given map cell type.
func (m *Map) Rune(c rl.Cell) (r rune) {
	switch c {
	case WallCell:
		r = '#'
	case FloorCell:
		r = '.'
	}
	return r
}

// placeMonsters spawns monsters in a given room.
func (m *Map) placeMonsters(g *Game, room Rect) {
	// Determine number of monsters for this room (e.g., 0 to maxMonstersPerRoom)
	numMonsters := rand.Intn(maxMonstersPerRoom + 1) // +1 because Intn is exclusive upper bound
	slog.Debug("Placing monsters in room", "numMonsters", numMonsters, "room", room)

	for i := 0; i < numMonsters; i++ {
		// Find a random walkable tile within the room bounds
		// Add +1 to x1, y1 and -1 to x2, y2 to avoid spawning on walls
		x := rand.Intn(room.X2-room.X1-1) + room.X1 + 1
		y := rand.Intn(room.Y2-room.Y1-1) + room.Y1 + 1
		pos := gruid.Point{X: x, Y: y}

		// Check if the tile is walkable and not already occupied
		if m.isWalkable(pos) && len(g.ecs.EntitiesAt(pos)) == 0 {
			g.SpawnMonster(pos)
		} else {
			// If tile is occupied or not walkable, we just skip spawning this monster for simplicity
			slog.Debug("Failed to spawn monster", "position", pos, "reason", "not walkable or occupied")
		}
	}
}

// placeItems spawns items in a given room.
func (m *Map) placeItems(g *Game, room Rect, items map[string]components.Item) {
	// 30% chance to spawn an item in each room
	if rand.Intn(100) < 30 {
		// Find a random walkable tile within the room bounds
		x := rand.Intn(room.X2-room.X1-1) + room.X1 + 1
		y := rand.Intn(room.Y2-room.Y1-1) + room.Y1 + 1
		pos := gruid.Point{X: x, Y: y}

		// Check if the tile is walkable and not already occupied
		if m.isWalkable(pos) && len(g.ecs.EntitiesAt(pos)) == 0 {
			// Get available items

			// Randomly select an item to spawn
			itemNames := []string{"Health Potion", "Iron Sword", "Leather Armor", "Gold Coin"}
			selectedName := itemNames[rand.Intn(len(itemNames))]
			selectedItem := items[selectedName]

			// Determine quantity
			quantity := 1
			if selectedItem.Stackable {
				quantity = rand.Intn(3) + 1 // 1-3 for stackable items
			}

			g.SpawnItem(selectedItem, quantity, pos)
			slog.Debug("Spawned item", "item", selectedItem.Name, "quantity", quantity, "position", pos)
		}
	}
}

// Constants like maxRooms, roomMinSize, roomMaxSize, maxMonstersPerRoom should be defined centrally (e.g., in game.go)
