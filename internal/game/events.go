package game

import (
	"fmt"
	"log/slog"
	"time"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui"
)

// Game Event System
// This extends the existing gruid.Msg event queue to handle game-specific events

// GameEvent represents various game events that can trigger consequences
type GameEvent interface {
	gruid.Msg // Implements gruid.Msg so it can be queued in the existing eventQueue
	EventType() string
	Execute(g *Game) error
}

// EntityDeathEvent is triggered when an entity dies
type EntityDeathEvent struct {
	VictimID   ecs.EntityID
	KillerID   ecs.EntityID // 0 if no killer (e.g., environmental death)
	VictimName string
	KillerName string
	Timestamp  time.Time
}

func (e EntityDeathEvent) EventType() string { return "entity_death" }

func (e EntityDeathEvent) Execute(g *Game) error {
	slog.Info("Processing death event", "victim", e.VictimName, "killer", e.KillerName)

	// Award experience to killer
	if e.KillerID != 0 && g.ecs.EntityExists(e.KillerID) {
		expSystem := NewExperienceSystem(g)
		if g.ecs.HasComponent(e.VictimID, components.CAITag) {
			xpReward := expSystem.GetExperienceForKill(e.KillerID, e.VictimID)
			expSystem.AwardExperience(e.KillerID, xpReward)
		}
	}

	// Check for special death consequences
	if e.VictimID == g.PlayerID {
		g.QueueEvent(GameOverEvent{Reason: "Player death", Timestamp: time.Now()})
	}

	// Check for quest completion or story triggers
	// This is where you'd add quest system integration

	return nil
}

// ItemPickupEvent is triggered when an item is picked up
type ItemPickupEvent struct {
	EntityID   ecs.EntityID
	ItemID     ecs.EntityID
	EntityName string
	ItemName   string
	Quantity   int
	Timestamp  time.Time
}

func (e ItemPickupEvent) EventType() string { return "item_pickup" }

func (e ItemPickupEvent) Execute(g *Game) error {
	slog.Debug("Processing pickup event", "entity", e.EntityName, "item", e.ItemName, "quantity", e.Quantity)

	// Check for auto-equip behavior
	if e.EntityID == g.PlayerID {
		if err := g.tryAutoEquip(e.ItemName); err != nil {
			slog.Debug("Auto-equip failed", "item", e.ItemName, "error", err)
		}
	}

	// Check for item-specific consequences
	switch e.ItemName {
	case "Cursed Ring", "cursed_ring":
		// Cursed items might have negative effects
		g.log.AddMessagef(ui.ColorStatusBad, "The %s feels unnaturally cold...", e.ItemName)
	case "Magic Scroll", "magic_scroll":
		// Picking up certain items might trigger events
		g.log.AddMessagef(ui.ColorStatusGood, "Ancient magic pulses through the %s.", e.ItemName)
	}

	return nil
}

// CombatEvent is triggered during combat actions
type CombatEvent struct {
	AttackerID   ecs.EntityID
	TargetID     ecs.EntityID
	AttackerName string
	TargetName   string
	Damage       int
	Critical     bool
	Timestamp    time.Time
}

func (e CombatEvent) EventType() string { return "combat" }

func (e CombatEvent) Execute(g *Game) error {
	slog.Debug("Processing combat event", "attacker", e.AttackerName, "target", e.TargetName, "damage", e.Damage, "critical", e.Critical)

	// Apply combat consequences
	if e.Critical {
		g.log.AddMessagef(ui.ColorCritical, "Critical hit!")
	}

	// Check for special combat triggers
	if e.TargetID == g.PlayerID && e.Damage > 5 {
		// High damage might trigger special effects
		g.QueueEvent(ScreenShakeEvent{Intensity: e.Damage, Duration: time.Millisecond * 200})
	}

	// Check for weapon/armor durability effects
	// This is where you'd implement equipment degradation

	return nil
}

// LevelUpEvent is triggered when an entity levels up
type LevelUpEvent struct {
	EntityID    ecs.EntityID
	EntityName  string
	NewLevel    int
	SkillPoints int
	StatPoints  int
	Timestamp   time.Time
}

func (e LevelUpEvent) EventType() string { return "level_up" }

func (e LevelUpEvent) Execute(g *Game) error {
	slog.Info("Processing level up event", "entity", e.EntityName, "level", e.NewLevel)

	if e.EntityID == g.PlayerID {
		g.log.AddMessagef(ui.ColorStatusGood, "Congratulations! You have reached level %d!", e.NewLevel)
		g.log.AddMessagef(ui.ColorStatusGood, "You gain %d skill points and %d stat points.", e.SkillPoints, e.StatPoints)

		// Queue UI notification for stat allocation
		g.QueueEvent(StatAllocationEvent{EntityID: e.EntityID, Points: e.StatPoints})
	}

	return nil
}

// GameOverEvent is triggered when the game ends
type GameOverEvent struct {
	Reason    string
	Timestamp time.Time
}

func (e GameOverEvent) EventType() string { return "game_over" }

func (e GameOverEvent) Execute(g *Game) error {
	slog.Info("Processing game over event", "reason", e.Reason)

	g.log.AddMessagef(ui.ColorCritical, "GAME OVER: %s", e.Reason)
	g.setGameOverState()

	return nil
}

