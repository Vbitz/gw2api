package main

// Build represents the current game build
type Build struct {
	ID int `json:"id"`
}

// Language represents supported API languages
type Language string

const (
	LanguageEnglish Language = "en"
	LanguageSpanish Language = "es"
	LanguageGerman  Language = "de"
	LanguageFrench  Language = "fr"
	LanguageChinese Language = "zh"
)

// PaginationResponse contains pagination metadata
type PaginationResponse struct {
	Page      int `json:"page"`
	PageSize  int `json:"page_size"`
	PageTotal int `json:"page_total"`
	Total     int `json:"total"`
}

// APIError represents an error response from the API
type APIError struct {
	Text string `json:"text"`
}

func (e APIError) Error() string {
	return e.Text
}

// Profession represents a Guild Wars 2 profession
type Profession struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Icon            string            `json:"icon"`
	IconBig         string            `json:"icon_big"`
	Specializations []int             `json:"specializations"`
	Weapons         map[string]Weapon `json:"weapons"`
	Flags           []string          `json:"flags"`
	Skills          []ProfessionSkill `json:"skills"`
	Training        []TrainingTrack   `json:"training"`
}

// Weapon represents weapon information for a profession
type Weapon struct {
	Specialization int           `json:"specialization,omitempty"`
	Flags          []string      `json:"flags"`
	Skills         []WeaponSkill `json:"skills"`
}

// WeaponSkill represents a skill available with a weapon
type WeaponSkill struct {
	ID         int    `json:"id"`
	Slot       string `json:"slot"`
	Offhand    string `json:"offhand,omitempty"`
	Attunement string `json:"attunement,omitempty"`
	Source     string `json:"source,omitempty"`
}

// ProfessionSkill represents skills available to a profession
type ProfessionSkill struct {
	Heal       []int `json:"heal"`
	Utility    []int `json:"utility"`
	Elite      []int `json:"elite"`
	Weapon     []int `json:"weapon"`
	Profession []int `json:"profession"`
}

// TrainingTrack represents a training track for a profession
type TrainingTrack struct {
	ID       int                `json:"id"`
	Category string             `json:"category"`
	Name     string             `json:"name"`
	Track    []TrainingCategory `json:"track"`
}

// TrainingCategory represents a category within a training track
type TrainingCategory struct {
	Cost    int    `json:"cost"`
	Type    string `json:"type"`
	SkillID int    `json:"skill_id,omitempty"`
	TraitID int    `json:"trait_id,omitempty"`
}

// Race represents a Guild Wars 2 race
type Race struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Skills []int  `json:"skills"`
}

// Title represents a character title
type Title struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Achievement  int    `json:"achievement,omitempty"`
	Achievements []int  `json:"achievements,omitempty"`
	APRequired   int    `json:"ap_required,omitempty"`
}

// File represents a game asset file
type File struct {
	ID   string `json:"id"`
	Icon string `json:"icon"`
}

// Color represents a dye color
type Color struct {
	ID         int           `json:"id"`
	Name       string        `json:"name"`
	BaseRGB    []int         `json:"base_rgb"`
	Cloth      ColorMaterial `json:"cloth"`
	Leather    ColorMaterial `json:"leather"`
	Metal      ColorMaterial `json:"metal"`
	Fur        ColorMaterial `json:"fur,omitempty"`
	Item       int           `json:"item,omitempty"`
	Categories []string      `json:"categories"`
}

// ColorMaterial represents color information for a specific material
type ColorMaterial struct {
	Brightness int     `json:"brightness"`
	Contrast   float64 `json:"contrast"`
	Hue        int     `json:"hue"`
	Saturation float64 `json:"saturation"`
	Lightness  float64 `json:"lightness"`
	RGB        []int   `json:"rgb"`
}

// Material represents a crafting material category
type Material struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Items []int  `json:"items"`
	Order int    `json:"order"`
}

// Dungeon represents a dungeon
type Dungeon struct {
	ID    string        `json:"id"`
	Paths []DungeonPath `json:"paths"`
}

// DungeonPath represents a path within a dungeon
type DungeonPath struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Raid represents a raid
type Raid struct {
	ID    string     `json:"id"`
	Wings []RaidWing `json:"wings"`
}

// RaidWing represents a wing within a raid
type RaidWing struct {
	ID     string          `json:"id"`
	Events []RaidEncounter `json:"events"`
}

// RaidEncounter represents an encounter within a raid wing
type RaidEncounter struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Mastery represents a mastery track
type Mastery struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Requirement string         `json:"requirement"`
	Order       int            `json:"order"`
	Background  string         `json:"background"`
	Region      string         `json:"region"`
	Levels      []MasteryLevel `json:"levels"`
}

// MasteryLevel represents a level within a mastery track
type MasteryLevel struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Instruction string `json:"instruction"`
	Icon        string `json:"icon"`
	PointCost   int    `json:"point_cost"`
	ExpCost     int    `json:"exp_cost"`
}

// Story represents a story chapter
type Story struct {
	ID          int            `json:"id"`
	Season      string         `json:"season"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Timeline    string         `json:"timeline"`
	Level       int            `json:"level"`
	Order       int            `json:"order"`
	Chapters    []StoryChapter `json:"chapters"`
	Races       []string       `json:"races,omitempty"`
	Flags       []string       `json:"flags,omitempty"`
}

// StoryChapter represents a chapter within a story
type StoryChapter struct {
	Name string `json:"name"`
}

// StorySeason represents a story season
type StorySeason struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Order   int    `json:"order"`
	Stories []int  `json:"stories"`
}

// Quest represents a quest/heart task
type Quest struct {
	ID    int         `json:"id"`
	Name  string      `json:"name"`
	Level int         `json:"level"`
	Story int         `json:"story"`
	Goals []QuestGoal `json:"goals"`
}

// QuestGoal represents an objective within a quest
type QuestGoal struct {
	Active   string `json:"active"`
	Complete string `json:"complete"`
}
