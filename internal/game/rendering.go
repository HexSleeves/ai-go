package game

import (
	"fmt"
	"log/slog"

	"codeberg.org/anaseto/gruid" // Needed for FOV type
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui" // For colors
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/utils"
)

// renderOrder is a type representing the priority of an entity rendering.
type renderOrder int

// Those constants represent distinct kinds of rendering priorities. In case
// two entities are at a given position, only the one with the highest priority
// gets displayed.
const (
	RONone renderOrder = iota
	ROCorpse
	ROItem
	ROActor
)

func RenderOrder(ecs *ecs.ECS, id ecs.EntityID) (ro renderOrder) {
	isPlayer := ecs.HasComponent(id, components.CPlayerTag)
	isMonster := ecs.HasComponent(id, components.CAITag)
	isCorpse := ecs.HasComponent(id, components.CCorpseTag)

	if isPlayer {
		ro = ROActor
	} else if isMonster {
		ro = ROActor
	} else if isCorpse {
		ro = ROCorpse
	}

	return ro
}

// Draw implements gruid.Model.Draw. It clears the grid, then renders the map
// and all entities using the RenderSystem, plus UI panels.
func (md *Model) Draw() gruid.Grid {
	g := md.game

	utils.Assert(g != nil, "Game is nil")
	utils.Assert(g.ecs != nil, "ECS is nil")
	utils.Assert(g.dungeon != nil, "Map is nil")

	// Clear the grid before drawing
	md.grid.Fill(gruid.Cell{Rune: ' '})

	// Handle different screen modes
	switch md.mode {
	case modeCharacterSheet:
		md.characterScreen.Render(md.grid, &gameDataAdapter{g})
		return md.grid

	case modeInventory:
		md.inventoryScreen.Render(md.grid, &gameDataAdapter{g})
		return md.grid

	case modeFullMessageLog:
		md.fullMessageScreen.Render(md.grid, g.MessageLog())
		return md.grid

	case modeNormal:
		// Normal game rendering
		break

	default:
		// Default to normal mode
		break
	}

	// Get player's FOV component using safe accessor
	playerFOVComp := g.ecs.GetFOVSafe(g.PlayerID)
	if playerFOVComp == nil {
		// Handle case where player FOV might be missing (though unlikely)
		slog.Error("Player entity missing FOV component in Draw", "id", g.PlayerID)
		return md.grid // Return the grid even if FOV is missing
	}

	// Update camera to follow player
	playerPos := g.GetPlayerPosition()
	md.camera.Update(playerPos.X, playerPos.Y)

	// Draw the map in the viewport
	md.drawMapViewport(g, playerFOVComp)

	// Render entities in the viewport
	md.renderEntitiesInViewport(g.ecs, playerFOVComp, g.dungeon.Width)

	// Draw debug overlays if enabled
	if md.debugLevel != DebugNone {
		md.drawDebugOverlays(g, playerFOVComp)
	}

	// Draw UI panels
	md.statsPanel.Render(md.grid, &gameDataAdapter{g})
	md.messagePanel.Render(md.grid, g.MessageLog())

	// Draw debug panels if enabled
	if md.showAIDebug || md.debugLevel == DebugAI || md.debugLevel == DebugFull {
		md.drawAIDebugPanel(g)
	}

	return md.grid
}

// drawMapViewport draws the map within the camera viewport
func (md *Model) drawMapViewport(g *Game, playerFOV *components.FOV) {
	minX, minY, maxX, maxY := md.camera.GetViewportBounds()

	it := g.dungeon.Grid.Iterator()
	for it.Next() {
		worldPos := it.P()

		// Skip if outside camera viewport
		if worldPos.X < minX || worldPos.X > maxX || worldPos.Y < minY || worldPos.Y > maxY {
			continue
		}

		isExplored := g.dungeon.IsExplored(worldPos)
		isVisible := playerFOV.IsVisible(worldPos, g.dungeon.Width)
		isWall := g.dungeon.IsWall(worldPos)

		// Check if FOV debug is enabled
		var style gruid.Style
		if md.showFOVDebug || md.debugLevel == DebugFOV || md.debugLevel == DebugFull {
			// Use FOV debug colors
			debugColor := GetFOVDebugColor(isVisible, isExplored)
			style = gruid.Style{Fg: debugColor}
		} else {
			// Use normal map style only for explored tiles
			if !isExplored {
				continue
			}
			style = ui.GetMapStyle(isWall, isVisible, isExplored)
		}

		// Convert world coordinates to screen coordinates
		screenX, screenY, visible := md.camera.WorldToScreen(worldPos.X, worldPos.Y)
		if visible {
			md.grid.Set(gruid.Point{X: screenX, Y: screenY}, gruid.Cell{
				Rune:  g.dungeon.Rune(it.Cell()),
				Style: style,
			})
		}
	}
}

