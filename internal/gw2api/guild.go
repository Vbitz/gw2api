package gw2api

import "time"

// Guild represents basic guild information
type Guild struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Tag    string `json:"tag"`
	Level  int    `json:"level"`
	Motd   string `json:"motd"`
	Influence int `json:"influence"`
	Aetherium int `json:"aetherium"`
	Favor    int    `json:"favor"`
	Resonance int  `json:"resonance"`
	Emblem   GuildEmblem `json:"emblem"`
}

// GuildEmblem represents a guild emblem
type GuildEmblem struct {
	Background   GuildEmblemLayer `json:"background"`
	Foreground   GuildEmblemLayer `json:"foreground"`
	Flags        []string         `json:"flags"`
}

// GuildEmblemLayer represents an emblem layer
type GuildEmblemLayer struct {
	ID     int   `json:"id"`
	Colors []int `json:"colors"`
}

// GuildLog represents a guild log entry
type GuildLog struct {
	ID        int       `json:"id"`
	Time      time.Time `json:"time"`
	User      string    `json:"user,omitempty"`
	Type      string    `json:"type"`
	Invited   string    `json:"invited,omitempty"`
	Kicked    string    `json:"kicked,omitempty"`
	ChangedTo string    `json:"changed_to,omitempty"`
	Operation string    `json:"operation,omitempty"`
	Activity  string    `json:"activity,omitempty"`
	TotalParticipants int `json:"total_participants,omitempty"`
	Participants []string `json:"participants,omitempty"`
}

// GuildMember represents a guild member
type GuildMember struct {
	Name   string    `json:"name"`
	Rank   string    `json:"rank"`
	Joined time.Time `json:"joined"`
}

// GuildRank represents a guild rank
type GuildRank struct {
	ID          string   `json:"id"`
	Order       int      `json:"order"`
	Permissions []string `json:"permissions"`
	Icon        string   `json:"icon"`
}

// GuildStash represents guild stash information
type GuildStash struct {
	UpgradeID int                `json:"upgrade_id"`
	Size      int                `json:"size"`
	Coins     int                `json:"coins"`
	Note      string             `json:"note"`
	Inventory []GuildStashSlot   `json:"inventory"`
}

// GuildStashSlot represents a slot in the guild stash
type GuildStashSlot struct {
	ID       int       `json:"id"`
	Count    int       `json:"count"`
	Binding  string    `json:"binding,omitempty"`
	BoundTo  string    `json:"bound_to,omitempty"`
	Stats    *ItemStat `json:"stats,omitempty"`
}

// GuildStorage represents guild storage
type GuildStorage struct {
	Item  int `json:"item"`
	Count int `json:"count"`
}

// GuildTeam represents a guild team
type GuildTeam struct {
	ID       int             `json:"id"`
	Members  []GuildTeamMember `json:"members"`
	Name     string          `json:"name"`
	Aggregate GuildTeamStats  `json:"aggregate"`
	Ladders  map[string]GuildTeamStats `json:"ladders"`
	Games    []GuildTeamGame `json:"games"`
	Seasons  []GuildTeamSeason `json:"seasons"`
}

// GuildTeamMember represents a team member
type GuildTeamMember struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

// GuildTeamStats represents team statistics
type GuildTeamStats struct {
	Wins       int `json:"wins"`
	Losses     int `json:"losses"`
	Desertions int `json:"desertions"`
	Byes       int `json:"byes"`
	Forfeits   int `json:"forfeits"`
}

// GuildTeamGame represents a team game
type GuildTeamGame struct {
	ID         string    `json:"id"`
	MapID      int       `json:"map_id"`
	Started    time.Time `json:"started"`
	Ended      time.Time `json:"ended"`
	Result     string    `json:"result"`
	Team       string    `json:"team"`
	RatingType string    `json:"rating_type"`
	RatingChange int     `json:"rating_change"`
	Season     string    `json:"season,omitempty"`
	Scores     PvPGameScores `json:"scores"`
}

// GuildTeamSeason represents team season data
type GuildTeamSeason struct {
	ID      string `json:"id"`
	Wins    int    `json:"wins"`
	Losses  int    `json:"losses"`
	Rating  int    `json:"rating"`
	Ranking int    `json:"ranking,omitempty"`
}

// GuildTreasury represents guild treasury items
type GuildTreasury struct {
	ItemID      int                      `json:"item_id"`
	Count       int                      `json:"count"`
	NeededBy    []GuildTreasuryNeededBy  `json:"needed_by"`
}

// GuildTreasuryNeededBy represents what needs treasury items
type GuildTreasuryNeededBy struct {
	UpgradeID int `json:"upgrade_id"`
	Count     int `json:"count"`
}

// GuildUpgrade represents guild upgrades
type GuildUpgrade struct {
	ID int `json:"id"`
}

// GuildPermission represents a guild permission
type GuildPermission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GuildSearch represents guild search functionality
type GuildSearch struct {
	// Placeholder for guild search functionality
}

// GuildUpgradeDetail represents detailed guild upgrade information
type GuildUpgradeDetail struct {
	ID           int                    `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	BuildTime    int                    `json:"build_time"`
	Icon         string                 `json:"icon"`
	Type         string                 `json:"type"`
	RequiredLevel int                   `json:"required_level"`
	Experience   int                    `json:"experience"`
	Prerequisites []int                 `json:"prerequisites"`
	Costs        []GuildUpgradeCost     `json:"costs"`
}

// GuildUpgradeCost represents the cost of a guild upgrade
type GuildUpgradeCost struct {
	Type     string `json:"type"`
	Name     string `json:"name,omitempty"`
	Count    int    `json:"count"`
	ItemID   int    `json:"item_id,omitempty"`
}