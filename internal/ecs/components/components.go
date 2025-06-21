package components

import (
	"reflect"

	"codeberg.org/anaseto/gruid"
)

// ComponentType is a string identifier for component types
type ComponentType string

// Component type constants
const (
	CAIComponent          ComponentType = "AIComponent"
	CAITag                ComponentType = "AITag"
	CBlocksMovement       ComponentType = "BlocksMovement"
	CCorpseTag            ComponentType = "CorpseTag"
	CEquipment            ComponentType = "Equipment"
	CFOV                  ComponentType = "FOV"
	CHealth               ComponentType = "Health"
	CInventory            ComponentType = "Inventory"
	CItemPickup           ComponentType = "ItemPickup"
	CName                 ComponentType = "Name"
	CPlayerTag            ComponentType = "PlayerTag"
	CPosition             ComponentType = "Position"
	CRenderable           ComponentType = "Renderable"
	CStats                ComponentType = "Stats"
	CExperience           ComponentType = "Experience"
	CSkills               ComponentType = "Skills"
	CCombat               ComponentType = "Combat"
	CMana                 ComponentType = "Mana"
	CStamina              ComponentType = "Stamina"
	CStatusEffects        ComponentType = "StatusEffects"
	CTurnActor            ComponentType = "TurnActor"
	CPathfindingComponent ComponentType = "PathfindingComponent"
)

var TypeToComponent = map[ComponentType]reflect.Type{
	CAIComponent:          reflect.TypeOf(AIComponent{}),
	CAITag:                reflect.TypeOf(AITag{}),
	CBlocksMovement:       reflect.TypeOf(BlocksMovement{}),
	CCorpseTag:            reflect.TypeOf(CorpseTag{}),
	CEquipment:            reflect.TypeOf(Equipment{}),
	CFOV:                  reflect.TypeOf((*FOV)(nil)),
	CHealth:               reflect.TypeOf(Health{}),
	CInventory:            reflect.TypeOf(Inventory{}),
	CItemPickup:           reflect.TypeOf(ItemPickup{}),
	CName:                 reflect.TypeOf(""),
	CPlayerTag:            reflect.TypeOf(PlayerTag{}),
	CPosition:             reflect.TypeOf(gruid.Point{}),
	CRenderable:           reflect.TypeOf(Renderable{}),
	CStats:                reflect.TypeOf(Stats{}),
	CExperience:           reflect.TypeOf(Experience{}),
	CSkills:               reflect.TypeOf(Skills{}),
	CCombat:               reflect.TypeOf(Combat{}),
	CMana:                 reflect.TypeOf(Mana{}),
	CStamina:              reflect.TypeOf(Stamina{}),
	CStatusEffects:        reflect.TypeOf(StatusEffects{}),
	CTurnActor:            reflect.TypeOf(TurnActor{}),
	CPathfindingComponent: reflect.TypeOf(PathfindingComponent{}),
}

// GetGoType returns the corresponding Go type for a ComponentType
func GetGoType(compType ComponentType) (reflect.Type, bool) {
	t, ok := TypeToComponent[compType]
	return t, ok
}

// Name component represents an entity's name
type Name struct {
	Name string
}

// Position component represents an entity's position in the game world
type Position struct {
	Point gruid.Point
}

// Renderable component represents how an entity is rendered
type Renderable struct {
	Glyph    rune
	Color    gruid.Color
	TileName string // Optional specific tile override for tile-based rendering
}

// Health component represents an entity's health points
type Health struct {
	CurrentHP int
	MaxHP     int
}

func NewHealth(maxHP int) Health {
	return Health{
		CurrentHP: maxHP,
		MaxHP:     maxHP,
	}
}

func (h *Health) IsDead() bool {
	return h.CurrentHP <= 0
}
