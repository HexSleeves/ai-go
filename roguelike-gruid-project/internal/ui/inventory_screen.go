package ui

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

// InventoryScreen handles the full-screen inventory display
type InventoryScreen struct {
	*Panel
	selectedIndex int
	scrollOffset  int
}

// NewInventoryScreen creates a new inventory screen
func NewInventoryScreen() *InventoryScreen {
	panel := NewPanel(
		0, 0,
		config.DungeonWidth,
		config.DungeonHeight,
		"Inventory",
		true,
	)
	
	return &InventoryScreen{
		Panel:         panel,
		selectedIndex: 0,
		scrollOffset:  0,
	}
}

// Render draws the inventory screen with item list and details
func (is *InventoryScreen) Render(grid gruid.Grid, gameData GameData) {
	// Clear and draw border
	is.Clear(grid)
	is.DrawBorder(grid)
	
	// Get content area
	contentX, contentY, contentWidth, contentHeight := is.GetContentArea()
	
	playerID := gameData.GetPlayerID()
	if !gameData.ECS().EntityExists(playerID) {
		return
	}
	
	if !gameData.ECS().HasInventorySafe(playerID) {
		is.drawText(grid, "No inventory available", contentX, contentY, ColorUIText)
		return
	}
	
	inventory := gameData.ECS().GetInventorySafe(playerID)
	
	// Calculate layout
	listWidth := contentWidth / 2
	detailWidth := contentWidth - listWidth - 1
	
	// Draw inventory list
	is.drawInventoryList(grid, &inventory, contentX, contentY, listWidth, contentHeight)

	// Draw item details
	if len(inventory.Items) > 0 && is.selectedIndex < len(inventory.Items) {
		is.drawItemDetails(grid, inventory.Items[is.selectedIndex],
			contentX+listWidth+1, contentY, detailWidth, contentHeight)
	}

	// Draw capacity info
	is.drawCapacityInfo(grid, &inventory, contentX, contentY+contentHeight-1)
	
	// Instructions at bottom
	is.drawInstructions(grid)
}

// drawInventoryList renders the list of items
func (is *InventoryScreen) drawInventoryList(grid gruid.Grid, inventory *components.Inventory, x, y, width, height int) {
	// Header
	is.drawText(grid, "Items:", x, y, ColorUITitle)
	y++
	
	if len(inventory.Items) == 0 {
		is.drawText(grid, "Empty", x, y, ColorUIText)
		return
	}
	
	// Calculate visible range
	maxVisible := height - 3 // Account for header and capacity info
	startIndex := is.scrollOffset
	endIndex := startIndex + maxVisible
	if endIndex > len(inventory.Items) {
		endIndex = len(inventory.Items)
	}
	
	// Draw items
	for i := startIndex; i < endIndex; i++ {
		item := inventory.Items[i]
		
		// Selection indicator
		prefix := "  "
		color := ColorUIText
		if i == is.selectedIndex {
			prefix = "> "
			color = ColorUIHighlight
		}
		
		// Item letter (a, b, c, etc.)
		itemLetter := string(rune('a' + i))
		
		// Format item text
		itemText := fmt.Sprintf("%s%s) %s", prefix, itemLetter, item.Item.Name)
		if item.Quantity > 1 {
			itemText += fmt.Sprintf(" (%d)", item.Quantity)
		}
		
		// Truncate if too long
		if len(itemText) > width {
			itemText = itemText[:width-3] + "..."
		}
		
		is.drawText(grid, itemText, x, y, color)
		y++
	}
	
	// Scroll indicators
	if is.scrollOffset > 0 {
		is.drawText(grid, "▲ More above", x, y-maxVisible-1, ColorUIHighlight)
	}
	if endIndex < len(inventory.Items) {
		is.drawText(grid, "▼ More below", x, y, ColorUIHighlight)
	}
}

// drawItemDetails renders detailed information about the selected item
func (is *InventoryScreen) drawItemDetails(grid gruid.Grid, itemStack components.ItemStack, x, y, width, height int) {
	item := itemStack.Item
	
	// Header
	is.drawText(grid, "Item Details:", x, y, ColorUITitle)
	y++
	
	// Item name
	is.drawText(grid, item.Name, x, y, ColorUIHighlight)
	y++
	
	// Item type
	typeText := is.getItemTypeString(item.Type)
	is.drawText(grid, fmt.Sprintf("Type: %s", typeText), x, y, ColorUIText)
	y++
	
	// Quantity
	if itemStack.Quantity > 1 {
		is.drawText(grid, fmt.Sprintf("Quantity: %d", itemStack.Quantity), x, y, ColorUIText)
		y++
	}
	
	// Value
	totalValue := item.Value * itemStack.Quantity
	if itemStack.Quantity > 1 {
		is.drawText(grid, fmt.Sprintf("Value: %d (%d each)", totalValue, item.Value), x, y, ColorUIText)
	} else {
		is.drawText(grid, fmt.Sprintf("Value: %d", item.Value), x, y, ColorUIText)
	}
	y++
	
	// Stackable info
	if item.Stackable {
		is.drawText(grid, fmt.Sprintf("Stackable (max %d)", item.MaxStack), x, y, ColorUIText)
		y++
	}
	
	y++ // Add spacing
	
	// Description (wrapped)
	is.drawText(grid, "Description:", x, y, ColorUIHighlight)
	y++
	
	descLines := is.wrapText(item.Description, width)
	for _, line := range descLines {
		if y >= is.Y+is.Height-3 { // Don't draw over instructions
			break
		}
		is.drawText(grid, line, x, y, ColorUIText)
		y++
	}
	
	y++ // Add spacing
	
	// Actions
	is.drawText(grid, "Actions:", x, y, ColorUIHighlight)
	y++
	
	actions := is.getAvailableActions(item)
	for _, action := range actions {
		if y >= is.Y+is.Height-3 {
			break
		}
		is.drawText(grid, action, x, y, ColorStatusGood)
		y++
	}
}

