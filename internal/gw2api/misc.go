package gw2api

// Adventure represents an adventure
type Adventure struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AdventureLeaderboard represents adventure leaderboard
type AdventureLeaderboard struct {
	// Placeholder for adventure leaderboard data
}

// BackstoryAnswer represents a backstory answer
type BackstoryAnswer struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Races       []string `json:"races,omitempty"`
	Professions []string `json:"professions,omitempty"`
}

// BackstoryQuestion represents a backstory question
type BackstoryQuestion struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Description string `json:"description"`
	Answers []string `json:"answers"`
	Order   int    `json:"order"`
	Races   []string `json:"races,omitempty"`
	Professions []string `json:"professions,omitempty"`
}

// CreateSubtoken represents create subtoken functionality
type CreateSubtoken struct {
	// Placeholder for subtoken creation
}

// DailyCrafting represents daily crafting items
type DailyCraftingItem struct {
	Item string `json:"item"`
}

// Emblem represents emblem information
type Emblem struct {
	// Placeholder for emblem data
}

// Emote represents an emote
type EmoteDetail struct {
	ID           string   `json:"id"`
	Commands     []string `json:"commands"`
	UnlockItems  []int    `json:"unlock_items,omitempty"`
}

// Event represents a world event
type Event struct {
	ID    string            `json:"id"`
	Name  string            `json:"name"`
	Level int               `json:"level"`
	MapID int               `json:"map_id"`
	Flags []string          `json:"flags"`
	Location EventLocation  `json:"location"`
}

// EventLocation represents event location
type EventLocation struct {
	Type   string      `json:"type"`
	Center []float64   `json:"center,omitempty"`
	Radius float64     `json:"radius,omitempty"`
	Points [][]float64 `json:"points,omitempty"`
	ZRange []float64   `json:"z_range,omitempty"`
}

// EventState represents world event state
type EventState struct {
	WorldID int    `json:"world_id"`
	MapID   int    `json:"map_id"`
	EventID string `json:"event_id"`
	State   string `json:"state"`
}

// FileDetail represents a file asset detail
type FileDetail struct {
	ID   string `json:"id"`
	Icon string `json:"icon"`
}

// GemstoreCatalog represents gemstore catalog
type GemstoreCatalog struct {
	// Placeholder for gemstore catalog data
}

// HomeInfo represents home instance information
type HomeInfo struct {
	// Placeholder for home instance data
}

// HomesteadInfo represents homestead information
type HomesteadInfo struct {
	// Placeholder for homestead data
}

// HomesteadDecorationCategory represents decoration categories
type HomesteadDecorationCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// HomesteadDecorationDetail represents decoration details
type HomesteadDecorationDetail struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Categories  []int    `json:"categories"`
	Icon        string   `json:"icon"`
	Vendor      string   `json:"vendor,omitempty"`
	CostItems   []int    `json:"cost_items,omitempty"`
}

// HomesteadGlyphDetail represents glyph details
type HomesteadGlyphDetail struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// LegendaryArmoryDetail represents legendary armory details
type LegendaryArmoryDetail struct {
	// Placeholder for legendary armory data
}

// Legend represents a revenant legend
type Legend struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Heal     int    `json:"heal"`
	Elite    int    `json:"elite"`
	Utilities []int `json:"utilities"`
}

// Logo represents a logo
type LogoDetail struct {
	// Placeholder for logo data
}

// MapDetail represents detailed map information
type MapDetail struct {
	ID            int              `json:"id"`
	Name          string           `json:"name"`
	MinLevel      int              `json:"min_level"`
	MaxLevel      int              `json:"max_level"`
	DefaultFloor  int              `json:"default_floor"`
	Type          string           `json:"type"`
	Floors        []int            `json:"floors"`
	RegionID      int              `json:"region_id"`
	RegionName    string           `json:"region_name"`
	ContinentID   int              `json:"continent_id"`
	ContinentName string           `json:"continent_name"`
	MapRect       [][]int          `json:"map_rect"`
	ContinentRect [][]int          `json:"continent_rect"`
}

// MountInfo represents mount information
type MountInfo struct {
	// Placeholder for mount data
}