// ScreenShakeEvent creates visual feedback
type ScreenShakeEvent struct {
	Intensity int
	Duration  time.Duration
	Timestamp time.Time
}

func (e ScreenShakeEvent) EventType() string { return "screen_shake" }

func (e ScreenShakeEvent) Execute(g *Game) error {
	// This would be handled by the UI system
	slog.Debug("Screen shake effect", "intensity", e.Intensity, "duration", e.Duration)
	return nil
}

// StatAllocationEvent prompts for stat allocation
type StatAllocationEvent struct {
	EntityID  ecs.EntityID
	Points    int
	Timestamp time.Time
}

func (e StatAllocationEvent) EventType() string { return "stat_allocation" }

func (e StatAllocationEvent) Execute(g *Game) error {
	// This would open a stat allocation UI
	slog.Debug("Stat allocation needed", "entity", e.EntityID, "points", e.Points)
	return nil
}

// Event Queue Management Extensions

// EventQueuer interface for queueing events
type EventQueuer interface {
	QueueEvent(msg gruid.Msg)
}

// QueueEvent adds a game event to the model's event queue
func (g *Game) QueueEvent(event GameEvent) {
	if g.model != nil {
		g.model.QueueEvent(event)
	} else {
		slog.Warn("Cannot queue game event: model not set", "type", event.EventType())
	}
}

// Helper functions for common event scenarios

// tryAutoEquip attempts to automatically equip items for the player
func (g *Game) tryAutoEquip(itemName string) error {
	inventory := g.ecs.GetInventorySafe(g.PlayerID)

	// Find the item in inventory
	var foundItem *components.Item
	for _, stack := range inventory.Items {
		if stack.Item.Name == itemName {
			foundItem = &stack.Item
			break
		}
	}

	if foundItem == nil {
		return fmt.Errorf("item not found in inventory: %s", itemName)
	}

	// Only auto-equip if no equipment in that slot
	switch foundItem.Type {
	case components.ItemTypeWeapon:
		equipment := g.ecs.GetEquipmentSafe(g.PlayerID)
		if equipment.Weapon == nil {
			g.log.AddMessagef(ui.ColorStatusGood, "Auto-equipping %s.", itemName)
			// Queue equip action
			actor, _ := g.ecs.GetTurnActor(g.PlayerID)
			actor.AddAction(EquipAction{EntityID: g.PlayerID, ItemName: itemName})
		}
	case components.ItemTypeArmor:
		equipment := g.ecs.GetEquipmentSafe(g.PlayerID)
		if equipment.Armor == nil {
			g.log.AddMessagef(ui.ColorStatusGood, "Auto-equipping %s.", itemName)
			// Queue equip action
			actor, _ := g.ecs.GetTurnActor(g.PlayerID)
			actor.AddAction(EquipAction{EntityID: g.PlayerID, ItemName: itemName})
		}
	}

	return nil
}

// TriggerDeathEvent creates and queues a death event
func (g *Game) TriggerDeathEvent(victimID, killerID ecs.EntityID) {
	victimName := g.ecs.GetNameSafe(victimID)
	killerName := ""
	if killerID != 0 {
		killerName = g.ecs.GetNameSafe(killerID)
	}

	event := EntityDeathEvent{
		VictimID:   victimID,
		KillerID:   killerID,
		VictimName: victimName,
		KillerName: killerName,
		Timestamp:  time.Now(),
	}

	g.QueueEvent(event)
}

// TriggerPickupEvent creates and queues a pickup event
func (g *Game) TriggerPickupEvent(entityID, itemID ecs.EntityID, quantity int) {
	entityName := g.ecs.GetNameSafe(entityID)
	itemName := g.ecs.GetNameSafe(itemID)

	event := ItemPickupEvent{
		EntityID:   entityID,
		ItemID:     itemID,
		EntityName: entityName,
		ItemName:   itemName,
		Quantity:   quantity,
		Timestamp:  time.Now(),
	}

	g.QueueEvent(event)
}

// TriggerCombatEvent creates and queues a combat event
func (g *Game) TriggerCombatEvent(attackerID, targetID ecs.EntityID, damage int, critical bool) {
	attackerName := g.ecs.GetNameSafe(attackerID)
	targetName := g.ecs.GetNameSafe(targetID)

	event := CombatEvent{
		AttackerID:   attackerID,
		TargetID:     targetID,
		AttackerName: attackerName,
		TargetName:   targetName,
		Damage:       damage,
		Critical:     critical,
		Timestamp:    time.Now(),
	}

	g.QueueEvent(event)
}

// TriggerLevelUpEvent creates and queues a level up event
func (g *Game) TriggerLevelUpEvent(entityID ecs.EntityID, newLevel, skillPoints, statPoints int) {
	entityName := g.ecs.GetNameSafe(entityID)

	event := LevelUpEvent{
		EntityID:    entityID,
		EntityName:  entityName,
		NewLevel:    newLevel,
		SkillPoints: skillPoints,
		StatPoints:  statPoints,
		Timestamp:   time.Now(),
	}

	g.QueueEvent(event)
}
