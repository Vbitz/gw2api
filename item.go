package main

// Item represents a Guild Wars 2 item
type Item struct {
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Type         string       `json:"type"`
	Level        int          `json:"level"`
	Rarity       string       `json:"rarity"`
	VendorValue  int          `json:"vendor_value"`
	DefaultSkin  int          `json:"default_skin,omitempty"`
	GameTypes    []string     `json:"game_types"`
	Flags        []string     `json:"flags"`
	Restrictions []string     `json:"restrictions"`
	Icon         string       `json:"icon"`
	ChatLink     string       `json:"chat_link"`
	Details      *ItemDetails `json:"details,omitempty"`
}

// ItemDetails contains type-specific details for items
type ItemDetails struct {
	// Common fields
	Type                string  `json:"type,omitempty"`
	WeightClass         string  `json:"weight_class,omitempty"`
	AttributeAdjustment float64 `json:"attribute_adjustment,omitempty"`

	// Armor/Weapon fields
	Defense               int            `json:"defense,omitempty"`
	InfusionSlots         []InfusionSlot `json:"infusion_slots,omitempty"`
	InfixUpgrade          *InfixUpgrade  `json:"infix_upgrade,omitempty"`
	SuffixItemID          string         `json:"suffix_item_id,omitempty"`
	SecondarySuffixItemID string         `json:"secondary_suffix_item_id,omitempty"`
	StatChoices           []int          `json:"stat_choices,omitempty"`

	// Weapon specific
	MinPower   int    `json:"min_power,omitempty"`
	MaxPower   int    `json:"max_power,omitempty"`
	DamageType string `json:"damage_type,omitempty"`

	// Consumable fields
	Duration       int    `json:"duration_ms,omitempty"`
	UnlockType     string `json:"unlock_type,omitempty"`
	ColorID        int    `json:"color_id,omitempty"`
	RecipeID       int    `json:"recipe_id,omitempty"`
	ExtraRecipeIDs []int  `json:"extra_recipe_ids,omitempty"`
	GuildUpgradeID int    `json:"guild_upgrade_id,omitempty"`
	ApplyCount     int    `json:"apply_count,omitempty"`
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`

	// Container fields
	NoSellOrSort bool `json:"no_sell_or_sort,omitempty"`

	// Bag fields
	Size int `json:"size,omitempty"`

	// Tool fields
	Charges int `json:"charges,omitempty"`

	// Trinket fields
	// Uses same InfusionSlots, InfixUpgrade, SuffixItemID, etc.

	// UpgradeComponent fields
	Flags   []string `json:"flags,omitempty"`
	Bonuses []string `json:"bonuses,omitempty"`

	// Salvage fields
	KitType string `json:"kit_type,omitempty"`
}

// InfusionSlot represents an infusion slot on an item
type InfusionSlot struct {
	Flags  []string `json:"flags"`
	ItemID int      `json:"item_id,omitempty"`
}

// InfixUpgrade represents the inherent stats on an item
type InfixUpgrade struct {
	ID         int         `json:"id"`
	Attributes []Attribute `json:"attributes"`
	Buff       *Buff       `json:"buff,omitempty"`
}

// Attribute represents a stat attribute
type Attribute struct {
	Attribute string `json:"attribute"`
	Modifier  int    `json:"modifier"`
}

// Buff represents a buff provided by an item
type Buff struct {
	SkillID     int    `json:"skill_id"`
	Description string `json:"description,omitempty"`
}

// ItemStat represents item stat combinations
type ItemStat struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	Attributes []Attribute `json:"attributes"`
}
