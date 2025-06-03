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
	scrollOffset int
}

// DrawableElement is a function that draws a part of the character screen.
type DrawableElement func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int)

// NewCharacterScreen creates a new character screen
func NewCharacterScreen() *CharacterScreen {
	panel := NewPanel(
		0, 0,
		config.DungeonWidth,
		config.DungeonHeight,
		"Character Sheet",
		true,
	)

	return &CharacterScreen{
		Panel:        panel,
		scrollOffset: 0,
	}
}

// generateDrawableElements collects all drawing operations for the character sheet.
func (cs *CharacterScreen) generateDrawableElements(gameData GameData, contentWidth int) []DrawableElement {
	var elements []DrawableElement

	playerID := gameData.GetPlayerID()
	if !gameData.ECS().EntityExists(playerID) {
		return elements
	}

	cs.appendBasicInfoElements(&elements, gameData.ECS(), playerID, contentWidth)
	cs.drawSpacer(&elements)

	cs.appendAttributesElements(&elements, gameData.ECS(), playerID, contentWidth)
	cs.drawSpacer(&elements)

	cs.appendCombatStatsElements(&elements, gameData.ECS(), playerID, contentWidth)
	cs.drawSpacer(&elements)

	cs.appendEquipmentElements(&elements, gameData.ECS(), playerID, contentWidth)
	cs.drawSpacer(&elements)

	cs.appendSkillsElements(&elements, gameData.ECS(), playerID, contentWidth)
	cs.drawSpacer(&elements)

	cs.appendStatusEffectsElements(&elements, gameData.ECS(), playerID, contentWidth)
	cs.drawSpacers(&elements, 2)

	return elements
}

// Render draws the character screen with detailed player information
func (cs *CharacterScreen) Render(grid gruid.Grid, gameData GameData) {
	cs.Clear(grid)
	cs.DrawBorder(grid)

	contentX, contentY, contentWidth, contentHeight := cs.GetContentArea()

	drawableElements := cs.generateDrawableElements(gameData, contentWidth)
	totalContentLines := len(drawableElements)
	displayHeight := contentHeight

	// Clamp scrollOffset
	if totalContentLines <= displayHeight {
		cs.scrollOffset = 0
	} else {
		maxScroll := totalContentLines - displayHeight
		if cs.scrollOffset > maxScroll {
			cs.scrollOffset = maxScroll
		}
		if cs.scrollOffset < 0 {
			cs.scrollOffset = 0
		}
	}

	// Draw visible elements
	for i := 0; i < displayHeight; i++ {
		contentLineIndex := cs.scrollOffset + i
		if contentLineIndex >= 0 && contentLineIndex < totalContentLines {
			yOnScreen := contentY + i
			element := drawableElements[contentLineIndex]
			element(cs, grid, contentX, yOnScreen)
		}
	}

	cs.drawScrollIndicators(grid, contentX, contentY, contentWidth, displayHeight, cs.scrollOffset, totalContentLines)
	cs.drawInstructions(grid)
}

func (cs *CharacterScreen) drawSpacer(elements *[]DrawableElement) {
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {}) // Spacing
}

func (cs *CharacterScreen) drawSpacers(elements *[]DrawableElement, numSpacers int) {
	for i := 0; i < numSpacers; i++ {
		cs.drawSpacer(elements)
	}
}