// MountSkinDetail represents mount skin details
type MountSkinDetail struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Icon   string   `json:"icon"`
	Mount  string   `json:"mount"`
	DyeSlots []MountDyeSlot `json:"dye_slots"`
}

// MountDyeSlot represents a mount dye slot
type MountDyeSlot struct {
	ColorID  int    `json:"color_id"`
	Material string `json:"material"`
}

// MountTypeDetail represents mount type details
type MountTypeDetail struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	DefaultSkin int           `json:"default_skin"`
	Skins    []int            `json:"skins"`
	Skills   []MountTypeSkill `json:"skills"`
}

// MountTypeSkill represents a mount type skill
type MountTypeSkill struct {
	ID   int    `json:"id"`
	Slot string `json:"slot"`
}

// NoveltyDetail represents novelty details
type NoveltyDetail struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Slot        string `json:"slot"`
}

// OutfitDetail represents outfit details
type OutfitDetail struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	UnlockItems []int    `json:"unlock_items,omitempty"`
}

// Pet represents a ranger pet
type Pet struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Description string `json:"description"`
	Icon   string `json:"icon"`
	Skills []PetSkill `json:"skills"`
}

// PetSkill represents a pet skill
type PetSkill struct {
	ID   int    `json:"id"`
	Slot string `json:"slot"`
}

// Quaggan represents a quaggan
type Quaggan struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// RecipeDetail represents recipe details
type RecipeDetail struct {
	ID            int              `json:"id"`
	Type          string           `json:"type"`
	OutputItemID  int              `json:"output_item_id"`
	OutputItemCount int            `json:"output_item_count"`
	TimeToCraftMS int              `json:"time_to_craft_ms"`
	Disciplines   []string         `json:"disciplines"`
	MinRating     int              `json:"min_rating"`
	Flags         []string         `json:"flags"`
	Ingredients   []RecipeIngredient `json:"ingredients"`
	GuildIngredients []RecipeIngredient `json:"guild_ingredients,omitempty"`
	OutputUpgradeID int             `json:"output_upgrade_id,omitempty"`
}

// RecipeIngredient represents a recipe ingredient
type RecipeIngredient struct {
	ItemID int `json:"item_id"`
	Count  int `json:"count"`
}

// RecipeSearch represents recipe search functionality
type RecipeSearch struct {
	// Placeholder for recipe search
}

// SkiffDetail represents skiff details
type SkiffDetail struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// SkinDetail represents skin details
type SkinDetail struct {
	ID           int             `json:"id"`
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Flags        []string        `json:"flags"`
	Restrictions []string        `json:"restrictions"`
	Icon         string          `json:"icon"`
	Rarity       string          `json:"rarity"`
	Race         []string        `json:"race,omitempty"`
	Description  string          `json:"description,omitempty"`
	Details      SkinDetailInfo  `json:"details,omitempty"`
}

// SkinDetailInfo represents detailed skin information
type SkinDetailInfo struct {
	Type        string `json:"type,omitempty"`
	WeightClass string `json:"weight_class,omitempty"`
	DyeSlots    []SkinDyeSlot `json:"dye_slots,omitempty"`
}

// SkinDyeSlot represents a skin dye slot
type SkinDyeSlot struct {
	ColorID  int    `json:"color_id"`
	Material string `json:"material"`
}

// TokenInfo represents API token information
type TokenInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// Vendor represents a vendor
type Vendor struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Sells    []VendorItem `json:"sells"`
	Buys     []VendorItem `json:"buys"`
}

// VendorItem represents an item sold by a vendor
type VendorItem struct {
	ItemID int         `json:"item_id"`
	Cost   []VendorCost `json:"cost"`
}

// VendorCost represents the cost of a vendor item
type VendorCost struct {
	ItemID   int `json:"item_id"`
	Quantity int `json:"quantity"`
}

// WizardsVaultListing represents wizard's vault listing
type WizardsVaultListingDetail struct {
	// Placeholder for wizard's vault listing data
}

// WizardsVaultObjective represents wizard's vault objective
type WizardsVaultObjective struct {
	// Placeholder for wizard's vault objective data
}

// WorldBossDetail represents world boss details
type WorldBossDetail struct {
	// Placeholder for world boss data
}