func (md *Model) drawMap(g *Game, playerFOV *components.FOV) {
	it := g.dungeon.Grid.Iterator()
	for it.Next() {
		p := it.P()

		isExplored := g.dungeon.IsExplored(p)
		if !isExplored {
			continue
		}

		isVisible := playerFOV.IsVisible(p, g.dungeon.Width)
		isWall := g.dungeon.IsWall(p)

		// Use the new helper function to get the appropriate style
		style := ui.GetMapStyle(isWall, isVisible, isExplored)

		md.grid.Set(p, gruid.Cell{
			Rune:  g.dungeon.Rune(it.Cell()),
			Style: style,
		})
	}
}

// renderEntitiesInViewport draws entities within the camera viewport
func (md *Model) renderEntitiesInViewport(world *ecs.ECS, playerFOV *components.FOV, mapWidth int) {
	utils.Assert(world != nil, "ECS is nil")
	utils.Assert(playerFOV != nil, "Player FOV is nil")
	utils.Assert(mapWidth > 0, "Map width is not positive")

	// Get all entities with Position and Renderable components
	entityIDs := world.GetEntitiesWithComponents(components.CPosition, components.CRenderable)

	// Create a map to cache render orders
	renderOrderCache := make(map[ecs.EntityID]renderOrder, len(entityIDs))

	// Filter entities by visibility and viewport, cache their render orders
	visibleEntities := make([]ecs.EntityID, 0, len(entityIDs))
	for _, id := range entityIDs {
		// Use safe accessor - no error handling needed!
		pos := world.GetPositionSafe(id)

		// Only process entities that are visible and in viewport
		if world.HasPositionSafe(id) &&
			playerFOV.IsVisible(pos, mapWidth) &&
			md.camera.IsInViewport(pos.X, pos.Y) {
			visibleEntities = append(visibleEntities, id)
			renderOrderCache[id] = RenderOrder(world, id)
		}
	}

	// Group entities by render order for efficient rendering
	orderBuckets := make(map[renderOrder][]ecs.EntityID)
	for _, id := range visibleEntities {
		ro := renderOrderCache[id]
		orderBuckets[ro] = append(orderBuckets[ro], id)
	}

	// Define render order priorities (lowest to highest)
	priorities := []renderOrder{RONone, ROCorpse, ROItem, ROActor}

	// Render entities in priority order
	for _, priority := range priorities {
		if bucket, ok := orderBuckets[priority]; ok {
			for _, id := range bucket {
				// Use safe accessor - no error handling needed!
				worldPos := world.GetPositionSafe(id)
				md.drawEntityInViewport(world, worldPos, id)
			}
		}
	}
}

// drawEntityInViewport draws an entity using camera coordinates
func (md *Model) drawEntityInViewport(ecs *ecs.ECS, worldPos gruid.Point, entityID ecs.EntityID) {
	// Use safe accessor - no error handling needed!
	renderable := ecs.GetRenderableSafe(entityID)

	// Only draw if entity actually has a renderable component
	if !ecs.HasRenderableSafe(entityID) {
		return
	}

	// Convert world coordinates to screen coordinates
	screenX, screenY, visible := md.camera.WorldToScreen(worldPos.X, worldPos.Y)
	if !visible {
		return
	}

	color := renderable.Color

	// Draw the entity with the appropriate color
	md.grid.Set(gruid.Point{X: screenX, Y: screenY}, gruid.Cell{Rune: renderable.Glyph, Style: gruid.Style{Fg: color}})
}

// When drawing an entity, check for HitFlash
func drawEntity(ecs *ecs.ECS, pos gruid.Point, entityID ecs.EntityID, grid gruid.Grid) {
	// Use safe accessor - no error handling needed!
	renderable := ecs.GetRenderableSafe(entityID)

	// Only draw if entity actually has a renderable component
	if !ecs.HasRenderableSafe(entityID) {
		return
	}

	color := renderable.Color

	// Draw the entity with the appropriate color
	grid.Set(pos, gruid.Cell{Rune: renderable.Glyph, Style: gruid.Style{Fg: color}})
}

// gameDataAdapter adapts Game to ui.GameData interface
type gameDataAdapter struct {
	game *Game
}

func (gda *gameDataAdapter) ECS() *ecs.ECS {
	return gda.game.ECS()
}

