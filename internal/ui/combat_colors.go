package ui

import (
	"codeberg.org/anaseto/gruid"
)

// Combat-related colors
var (
	// Attack colors
	ColorPlayerAttack  gruid.Color // Player attacks monsters
	ColorEnemyAttack   gruid.Color // Monsters attack player
	ColorNeutralAttack gruid.Color // Monster attacks another monster

	// Status colors
	ColorDeath    gruid.Color // For death messages
	ColorCorpse   gruid.Color // For corpse messages
	ColorCritical gruid.Color // For critical messages
)

func init() {
	// Initialize combat colors
	ColorPlayerAttack = ColorBlue    // Same as player color
	ColorEnemyAttack = ColorOrange   // Same as monster color
	ColorNeutralAttack = ColorYellow // Neutral color for monster-monster

	ColorDeath = ColorRed                  // Death messages are red
	ColorCorpse = ColorForegroundSecondary // Corpse messages are corpse color
	ColorCritical = ColorRed               // Critical messages are bright white
}
