package gw2api

import "time"

// Account represents basic account information
type Account struct {
	ID                string    `json:"id"`
	Age               int       `json:"age"`
	Name              string    `json:"name"`
	World             int       `json:"world"`
	Guilds            []string  `json:"guilds"`
	GuildLeader       []string  `json:"guild_leader"`
	Created           time.Time `json:"created"`
	Access            []string  `json:"access"`
	Commander         bool      `json:"commander"`
	FractalLevel      int       `json:"fractal_level"`
	DailyAP           int       `json:"daily_ap"`
	MonthlyAP         int       `json:"monthly_ap"`
	WvwRank           int       `json:"wvw_rank"`
	LastModified      time.Time `json:"last_modified"`
	BuildStorageSlots int       `json:"build_storage_slots"`
}

// AccountAchievement represents an account achievement
type AccountAchievement struct {
	ID      int           `json:"id"`
	Bits    []int         `json:"bits,omitempty"`
	Current []int         `json:"current,omitempty"`
	Max     []int         `json:"max,omitempty"`
	Done    bool          `json:"done"`
	Unlocked bool         `json:"unlocked,omitempty"`
	Repeated int          `json:"repeated,omitempty"`
}

// BankSlot represents an item in the account bank
type BankSlot struct {
	ID     int              `json:"id"`
	Count  int              `json:"count"`
	Skin   int              `json:"skin,omitempty"`
	Dyes   []int            `json:"dyes,omitempty"`
	Upgrades []int          `json:"upgrades,omitempty"`
	Infusions []int         `json:"infusions,omitempty"`
	Binding string          `json:"binding,omitempty"`
	BoundTo string          `json:"bound_to,omitempty"`
	Stats   *ItemStat       `json:"stats,omitempty"`
}

// BuildStorage represents a build template
type BuildStorage struct {
	Name           string                    `json:"name"`
	Profession     string                    `json:"profession"`
	Specializations []BuildSpecialization    `json:"specializations"`
	Skills         BuildSkills              `json:"skills"`
	AquaticSkills  BuildSkills              `json:"aquatic_skills"`
	Legends        []string                 `json:"legends,omitempty"`
	AquaticLegends []string                 `json:"aquatic_legends,omitempty"`
}

// BuildSpecialization represents a specialization in a build
type BuildSpecialization struct {
	ID     int   `json:"id"`
	Traits []int `json:"traits"`
}

// BuildSkills represents skills in a build
type BuildSkills struct {
	Heal     int   `json:"heal"`
	Utility  []int `json:"utility"`
	Elite    int   `json:"elite"`
	Legends  []string `json:"legends,omitempty"`
}

// DailyCrafting represents daily crafting progress
type DailyCrafting struct {
	Item string `json:"item"`
}

// Dye represents an unlocked dye
type Dye int

// Emote represents an unlocked emote
type Emote string

// Finisher represents an unlocked finisher
type Finisher struct {
	ID       int    `json:"id"`
	Permanent bool  `json:"permanent,omitempty"`
	Quantity int   `json:"quantity,omitempty"`
}

// Glider represents an unlocked glider
type Glider int

// HomeCat represents a home instance cat
type HomeCat struct {
	ID   int    `json:"id"`
	Hint string `json:"hint"`
}

// HomeNode represents a home instance gathering node
type HomeNode string

// Homestead represents homestead information
type Homestead struct {
	// Basic homestead data structure
}

// HomesteadDecoration represents a homestead decoration
type HomesteadDecoration struct {
	ID int `json:"id"`
}

// HomesteadGlyph represents a homestead glyph
type HomesteadGlyph struct {
	ID int `json:"id"`
}

// InventorySlot represents an item in shared inventory
type InventorySlot struct {
	ID       int       `json:"id"`
	Count    int       `json:"count"`
	Binding  string    `json:"binding,omitempty"`
	BoundTo  string    `json:"bound_to,omitempty"`
	Stats    *ItemStat `json:"stats,omitempty"`
}

// JadeBot represents an unlocked jade bot
type JadeBot int

// LegendaryArmory represents legendary armory information
type LegendaryArmory struct {
	ID    int `json:"id"`
	Count int `json:"count"`
}

// Luck represents account luck
type Luck struct {
	ID    string `json:"id"`
	Value int    `json:"value"`
}

// Mail represents a mail message
type Mail struct {
	ID         int       `json:"id"`
	Subject    string    `json:"subject"`
	From       string    `json:"from"`
	Timestamp  time.Time `json:"timestamp"`
	Content    string    `json:"content"`
	Attachment MailItem  `json:"attachment,omitempty"`
}

// MailItem represents an item attached to mail
type MailItem struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
	Count int   `json:"count"`
}

// MailCarrier represents an unlocked mail carrier
type MailCarrier int

// MapChest represents opened map chests
type MapChest string

// Mastery represents account mastery progress
type AccountMastery struct {
	ID    int `json:"id"`
	Level int `json:"level"`
}

// MasteryPoint represents mastery points
type MasteryPoint struct {
	Region   string `json:"region"`
	Spent    int    `json:"spent"`
	Earned   int    `json:"earned"`
}

// MaterialSlot represents materials storage
type MaterialSlot struct {
	ID       int `json:"id"`
	Category int `json:"category"`
	Binding  string `json:"binding,omitempty"`
	Count    int `json:"count"`
}

// Mini represents an unlocked miniature
type Mini int

// MountSkin represents an unlocked mount skin
type MountSkin int

// MountType represents an unlocked mount type
type MountType string

// Novelty represents an unlocked novelty
type Novelty int

// Outfit represents an unlocked outfit
type Outfit int

// Progression represents account progression
type Progression string

// AccountPvPHero represents an unlocked PvP hero
type AccountPvPHero int

// Recipe represents an unlocked recipe
type Recipe int

// Skiff represents an unlocked skiff
type Skiff int

// Skin represents an unlocked skin
type Skin int

// Title represents an unlocked title
type UnlockedTitle int

// WalletCurrency represents currency in the account wallet
type WalletCurrency struct {
	ID    int `json:"id"`
	Value int `json:"value"`
}

// WizardsVaultDaily represents daily wizard's vault objectives
type WizardsVaultDaily struct {
	// Placeholder structure
}

// WizardsVaultListing represents wizard's vault listings
type WizardsVaultListing struct {
	// Placeholder structure
}

// WizardsVaultSpecial represents special wizard's vault objectives
type WizardsVaultSpecial struct {
	// Placeholder structure
}

// WizardsVaultWeekly represents weekly wizard's vault objectives
type WizardsVaultWeekly struct {
	// Placeholder structure
}

// WorldBoss represents defeated world bosses
type WorldBoss string

// WvWInfo represents WvW account information
type WvWInfo struct {
	// Placeholder structure
}