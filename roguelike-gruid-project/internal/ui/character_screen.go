package ui

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
)

// CharacterScreen handles the full-screen character information display
type CharacterScreen struct {
	*Panel
}

// NewCharacterScreen creates a new character screen
func NewCharacterScreen() *CharacterScreen {
	panel := NewPanel(
		0, 0,
		config.DungeonWidth,
		config.DungeonHeight,
		"Character Sheet",
		true,
	)
	
	return &CharacterScreen{Panel: panel}
}

// Render draws the character screen with detailed player information
func (cs *CharacterScreen) Render(grid gruid.Grid, gameData GameData) {
	// Clear and draw border
	cs.Clear(grid)
	cs.DrawBorder(grid)
	
	// Get content area
	contentX, contentY, contentWidth, _ := cs.GetContentArea()
	
	playerID := gameData.GetPlayerID()
	if !gameData.ECS().EntityExists(playerID) {
		return
	}
	
	currentY := contentY
	
	// Character name and basic info
	currentY = cs.drawBasicInfo(grid, gameData.ECS(), playerID, contentX, currentY, contentWidth)
	currentY++ // Add spacing
	
	// Attributes section
	currentY = cs.drawAttributes(grid, gameData.ECS(), playerID, contentX, currentY, contentWidth)
	currentY++ // Add spacing
	
	// Combat statistics
	currentY = cs.drawCombatStats(grid, gameData.ECS(), playerID, contentX, currentY, contentWidth)
	currentY++ // Add spacing
	
	// Equipment section
	currentY = cs.drawEquipment(grid, gameData.ECS(), playerID, contentX, currentY, contentWidth)
	currentY++ // Add spacing
	
	// Skills section
	currentY = cs.drawSkills(grid, gameData.ECS(), playerID, contentX, currentY, contentWidth)
	currentY++ // Add spacing
	
	// Status effects
	currentY = cs.drawStatusEffects(grid, gameData.ECS(), playerID, contentX, currentY, contentWidth)
	
	// Instructions at bottom
	cs.drawInstructions(grid)
}

// drawBasicInfo renders character name, level, and experience
func (cs *CharacterScreen) drawBasicInfo(grid gruid.Grid, ecs *ecs.ECS, playerID ecs.EntityID, x, y, width int) int {
	// Character name (placeholder - using "Player" for now)
	cs.drawLine(grid, "Name: Player", x, y, ColorUITitle)
	y++
	
	// Level and experience
	if ecs.HasExperienceSafe(playerID) {
		exp := ecs.GetExperienceSafe(playerID)
		cs.drawLine(grid, fmt.Sprintf("Level: %d", exp.Level), x, y, ColorUIText)
		y++
		
		cs.drawLine(grid, fmt.Sprintf("Experience: %d / %d", exp.CurrentXP, exp.XPToNextLevel), x, y, ColorUIText)
		y++
		
		// XP progress bar
		barWidth := width - 15
		if barWidth > 0 {
			cs.drawText(grid, "XP Progress: ", x, y, ColorUIText)
			cs.DrawProgressBar(grid, x+13, y, barWidth, exp.CurrentXP, exp.XPToNextLevel,
				gruid.Style{Fg: ColorStatusGood, Bg: ColorUIBackground})
		}
		y++
		
		cs.drawLine(grid, fmt.Sprintf("Total XP: %d", exp.TotalXP), x, y, ColorUIText)
		y++
		
		if exp.SkillPoints > 0 || exp.AttributePoints > 0 {
			cs.drawLine(grid, fmt.Sprintf("Skill Points: %d | Attribute Points: %d", 
				exp.SkillPoints, exp.AttributePoints), x, y, ColorStatusGood)
			y++
		}
	}
	
	// Health, Mana, Stamina
	if ecs.HasHealthSafe(playerID) {
		health := ecs.GetHealthSafe(playerID)
		cs.drawLine(grid, fmt.Sprintf("Health: %d / %d", health.CurrentHP, health.MaxHP), x, y, ColorUIText)
		y++
	}
	
	if ecs.HasManaSafe(playerID) {
		mana := ecs.GetManaSafe(playerID)
		cs.drawLine(grid, fmt.Sprintf("Mana: %d / %d", mana.CurrentMP, mana.MaxMP), x, y, ColorUIText)
		y++
	}
	
	if ecs.HasStaminaSafe(playerID) {
		stamina := ecs.GetStaminaSafe(playerID)
		cs.drawLine(grid, fmt.Sprintf("Stamina: %d / %d", stamina.CurrentSP, stamina.MaxSP), x, y, ColorUIText)
		y++
	}
	
	return y
}

