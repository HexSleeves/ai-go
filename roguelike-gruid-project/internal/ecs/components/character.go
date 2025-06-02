package components

// Stats represents character attributes
type Stats struct {
	Strength     int // Affects melee damage and carrying capacity
	Dexterity    int // Affects accuracy and dodge chance
	Constitution int // Affects health and stamina
	Intelligence int // Affects mana and spell effectiveness
	Wisdom       int // Affects perception and mana regeneration
	Charisma     int // Affects social interactions and leadership
}

// NewStats creates default stats
func NewStats() Stats {
	return Stats{
		Strength:     10,
		Dexterity:    10,
		Constitution: 10,
		Intelligence: 10,
		Wisdom:       10,
		Charisma:     10,
	}
}

// Experience component for character progression
type Experience struct {
	Level           int
	CurrentXP       int
	XPToNextLevel   int
	TotalXP         int
	SkillPoints     int
	AttributePoints int
}

// NewExperience creates a new experience component
func NewExperience() Experience {
	return Experience{
		Level:         1,
		CurrentXP:     0,
		XPToNextLevel: 100,
		TotalXP:       0,
		SkillPoints:   0,
	}
}

// AddXP adds experience points and handles level ups
func (exp *Experience) AddXP(amount int) bool {
	exp.CurrentXP += amount
	exp.TotalXP += amount
	
	if exp.CurrentXP >= exp.XPToNextLevel {
		return exp.levelUp()
	}
	return false
}

// levelUp handles leveling up
func (exp *Experience) levelUp() bool {
	if exp.CurrentXP >= exp.XPToNextLevel {
		exp.CurrentXP -= exp.XPToNextLevel
		exp.Level++
		exp.XPToNextLevel = calculateXPForLevel(exp.Level + 1)
		exp.SkillPoints += 2
		exp.AttributePoints += 1
		return true
	}
	return false
}

// calculateXPForLevel calculates XP required for a given level
func calculateXPForLevel(level int) int {
	// Simple exponential formula: 100 * level^1.5
	return int(100 * float64(level) * float64(level) * 0.5)
}

// Skills represents character skills
type Skills struct {
	// Combat Skills
	MeleeWeapons int
	RangedWeapons int
	Defense      int
	
	// Magic Skills
	Evocation    int // Destructive magic
	Conjuration  int // Summoning magic
	Enchantment  int // Buff/debuff magic
	Divination   int // Information magic
	
	// Utility Skills
	Stealth      int
	Lockpicking  int
	Perception   int
	Medicine     int
	Crafting     int
}

// NewSkills creates default skills
func NewSkills() Skills {
	return Skills{
		MeleeWeapons:  1,
		RangedWeapons: 1,
		Defense:       1,
		Evocation:     0,
		Conjuration:   0,
		Enchantment:   0,
		Divination:    0,
		Stealth:       1,
		Lockpicking:   0,
		Perception:    1,
		Medicine:      0,
		Crafting:      0,
	}
}

// Combat represents combat-related stats
type Combat struct {
	AttackPower    int
	Defense        int
	Accuracy       int
	DodgeChance    int
	CriticalChance int
	CriticalDamage int
}

// NewCombat creates default combat stats
func NewCombat() Combat {
	return Combat{
		AttackPower:    1,
		Defense:        0,
		Accuracy:       75,
		DodgeChance:    5,
		CriticalChance: 5,
		CriticalDamage: 150, // 150% damage on crit
	}
}

// Mana represents magical energy
type Mana struct {
	CurrentMP int
	MaxMP     int
	RegenRate int // MP regenerated per turn
}

// NewMana creates a new mana component
func NewMana(maxMP int) Mana {
	return Mana{
		CurrentMP: maxMP,
		MaxMP:     maxMP,
		RegenRate: 1,
	}
}

// RegenerateMana restores mana over time
func (m *Mana) RegenerateMana() {
	m.CurrentMP += m.RegenRate
	if m.CurrentMP > m.MaxMP {
		m.CurrentMP = m.MaxMP
	}
}