// appendBasicInfoElements appends drawing functions for basic info to the elements slice
func (cs *CharacterScreen) appendBasicInfoElements(elements *[]DrawableElement, ecs *ecs.ECS, playerID ecs.EntityID, contentWidth int) {
	// Character name
	nameText := "Name: Player" // Placeholder
	nameColor := ColorUITitle
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, nameText, drawX, drawY, nameColor)
	})

	// Level and experience
	if ecs.HasExperienceSafe(playerID) {
		exp := ecs.GetExperienceSafe(playerID)
		levelText := fmt.Sprintf("Level: %d", exp.Level)
		textColor := ColorUIText
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, levelText, drawX, drawY, textColor)
		})

		expText := fmt.Sprintf("Experience: %d / %d", exp.CurrentXP, exp.XPToNextLevel)
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, expText, drawX, drawY, textColor)
		})

		barWidth := contentWidth - 15 // Adjusted from original x + 13 relative positioning
		if barWidth > 0 {
			xpProgressText := "XP Progress: "
			currentXP := exp.CurrentXP
			xpToNext := exp.XPToNextLevel
			barStyle := gruid.Style{Fg: ColorStatusGood, Bg: ColorUIBackground}
			*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
				cs.drawText(grid, xpProgressText, drawX, drawY, textColor)
				cs.DrawProgressBar(grid, drawX+13, drawY, barWidth, currentXP, xpToNext, barStyle)
			})
		}

		totalXPText := fmt.Sprintf("Total XP: %d", exp.TotalXP)
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, totalXPText, drawX, drawY, textColor)
		})

		if exp.SkillPoints > 0 || exp.AttributePoints > 0 {
			pointsText := fmt.Sprintf("Skill Points: %d | Attribute Points: %d", exp.SkillPoints, exp.AttributePoints)
			pointsColor := ColorStatusGood
			*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
				cs.drawText(grid, pointsText, drawX, drawY, pointsColor)
			})
		}
	}

	// Health, Mana, Stamina
	if ecs.HasHealthSafe(playerID) {
		health := ecs.GetHealthSafe(playerID)
		healthText := fmt.Sprintf("Health: %d / %d", health.CurrentHP, health.MaxHP)
		textColor := ColorUIText
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, healthText, drawX, drawY, textColor)
		})
	}

	if ecs.HasManaSafe(playerID) {
		mana := ecs.GetManaSafe(playerID)
		manaText := fmt.Sprintf("Mana: %d / %d", mana.CurrentMP, mana.MaxMP)
		textColor := ColorUIText
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, manaText, drawX, drawY, textColor)
		})
	}

	if ecs.HasStaminaSafe(playerID) {
		stamina := ecs.GetStaminaSafe(playerID)
		staminaText := fmt.Sprintf("Stamina: %d / %d", stamina.CurrentSP, stamina.MaxSP)
		textColor := ColorUIText
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, staminaText, drawX, drawY, textColor)
		})
	}
}

// appendAttributesElements appends drawing functions for attributes
func (cs *CharacterScreen) appendAttributesElements(elements *[]DrawableElement, ecs *ecs.ECS, playerID ecs.EntityID, contentWidth int) {
	if !ecs.HasStatsSafe(playerID) {
		return
	}
	stats := ecs.GetStatsSafe(playerID)
	titleText := "=== ATTRIBUTES ==="
	titleColor := ColorUITitle
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, titleText, drawX, drawY, titleColor)
	})

	line1Text := fmt.Sprintf("Strength:     %2d    Dexterity:    %2d", stats.Strength, stats.Dexterity)
	textColor := ColorUIText
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, line1Text, drawX, drawY, textColor)
	})

	line2Text := fmt.Sprintf("Constitution: %2d    Intelligence: %2d", stats.Constitution, stats.Intelligence)
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, line2Text, drawX, drawY, textColor)
	})

	line3Text := fmt.Sprintf("Wisdom:       %2d    Charisma:     %2d", stats.Wisdom, stats.Charisma)
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, line3Text, drawX, drawY, textColor)
	})
}

// appendCombatStatsElements appends drawing functions for combat stats
func (cs *CharacterScreen) appendCombatStatsElements(elements *[]DrawableElement, ecs *ecs.ECS, playerID ecs.EntityID, contentWidth int) {
	if !ecs.HasCombatSafe(playerID) {
		return
	}
	combat := ecs.GetCombatSafe(playerID)
	titleText := "=== COMBAT STATS ==="
	titleColor := ColorUITitle
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, titleText, drawX, drawY, titleColor)
	})

	line1Text := fmt.Sprintf("Attack Power: %2d    Defense:      %2d", combat.AttackPower, combat.Defense)
	textColor := ColorUIText
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, line1Text, drawX, drawY, textColor)
	})

	line2Text := fmt.Sprintf("Accuracy:     %2d%%   Dodge Chance: %2d%%", combat.Accuracy, combat.DodgeChance)
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, line2Text, drawX, drawY, textColor)
	})

	line3Text := fmt.Sprintf("Critical:     %2d%%   Crit Damage:  %2d%%", combat.CriticalChance, combat.CriticalDamage)
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, line3Text, drawX, drawY, textColor)
	})
}

