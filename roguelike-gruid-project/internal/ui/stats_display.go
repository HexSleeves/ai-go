package ui

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
)

// GameData interface to avoid import cycles
type GameData interface {
	ECS() *ecs.ECS
	GetPlayerID() ecs.EntityID
	GetDepth() int
	Stats() GameStats
}

// GameStats interface for game statistics
type GameStats interface {
	GetMonstersKilled() int
}

// StatsPanel handles the display of player statistics
type StatsPanel struct {
	*Panel
}

// NewStatsPanel creates a new stats panel
func NewStatsPanel() *StatsPanel {
	panel := NewPanel(
		config.StatsPanelX,
		config.StatsPanelY,
		config.StatsPanelWidth,
		config.StatsPanelHeight,
		"Stats",
		true,
	)

	return &StatsPanel{Panel: panel}
}

// Render draws the stats panel with current player information
func (sp *StatsPanel) Render(grid gruid.Grid, gameData GameData) {
	// Clear and draw border
	sp.Clear(grid)
	sp.DrawBorder(grid)
	
	// Get content area
	contentX, contentY, contentWidth, _ := sp.GetContentArea()

	playerID := gameData.GetPlayerID()
	if !gameData.ECS().EntityExists(playerID) {
		return
	}

	currentY := contentY

	// Health display
	if gameData.ECS().HasHealthSafe(playerID) {
		currentY = sp.drawHealth(grid, gameData.ECS(), playerID, contentX, currentY, contentWidth)
	}

	// Level and experience
	if gameData.ECS().HasExperienceSafe(playerID) {
		currentY = sp.drawExperience(grid, gameData.ECS(), playerID, contentX, currentY, contentWidth)
	}

	// Equipment
	if gameData.ECS().HasEquipmentSafe(playerID) {
		currentY = sp.drawEquipment(grid, gameData.ECS(), playerID, contentX, currentY, contentWidth)
	}

	// Game stats
	currentY = sp.drawGameStats(grid, gameData, contentX, currentY, contentWidth)
}

// drawHealth renders the health bar and text
func (sp *StatsPanel) drawHealth(grid gruid.Grid, ecs *ecs.ECS, playerID ecs.EntityID, x, y, width int) int {
	health := ecs.GetHealthSafe(playerID)
	
	// Health text
	healthText := fmt.Sprintf("HP: %d/%d", health.CurrentHP, health.MaxHP)
	sp.drawLine(grid, healthText, x, y, ColorUIText)
	y++
	
	// Health bar
	healthPercent := float64(health.CurrentHP) / float64(health.MaxHP)
	var healthColor gruid.Color
	if healthPercent > 0.6 {
		healthColor = ColorHealthOk
	} else if healthPercent > 0.3 {
		healthColor = ColorHealthWounded
	} else {
		healthColor = ColorHealthCritical
	}
	
	barWidth := width - 2
	if barWidth > 0 {
		sp.DrawProgressBar(grid, x, y, barWidth, health.CurrentHP, health.MaxHP, 
			gruid.Style{Fg: healthColor, Bg: ColorUIBackground})
	}
	y++
	
	return y
}

// drawExperience renders level and XP information
func (sp *StatsPanel) drawExperience(grid gruid.Grid, ecs *ecs.ECS, playerID ecs.EntityID, x, y, width int) int {
	exp := ecs.GetExperienceSafe(playerID)
	
	// Level
	levelText := fmt.Sprintf("Level: %d", exp.Level)
	sp.drawLine(grid, levelText, x, y, ColorUIText)
	y++
	
	// XP text
	xpText := fmt.Sprintf("XP: %d/%d", exp.CurrentXP, exp.XPToNextLevel)
	sp.drawLine(grid, xpText, x, y, ColorUIText)
	y++
	
	// XP bar
	barWidth := width - 2
	if barWidth > 0 && exp.XPToNextLevel > 0 {
		sp.DrawProgressBar(grid, x, y, barWidth, exp.CurrentXP, exp.XPToNextLevel,
			gruid.Style{Fg: ColorStatusGood, Bg: ColorUIBackground})
	}
	y++
	
	return y
}

// drawEquipment renders equipped items
func (sp *StatsPanel) drawEquipment(grid gruid.Grid, ecs *ecs.ECS, playerID ecs.EntityID, x, y, width int) int {
	equipment := ecs.GetEquipmentSafe(playerID)
	
	// Weapon
	if equipment.Weapon != nil {
		weaponText := fmt.Sprintf("Wpn: %s", sp.truncateText(equipment.Weapon.Name, width-5))
		sp.drawLine(grid, weaponText, x, y, ColorUIText)
	} else {
		sp.drawLine(grid, "Wpn: None", x, y, ColorUIText)
	}
	y++
	
	// Armor
	if equipment.Armor != nil {
		armorText := fmt.Sprintf("Arm: %s", sp.truncateText(equipment.Armor.Name, width-5))
		sp.drawLine(grid, armorText, x, y, ColorUIText)
	} else {
		sp.drawLine(grid, "Arm: None", x, y, ColorUIText)
	}
	y++
	
	return y
}

// drawGameStats renders game statistics
func (sp *StatsPanel) drawGameStats(grid gruid.Grid, gameData GameData, x, y, width int) int {
	// Add spacing
	y++

	// Depth
	depthText := fmt.Sprintf("Depth: %d", gameData.GetDepth())
	sp.drawLine(grid, depthText, x, y, ColorUIText)
	y++

	// Game stats if available
	stats := gameData.Stats()
	if stats != nil {
		killsText := fmt.Sprintf("Kills: %d", stats.GetMonstersKilled())
		sp.drawLine(grid, killsText, x, y, ColorUIText)
		y++
	}

	return y
}

// drawLine draws a single line of text
func (sp *StatsPanel) drawLine(grid gruid.Grid, text string, x, y int, color gruid.Color) {
	style := gruid.Style{Fg: color, Bg: ColorUIBackground}
	
	for i, r := range text {
		if x+i >= sp.X+sp.Width-1 { // Don't draw over border
			break
		}
		if x+i < grid.Size().X && y < grid.Size().Y {
			grid.Set(gruid.Point{X: x + i, Y: y}, gruid.Cell{Rune: r, Style: style})
		}
	}
}

// truncateText truncates text to fit within the specified width
func (sp *StatsPanel) truncateText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return text[:maxWidth]
	}
	return text[:maxWidth-3] + "..."
}