func (gda *gameDataAdapter) GetPlayerID() ecs.EntityID {
	return gda.game.GetPlayerID()
}

// drawDebugOverlays draws debug visualization overlays
func (md *Model) drawDebugOverlays(g *Game, playerFOV *components.FOV) {
	// Draw pathfinding debug if enabled
	if md.showPathfindingDebug || md.debugLevel == DebugPathfinding || md.debugLevel == DebugFull {
		md.drawPathfindingDebug(g)
	}
}

// drawPathfindingDebug draws pathfinding debug visualization with AI behavior colors
func (md *Model) drawPathfindingDebug(g *Game) {
	if md.pathfindingDebugInfo == nil {
		return
	}

	// Draw entity paths with behavior-based colors
	for entityID, path := range md.pathfindingDebugInfo.EntityPaths {
		if len(path) < 2 {
			continue
		}

		// Get AI state for color determination
		var pathColor gruid.Color = ui.ColorForeground
		if aiState, exists := md.pathfindingDebugInfo.AIStates[entityID]; exists {
			pathColor = GetPathfindingDebugColor(aiState)
		} else {
			// Fallback to getting AI component directly
			if g.ecs.HasAIComponentSafe(entityID) {
				aiComp := g.ecs.GetAIComponentSafe(entityID)
				pathColor = GetPathfindingDebugColor(aiComp.State)
			}
		}

		// Draw path segments
		for i := 0; i < len(path)-1; i++ {
			currentPos := path[i]
			nextPos := path[i+1]

			// Only draw if in viewport
			if !md.IsInViewport(currentPos) {
				continue
			}

			// Convert to screen coordinates
			screenX, screenY, visible := md.camera.WorldToScreen(currentPos.X, currentPos.Y)
			if !visible {
				continue
			}

			// Get appropriate path character
			pathChar := GetPathCharacter(currentPos, nextPos)

			// Draw the path segment
			md.grid.Set(gruid.Point{X: screenX, Y: screenY}, gruid.Cell{
				Rune:  pathChar,
				Style: gruid.Style{Fg: pathColor},
			})
		}

		// Draw target destination if path exists
		if len(path) > 0 {
			targetPos := path[len(path)-1]
			if md.IsInViewport(targetPos) {
				screenX, screenY, visible := md.camera.WorldToScreen(targetPos.X, targetPos.Y)
				if visible {
					md.grid.Set(gruid.Point{X: screenX, Y: screenY}, gruid.Cell{
						Rune:  '◆', // Diamond marker for target
						Style: gruid.Style{Fg: ui.ColorDebugTarget},
					})
				}
			}
		}
	}

	// Draw failed pathfinding attempts
	for entityID, failedTarget := range md.pathfindingDebugInfo.FailedPaths {
		if !md.IsInViewport(failedTarget) {
			continue
		}

		screenX, screenY, visible := md.camera.WorldToScreen(failedTarget.X, failedTarget.Y)
		if visible {
			md.grid.Set(gruid.Point{X: screenX, Y: screenY}, gruid.Cell{
				Rune:  '×', // X marker for failed paths
				Style: gruid.Style{Fg: ui.ColorRed},
			})
		}

		// Suppress unused variable warning
		_ = entityID
	}
}

