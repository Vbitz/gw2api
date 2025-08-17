package gw2api

import "time"

// WvWAbility represents a WvW ability
type WvWAbility struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Ranks       []WvWAbilityRank `json:"ranks"`
}

// WvWAbilityRank represents a rank within a WvW ability
type WvWAbilityRank struct {
	Cost   int    `json:"cost"`
	Effect string `json:"effect"`
}

// WvWGuild represents WvW guild information
type WvWGuild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

// WvWGuildRegion represents WvW guilds by region
type WvWGuildRegion struct {
	// Placeholder for regional guild data
}

// WvWMatch represents a WvW match
type WvWMatch struct {
	ID            string           `json:"id"`
	StartTime     time.Time        `json:"start_time"`
	EndTime       time.Time        `json:"end_time"`
	Scores        WvWMatchScores   `json:"scores"`
	Worlds        WvWMatchWorlds   `json:"worlds"`
	AllWorlds     WvWMatchAllWorlds `json:"all_worlds"`
	Deaths        WvWMatchDeaths   `json:"deaths"`
	Kills         WvWMatchKills    `json:"kills"`
	VictoryPoints WvWMatchVictoryPoints `json:"victory_points"`
	Skirmishes    []WvWMatchSkirmish `json:"skirmishes"`
	Maps          []WvWMatchMap    `json:"maps"`
}

// WvWMatchScores represents match scores
type WvWMatchScores struct {
	Red   int `json:"red"`
	Blue  int `json:"blue"`
	Green int `json:"green"`
}

// WvWMatchWorlds represents match worlds
type WvWMatchWorlds struct {
	Red   []int `json:"red"`
	Blue  []int `json:"blue"`
	Green []int `json:"green"`
}

// WvWMatchAllWorlds represents all worlds in a match
type WvWMatchAllWorlds struct {
	Red   []int `json:"red"`
	Blue  []int `json:"blue"`
	Green []int `json:"green"`
}

// WvWMatchDeaths represents match deaths
type WvWMatchDeaths struct {
	Red   int `json:"red"`
	Blue  int `json:"blue"`
	Green int `json:"green"`
}

// WvWMatchKills represents match kills
type WvWMatchKills struct {
	Red   int `json:"red"`
	Blue  int `json:"blue"`
	Green int `json:"green"`
}

// WvWMatchVictoryPoints represents victory points
type WvWMatchVictoryPoints struct {
	Red   int `json:"red"`
	Blue  int `json:"blue"`
	Green int `json:"green"`
}

// WvWMatchSkirmish represents a match skirmish
type WvWMatchSkirmish struct {
	ID            int                   `json:"id"`
	Scores        WvWMatchScores        `json:"scores"`
	MapScores     []WvWMatchMapScore    `json:"map_scores"`
}

// WvWMatchMapScore represents a map score in a skirmish
type WvWMatchMapScore struct {
	Type   string         `json:"type"`
	Scores WvWMatchScores `json:"scores"`
}

// WvWMatchMap represents a WvW map in a match
type WvWMatchMap struct {
	ID            int                 `json:"id"`
	Type          string              `json:"type"`
	Scores        WvWMatchScores      `json:"scores"`
	Objectives    []WvWMatchObjective `json:"objectives"`
	Deaths        WvWMatchDeaths      `json:"deaths"`
	Kills         WvWMatchKills       `json:"kills"`
	Bonuses       []WvWMatchBonus     `json:"bonuses"`
}

// WvWMatchObjective represents a match objective
type WvWMatchObjective struct {
	ID                    string                `json:"id"`
	Type                  string                `json:"type"`
	Owner                 string                `json:"owner"`
	LastFlipped           time.Time             `json:"last_flipped"`
	ClaimedBy             string                `json:"claimed_by,omitempty"`
	ClaimedAt             *time.Time            `json:"claimed_at,omitempty"`
	PointsTick            int                   `json:"points_tick"`
	PointsCapture         int                   `json:"points_capture"`
	GuildUpgrades         []int                 `json:"guild_upgrades,omitempty"`
	YaksDelivered         int                   `json:"yaks_delivered,omitempty"`
}