// drawAttributes renders character attributes
func (cs *CharacterScreen) drawAttributes(grid gruid.Grid, ecs *ecs.ECS, playerID ecs.EntityID, x, y, width int) int {
	if !ecs.HasStatsSafe(playerID) {
		return y
	}
	
	stats := ecs.GetStatsSafe(playerID)
	
	cs.drawLine(grid, "=== ATTRIBUTES ===", x, y, ColorUITitle)
	y++
	
	cs.drawLine(grid, fmt.Sprintf("Strength:     %2d    Dexterity:    %2d", 
		stats.Strength, stats.Dexterity), x, y, ColorUIText)
	y++
	
	cs.drawLine(grid, fmt.Sprintf("Constitution: %2d    Intelligence: %2d", 
		stats.Constitution, stats.Intelligence), x, y, ColorUIText)
	y++
	
	cs.drawLine(grid, fmt.Sprintf("Wisdom:       %2d    Charisma:     %2d", 
		stats.Wisdom, stats.Charisma), x, y, ColorUIText)
	y++
	
	return y
}

// drawCombatStats renders combat-related statistics
func (cs *CharacterScreen) drawCombatStats(grid gruid.Grid, ecs *ecs.ECS, playerID ecs.EntityID, x, y, width int) int {
	if !ecs.HasCombatSafe(playerID) {
		return y
	}
	
	combat := ecs.GetCombatSafe(playerID)
	
	cs.drawLine(grid, "=== COMBAT STATS ===", x, y, ColorUITitle)
	y++
	
	cs.drawLine(grid, fmt.Sprintf("Attack Power: %2d    Defense:      %2d", 
		combat.AttackPower, combat.Defense), x, y, ColorUIText)
	y++
	
	cs.drawLine(grid, fmt.Sprintf("Accuracy:     %2d%%   Dodge Chance: %2d%%", 
		combat.Accuracy, combat.DodgeChance), x, y, ColorUIText)
	y++
	
	cs.drawLine(grid, fmt.Sprintf("Critical:     %2d%%   Crit Damage:  %2d%%", 
		combat.CriticalChance, combat.CriticalDamage), x, y, ColorUIText)
	y++
	
	return y
}

// drawEquipment renders equipped items with detailed stats
func (cs *CharacterScreen) drawEquipment(grid gruid.Grid, ecs *ecs.ECS, playerID ecs.EntityID, x, y, width int) int {
	if !ecs.HasEquipmentSafe(playerID) {
		return y
	}
	
	equipment := ecs.GetEquipmentSafe(playerID)
	
	cs.drawLine(grid, "=== EQUIPMENT ===", x, y, ColorUITitle)
	y++
	
	// Weapon
	if equipment.Weapon != nil {
		cs.drawLine(grid, fmt.Sprintf("Weapon: %s", equipment.Weapon.Name), x, y, ColorUIText)
		y++
		cs.drawLine(grid, fmt.Sprintf("  %s", equipment.Weapon.Description), x, y, ColorUIHighlight)
		y++
	} else {
		cs.drawLine(grid, "Weapon: None", x, y, ColorUIText)
		y++
	}
	
	// Armor
	if equipment.Armor != nil {
		cs.drawLine(grid, fmt.Sprintf("Armor:  %s", equipment.Armor.Name), x, y, ColorUIText)
		y++
		cs.drawLine(grid, fmt.Sprintf("  %s", equipment.Armor.Description), x, y, ColorUIHighlight)
		y++
	} else {
		cs.drawLine(grid, "Armor:  None", x, y, ColorUIText)
		y++
	}
	
	// Accessory
	if equipment.Accessory != nil {
		cs.drawLine(grid, fmt.Sprintf("Accessory: %s", equipment.Accessory.Name), x, y, ColorUIText)
		y++
		cs.drawLine(grid, fmt.Sprintf("  %s", equipment.Accessory.Description), x, y, ColorUIHighlight)
		y++
	} else {
		cs.drawLine(grid, "Accessory: None", x, y, ColorUIText)
		y++
	}
	
	return y
}

