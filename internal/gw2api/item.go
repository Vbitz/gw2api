package gw2api

// Item represents a Guild Wars 2 item
type Item struct {
	ID           int                `json:"id"`
	ChatLink     string             `json:"chat_link"`
	Name         string             `json:"name"`
	Icon         string             `json:"icon,omitempty"`
	Description  string             `json:"description,omitempty"`
	Type         string             `json:"type"`
	Rarity       string             `json:"rarity"`
	Level        int                `json:"level"`
	VendorValue  int                `json:"vendor_value"`
	DefaultSkin  int                `json:"default_skin,omitempty"`
	Flags        []string           `json:"flags"`
	GameTypes    []string           `json:"game_types"`
	Restrictions []string           `json:"restrictions"`
	UpgradesInto []ItemUpgrade      `json:"upgrades_into,omitempty"`
	UpgradesFrom []ItemUpgrade      `json:"upgrades_from,omitempty"`
	Details      *ItemDetails       `json:"details,omitempty"`
}

// ItemUpgrade represents upgrade information
type ItemUpgrade struct {
	Upgrade string `json:"upgrade"`
	ItemID  int    `json:"item_id"`
}

// ItemDetails contains type-specific details for items
type ItemDetails struct {
	// Common fields
	Type string `json:"type,omitempty"`

	// Armor fields
	WeightClass         string         `json:"weight_class,omitempty"`
	Defense             int            `json:"defense,omitempty"`
	InfusionSlots       []InfusionSlot `json:"infusion_slots,omitempty"`
	AttributeAdjustment int            `json:"attribute_adjustment,omitempty"`
	InfixUpgrade        *InfixUpgrade  `json:"infix_upgrade,omitempty"`
	SuffixItemID        int            `json:"suffix_item_id,omitempty"`
	SecondarySuffixItemID string       `json:"secondary_suffix_item_id,omitempty"`
	StatChoices         []int          `json:"stat_choices,omitempty"`

	// Weapon fields
	DamageType string `json:"damage_type,omitempty"`
	MinPower   int    `json:"min_power,omitempty"`
	MaxPower   int    `json:"max_power,omitempty"`

	// Bag fields
	Size         int  `json:"size,omitempty"`
	NoSellOrSort bool `json:"no_sell_or_sort,omitempty"`

	// Consumable fields
	Description        string `json:"description,omitempty"`
	DurationMs         int    `json:"duration_ms,omitempty"`
	UnlockType         string `json:"unlock_type,omitempty"`
	ColorID            int    `json:"color_id,omitempty"`
	RecipeID           int    `json:"recipe_id,omitempty"`
	ExtraRecipeIDs     []int  `json:"extra_recipe_ids,omitempty"`
	GuildUpgradeID     int    `json:"guild_upgrade_id,omitempty"`
	ApplyCount         int    `json:"apply_count,omitempty"`
	Name               string `json:"name,omitempty"`
	Icon               string `json:"icon,omitempty"`
	Skins              []int  `json:"skins,omitempty"`

	// Container fields (no additional fields)

	// Gathering fields (no additional fields)

	// Gizmo fields
	VendorIDs  []int  `json:"vendor_ids,omitempty"`

	// Miniature fields
	MinipetID int `json:"minipet_id,omitempty"`

	// Tool (Salvage kit) fields
	Charges  int    `json:"charges,omitempty"`

	// Trinket fields (uses same InfusionSlots, InfixUpgrade, etc.)

	// UpgradeComponent fields
	UpgradeFlags         []string `json:"flags,omitempty"`
	InfusionUpgradeFlags []string `json:"infusion_upgrade_flags,omitempty"`
	Suffix               string   `json:"suffix,omitempty"`
	Bonuses              []string `json:"bonuses,omitempty"`
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