// drawCapacityInfo renders inventory capacity information
func (is *InventoryScreen) drawCapacityInfo(grid gruid.Grid, inventory *components.Inventory, x, y int) {
	capacityText := fmt.Sprintf("Capacity: %d / %d", len(inventory.Items), inventory.Capacity)
	is.drawText(grid, capacityText, x, y, ColorUIText)
}

// drawInstructions renders control instructions at the bottom
func (is *InventoryScreen) drawInstructions(grid gruid.Grid) {
	instructionY := is.Y + is.Height - 2
	instructions := "↑↓/jk: Select | u: Use | e: Equip | d: Drop | [ESC]/q: Close"
	
	// Center the instructions
	startX := is.X + (is.Width-len(instructions))/2
	if startX < is.X+1 {
		startX = is.X + 1
	}
	
	is.drawText(grid, instructions, startX, instructionY, ColorUIHighlight)
}

// getItemTypeString returns a human-readable item type string
func (is *InventoryScreen) getItemTypeString(itemType components.ItemType) string {
	switch itemType {
	case components.ItemTypeWeapon:
		return "Weapon"
	case components.ItemTypeArmor:
		return "Armor"
	case components.ItemTypeConsumable:
		return "Consumable"
	case components.ItemTypeMisc:
		return "Miscellaneous"
	default:
		return "Unknown"
	}
}

// getAvailableActions returns a list of available actions for an item
func (is *InventoryScreen) getAvailableActions(item components.Item) []string {
	var actions []string
	
	switch item.Type {
	case components.ItemTypeWeapon, components.ItemTypeArmor:
		actions = append(actions, "e) Equip")
		actions = append(actions, "d) Drop")
	case components.ItemTypeConsumable:
		actions = append(actions, "u) Use")
		actions = append(actions, "d) Drop")
	default:
		actions = append(actions, "d) Drop")
	}
	
	return actions
}

// wrapText wraps text to fit within the specified width
func (is *InventoryScreen) wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{}
	}
	
	words := []string{}
	for _, word := range []rune(text) {
		words = append(words, string(word))
	}
	
	var lines []string
	var currentLine string
	
	for _, word := range words {
		if len(currentLine)+len(word) <= width {
			currentLine += word
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}
	
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	
	return lines
}

// ScrollUp moves selection up
func (is *InventoryScreen) ScrollUp() {
	if is.selectedIndex > 0 {
		is.selectedIndex--
		
		// Adjust scroll offset if needed
		if is.selectedIndex < is.scrollOffset {
			is.scrollOffset = is.selectedIndex
		}
	}
}

// ScrollDown moves selection down
func (is *InventoryScreen) ScrollDown(maxItems int) {
	if is.selectedIndex < maxItems-1 {
		is.selectedIndex++
		
		// Adjust scroll offset if needed
		maxVisible := 15 // Approximate visible items
		if is.selectedIndex >= is.scrollOffset+maxVisible {
			is.scrollOffset = is.selectedIndex - maxVisible + 1
		}
	}
}

// GetSelectedIndex returns the currently selected item index
func (is *InventoryScreen) GetSelectedIndex() int {
	return is.selectedIndex
}

// ResetSelection resets the selection to the first item
func (is *InventoryScreen) ResetSelection() {
	is.selectedIndex = 0
	is.scrollOffset = 0
}

// drawText draws text at the specified position
func (is *InventoryScreen) drawText(grid gruid.Grid, text string, x, y int, color gruid.Color) {
	style := gruid.Style{Fg: color, Bg: ColorUIBackground}
	
	for i, r := range text {
		if x+i >= is.X+is.Width-1 { // Don't draw over border
			break
		}
		if x+i < grid.Size().X && y < grid.Size().Y {
			grid.Set(gruid.Point{X: x + i, Y: y}, gruid.Cell{Rune: r, Style: style})
		}
	}
}
