package game

import "codeberg.org/anaseto/gruid"

var KEYS_NORMAL = map[gruid.Key]playerAction{
	gruid.KeyArrowLeft:  ActionW,
	gruid.KeyArrowDown:  ActionS,
	gruid.KeyArrowUp:    ActionN,
	gruid.KeyArrowRight: ActionE,
	"h":                 ActionW,
	"j":                 ActionS,
	"k":                 ActionN,
	"l":                 ActionE,
	"a":                 ActionW,
	"s":                 ActionS,
	"w":                 ActionN,
	"d":                 ActionE,
	"4":                 ActionW,
	"2":                 ActionS,
	"8":                 ActionN,
	"6":                 ActionE,
	"Q":                 ActionQuit,
	"g":                 ActionPickup,
	"D":                 ActionDrop,
	"i":                 ActionInventory,
	"u":                 ActionUseItem,
	"e":                 ActionEquip,
	".":                 ActionWait,
	gruid.KeySpace:      ActionWait,
	"S":                 ActionSave,
	"L":                 ActionLoad,
	"C":                 ActionCharacterSheet,
	"?":                 ActionHelp,
	gruid.KeyPageUp:     ActionScrollMessagesUp,
	gruid.KeyPageDown:   ActionScrollMessagesDown,
	"M":                 ActionScrollMessagesBottom,
	"V":                 ActionFullMessageLog,
	"T":                 ActionToggleTiles,
}

// KEYS_SCREEN defines key bindings for full-screen UI modes
var KEYS_SCREEN = map[gruid.Key]playerAction{
	gruid.KeyEscape:   ActionCloseScreen,
	"q":               ActionCloseScreen,
	gruid.KeyPageUp:   ActionScrollMessagesUp,
	gruid.KeyPageDown: ActionScrollMessagesDown,
	"k":               ActionScrollMessagesUp,
	"j":               ActionScrollMessagesDown,
}

func keyToDir(k playerAction) (p gruid.Point) {
	switch k {
	case ActionW:
		p = gruid.Point{X: -1, Y: 0}
	case ActionE:
		p = gruid.Point{X: 1, Y: 0}
	case ActionS:
		p = gruid.Point{X: 0, Y: 1}
	case ActionN:
		p = gruid.Point{X: 0, Y: -1}
	}
	return p
}