// drawSkills renders character skills
func (cs *CharacterScreen) drawSkills(grid gruid.Grid, ecs *ecs.ECS, playerID ecs.EntityID, x, y, width int) int {
	if !ecs.HasSkillsSafe(playerID) {
		return y
	}
	
	skills := ecs.GetSkillsSafe(playerID)
	
	cs.drawLine(grid, "=== SKILLS ===", x, y, ColorUITitle)
	y++
	
	// Combat skills
	cs.drawLine(grid, "Combat:", x, y, ColorUIHighlight)
	y++
	cs.drawLine(grid, fmt.Sprintf("  Melee Weapons: %2d    Ranged Weapons: %2d    Defense: %2d", 
		skills.MeleeWeapons, skills.RangedWeapons, skills.Defense), x, y, ColorUIText)
	y++
	
	// Magic skills
	cs.drawLine(grid, "Magic:", x, y, ColorUIHighlight)
	y++
	cs.drawLine(grid, fmt.Sprintf("  Evocation: %2d    Conjuration: %2d    Enchantment: %2d    Divination: %2d", 
		skills.Evocation, skills.Conjuration, skills.Enchantment, skills.Divination), x, y, ColorUIText)
	y++
	
	// Utility skills
	cs.drawLine(grid, "Utility:", x, y, ColorUIHighlight)
	y++
	cs.drawLine(grid, fmt.Sprintf("  Stealth: %2d    Lockpicking: %2d    Perception: %2d", 
		skills.Stealth, skills.Lockpicking, skills.Perception), x, y, ColorUIText)
	y++
	cs.drawLine(grid, fmt.Sprintf("  Medicine: %2d   Crafting: %2d", 
		skills.Medicine, skills.Crafting), x, y, ColorUIText)
	y++
	
	return y
}

// drawStatusEffects renders active status effects
func (cs *CharacterScreen) drawStatusEffects(grid gruid.Grid, ecs *ecs.ECS, playerID ecs.EntityID, x, y, width int) int {
	if !ecs.HasStatusEffectsSafe(playerID) {
		return y
	}
	
	statusEffects := ecs.GetStatusEffectsSafe(playerID)
	
	cs.drawLine(grid, "=== STATUS EFFECTS ===", x, y, ColorUITitle)
	y++
	
	if len(statusEffects.Effects) == 0 {
		cs.drawLine(grid, "None", x, y, ColorUIText)
		y++
	} else {
		for _, effect := range statusEffects.Effects {
			effectText := fmt.Sprintf("%s (%d turns)", effect.Name, effect.Duration)
			cs.drawLine(grid, effectText, x, y, ColorStatusNeutral)
			y++
		}
	}
	
	return y
}

// drawInstructions renders control instructions at the bottom
func (cs *CharacterScreen) drawInstructions(grid gruid.Grid) {
	instructionY := cs.Y + cs.Height - 2
	instructions := "Press [ESC] or [q] to close"
	
	// Center the instructions
	startX := cs.X + (cs.Width-len(instructions))/2
	cs.drawText(grid, instructions, startX, instructionY, ColorUIHighlight)
}

// drawLine draws a single line of text
func (cs *CharacterScreen) drawLine(grid gruid.Grid, text string, x, y int, color gruid.Color) {
	cs.drawText(grid, text, x, y, color)
}

// drawText draws text at the specified position
func (cs *CharacterScreen) drawText(grid gruid.Grid, text string, x, y int, color gruid.Color) {
	style := gruid.Style{Fg: color, Bg: ColorUIBackground}

	for i, r := range text {
		if x+i >= cs.X+cs.Width-1 { // Don't draw over border
			break
		}
		if x+i < grid.Size().X && y < grid.Size().Y {
			grid.Set(gruid.Point{X: x + i, Y: y}, gruid.Cell{Rune: r, Style: style})
		}
	}
}

// ResetSelection resets any selection state (placeholder for future use)
func (cs *CharacterScreen) ResetSelection() {
	// Character screen doesn't have selection state currently
	// This method is here for consistency with other screens
}
