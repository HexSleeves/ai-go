package game

import (
	"fmt"
	"log/slog"

	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui"
)

// PickupAction represents an action to pick up an item
type PickupAction struct {
	EntityID ecs.EntityID
	ItemID   ecs.EntityID
}

func (a PickupAction) Execute(g *Game) (cost uint, err error) {
	// Check if entity has inventory
	if !g.ecs.HasInventorySafe(a.EntityID) {
		return 0, fmt.Errorf("entity %d has no inventory", a.EntityID)
	}

	// Check if item exists and has ItemPickup component
	if !g.ecs.EntityExists(a.ItemID) || !g.ecs.HasItemPickupSafe(a.ItemID) {
		return 0, fmt.Errorf("item %d does not exist or is not pickupable", a.ItemID)
	}

	// Get components
	inventory := g.ecs.GetInventorySafe(a.EntityID)
	pickup := g.ecs.GetItemPickupSafe(a.ItemID)
	entityName := g.ecs.GetNameSafe(a.EntityID)

	// Try to add item to inventory
	if inventory.AddItem(pickup.Item, pickup.Quantity) {
		// Update inventory component
		g.ecs.AddComponent(a.EntityID, components.CInventory, inventory)

		// Grab position before we delete the entity
		itemPos := g.ecs.GetPositionSafe(a.ItemID)

		// Remove from world & grid
		g.ecs.RemoveEntity(a.ItemID)
		g.spatialGrid.Remove(a.ItemID, itemPos)

		// Track item collection statistics
		if a.EntityID == g.PlayerID {
			g.IncrementItemsCollected()
		}

		// Trigger pickup event
		g.TriggerPickupEvent(a.EntityID, a.ItemID, pickup.Quantity)

		// Log message
		if a.EntityID == g.PlayerID {
			g.log.AddMessagef(ui.ColorStatusGood, "You pick up %s.", pickup.Item.Name)
		}

		slog.Debug("Picked up item", "entity", entityName, "item", pickup.Item.Name, "quantity", pickup.Quantity)
		return 100, nil // Standard action cost
	}

	// Inventory full
	if a.EntityID == g.PlayerID {
		g.log.AddMessagef(ui.ColorStatusBad, "Your inventory is full!")
	}
	return 0, fmt.Errorf("inventory full") // No cost if inventory is full and pickup fails
}

// DropAction represents an action to drop an item
type DropAction struct {
	EntityID ecs.EntityID
	ItemName string
	Quantity int
}

func (a DropAction) Execute(g *Game) (cost uint, err error) {
	// Check if entity has inventory
	if !g.ecs.HasInventorySafe(a.EntityID) {
		return 0, fmt.Errorf("entity %d has no inventory", a.EntityID)
	}

	// Get components
	inventory := g.ecs.GetInventorySafe(a.EntityID)
	entityPos := g.ecs.GetPositionSafe(a.EntityID)
	entityName := g.ecs.GetNameSafe(a.EntityID)

	// Check if entity has the item
	if !inventory.HasItem(a.ItemName, a.Quantity) {
		if a.EntityID == g.PlayerID {
			g.log.AddMessagef(ui.ColorStatusBad, "You don't have enough %s to drop.", a.ItemName)
		}
		return 0, fmt.Errorf("not enough items to drop")
	}

	// Find the item in inventory to get its details
	var itemToDrop components.Item
	found := false
	for _, stack := range inventory.Items {
		if stack.Item.Name == a.ItemName {
			itemToDrop = stack.Item
			found = true
			break
		}
	}

	if !found {
		return 0, fmt.Errorf("item %s not found in inventory", a.ItemName)
	}

	// Remove item from inventory
	if inventory.RemoveItem(a.ItemName, a.Quantity) {
		// Update inventory component
		g.ecs.AddComponent(a.EntityID, components.CInventory, inventory)

		// Create item pickup entity
		itemEntity := g.ecs.AddEntity()
		pickup := components.NewItemPickup(itemToDrop, a.Quantity)

		g.ecs.AddComponents(itemEntity,
			entityPos,
			pickup,
			components.Renderable{Glyph: itemToDrop.Glyph, Color: itemToDrop.Color},
			components.Name{Name: itemToDrop.Name},
		)

		// Add to spatial grid
		g.spatialGrid.Add(itemEntity, entityPos)

		// Log message
		if a.EntityID == g.PlayerID {
			g.log.AddMessagef(ui.ColorStatusGood, "You drop %s.", a.ItemName)
		}

		slog.Debug("Dropped item", "entity", entityName, "item", a.ItemName, "quantity", a.Quantity)
		return 100, nil
	}

	return 0, fmt.Errorf("failed to remove item from inventory")
}

// UseItemAction represents an action to use a consumable item
type UseItemAction struct {
	EntityID ecs.EntityID
	ItemName string
}

