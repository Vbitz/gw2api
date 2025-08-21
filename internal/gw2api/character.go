package gw2api

import "time"

// Character represents a character name
type Character string

// CharacterBackstory represents character backstory choices
type CharacterBackstory struct {
	Backstory []string `json:"backstory"`
}

// CharacterBuildTab represents a character build tab
type CharacterBuildTab struct {
	Tab              int                   `json:"tab"`
	IsActive         bool                  `json:"is_active"`
	Build            CharacterBuild        `json:"build"`
}

// CharacterBuild represents a character build
type CharacterBuild struct {
	Name            string                    `json:"name"`
	Profession      string                    `json:"profession"`
	Specializations []BuildSpecialization     `json:"specializations"`
	Skills          BuildSkills               `json:"skills"`
	AquaticSkills   BuildSkills               `json:"aquatic_skills"`
	Legends         []string                  `json:"legends,omitempty"`
	AquaticLegends  []string                  `json:"aquatic_legends,omitempty"`
}

// CharacterBuildTabActive represents the active build tab
type CharacterBuildTabActive struct {
	Tab int `json:"tab"`
}

// CharacterCore represents core character information
type CharacterCore struct {
	Name         string    `json:"name"`
	Race         string    `json:"race"`
	Gender       string    `json:"gender"`
	Profession   string    `json:"profession"`
	Level        int       `json:"level"`
	Guild        string    `json:"guild,omitempty"`
	Age          int       `json:"age"`
	LastModified time.Time `json:"last_modified,omitempty"`
	Created      time.Time `json:"created"`
	Deaths       int       `json:"deaths"`
	Title        int       `json:"title,omitempty"`
}

// CharacterCrafting represents character crafting disciplines
type CharacterCrafting struct {
	Discipline string `json:"discipline"`
	Rating     int    `json:"rating"`
	Active     bool   `json:"active"`
}

// CharacterWvWAbility represents WvW abilities
type CharacterWvWAbility struct {
	ID   int `json:"id"`
	Rank int `json:"rank"`
}

// CharacterDungeon represents completed dungeon paths
type CharacterDungeon struct {
	ID    string `json:"id"`
	Paths []CharacterDungeonPath `json:"paths"`
}

// CharacterDungeonPath represents a completed dungeon path
type CharacterDungeonPath struct {
	ID   string `json:"id"`
	Mode string `json:"mode"`
}

// CharacterEquipment represents equipped items
type CharacterEquipment struct {
	ID        int                      `json:"id"`
	Slot      string                   `json:"slot,omitempty"`
	Infusions []int                    `json:"infusions,omitempty"`
	Upgrades  []int                    `json:"upgrades,omitempty"`
	Skin      int                      `json:"skin,omitempty"`
	Stats     *CharacterEquipmentStats `json:"stats,omitempty"`
	Binding   string                   `json:"binding,omitempty"`
	Location  string                   `json:"location,omitempty"`
	Tabs      []int                    `json:"tabs,omitempty"`
	Charges   int                      `json:"charges,omitempty"`
	BoundTo   string                   `json:"bound_to,omitempty"`
	Dyes      []*int                   `json:"dyes,omitempty"`
}

// CharacterEquipmentStats represents equipment stats
type CharacterEquipmentStats struct {
	ID         int                              `json:"id"`
	Attributes CharacterEquipmentStatsAttributes `json:"attributes"`
}

// CharacterEquipmentStatsAttributes represents the stat attributes
type CharacterEquipmentStatsAttributes struct {
	Power              int `json:"Power,omitempty"`
	Precision          int `json:"Precision,omitempty"`
	Toughness          int `json:"Toughness,omitempty"`
	Vitality           int `json:"Vitality,omitempty"`
	ConditionDamage    int `json:"ConditionDamage,omitempty"`
	ConditionDuration  int `json:"ConditionDuration,omitempty"`
	Healing            int `json:"Healing,omitempty"`
	BoonDuration       int `json:"BoonDuration,omitempty"`
}

// CharacterEquipmentTab represents an equipment tab
type CharacterEquipmentTab struct {
	Tab        int                   `json:"tab"`
	Name       string                `json:"name"`
	IsActive   bool                  `json:"is_active"`
	Equipment  []CharacterEquipment  `json:"equipment"`
	EquipmentPvp CharacterEquipmentPvp `json:"equipment_pvp,omitempty"`
}

// CharacterEquipmentPvp represents PvP equipment
type CharacterEquipmentPvp struct {
	Amulet    int   `json:"amulet,omitempty"`
	Rune      int   `json:"rune,omitempty"`
	Sigils    []int `json:"sigils,omitempty"`
}

// CharacterEquipmentTabActive represents the active equipment tab
type CharacterEquipmentTabActive struct {
	Tab int `json:"tab"`
}

// CharacterHeroPoint represents unlocked hero points
type CharacterHeroPoint string

// CharacterInventory represents character inventory
type CharacterInventory struct {
	Bags []CharacterBag `json:"bags"`
}

// CharacterBag represents an inventory bag
type CharacterBag struct {
	ID        int                    `json:"id"`
	Size      int                    `json:"size"`
	Inventory []CharacterInventorySlot `json:"inventory"`
}

// CharacterInventorySlot represents an inventory slot
type CharacterInventorySlot struct {
	ID        int       `json:"id"`
	Count     int       `json:"count"`
	Binding   string    `json:"binding,omitempty"`
	BoundTo   string    `json:"bound_to,omitempty"`
	Stats     *ItemStat `json:"stats,omitempty"`
}

// CharacterQuest represents character quests
type CharacterQuest struct {
	ID    int    `json:"id"`
	Step  int    `json:"step"`
	Goals []CharacterQuestGoal `json:"goals"`
}

// CharacterQuestGoal represents a quest goal
type CharacterQuestGoal struct {
	Active   string `json:"active"`
	Complete string `json:"complete"`
}

// CharacterRecipe represents unlocked recipes
type CharacterRecipe struct {
	Recipes []int `json:"recipes"`
}

// CharacterSAB represents Super Adventure Box progress
type CharacterSAB struct {
	Zones []CharacterSABZone `json:"zones"`
	Unlocks []CharacterSABUnlock `json:"unlocks"`
	Songs []CharacterSABSong `json:"songs"`
}

// CharacterSABZone represents SAB zone progress
type CharacterSABZone struct {
	ID    int                  `json:"id"`
	Mode  string               `json:"mode"`
	World int                  `json:"world"`
	Zone  int                  `json:"zone"`
}

// CharacterSABUnlock represents SAB unlocks
type CharacterSABUnlock struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CharacterSABSong represents SAB songs
type CharacterSABSong struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CharacterSkills represents character skills
type CharacterSkills struct {
	Heal     int      `json:"heal"`
	Utility  []int    `json:"utility"`
	Elite    int      `json:"elite"`
	Legends  []string `json:"legends,omitempty"`
}

// CharacterSpecialization represents character specializations
type CharacterSpecialization struct {
	ID     int   `json:"id"`
	Traits []int `json:"traits"`
}

// CharacterTraining represents character training
type CharacterTraining struct {
	ID    int  `json:"id"`
	Spent int  `json:"spent"`
	Done  bool `json:"done"`
}