// drawAIDebugPanel draws the AI debug information panel
func (md *Model) drawAIDebugPanel(g *Game) {
	if md.aiDebugInfo == nil {
		return
	}

	// Panel configuration
	panelX := md.grid.Size().X - 25 // Right side of screen
	panelY := 1
	panelWidth := 24
	panelHeight := 15

	// Ensure panel fits on screen
	if panelX < 0 {
		return
	}

	// Draw panel border
	borderStyle := gruid.Style{Fg: ui.ColorDebugAIPanel}

	// Top border
	for x := panelX; x < panelX+panelWidth; x++ {
		if x < md.grid.Size().X {
			md.grid.Set(gruid.Point{X: x, Y: panelY}, gruid.Cell{Rune: '─', Style: borderStyle})
		}
	}

	// Bottom border
	for x := panelX; x < panelX+panelWidth; x++ {
		if x < md.grid.Size().X && panelY+panelHeight-1 < md.grid.Size().Y {
			md.grid.Set(gruid.Point{X: x, Y: panelY + panelHeight - 1}, gruid.Cell{Rune: '─', Style: borderStyle})
		}
	}

	// Side borders
	for y := panelY; y < panelY+panelHeight; y++ {
		if y < md.grid.Size().Y {
			if panelX < md.grid.Size().X {
				md.grid.Set(gruid.Point{X: panelX, Y: y}, gruid.Cell{Rune: '│', Style: borderStyle})
			}
			if panelX+panelWidth-1 < md.grid.Size().X {
				md.grid.Set(gruid.Point{X: panelX + panelWidth - 1, Y: y}, gruid.Cell{Rune: '│', Style: borderStyle})
			}
		}
	}

	// Panel title
	title := "AI Debug"
	titleX := panelX + (panelWidth-len(title))/2
	if titleX >= 0 && titleX+len(title) < md.grid.Size().X && panelY < md.grid.Size().Y {
		for i, r := range title {
			if titleX+i < md.grid.Size().X {
				md.grid.Set(gruid.Point{X: titleX + i, Y: panelY}, gruid.Cell{Rune: r, Style: gruid.Style{Fg: ui.ColorUITitle}})
			}
		}
	}

	// Content area
	contentY := panelY + 2
	lineNum := 0
	maxLines := panelHeight - 3

	// Get player position for distance calculations
	playerPos := g.GetPlayerPosition()

	// Sort entities by distance to player (closest first)
	type entityDistance struct {
		entityID ecs.EntityID
		distance int
		debug    AIEntityDebug
	}

	var sortedEntities []entityDistance
	for entityID, debugInfo := range md.aiDebugInfo.EntityStates {
		distance := manhattanDistance(debugInfo.Position, playerPos)
		sortedEntities = append(sortedEntities, entityDistance{
			entityID: entityID,
			distance: distance,
			debug:    debugInfo,
		})
	}

	// Simple sort by distance (bubble sort for simplicity)
	for i := 0; i < len(sortedEntities); i++ {
		for j := i + 1; j < len(sortedEntities); j++ {
			if sortedEntities[i].distance > sortedEntities[j].distance {
				sortedEntities[i], sortedEntities[j] = sortedEntities[j], sortedEntities[i]
			}
		}
	}

	// Display AI entities (limit to visible ones)
	for _, entity := range sortedEntities {
		if lineNum >= maxLines {
			break
		}

		debugInfo := entity.debug

		// Only show entities within reasonable distance
		if entity.distance > 20 {
			continue
		}

		// Entity header line
		entityLine := fmt.Sprintf("E%d [%s]", debugInfo.EntityID, GetAIStateString(debugInfo.State))
		if len(entityLine) > panelWidth-3 {
			entityLine = entityLine[:panelWidth-3]
		}

		md.drawPanelText(entityLine, panelX+1, contentY+lineNum, ui.ColorUIText)
		lineNum++

		if lineNum >= maxLines {
			break
		}

		// Details line
		healthPercent := int(debugInfo.HealthPercent * 100)
		canSeeText := "No"
		if debugInfo.CanSeePlayer {
			canSeeText = "Yes"
		}

		detailLine := fmt.Sprintf("D:%d HP:%d%% See:%s", debugInfo.DistanceToPlayer, healthPercent, canSeeText)
		if len(detailLine) > panelWidth-3 {
			detailLine = detailLine[:panelWidth-3]
		}

		md.drawPanelText(detailLine, panelX+1, contentY+lineNum, ui.ColorDebugAIPanel)
		lineNum++

		// Add spacing between entities
		lineNum++
		if lineNum >= maxLines {
			break
		}
	}

	// If no entities to show
	if len(sortedEntities) == 0 {
		noEntitiesText := "No AI entities"
		md.drawPanelText(noEntitiesText, panelX+1, contentY, ui.ColorDebugAIPanel)
	}
}

// drawPanelText draws text within the panel bounds
func (md *Model) drawPanelText(text string, x, y int, color gruid.Color) {
	if y >= md.grid.Size().Y {
		return
	}

	for i, r := range text {
		if x+i >= md.grid.Size().X {
			break
		}
		md.grid.Set(gruid.Point{X: x + i, Y: y}, gruid.Cell{Rune: r, Style: gruid.Style{Fg: color}})
	}
}

func (gda *gameDataAdapter) GetDepth() int {
	return gda.game.GetDepth()
}

func (gda *gameDataAdapter) Stats() ui.GameStats {
	return &gameStatsAdapter{gda.game.Stats()}
}

// gameStatsAdapter adapts GameStats to ui.GameStats interface
type gameStatsAdapter struct {
	stats *GameStats
}

func (gsa *gameStatsAdapter) GetMonstersKilled() int {
	if gsa.stats == nil {
		return 0
	}
	return gsa.stats.GetMonstersKilled()
}
