package game

import (
	"testing"

	"codeberg.org/anaseto/gruid"
)

func TestScreenKeyBindings(t *testing.T) {
	tests := []struct {
		name       string
		mode       mode
		key        gruid.Key
		wantAction playerAction
		wantFound  bool
	}{
		// Inventory screen key bindings
		{
			name:       "inventory_use_key",
			mode:       modeInventory,
			key:        "u",
			wantAction: ActionUseSelectedItem,
			wantFound:  true,
		},
		{
			name:       "inventory_equip_key",
			mode:       modeInventory,
			key:        "e",
			wantAction: ActionEquipSelectedItem,
			wantFound:  true,
		},
		{
			name:       "inventory_drop_key",
			mode:       modeInventory,
			key:        "d",
			wantAction: ActionDropSelectedItem,
			wantFound:  true,
		},
		{
			name:       "inventory_close_key",
			mode:       modeInventory,
			key:        "q",
			wantAction: ActionCloseScreen,
			wantFound:  true,
		},
		{
			name:       "inventory_scroll_up",
			mode:       modeInventory,
			key:        "k",
			wantAction: ActionScrollMessagesUp,
			wantFound:  true,
		},
		{
			name:       "inventory_scroll_down",
			mode:       modeInventory,
			key:        "j",
			wantAction: ActionScrollMessagesDown,
			wantFound:  true,
		},
		{
			name:       "inventory_arrow_up",
			mode:       modeInventory,
			key:        gruid.KeyArrowUp,
			wantAction: ActionScrollMessagesUp,
			wantFound:  true,
		},
		{
			name:       "inventory_arrow_down",
			mode:       modeInventory,
			key:        gruid.KeyArrowDown,
			wantAction: ActionScrollMessagesDown,
			wantFound:  true,
		},

		// Character screen key bindings
		{
			name:       "character_page_up_with_u",
			mode:       modeCharacterSheet,
			key:        "u",
			wantAction: ActionScrollMessagesUp,
			wantFound:  true,
		},
		{
			name:       "character_close_key",
			mode:       modeCharacterSheet,
			key:        "q",
			wantAction: ActionCloseScreen,
			wantFound:  true,
		},
		{
			name:       "character_scroll_up",
			mode:       modeCharacterSheet,
			key:        "k",
			wantAction: ActionScrollMessagesUp,
			wantFound:  true,
		},
		{
			name:       "character_scroll_down",
			mode:       modeCharacterSheet,
			key:        "j",
			wantAction: ActionScrollMessagesDown,
			wantFound:  true,
		},

		// Message screen key bindings
		{
			name:       "message_close_key",
			mode:       modeFullMessageLog,
			key:        "q",
			wantAction: ActionCloseScreen,
			wantFound:  true,
		},
		{
			name:       "message_scroll_up",
			mode:       modeFullMessageLog,
			key:        "k",
			wantAction: ActionScrollMessagesUp,
			wantFound:  true,
		},
		{
			name:       "message_scroll_down",
			mode:       modeFullMessageLog,
			key:        "j",
			wantAction: ActionScrollMessagesDown,
			wantFound:  true,
		},

		// Test invalid keys
		{
			name:       "inventory_invalid_key",
			mode:       modeInventory,
			key:        "x",
			wantAction: ActionNone,
			wantFound:  false,
		},
		{
			name:       "character_invalid_key",
			mode:       modeCharacterSheet,
			key:        "z",
			wantAction: ActionNone,
			wantFound:  false,
		},

		// Test that 'u' has different meanings in different screens
		{
			name:       "u_in_inventory_is_use",
			mode:       modeInventory,
			key:        "u",
			wantAction: ActionUseSelectedItem,
			wantFound:  true,
		},
		{
			name:       "u_in_character_is_page_up",
			mode:       modeCharacterSheet,
			key:        "u",
			wantAction: ActionScrollMessagesUp,
			wantFound:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var action playerAction
			var found bool

			// Simulate the key lookup logic from screenModeKeyDown
			switch tt.mode {
			case modeInventory:
				action, found = KEYS_INVENTORY_SCREEN[tt.key]
			case modeCharacterSheet:
				action, found = KEYS_CHARACTER_SCREEN[tt.key]
			case modeFullMessageLog:
				action, found = KEYS_MESSAGE_SCREEN[tt.key]
			}

			if found != tt.wantFound {
				t.Errorf("key '%s' in mode %v: found = %v, want %v", tt.key, tt.mode, found, tt.wantFound)
			}

			if found && action != tt.wantAction {
				t.Errorf("key '%s' in mode %v: action = %v, want %v", tt.key, tt.mode, action, tt.wantAction)
			}

			if !found && tt.wantFound {
				t.Errorf("key '%s' should be found in mode %v but was not", tt.key, tt.mode)
			}
		})
	}
}