// UseMana consumes mana for spells
func (m *Mana) UseMana(amount int) bool {
	if m.CurrentMP >= amount {
		m.CurrentMP -= amount
		return true
	}
	return false
}

// Stamina represents physical energy
type Stamina struct {
	CurrentSP int
	MaxSP     int
	RegenRate int // SP regenerated per turn
}

// NewStamina creates a new stamina component
func NewStamina(maxSP int) Stamina {
	return Stamina{
		CurrentSP: maxSP,
		MaxSP:     maxSP,
		RegenRate: 2,
	}
}

// RegenerateStamina restores stamina over time
func (s *Stamina) RegenerateStamina() {
	s.CurrentSP += s.RegenRate
	if s.CurrentSP > s.MaxSP {
		s.CurrentSP = s.MaxSP
	}
}

// UseStamina consumes stamina for actions
func (s *Stamina) UseStamina(amount int) bool {
	if s.CurrentSP >= amount {
		s.CurrentSP -= amount
		return true
	}
	return false
}

// StatusEffect represents temporary effects on entities
type StatusEffect struct {
	Name        string
	Duration    int    // Turns remaining
	Type        string // "buff", "debuff", "neutral"
	Description string
	
	// Stat modifiers
	StrengthMod     int
	DexterityMod    int
	ConstitutionMod int
	IntelligenceMod int
	WisdomMod       int
	CharismaMod     int
	
	// Combat modifiers
	AttackMod    int
	DefenseMod   int
	AccuracyMod  int
	DodgeMod     int
	
	// Special effects
	Poisoned    bool
	Regenerating bool
	Paralyzed   bool
	Confused    bool
}

// StatusEffects component holds all active status effects
type StatusEffects struct {
	Effects []StatusEffect
}

// NewStatusEffects creates a new status effects component
func NewStatusEffects() StatusEffects {
	return StatusEffects{
		Effects: make([]StatusEffect, 0),
	}
}

// AddEffect adds a new status effect
func (se *StatusEffects) AddEffect(effect StatusEffect) {
	// Check if effect already exists and refresh duration
	for i := range se.Effects {
		if se.Effects[i].Name == effect.Name {
			se.Effects[i].Duration = effect.Duration
			return
		}
	}
	
	// Add new effect
	se.Effects = append(se.Effects, effect)
}

// RemoveEffect removes a status effect by name
func (se *StatusEffects) RemoveEffect(name string) {
	for i := range se.Effects {
		if se.Effects[i].Name == name {
			se.Effects = append(se.Effects[:i], se.Effects[i+1:]...)
			return
		}
	}
}

// UpdateEffects decreases duration and removes expired effects
func (se *StatusEffects) UpdateEffects() []StatusEffect {
	var expired []StatusEffect
	var active []StatusEffect
	
	for _, effect := range se.Effects {
		effect.Duration--
		if effect.Duration <= 0 {
			expired = append(expired, effect)
		} else {
			active = append(active, effect)
		}
	}
	
	se.Effects = active
	return expired
}

// HasEffect checks if a specific effect is active
func (se *StatusEffects) HasEffect(name string) bool {
	for _, effect := range se.Effects {
		if effect.Name == name {
			return true
		}
	}
	return false
}

// GetTotalModifiers calculates total stat modifiers from all effects
func (se *StatusEffects) GetTotalModifiers() (Stats, Combat) {
	var statMods Stats
	var combatMods Combat
	
	for _, effect := range se.Effects {
		statMods.Strength += effect.StrengthMod
		statMods.Dexterity += effect.DexterityMod
		statMods.Constitution += effect.ConstitutionMod
		statMods.Intelligence += effect.IntelligenceMod
		statMods.Wisdom += effect.WisdomMod
		statMods.Charisma += effect.CharismaMod
		
		combatMods.AttackPower += effect.AttackMod
		combatMods.Defense += effect.DefenseMod
		combatMods.Accuracy += effect.AccuracyMod
		combatMods.DodgeChance += effect.DodgeMod
	}
	
	return statMods, combatMods
}
