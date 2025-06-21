package game

import (
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui"
	"github.com/sirupsen/logrus"
)

// ExperienceSystem handles experience gain and leveling
type ExperienceSystem struct {
	game *Game
}

// NewExperienceSystem creates a new experience system
func NewExperienceSystem(game *Game) *ExperienceSystem {
	return &ExperienceSystem{game: game}
}

// AwardExperience gives experience to an entity and handles level ups
func (es *ExperienceSystem) AwardExperience(entityID ecs.EntityID, amount int) {
	if !es.game.ecs.HasExperienceSafe(entityID) {
		return
	}

	experience := es.game.ecs.GetExperienceSafe(entityID)
	entityName := es.game.ecs.GetNameSafe(entityID)

	// Add experience
	leveledUp := experience.AddXP(amount)

	// Update the component
	es.game.ecs.AddComponent(entityID, components.CExperience, experience)

	// Log experience gain
	if entityID == es.game.PlayerID {
		es.game.log.AddMessagef(ui.ColorStatusGood, "You gain %d experience points!", amount)
	}

	logrus.Debugf("%s gained %d XP (Total: %d, Level: %d)", entityName, amount, experience.TotalXP, experience.Level)

	// Handle level up
	if leveledUp {
		es.handleLevelUp(entityID)
	}
}

// handleLevelUp processes a level up for an entity
func (es *ExperienceSystem) handleLevelUp(entityID ecs.EntityID) {
	experience := es.game.ecs.GetExperienceSafe(entityID)
	entityName := es.game.ecs.GetNameSafe(entityID)

	// Log level up
	if entityID == es.game.PlayerID {
		es.game.log.AddMessagef(ui.ColorStatusGood, "Level up! You are now level %d!", experience.Level)
		es.game.log.AddMessagef(ui.ColorStatusGood, "You gained %d skill points and %d attribute points!",
			2, 1) // From Experience.levelUp()
	}

	logrus.Infof("%s leveled up to level %d!", entityName, experience.Level)

	// Increase health on level up
	if es.game.ecs.HasHealthSafe(entityID) {
		health := es.game.ecs.GetHealthSafe(entityID)
		healthIncrease := 2 + (experience.Level / 3) // 2-5 HP per level
		health.MaxHP += healthIncrease
		health.CurrentHP += healthIncrease // Full heal on level up
		es.game.ecs.AddComponent(entityID, components.CHealth, health)

		if entityID == es.game.PlayerID {
			es.game.log.AddMessagef(ui.ColorStatusGood, "Your maximum health increased by %d!", healthIncrease)
		}
	}

	// Increase mana on level up
	if es.game.ecs.HasManaSafe(entityID) {
		mana := es.game.ecs.GetManaSafe(entityID)
		manaIncrease := 1 + (experience.Level / 4) // 1-3 MP per level
		mana.MaxMP += manaIncrease
		mana.CurrentMP += manaIncrease // Full mana restore on level up
		es.game.ecs.AddComponent(entityID, components.CMana, mana)

		if entityID == es.game.PlayerID {
			es.game.log.AddMessagef(ui.ColorStatusGood, "Your maximum mana increased by %d!", manaIncrease)
		}
	}

	// Increase stamina on level up
	if es.game.ecs.HasStaminaSafe(entityID) {
		stamina := es.game.ecs.GetStaminaSafe(entityID)
		staminaIncrease := 3 + (experience.Level / 2) // 3-8 SP per level
		stamina.MaxSP += staminaIncrease
		stamina.CurrentSP += staminaIncrease // Full stamina restore on level up
		es.game.ecs.AddComponent(entityID, components.CStamina, stamina)

		if entityID == es.game.PlayerID {
			es.game.log.AddMessagef(ui.ColorStatusGood, "Your maximum stamina increased by %d!", staminaIncrease)
		}
	}

	// Auto-improve combat stats slightly
	if es.game.ecs.HasCombatSafe(entityID) {
		combat := es.game.ecs.GetCombatSafe(entityID)
		combat.AttackPower += 1
		combat.Accuracy += 2
		if experience.Level%3 == 0 {
			combat.Defense += 1
		}
		if experience.Level%5 == 0 {
			combat.CriticalChance += 1
		}
		es.game.ecs.AddComponent(entityID, components.CCombat, combat)

		if entityID == es.game.PlayerID {
			es.game.log.AddMessagef(ui.ColorStatusGood, "Your combat abilities improved!")
		}
	}
}