// appendEquipmentElements appends drawing functions for equipment
func (cs *CharacterScreen) appendEquipmentElements(elements *[]DrawableElement, ecs *ecs.ECS, playerID ecs.EntityID, contentWidth int) {
	if !ecs.HasEquipmentSafe(playerID) {
		return
	}
	equipment := ecs.GetEquipmentSafe(playerID)
	titleText := "=== EQUIPMENT ==="
	titleColor := ColorUITitle
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, titleText, drawX, drawY, titleColor)
	})

	textColor := ColorUIText
	highlightColor := ColorUIHighlight

	// Weapon
	if equipment.Weapon != nil {
		weaponName := fmt.Sprintf("Weapon: %s", equipment.Weapon.Name)
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, weaponName, drawX, drawY, textColor)
		})
		weaponDesc := fmt.Sprintf("  %s", equipment.Weapon.Description)
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, weaponDesc, drawX, drawY, highlightColor)
		})
	} else {
		weaponNone := "Weapon: None"
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, weaponNone, drawX, drawY, textColor)
		})
	}

	// Armor
	if equipment.Armor != nil {
		armorName := fmt.Sprintf("Armor:  %s", equipment.Armor.Name)
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, armorName, drawX, drawY, textColor)
		})
		armorDesc := fmt.Sprintf("  %s", equipment.Armor.Description)
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, armorDesc, drawX, drawY, highlightColor)
		})
	} else {
		armorNone := "Armor:  None"
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, armorNone, drawX, drawY, textColor)
		})
	}

	// Accessory
	if equipment.Accessory != nil {
		accessoryName := fmt.Sprintf("Accessory: %s", equipment.Accessory.Name)
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, accessoryName, drawX, drawY, textColor)
		})
		accessoryDesc := fmt.Sprintf("  %s", equipment.Accessory.Description)
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, accessoryDesc, drawX, drawY, highlightColor)
		})
	} else {
		accessoryNone := "Accessory: None"
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, accessoryNone, drawX, drawY, textColor)
		})
	}
}

// appendSkillsElements appends drawing functions for skills
func (cs *CharacterScreen) appendSkillsElements(elements *[]DrawableElement, ecs *ecs.ECS, playerID ecs.EntityID, contentWidth int) {
	if !ecs.HasSkillsSafe(playerID) {
		return
	}
	skills := ecs.GetSkillsSafe(playerID)
	titleText := "=== SKILLS ==="
	titleColor := ColorUITitle
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, titleText, drawX, drawY, titleColor)
	})

	highlightColor := ColorUIHighlight
	textColor := ColorUIText

	// Combat skills
	combatTitle := "Combat:"
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, combatTitle, drawX, drawY, highlightColor)
	})
	combatSkillsText := fmt.Sprintf("  Melee Weapons: %2d    Ranged Weapons: %2d    Defense: %2d",
		skills.MeleeWeapons, skills.RangedWeapons, skills.Defense)
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, combatSkillsText, drawX, drawY, textColor)
	})

	// Magic skills
	magicTitle := "Magic:"
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, magicTitle, drawX, drawY, highlightColor)
	})
	magicSkillsText := fmt.Sprintf("  Evocation: %2d    Conjuration: %2d    Enchantment: %2d    Divination: %2d",
		skills.Evocation, skills.Conjuration, skills.Enchantment, skills.Divination)
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, magicSkillsText, drawX, drawY, textColor)
	})

	// Utility skills
	utilityTitle := "Utility:"
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, utilityTitle, drawX, drawY, highlightColor)
	})
	utilitySkills1Text := fmt.Sprintf("  Stealth: %2d    Lockpicking: %2d    Perception: %2d",
		skills.Stealth, skills.Lockpicking, skills.Perception)
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, utilitySkills1Text, drawX, drawY, textColor)
	})
	utilitySkills2Text := fmt.Sprintf("  Medicine: %2d   Crafting: %2d",
		skills.Medicine, skills.Crafting)
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, utilitySkills2Text, drawX, drawY, textColor)
	})
}

// appendStatusEffectsElements appends drawing functions for status effects
func (cs *CharacterScreen) appendStatusEffectsElements(elements *[]DrawableElement, ecs *ecs.ECS, playerID ecs.EntityID, contentWidth int) {
	if !ecs.HasStatusEffectsSafe(playerID) {
		return
	}

	statusEffects := ecs.GetStatusEffectsSafe(playerID)
	titleText := "=== STATUS EFFECTS ==="
	titleColor := ColorUITitle
	*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
		cs.drawText(grid, titleText, drawX, drawY, titleColor)
	})

	if len(statusEffects.Effects) == 0 {
		noneText := "None"
		textColor := ColorUIText
		*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
			cs.drawText(grid, noneText, drawX, drawY, textColor)
		})
	} else {
		for _, effect := range statusEffects.Effects {
			effectText := fmt.Sprintf("%s (%d turns)", effect.Name, effect.Duration)
			effectColor := ColorStatusNeutral // Or derive from effect type
			// Capture effectText and effectColor in the closure
			currentEffectText := effectText
			currentEffectColor := effectColor
			*elements = append(*elements, func(cs *CharacterScreen, grid gruid.Grid, drawX, drawY int) {
				cs.drawText(grid, currentEffectText, drawX, drawY, currentEffectColor)
			})
		}
	}
}

