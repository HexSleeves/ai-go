package game

import (
	"codeberg.org/anaseto/gruid" // Needed for FOV type
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui" // For colors
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/utils"
	"github.com/sirupsen/logrus"
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
		logrus.Errorf("Player entity %d missing FOV component in Draw", g.PlayerID)
		return md.grid // Return the grid even if FOV is missing
	}

	// Update camera to follow player
	playerPos := g.GetPlayerPosition()
	md.camera.Update(playerPos.X, playerPos.Y)

	// Draw the map in the viewport
	md.drawMapViewport(g, playerFOVComp)

	// Render entities in the viewport
	md.renderEntitiesInViewport(g.ecs, playerFOVComp, g.dungeon.Width)

	// Draw UI panels
	md.statsPanel.Render(md.grid, &gameDataAdapter{g})
	md.messagePanel.Render(md.grid, g.MessageLog())

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
		if !isExplored {
			continue
		}

		isVisible := playerFOV.IsVisible(worldPos, g.dungeon.Width)
		isWall := g.dungeon.IsWall(worldPos)

		// Use the new helper function to get the appropriate style
		style := ui.GetMapStyle(isWall, isVisible, isExplored)

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
