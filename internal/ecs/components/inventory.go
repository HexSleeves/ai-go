package components

import "codeberg.org/anaseto/gruid"

// ItemType represents different categories of items
type ItemType int

const (
	ItemTypeWeapon ItemType = iota
	ItemTypeArmor
	ItemTypeConsumable
	ItemTypeMisc
)

// Item represents a game item
type Item struct {
	Name        string
	Description string
	Type        ItemType
	Glyph       rune
	Color       gruid.Color
	Value       int
	Stackable   bool
	MaxStack    int
}

// ItemStack represents a stack of items in inventory
type ItemStack struct {
	Item     Item
	Quantity int
}

// Inventory component for entities that can carry items
type Inventory struct {
	Items    []ItemStack
	Capacity int
}

// NewInventory creates a new inventory with the specified capacity
func NewInventory(capacity int) Inventory {
	return Inventory{
		Items:    make([]ItemStack, 0),
		Capacity: capacity,
	}
}

// AddItem attempts to add an item to the inventory
func (inv *Inventory) AddItem(item Item, quantity int) bool {
	// Try to stack with existing items if stackable
	if item.Stackable {
		for i := range inv.Items {
			if inv.Items[i].Item.Name == item.Name {
				spaceLeft := item.MaxStack - inv.Items[i].Quantity
				if spaceLeft >= quantity {
					inv.Items[i].Quantity += quantity
					return true
				} else if spaceLeft > 0 {
					inv.Items[i].Quantity = item.MaxStack
					quantity -= spaceLeft
				}
			}
		}
	}

	// Add as new stack if there's space
	if len(inv.Items) < inv.Capacity {
		inv.Items = append(inv.Items, ItemStack{
			Item:     item,
			Quantity: quantity,
		})
		return true
	}

	return false // Inventory full
}

// RemoveItem removes a specified quantity of an item
func (inv *Inventory) RemoveItem(itemName string, quantity int) bool {
	for i := range inv.Items {
		if inv.Items[i].Item.Name == itemName {
			if inv.Items[i].Quantity >= quantity {
				inv.Items[i].Quantity -= quantity
				if inv.Items[i].Quantity == 0 {
					// Remove empty stack
					inv.Items = append(inv.Items[:i], inv.Items[i+1:]...)
				}
				return true
			}
		}
	}
	return false
}

// HasItem checks if inventory contains at least the specified quantity of an item
func (inv *Inventory) HasItem(itemName string, quantity int) bool {
	for _, stack := range inv.Items {
		if stack.Item.Name == itemName && stack.Quantity >= quantity {
			return true
		}
	}
	return false
}

// GetItemCount returns the total quantity of a specific item
func (inv *Inventory) GetItemCount(itemName string) int {
	total := 0
	for _, stack := range inv.Items {
		if stack.Item.Name == itemName {
			total += stack.Quantity
		}
	}
	return total
}

// IsFull returns true if the inventory cannot accept any more item stacks
func (inv *Inventory) IsFull() bool {
	return len(inv.Items) >= inv.Capacity
}

// Equipment component for entities that can equip items
type Equipment struct {
	Weapon    *Item
	Armor     *Item
	Accessory *Item
}

// NewEquipment creates a new empty equipment set
func NewEquipment() Equipment {
	return Equipment{}
}

// EquipItem equips an item in the appropriate slot
func (eq *Equipment) EquipItem(item Item) *Item {
	var oldItem *Item

	switch item.Type {
	case ItemTypeWeapon:
		oldItem = eq.Weapon
		eq.Weapon = &item
	case ItemTypeArmor:
		oldItem = eq.Armor
		eq.Armor = &item
	default:
		return nil // Cannot equip this item type
	}

	return oldItem
}

// UnequipItem removes an item from the specified slot
func (eq *Equipment) UnequipItem(itemType ItemType) *Item {
	var item *Item

	switch itemType {
	case ItemTypeWeapon:
		item = eq.Weapon
		eq.Weapon = nil
	case ItemTypeArmor:
		item = eq.Armor
		eq.Armor = nil
	}

	return item
}

// GetEquippedItem returns the item in the specified slot
func (eq *Equipment) GetEquippedItem(itemType ItemType) *Item {
	switch itemType {
	case ItemTypeWeapon:
		return eq.Weapon
	case ItemTypeArmor:
		return eq.Armor
	}
	return nil
}

// ItemPickup component for items lying on the ground
type ItemPickup struct {
	Item     Item
	Quantity int
}

// NewItemPickup creates a new item pickup
func NewItemPickup(item Item, quantity int) ItemPickup {
	return ItemPickup{
		Item:     item,
		Quantity: quantity,
	}
}
