package game

import (
	"testing"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
)

func TestNewMap(t *testing.T) {
	width, height := 80, 24
	m := NewMap(width, height)

	if m.Width != width {
		t.Errorf("Expected width %d, got %d", width, m.Width)
	}

	if m.Height != height {
		t.Errorf("Expected height %d, got %d", height, m.Height)
	}

	expectedSize := gruid.Point{X: width, Y: height}
	if m.Grid.Size() != expectedSize {
		t.Error("Grid size should match map dimensions")
	}
}

func TestMap_InBounds(t *testing.T) {
	m := NewMap(10, 10)

	testCases := []struct {
		point    gruid.Point
		expected bool
		name     string
	}{
		{gruid.Point{X: 0, Y: 0}, true, "top-left corner"},
		{gruid.Point{X: 9, Y: 9}, true, "bottom-right corner"},
		{gruid.Point{X: 5, Y: 5}, true, "center"},
		{gruid.Point{X: -1, Y: 5}, false, "negative X"},
		{gruid.Point{X: 5, Y: -1}, false, "negative Y"},
		{gruid.Point{X: 10, Y: 5}, false, "X out of bounds"},
		{gruid.Point{X: 5, Y: 10}, false, "Y out of bounds"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := m.InBounds(tc.point)
			if result != tc.expected {
				t.Errorf("InBounds(%v) = %v, expected %v", tc.point, result, tc.expected)
			}
		})
	}
}

func TestMap_IsWall(t *testing.T) {
	m := NewMap(10, 10)

	// Initially all cells should be walls
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			point := gruid.Point{X: x, Y: y}
			if !m.IsWall(point) {
				t.Errorf("Point %v should be a wall initially", point)
			}
		}
	}
}

func TestMap_SetExplored(t *testing.T) {
	m := NewMap(10, 10)

	point := gruid.Point{X: 5, Y: 5}

	// Initially should not be explored
	if m.IsExplored(point) {
		t.Error("Point should not be explored initially")
	}

	// Set as explored
	m.SetExplored(point)

	// Should now be explored
	if !m.IsExplored(point) {
		t.Error("Point should be explored after SetExplored")
	}
}

func TestMap_GenerateMap(t *testing.T) {
	g := NewGame()
	g.InitLevel()

	// Check that player start position is valid
	playerPos := g.GetPlayerPosition()

	if !g.dungeon.InBounds(playerPos) {
		t.Error("Player start position should be in bounds")
	}

	if !g.dungeon.isWalkable(playerPos) {
		t.Error("Player start position should be walkable")
	}

	// Check that there are some floor tiles
	floorCount := 0
	for y := 0; y < g.dungeon.Height; y++ {
		for x := 0; x < g.dungeon.Width; x++ {
			point := gruid.Point{X: x, Y: y}
			if g.dungeon.isWalkable(point) {
				floorCount++
			}
		}
	}

	if floorCount == 0 {
		t.Error("Map should have some walkable floor tiles")
	}

	// Floor should be a reasonable percentage of the map
	totalTiles := g.dungeon.Width * g.dungeon.Height
	floorPercentage := float64(floorCount) / float64(totalTiles)

	if floorPercentage < 0.1 || floorPercentage > 0.8 {
		t.Errorf("Floor percentage %f seems unreasonable", floorPercentage)
	}
}

func TestRect_Center(t *testing.T) {
	rect := Rect{X1: 0, Y1: 0, X2: 10, Y2: 10}
	center := rect.Center()

	expected := gruid.Point{X: 5, Y: 5}
	if center != expected {
		t.Errorf("Expected center %v, got %v", expected, center)
	}
}

func TestRect_Intersects(t *testing.T) {
	rect1 := Rect{X1: 0, Y1: 0, X2: 5, Y2: 5}
	rect2 := Rect{X1: 3, Y1: 3, X2: 8, Y2: 8}
	rect3 := Rect{X1: 6, Y1: 6, X2: 10, Y2: 10}

	if !rect1.Intersects(rect2) {
		t.Error("rect1 and rect2 should intersect")
	}

	if rect1.Intersects(rect3) {
		t.Error("rect1 and rect3 should not intersect")
	}
}

func BenchmarkMap_IsWall(b *testing.B) {
	m := NewMap(config.DungeonWidth, config.DungeonHeight)
	point := gruid.Point{X: 5, Y: 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.IsWall(point)
	}
}

func BenchmarkMap_GenerateMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		g := NewGame()
		g.InitLevel()
	}
}