// WvWMatchBonus represents a match bonus
type WvWMatchBonus struct {
	Type  string `json:"type"`
	Owner string `json:"owner"`
}

// WvWMatchOverview represents match overview information
type WvWMatchOverview struct {
	ID         string           `json:"id"`
	Worlds     WvWMatchWorlds   `json:"worlds"`
	AllWorlds  WvWMatchAllWorlds `json:"all_worlds"`
	StartTime  time.Time        `json:"start_time"`
	EndTime    time.Time        `json:"end_time"`
}

// WvWMatchStats represents match statistics
type WvWMatchStats struct {
	ID        string              `json:"id"`
	Deaths    WvWMatchDeaths      `json:"deaths"`
	Kills     WvWMatchKills       `json:"kills"`
	Maps      []WvWMatchStatsMap  `json:"maps"`
}

// WvWMatchStatsMap represents map statistics
type WvWMatchStatsMap struct {
	ID     int            `json:"id"`
	Type   string         `json:"type"`
	Deaths WvWMatchDeaths `json:"deaths"`
	Kills  WvWMatchKills  `json:"kills"`
}

// WvWMatchStatsGuild represents guild statistics for a match
type WvWMatchStatsGuild struct {
	// Placeholder for guild match stats
}

// WvWMatchStatsTop represents top players statistics
type WvWMatchStatsTop struct {
	// Placeholder for top player stats
}

// WvWObjective represents a WvW objective
type WvWObjective struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Type          string             `json:"type"`
	SectorID      int                `json:"sector_id"`
	MapID         int                `json:"map_id"`
	MapType       string             `json:"map_type"`
	Coord         []float64          `json:"coord"`
	LabelCoord    []float64          `json:"label_coord,omitempty"`
	Marker        string             `json:"marker"`
	ChatLink      string             `json:"chat_link"`
	Upgrade       []WvWObjectiveUpgrade `json:"upgrade,omitempty"`
}

// WvWObjectiveUpgrade represents an objective upgrade
type WvWObjectiveUpgrade struct {
	YaksRequired []WvWObjectiveYakRequirement `json:"yaks_required"`
	Upgrades     []WvWUpgrade                 `json:"upgrades"`
}

// WvWObjectiveYakRequirement represents yak delivery requirements
type WvWObjectiveYakRequirement struct {
	YaksRequired int `json:"yaks_required"`
	UpgradeID    int `json:"upgrade_id"`
}

// WvWRank represents a WvW rank
type WvWRank struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	MinRank  int    `json:"min_rank"`
}

// WvWRewardTrack represents a WvW reward track
type WvWRewardTrack struct {
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Icon        string                   `json:"icon"`
	Type        string                   `json:"type"`
	Rewards     []WvWRewardTrackReward   `json:"rewards"`
	Periods     []WvWRewardTrackPeriod   `json:"periods,omitempty"`
}

// WvWRewardTrackReward represents a reward in a WvW track
type WvWRewardTrackReward struct {
	Type   string `json:"type"`
	ID     int    `json:"id,omitempty"`
	Count  int    `json:"count,omitempty"`
	Region string `json:"region,omitempty"`
}

// WvWRewardTrackPeriod represents a period in a WvW track
type WvWRewardTrackPeriod struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// WvWTimer represents WvW timers
type WvWTimer struct {
	// Placeholder for WvW timer data
}

// WvWTimerLockout represents WvW lockout timers
type WvWTimerLockout struct {
	// Placeholder for WvW lockout data
}

// WvWTimerTeamAssignment represents team assignment timers
type WvWTimerTeamAssignment struct {
	// Placeholder for team assignment data
}

// WvWUpgrade represents a WvW upgrade
type WvWUpgrade struct {
	ID          int                `json:"id"`
	Tiers       []WvWUpgradeTier   `json:"tiers"`
}

// WvWUpgradeTier represents a tier within a WvW upgrade
type WvWUpgradeTier struct {
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	Icon         string                  `json:"icon"`
	YaksRequired int                     `json:"yaks_required"`
	Upgrades     []WvWUpgradeBonus       `json:"upgrades"`
}

// WvWUpgradeBonus represents a bonus from an upgrade
type WvWUpgradeBonus struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}