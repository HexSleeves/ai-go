package game

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/rl"
)

// Rect represents a rectangular area on the map.
type Rect struct {
	X1, Y1, X2, Y2 int
}

// NewRect creates a new Rect.
func NewRect(x, y, w, h int) Rect {
	return Rect{X1: x, Y1: y, X2: x + w, Y2: y + h}
}

// Center returns the center point of the rectangle.
func (r Rect) Center() gruid.Point {
	centerX := (r.X1 + r.X2) / 2
	centerY := (r.Y1 + r.Y2) / 2
	return gruid.Point{X: centerX, Y: centerY}
}

// Intersects checks if this rectangle intersects with another one.
func (r Rect) Intersects(other Rect) bool {
	return r.X1 <= other.X2 && r.X2 >= other.X1 &&
		r.Y1 <= other.Y2 && r.Y2 >= other.Y1
}

// Constants for map generation (roomMaxSize, roomMinSize, maxRooms)
// are now defined centrally (e.g., in game.go)

// createRoom carves a rectangular room onto the grid.
// It now sets floor cells with the TileFloor attribute.
func createRoom(grid rl.Grid, room Rect) {
	for y := room.Y1 + 1; y < room.Y2; y++ {
		for x := room.X1 + 1; x < room.X2; x++ {
			grid.Set(gruid.Point{X: x, Y: y}, FloorCell)
		}
	}
}

// createHTunnel carves a horizontal tunnel between two points.
func createHTunnel(grid rl.Grid, x1, x2, y int) {
	startX := min(x1, x2)
	endX := max(x1, x2)
	for x := startX; x <= endX; x++ {
		grid.Set(gruid.Point{X: x, Y: y}, FloorCell)
	}
}

// createVTunnel carves a vertical tunnel between two points.
func createVTunnel(grid rl.Grid, y1, y2, x int) {
	startY := min(y1, y2)
	endY := max(y1, y2)
	for y := startY; y <= endY; y++ {
		grid.Set(gruid.Point{X: x, Y: y}, FloorCell)
	}
}