// GetExperienceForKill calculates experience reward for killing an entity
func (es *ExperienceSystem) GetExperienceForKill(killerID, victimID ecs.EntityID) int {
	baseXP := 10

	// Bonus XP based on victim's level if they have experience
	if es.game.ecs.HasExperienceSafe(victimID) {
		victimExp := es.game.ecs.GetExperienceSafe(victimID)
		baseXP += victimExp.Level * 5
	}

	// Bonus XP based on victim's max health
	if es.game.ecs.HasHealthSafe(victimID) {
		victimHealth := es.game.ecs.GetHealthSafe(victimID)
		baseXP += victimHealth.MaxHP
	}

	// Level difference modifier
	if es.game.ecs.HasExperienceSafe(killerID) {
		killerExp := es.game.ecs.GetExperienceSafe(killerID)
		if es.game.ecs.HasExperienceSafe(victimID) {
			victimExp := es.game.ecs.GetExperienceSafe(victimID)
			levelDiff := victimExp.Level - killerExp.Level

			if levelDiff > 0 {
				// Bonus for killing higher level enemies
				baseXP += levelDiff * 3
			} else if levelDiff < -3 {
				// Reduced XP for killing much lower level enemies
				baseXP = baseXP / 2
			}
		}
	}

	// Minimum XP
	if baseXP < 1 {
		baseXP = 1
	}

	return baseXP
}

// SpendSkillPoint allows spending a skill point to improve a skill
func (es *ExperienceSystem) SpendSkillPoint(entityID ecs.EntityID, skillName string) bool {
	if !es.game.ecs.HasExperienceSafe(entityID) || !es.game.ecs.HasSkillsSafe(entityID) {
		return false
	}

	experience := es.game.ecs.GetExperienceSafe(entityID)
	if experience.SkillPoints <= 0 {
		if entityID == es.game.PlayerID {
			es.game.log.AddMessagef(ui.ColorStatusBad, "You have no skill points to spend!")
		}
		return false
	}

	skills := es.game.ecs.GetSkillsSafe(entityID)

	// Improve the specified skill
	improved := false
	switch skillName {
	case "melee":
		skills.MeleeWeapons++
		improved = true
	case "ranged":
		skills.RangedWeapons++
		improved = true
	case "defense":
		skills.Defense++
		improved = true
	case "stealth":
		skills.Stealth++
		improved = true
	case "perception":
		skills.Perception++
		improved = true
	case "medicine":
		skills.Medicine++
		improved = true
	case "crafting":
		skills.Crafting++
		improved = true
	case "evocation":
		skills.Evocation++
		improved = true
	case "conjuration":
		skills.Conjuration++
		improved = true
	case "enchantment":
		skills.Enchantment++
		improved = true
	case "divination":
		skills.Divination++
		improved = true
	case "lockpicking":
		skills.Lockpicking++
		improved = true
	}

	if improved {
		// Spend the skill point
		experience.SkillPoints--

		// Update components
		es.game.ecs.AddComponent(entityID, components.CExperience, experience)
		es.game.ecs.AddComponent(entityID, components.CSkills, skills)

		if entityID == es.game.PlayerID {
			es.game.log.AddMessagef(ui.ColorStatusGood, "You improved your %s skill!", skillName)
		}

		logrus.Debugf("%s improved %s skill (remaining skill points: %d)",
			es.game.ecs.GetNameSafe(entityID), skillName, experience.SkillPoints)
		return true
	}

	return false
}

// SpendAttributePoint allows spending an attribute point to improve a stat
func (es *ExperienceSystem) SpendAttributePoint(entityID ecs.EntityID, attributeName string) bool {
	if !es.game.ecs.HasExperienceSafe(entityID) || !es.game.ecs.HasStatsSafe(entityID) {
		return false
	}

	experience := es.game.ecs.GetExperienceSafe(entityID)
	if experience.AttributePoints <= 0 {
		if entityID == es.game.PlayerID {
			es.game.log.AddMessagef(ui.ColorStatusBad, "You have no attribute points to spend!")
		}
		return false
	}

	stats := es.game.ecs.GetStatsSafe(entityID)

	// Improve the specified attribute
	improved := false
	switch attributeName {
	case "strength":
		stats.Strength++
		improved = true
	case "dexterity":
		stats.Dexterity++
		improved = true
	case "constitution":
		stats.Constitution++
		improved = true
	case "intelligence":
		stats.Intelligence++
		improved = true
	case "wisdom":
		stats.Wisdom++
		improved = true
	case "charisma":
		stats.Charisma++
		improved = true
	}

	if improved {
		// Spend the attribute point
		experience.AttributePoints--

		// Update components
		es.game.ecs.AddComponent(entityID, components.CExperience, experience)
		es.game.ecs.AddComponent(entityID, components.CStats, stats)

		if entityID == es.game.PlayerID {
			es.game.log.AddMessagef(ui.ColorStatusGood, "You increased your %s!", attributeName)
		}

		logrus.Debugf("%s increased %s (remaining attribute points: %d)",
			es.game.ecs.GetNameSafe(entityID), attributeName, experience.AttributePoints)
		return true
	}

	return false
}
