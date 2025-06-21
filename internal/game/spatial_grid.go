package game

import (
	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
)

// SpatialGrid provides efficient spatial partitioning for entity positions
type SpatialGrid struct {
	cells    map[gruid.Point][]ecs.EntityID
	width    int
	height   int
	capacity int // Initial capacity for entity slices
}

// NewSpatialGrid creates a new spatial grid with the given dimensions
func NewSpatialGrid(width, height int) *SpatialGrid {
	return &SpatialGrid{
		cells:    make(map[gruid.Point][]ecs.EntityID),
		width:    width,
		height:   height,
		capacity: 4, // Default initial capacity per cell
	}
}

// Clear removes all entities from the grid
func (sg *SpatialGrid) Clear() {
	sg.cells = make(map[gruid.Point][]ecs.EntityID)
}

// Add adds an entity to the grid at the specified position
func (sg *SpatialGrid) Add(id ecs.EntityID, pos gruid.Point) {
	if !sg.isValidPosition(pos) {
		return
	}

	if _, exists := sg.cells[pos]; !exists {
		sg.cells[pos] = make([]ecs.EntityID, 0, sg.capacity)
	}
	sg.cells[pos] = append(sg.cells[pos], id)
}

// Remove removes an entity from its position in the grid
func (sg *SpatialGrid) Remove(id ecs.EntityID, pos gruid.Point) {
	if !sg.isValidPosition(pos) {
		return
	}

	entities, exists := sg.cells[pos]
	if !exists {
		return
	}

	// Find and remove the entity
	for i, eid := range entities {
		if eid == id {
			// Remove the entity by swapping with the last element and truncating
			lastIdx := len(entities) - 1
			entities[i] = entities[lastIdx]
			sg.cells[pos] = entities[:lastIdx]

			// If the cell is empty, remove it from the map
			if len(sg.cells[pos]) == 0 {
				delete(sg.cells, pos)
			}
			return
		}
	}
}

// Move updates an entity's position in the grid
func (sg *SpatialGrid) Move(id ecs.EntityID, oldPos, newPos gruid.Point) {
	sg.Remove(id, oldPos)
	sg.Add(id, newPos)
}

// GetEntitiesAt returns all entities at the specified position
func (sg *SpatialGrid) GetEntitiesAt(pos gruid.Point) []ecs.EntityID {
	if entities, exists := sg.cells[pos]; exists {
		return entities
	}
	return nil
}

// isValidPosition checks if a position is within the grid bounds
func (sg *SpatialGrid) isValidPosition(pos gruid.Point) bool {
	return pos.X >= 0 && pos.X < sg.width && pos.Y >= 0 && pos.Y < sg.height
}

// GetVisibleEntities returns all entities at visible positions
func (sg *SpatialGrid) GetVisibleEntities(visiblePoints []gruid.Point) []ecs.EntityID {
	result := make([]ecs.EntityID, 0)
	seen := make(map[ecs.EntityID]bool)

	for _, pos := range visiblePoints {
		if entities := sg.GetEntitiesAt(pos); entities != nil {
			for _, id := range entities {
				if !seen[id] {
					seen[id] = true
					result = append(result, id)
				}
			}
		}
	}

	return result
}
