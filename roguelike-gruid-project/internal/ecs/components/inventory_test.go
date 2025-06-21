package components

import (
	"testing"

	"codeberg.org/anaseto/gruid"
)

func TestInventory_AddItem(t *testing.T) {
	inv := NewInventory(5)

	// Test adding a non-stackable item
	sword := Item{
		Name:      "Iron Sword",
		Type:      ItemTypeWeapon,
		Stackable: false,
		Value:     100,
	}

	if !inv.AddItem(sword, 1) {
		t.Error("Should be able to add item to empty inventory")
	}

	if len(inv.Items) != 1 {
		t.Errorf("Expected 1 item in inventory, got %d", len(inv.Items))
	}

	// Test adding a stackable item
	potion := Item{
		Name:      "Health Potion",
		Type:      ItemTypeConsumable,
		Stackable: true,
		MaxStack:  10,
		Value:     50,
	}

	if !inv.AddItem(potion, 3) {
		t.Error("Should be able to add stackable item")
	}

	if len(inv.Items) != 2 {
		t.Errorf("Expected 2 different items in inventory, got %d", len(inv.Items))
	}

	// Test stacking same item
	if !inv.AddItem(potion, 2) {
		t.Error("Should be able to stack same item")
	}

	if len(inv.Items) != 2 {
		t.Errorf("Expected 2 different items after stacking, got %d", len(inv.Items))
	}

	// Check quantity
	if inv.GetItemCount("Health Potion") != 5 {
		t.Errorf("Expected 5 potions, got %d", inv.GetItemCount("Health Potion"))
	}
}

func TestInventory_RemoveItem(t *testing.T) {
	inv := NewInventory(5)

	potion := Item{
		Name:      "Health Potion",
		Type:      ItemTypeConsumable,
		Stackable: true,
		MaxStack:  10,
		Value:     50,
	}

	inv.AddItem(potion, 5)

	// Remove some items
	if !inv.RemoveItem("Health Potion", 2) {
		t.Error("Should be able to remove items")
	}

	if inv.GetItemCount("Health Potion") != 3 {
		t.Errorf("Expected 3 potions after removal, got %d", inv.GetItemCount("Health Potion"))
	}

	// Remove all remaining items
	if !inv.RemoveItem("Health Potion", 3) {
		t.Error("Should be able to remove all remaining items")
	}

	if len(inv.Items) != 0 {
		t.Errorf("Expected empty inventory after removing all items, got %d items", len(inv.Items))
	}

	// Try to remove non-existent item
	if inv.RemoveItem("Non-existent", 1) {
		t.Error("Should not be able to remove non-existent item")
	}
}

func TestInventory_HasItem(t *testing.T) {
	inv := NewInventory(5)

	potion := Item{
		Name:      "Health Potion",
		Type:      ItemTypeConsumable,
		Stackable: true,
		MaxStack:  10,
		Value:     50,
	}

	inv.AddItem(potion, 3)

	if !inv.HasItem("Health Potion", 2) {
		t.Error("Should have at least 2 potions")
	}

	if !inv.HasItem("Health Potion", 3) {
		t.Error("Should have exactly 3 potions")
	}

	if inv.HasItem("Health Potion", 4) {
		t.Error("Should not have 4 potions")
	}

	if inv.HasItem("Non-existent", 1) {
		t.Error("Should not have non-existent item")
	}
}

func TestInventory_IsFull(t *testing.T) {
	inv := NewInventory(2)

	sword := Item{Name: "Sword", Type: ItemTypeWeapon, Stackable: false}
	shield := Item{Name: "Shield", Type: ItemTypeArmor, Stackable: false}
	helmet := Item{Name: "Helmet", Type: ItemTypeArmor, Stackable: false}

	if inv.IsFull() {
		t.Error("Empty inventory should not be full")
	}

	inv.AddItem(sword, 1)
	if inv.IsFull() {
		t.Error("Inventory with 1/2 items should not be full")
	}

	inv.AddItem(shield, 1)
	if !inv.IsFull() {
		t.Error("Inventory with 2/2 items should be full")
	}

	// Try to add another item
	if inv.AddItem(helmet, 1) {
		t.Error("Should not be able to add item to full inventory")
	}
}

func TestEquipment_EquipItem(t *testing.T) {
	eq := NewEquipment()

	sword := Item{Name: "Iron Sword", Type: ItemTypeWeapon, Value: 100}
	armor := Item{Name: "Leather Armor", Type: ItemTypeArmor, Value: 50}
	misc := Item{Name: "Key", Type: ItemTypeMisc, Value: 1}

	// Equip weapon
	oldWeapon := eq.EquipItem(sword)
	if oldWeapon != nil {
		t.Error("Should not have old weapon when equipping to empty slot")
	}

	if eq.Weapon == nil || eq.Weapon.Name != "Iron Sword" {
		t.Error("Weapon should be equipped")
	}

	// Equip armor
	oldArmor := eq.EquipItem(armor)
	if oldArmor != nil {
		t.Error("Should not have old armor when equipping to empty slot")
	}

	if eq.Armor == nil || eq.Armor.Name != "Leather Armor" {
		t.Error("Armor should be equipped")
	}

	// Try to equip misc item (should fail)
	oldItem := eq.EquipItem(misc)
	if oldItem != nil {
		t.Error("Should not be able to equip misc item")
	}

	// Replace weapon
	betterSword := Item{Name: "Steel Sword", Type: ItemTypeWeapon, Value: 200}
	oldWeapon = eq.EquipItem(betterSword)
	if oldWeapon == nil || oldWeapon.Name != "Iron Sword" {
		t.Error("Should return old weapon when replacing")
	}

	if eq.Weapon.Name != "Steel Sword" {
		t.Error("New weapon should be equipped")
	}
}

func TestEquipment_UnequipItem(t *testing.T) {
	eq := NewEquipment()

	sword := Item{Name: "Iron Sword", Type: ItemTypeWeapon, Value: 100}
	eq.EquipItem(sword)

	// Unequip weapon
	unequipped := eq.UnequipItem(ItemTypeWeapon)
	if unequipped == nil || unequipped.Name != "Iron Sword" {
		t.Error("Should return unequipped weapon")
	}

	if eq.Weapon != nil {
		t.Error("Weapon slot should be empty after unequipping")
	}

	// Try to unequip from empty slot
	unequipped = eq.UnequipItem(ItemTypeWeapon)
	if unequipped != nil {
		t.Error("Should not return item when unequipping from empty slot")
	}
}

func TestItemPickup_NewItemPickup(t *testing.T) {
	item := Item{
		Name:      "Gold Coin",
		Type:      ItemTypeMisc,
		Stackable: true,
		MaxStack:  100,
		Value:     1,
		Glyph:     '$',
		Color:     gruid.Color(0xFFFF00),
	}

	pickup := NewItemPickup(item, 25)

	if pickup.Item.Name != "Gold Coin" {
		t.Error("Pickup should contain the correct item")
	}

	if pickup.Quantity != 25 {
		t.Errorf("Expected quantity 25, got %d", pickup.Quantity)
	}
}