// drawScrollIndicators shows scroll position and availability
func (cs *CharacterScreen) drawScrollIndicators(grid gruid.Grid, x, y, width, displayHeight, scrollOffset, totalLines int) {
	// Scroll up indicator
	if scrollOffset > 0 {
		indicator := "▲ more above"
		// Draw at the top of the content area, possibly slightly offset
		cs.drawText(grid, indicator, x+(width-len(indicator))/2, y, ColorUIHighlight)
	}

	// Scroll down indicator
	if scrollOffset+displayHeight < totalLines {
		indicator := "▼ more below"
		// Draw at the bottom of the content area
		cs.drawText(grid, indicator, x+(width-len(indicator))/2, y+displayHeight-1, ColorUIHighlight)
	}
}

// drawInstructions renders control instructions at the bottom
func (cs *CharacterScreen) drawInstructions(grid gruid.Grid) {
	instructionY := cs.Y + cs.Height - 2 // Positioned at the bottom of the panel
	instructions := "↑↓/jk/PgUp/PgDn: Scroll | Home: Top | End: Bottom | [ESC]/q: Close"

	// Center the instructions
	startX := cs.X + (cs.Width-len(instructions))/2
	if startX < cs.X+1 { // Ensure it's within panel bounds
		startX = cs.X + 1
	}
	cs.drawText(grid, instructions, startX, instructionY, ColorUIHighlight)
}

// ScrollUp scrolls the content up (view moves towards the top of the sheet)
func (cs *CharacterScreen) ScrollUp(lines int) {
	cs.scrollOffset -= lines
	if cs.scrollOffset < 0 {
		cs.scrollOffset = 0
	}
}

// ScrollDown scrolls the content down (view moves towards the bottom of the sheet)
// It requires gameData to determine the actual maximum scroll extent dynamically.
func (cs *CharacterScreen) ScrollDown(lines int, gameData GameData) {
	cs.scrollOffset += lines // Tentatively update offset

	// Determine current content dimensions for accurate clamping
	_, _, contentWidth, contentHeight := cs.GetContentArea()
	drawableElements := cs.generateDrawableElements(gameData, contentWidth)
	totalContentLines := len(drawableElements)
	displayHeight := contentHeight

	maxScroll := 0
	if totalContentLines > displayHeight { // Can only scroll if content is larger than display area
		maxScroll = totalContentLines - displayHeight
	}

	if cs.scrollOffset > maxScroll {
		cs.scrollOffset = maxScroll
	}
	// Safety clamp: ensure scrollOffset is not negative, though Render usually handles this too.
	if cs.scrollOffset < 0 {
		cs.scrollOffset = 0
	}
}

// ScrollToTop scrolls to the top of the content
func (cs *CharacterScreen) ScrollToTop() {
	cs.scrollOffset = 0
}

// ScrollToBottom scrolls to the bottom of the content
// It requires gameData to determine the actual maximum scroll extent dynamically.
func (cs *CharacterScreen) ScrollToBottom(gameData GameData) {
	// Determine current content dimensions for accurate clamping
	_, _, contentWidth, contentHeight := cs.GetContentArea()
	drawableElements := cs.generateDrawableElements(gameData, contentWidth)
	totalContentLines := len(drawableElements)
	displayHeight := contentHeight

	maxScroll := 0
	if totalContentLines > displayHeight { // Can only scroll if content is larger than display area
		maxScroll = totalContentLines - displayHeight
	}

	if maxScroll < 0 {
		maxScroll = 0
	}
	cs.scrollOffset = maxScroll
}

// GetScrollOffset returns the current vertical scroll offset.
func (cs *CharacterScreen) GetScrollOffset() int {
	return cs.scrollOffset
}

// IsAtTop returns true if the character screen is scrolled to the top.
func (cs *CharacterScreen) IsAtTop() bool {
	return cs.scrollOffset == 0
}

// IsAtBottom returns true if the character screen is scrolled to the bottom.
// It requires gameData to determine the actual maximum scroll extent dynamically.
func (cs *CharacterScreen) IsAtBottom(gameData GameData) bool {
	_, _, contentWidth, contentHeight := cs.GetContentArea()
	drawableElements := cs.generateDrawableElements(gameData, contentWidth)
	totalContentLines := len(drawableElements)
	displayHeight := contentHeight

	maxScroll := 0
	if totalContentLines > displayHeight {
		maxScroll = totalContentLines - displayHeight
	}

	return cs.scrollOffset >= maxScroll // Use >= for robustness
}

// drawLine is a helper, can be removed if not used elsewhere or kept for potential direct drawing
// func (cs *CharacterScreen) drawLine(grid gruid.Grid, text string, x, y int, color gruid.Color) {
// 	cs.drawText(grid, text, x, y, color)
// }

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