func (a UseItemAction) Execute(g *Game) (cost uint, err error) {
	// Check if entity has inventory
	if !g.ecs.HasInventorySafe(a.EntityID) {
		return 0, fmt.Errorf("entity %d has no inventory", a.EntityID)
	}

	// Get components
	inventory := g.ecs.GetInventorySafe(a.EntityID)
	entityName := g.ecs.GetNameSafe(a.EntityID)

	// Check if entity has the item
	if !inventory.HasItem(a.ItemName, 1) {
		if a.EntityID == g.PlayerID {
			g.log.AddMessagef(ui.ColorStatusBad, "You don't have %s.", a.ItemName)
		}
		return 0, fmt.Errorf("item not found in inventory")
	}

	// Find the item to check if it's consumable
	var itemToUse components.Item
	found := false
	for _, stack := range inventory.Items {
		if stack.Item.Name == a.ItemName {
			itemToUse = stack.Item
			found = true
			break
		}
	}

	if !found {
		return 0, fmt.Errorf("item %s not found in inventory", a.ItemName)
	}

	// Check if item is consumable
	if itemToUse.Type != components.ItemTypeConsumable {
		if a.EntityID == g.PlayerID {
			g.log.AddMessagef(ui.ColorStatusBad, "You can't use %s.", a.ItemName)
		}
		return 0, fmt.Errorf("item is not consumable")
	}

	// Apply item effects (simplified - just health potion for now)
	if itemToUse.Name == "Health Potion" {
		if g.ecs.HasHealthSafe(a.EntityID) {
			health := g.ecs.GetHealthSafe(a.EntityID)
			healAmount := 10 // Fixed heal amount for now
			health.CurrentHP += healAmount
			if health.CurrentHP > health.MaxHP {
				health.CurrentHP = health.MaxHP
			}
			g.ecs.AddComponent(a.EntityID, components.CHealth, health)

			if a.EntityID == g.PlayerID {
				g.log.AddMessagef(ui.ColorStatusGood, "You feel better! (+%d HP)", healAmount)
			}
		}
	}

	// Remove item from inventory
	if inventory.RemoveItem(a.ItemName, 1) {
		g.ecs.AddComponent(a.EntityID, components.CInventory, inventory)

		if a.EntityID == g.PlayerID {
			g.log.AddMessagef(ui.ColorStatusGood, "You use %s.", a.ItemName)
		}

		slog.Debug("Used item", "entity", entityName, "item", a.ItemName)
		return 100, nil
	}

	return 0, fmt.Errorf("failed to remove item from inventory")
}

// EquipAction represents an action to equip an item
type EquipAction struct {
	EntityID ecs.EntityID
	ItemName string
}

func (a EquipAction) Execute(g *Game) (cost uint, err error) {
	// Check if entity has inventory and equipment
	if !g.ecs.HasInventorySafe(a.EntityID) || !g.ecs.HasEquipmentSafe(a.EntityID) {
		return 0, fmt.Errorf("entity %d missing inventory or equipment", a.EntityID)
	}

	// Get components
	inventory := g.ecs.GetInventorySafe(a.EntityID)
	equipment := g.ecs.GetEquipmentSafe(a.EntityID)
	entityName := g.ecs.GetNameSafe(a.EntityID)

	// Find item in inventory
	var itemToEquip components.Item
	found := false
	for _, stack := range inventory.Items {
		if stack.Item.Name == a.ItemName {
			itemToEquip = stack.Item
			found = true
			break
		}
	}

	if !found {
		if a.EntityID == g.PlayerID {
			g.log.AddMessagef(ui.ColorStatusBad, "You don't have %s.", a.ItemName)
		}
		return 0, fmt.Errorf("item not found in inventory")
	}

	// Check if item is equippable
	if itemToEquip.Type != components.ItemTypeWeapon && itemToEquip.Type != components.ItemTypeArmor {
		if a.EntityID == g.PlayerID {
			g.log.AddMessagef(ui.ColorStatusBad, "You can't equip %s.", a.ItemName)
		}
		return 0, fmt.Errorf("item is not equippable")
	}

	// Try to equip item
	oldItem := equipment.EquipItem(itemToEquip)

	// Remove item from inventory
	if !inventory.RemoveItem(a.ItemName, 1) {
		return 0, fmt.Errorf("failed to remove item from inventory")
	}

	// If there was an old item, add it back to inventory
	if oldItem != nil {
		if !inventory.AddItem(*oldItem, 1) {
			// Inventory full, drop the old item
			dropAction := DropAction{
				EntityID: a.EntityID,
				ItemName: oldItem.Name,
				Quantity: 1,
			}

			// Inventory is full – just drop the old item directly
			if _, err := dropAction.Execute(g); err != nil {
				return 0, fmt.Errorf("failed to drop old item: %w", err)
			}
		}
	}

	// Update components
	g.ecs.AddComponent(a.EntityID, components.CInventory, inventory)
	g.ecs.AddComponent(a.EntityID, components.CEquipment, equipment)

	// Log message
	if a.EntityID == g.PlayerID {
		if oldItem != nil {
			g.log.AddMessagef(ui.ColorStatusGood, "You equip %s (unequipped %s).", a.ItemName, oldItem.Name)
		} else {
			g.log.AddMessagef(ui.ColorStatusGood, "You equip %s.", a.ItemName)
		}
	}

	slog.Debug("Equipped item", "entity", entityName, "item", a.ItemName)
	return 100, nil
}