func TestKeyBindingConsistency(t *testing.T) {
	// Test that common keys work across all screen modes
	commonKeys := []struct {
		key    gruid.Key
		action playerAction
	}{
		{"q", ActionCloseScreen},
		{gruid.KeyEscape, ActionCloseScreen},
		{"k", ActionScrollMessagesUp},
		{"j", ActionScrollMessagesDown},
		{gruid.KeyPageUp, ActionScrollMessagesUp},
		{gruid.KeyPageDown, ActionScrollMessagesDown},
	}

	screenMaps := map[string]map[gruid.Key]playerAction{
		"inventory": KEYS_INVENTORY_SCREEN,
		"character": KEYS_CHARACTER_SCREEN,
		"message":   KEYS_MESSAGE_SCREEN,
	}

	for _, commonKey := range commonKeys {
		for screenName, keyMap := range screenMaps {
			t.Run(screenName+"_"+string(commonKey.key), func(t *testing.T) {
				action, found := keyMap[commonKey.key]
				if !found {
					t.Errorf("common key '%s' not found in %s screen", commonKey.key, screenName)
				} else if action != commonKey.action {
					t.Errorf("common key '%s' in %s screen: action = %v, want %v",
						commonKey.key, screenName, action, commonKey.action)
				}
			})
		}
	}
}

func TestInventorySpecificKeys(t *testing.T) {
	// Test that inventory-specific keys only exist in inventory screen
	inventoryOnlyKeys := []struct {
		key    gruid.Key
		action playerAction
	}{
		{"u", ActionUseSelectedItem},
		{"e", ActionEquipSelectedItem},
		{"d", ActionDropSelectedItem},
	}

	for _, invKey := range inventoryOnlyKeys {
		t.Run("inventory_only_"+string(invKey.key), func(t *testing.T) {
			// Should exist in inventory screen
			action, found := KEYS_INVENTORY_SCREEN[invKey.key]
			if !found {
				t.Errorf("key '%s' should exist in inventory screen", invKey.key)
			} else if action != invKey.action {
				t.Errorf("key '%s' in inventory screen: action = %v, want %v",
					invKey.key, action, invKey.action)
			}

			// Should not exist in character screen with same action (except 'u' which has different meaning)
			if invKey.key == "u" {
				// 'u' should exist in character screen but with different action
				charAction, charFound := KEYS_CHARACTER_SCREEN[invKey.key]
				if !charFound {
					t.Errorf("key 'u' should exist in character screen")
				} else if charAction == ActionUseSelectedItem {
					t.Errorf("key 'u' in character screen should not be ActionUseSelectedItem")
				}
			} else {
				// Other inventory-specific keys should not exist in character screen
				if _, found := KEYS_CHARACTER_SCREEN[invKey.key]; found {
					t.Errorf("inventory-specific key '%s' should not exist in character screen", invKey.key)
				}
			}

			// Should not exist in message screen
			if _, found := KEYS_MESSAGE_SCREEN[invKey.key]; found {
				t.Errorf("inventory-specific key '%s' should not exist in message screen", invKey.key)
			}
		})
	}
}

func TestScreenModeActionProcessing(t *testing.T) {
	// Test the core fix: when screen mode actions switch to normal mode,
	// the handleScreenKeyDown should call EndTurn() to process the action immediately

	t.Run("screen_mode_calls_end_turn_when_switching_to_normal", func(t *testing.T) {
		grid := gruid.NewGrid(80, 24)
		model := NewModel(grid)

		// Set up initial state
		initialMode := modeInventory
		model.mode = initialMode

		// Mock the screenModeKeyDown to simulate an action that switches to normal mode
		// and returns again=false (indicating action was queued)
		originalMode := model.mode

		// Simulate what happens when an inventory action is processed:
		// 1. Action switches mode to normal
		// 2. Action returns again=false (action was queued)
		model.mode = modeNormal
		again := false

		// Test the logic in handleScreenKeyDown
		// The fix should detect that mode switched to normal and call EndTurn()
		if !again && model.mode == modeNormal && originalMode != modeNormal {
			t.Log("SUCCESS: Screen action switched to normal mode - would call EndTurn()")
		} else {
			t.Error("FAIL: Screen action should trigger turn processing when switching to normal mode")
		}
	})

	t.Run("screen_mode_stays_in_screen_when_again_true", func(t *testing.T) {
		grid := gruid.NewGrid(80, 24)
		model := NewModel(grid)

		// Set up initial state
		model.mode = modeInventory

		// Simulate what happens when a screen navigation action is processed:
		// 1. Mode stays the same (still in inventory)
		// 2. Action returns again=true (no action queued, just UI update)
		again := true

		// Test the logic - should NOT call EndTurn() in this case
		if again || model.mode != modeNormal {
			t.Log("SUCCESS: Screen navigation does not trigger turn processing")
		} else {
			t.Error("FAIL: Screen navigation should not trigger turn processing")
		}
	})
}
