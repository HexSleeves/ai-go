package ui

import (
	"image/color"

	"codeberg.org/anaseto/gruid"
)

// Base color palette - consistent with terminals using 16-color palette.
// These are the foundation colors that will be used to derive specific game element colors.
const (
	ColorBackground          gruid.Color = gruid.ColorDefault // background
	ColorBackgroundSecondary gruid.Color = 1 + 0              // black
	ColorForeground          gruid.Color = gruid.ColorDefault // foreground
	ColorForegroundSecondary gruid.Color = 1 + 7              // white
	ColorForegroundEmph      gruid.Color = 1 + 15             // bright white
	ColorYellow              gruid.Color = 1 + 3              // yellow
	ColorOrange              gruid.Color = 1 + 1              // red (used as orange)
	ColorRed                 gruid.Color = 1 + 9              // bright red
	ColorMagenta             gruid.Color = 1 + 5              // magenta
	ColorViolet              gruid.Color = 1 + 12             // bright blue (used as violet)
	ColorBlue                gruid.Color = 1 + 4              // blue
	ColorCyan                gruid.Color = 1 + 6              // cyan
	ColorGreen               gruid.Color = 1 + 2              // green
)

// Game element colors - each element in the game gets a semantic color.
var (
	ColorBg,
	ColorBgDark,
	ColorBgLOS,

	// Map colors
	ColorWall,
	ColorFloor,
	ColorExploredWall,
	ColorExploredFloor,
	ColorVisibleWall,
	ColorVisibleFloor,

	// Entity colors
	ColorPlayer,
	ColorMonster,
	ColorSleepingMonster,
	ColorConfusedMonster,
	ColorParalyzedMonster,
	ColorItem,
	ColorSpecialItem,

	// UI colors
	ColorUIBackground,
	ColorUIBorder,
	ColorUIText,
	ColorUITitle,
	ColorUIHighlight,

	// Status colors
	ColorHealthOk,
	ColorHealthWounded,
	ColorHealthCritical,
	ColorStatusGood,
	ColorStatusBad,
	ColorStatusNeutral,

	// Debug colors
	ColorDebugFOVVisible,
	ColorDebugFOVExplored,
	ColorDebugFOVUnexplored,
	ColorDebugPathChasing,
	ColorDebugPathFleeing,
	ColorDebugPathPatrolling,
	ColorDebugPathSearching,
	ColorDebugTarget,
	ColorDebugWaypoint,
	ColorDebugAIPanel gruid.Color
)

// Style attributes
const (
	AttrNone    gruid.AttrMask = 0
	AttrReverse gruid.AttrMask = 1 << iota
	AttrBlink
	AttrUnderline
	AttrBold
)

// Initialize colors with their semantic mappings
func init() {
	ColorBg = ColorBackground
	ColorBgDark = ColorBackground
	ColorBgLOS = ColorBackgroundSecondary

	// Map colors
	ColorWall = ColorForegroundSecondary
	ColorFloor = ColorBackgroundSecondary
	ColorExploredWall = ColorBackgroundSecondary
	ColorExploredFloor = ColorBackground
	ColorVisibleWall = ColorForegroundEmph
	ColorVisibleFloor = ColorForeground

	// Entity colors
	ColorPlayer = ColorBlue
	ColorMonster = ColorRed
	ColorSleepingMonster = ColorViolet
	ColorConfusedMonster = ColorGreen
	ColorParalyzedMonster = ColorCyan
	ColorItem = ColorYellow
	ColorSpecialItem = ColorMagenta

	// UI colors
	ColorUIBackground = ColorBackground
	ColorUIBorder = ColorForegroundSecondary
	ColorUIText = ColorForeground
	ColorUITitle = ColorForegroundEmph
	ColorUIHighlight = ColorYellow

	// Status colors
	ColorHealthOk = ColorGreen
	ColorHealthWounded = ColorYellow
	ColorHealthCritical = ColorRed
	ColorStatusGood = ColorBlue
	ColorStatusBad = ColorRed
	ColorStatusNeutral = ColorYellow

	// Debug colors
	ColorDebugFOVVisible = ColorForegroundEmph     // Bright white for currently visible
	ColorDebugFOVExplored = ColorBackgroundSecondary // Dim gray for explored
	ColorDebugFOVUnexplored = ColorBackground      // Very dark for unexplored
	ColorDebugPathChasing = ColorRed               // Red for aggressive/chase behavior
	ColorDebugPathFleeing = ColorYellow            // Yellow for flee/retreat behavior
	ColorDebugPathPatrolling = ColorBlue           // Blue for patrol/wander behavior
	ColorDebugPathSearching = ColorGreen           // Green for searching/other behaviors
	ColorDebugTarget = ColorMagenta                // Magenta for target destinations
	ColorDebugWaypoint = ColorCyan                 // Cyan for waypoints
	ColorDebugAIPanel = ColorForegroundSecondary   // Gray for AI debug panel
}

// Helper function to get a style for a map cell based on explored/visible state
func GetMapStyle(isWall bool, isVisible bool, isExplored bool) gruid.Style {
	if !isExplored {
		// Unexplored cells should not be rendered
		return gruid.Style{}
	}

	if isVisible {
		// Currently visible
		if isWall {
			return gruid.Style{Fg: ColorVisibleWall}
		}
		return gruid.Style{Fg: ColorVisibleFloor}
	}

	// Explored but not visible
	if isWall {
		return gruid.Style{Fg: ColorExploredWall}
	}
	return gruid.Style{Fg: ColorExploredFloor}
}

func ColorToRGBA(c gruid.Color, fg bool) color.RGBA {
	var cl color.RGBA
	opaque := uint8(255)

	// Foreground colors
	switch c {
	case ColorRed:
		cl = color.RGBA{220, 50, 47, opaque}
	case ColorGreen:
		cl = color.RGBA{133, 153, 0, opaque}
	case ColorYellow:
		cl = color.RGBA{181, 137, 0, opaque}
	case ColorBlue:
		cl = color.RGBA{38, 139, 210, opaque}
	case ColorMagenta:
		cl = color.RGBA{211, 54, 130, opaque}
	case ColorCyan:
		cl = color.RGBA{42, 161, 152, opaque}
	case ColorOrange:
		cl = color.RGBA{203, 75, 22, opaque}
	case ColorViolet:
		cl = color.RGBA{108, 113, 196, opaque}

	case ColorBackgroundSecondary:
		cl = color.RGBA{7, 54, 66, opaque}
	case ColorForegroundEmph:
		cl = color.RGBA{147, 161, 161, opaque}
	case ColorForegroundSecondary:
		cl = color.RGBA{88, 110, 117, opaque}

	default:
		cl = color.RGBA{0, 43, 54, opaque}
		if fg {
			cl = color.RGBA{131, 148, 150, opaque}
		}
	}

	return cl
}
