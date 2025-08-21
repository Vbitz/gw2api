package gw2api

import "time"

// PvPAmulet represents a PvP amulet
type PvPAmulet struct {
	ID         int              `json:"id"`
	Name       string           `json:"name"`
	Icon       string           `json:"icon"`
	Attributes map[string]int   `json:"attributes"`
}

// PvPGame represents a PvP game
type PvPGame struct {
	ID         string           `json:"id"`
	MapID      int              `json:"map_id"`
	Started    time.Time        `json:"started"`
	Ended      time.Time        `json:"ended"`
	Result     string           `json:"result"`
	Team       string           `json:"team"`
	Profession string           `json:"profession"`
	Scores     PvPGameScores    `json:"scores"`
	RatingType string           `json:"rating_type"`
	RatingChange int            `json:"rating_change"`
	Season     string           `json:"season,omitempty"`
}

// PvPGameScores represents game scores
type PvPGameScores struct {
	Red  int `json:"red"`
	Blue int `json:"blue"`
}

// PvPHero represents a PvP hero
type PvPHero struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        string        `json:"type"`
	Stats       PvPHeroStats  `json:"stats"`
	Overlay     string        `json:"overlay"`
	Underlay    string        `json:"underlay"`
	Skins       []PvPHeroSkin `json:"skins"`
}

// PvPHeroStats represents hero stats
type PvPHeroStats struct {
	Offense int `json:"offense"`
	Defense int `json:"defense"`
	Speed   int `json:"speed"`
}

// PvPHeroSkin represents a hero skin
type PvPHeroSkin struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Icon    string `json:"icon"`
	Default bool   `json:"default"`
	Unlock  string `json:"unlock"`
}

// PvPRank represents a PvP rank
type PvPRank struct {
	ID          int           `json:"id"`
	FinisherID  int           `json:"finisher_id"`
	Name        string        `json:"name"`
	Icon        string        `json:"icon"`
	MinRank     int           `json:"min_rank"`
	MaxRank     int           `json:"max_rank"`
	Levels      []PvPRankLevel `json:"levels"`
}

// PvPRankLevel represents a rank level
type PvPRankLevel struct {
	MinRank int `json:"min_rank"`
	MaxRank int `json:"max_rank"`
	Points  int `json:"points"`
}

// PvPRewardTrack represents a PvP reward track
type PvPRewardTrack struct {
	ID          int                   `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Icon        string                `json:"icon"`
	Type        string                `json:"type"`
	Rewards     []PvPRewardTrackReward `json:"rewards"`
	Periods     []PvPRewardTrackPeriod `json:"periods,omitempty"`
}

// PvPRewardTrackReward represents a reward in a track
type PvPRewardTrackReward struct {
	Type     string `json:"type"`
	ID       int    `json:"id,omitempty"`
	Count    int    `json:"count,omitempty"`
	Region   string `json:"region,omitempty"`
}

// PvPRewardTrackPeriod represents a period in a track
type PvPRewardTrackPeriod struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// PvPRune represents a PvP rune
type PvPRune struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Runes  []PvPRuneBonus `json:"runes"`
}

// PvPRuneBonus represents a rune bonus
type PvPRuneBonus struct {
	Count       int    `json:"count"`
	Description string `json:"description"`
}

// PvPSeason represents a PvP season
type PvPSeason struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Start        time.Time          `json:"start"`
	End          time.Time          `json:"end"`
	Active       bool               `json:"active"`
	Divisions    []PvPSeasonDivision `json:"divisions"`
	Leaderboards PvPSeasonLeaderboards `json:"leaderboards"`
	Ranks        []PvPSeasonRank    `json:"ranks"`
}

// PvPSeasonDivision represents a season division
type PvPSeasonDivision struct {
	Name         string              `json:"name"`
	Flags        []string            `json:"flags"`
	LargeIcon    string              `json:"large_icon"`
	SmallIcon    string              `json:"small_icon"`
	PipIcon      string              `json:"pip_icon"`
	Tiers        []PvPSeasonTier     `json:"tiers"`
}

// PvPSeasonTier represents a season tier
type PvPSeasonTier struct {
	Rating int `json:"rating"`
}

// PvPSeasonLeaderboards represents season leaderboards
type PvPSeasonLeaderboards struct {
	Ladder PvPLeaderboard `json:"ladder"`
	Guild  PvPLeaderboard `json:"guild"`
}

// PvPLeaderboard represents a leaderboard
type PvPLeaderboard struct {
	Settings PvPLeaderboardSettings `json:"settings"`
	Scoring  []PvPLeaderboardScoring `json:"scoring"`
}

// PvPLeaderboardSettings represents leaderboard settings
type PvPLeaderboardSettings struct {
	Name         string `json:"name"`
	Duration     int    `json:"duration"`
	Scoring      string `json:"scoring"`
	Tiers        []PvPLeaderboardTier `json:"tiers"`
}

// PvPLeaderboardScoring represents leaderboard scoring
type PvPLeaderboardScoring struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Ordering    string `json:"ordering"`
}

// PvPLeaderboardTier represents a leaderboard tier
type PvPLeaderboardTier struct {
	Color string `json:"color"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Range []int  `json:"range"`
}

// PvPSeasonRank represents a season rank
type PvPSeasonRank struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Overlay     string `json:"overlay"`
	Underlay    string `json:"underlay"`
	Tier        int    `json:"tier"`
	Division    int    `json:"division"`
}

// PvPSigil represents a PvP sigil
type PvPSigil struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
}

// PvPStandings represents PvP standings
type PvPStandings struct {
	Current PvPStandingsCurrent `json:"current"`
	Best    PvPStandingsBest    `json:"best"`
	SeasonID string             `json:"season_id"`
}

// PvPStandingsCurrent represents current standings
type PvPStandingsCurrent struct {
	TotalPoints   int    `json:"total_points"`
	Division      int    `json:"division"`
	Tier          int    `json:"tier"`
	Points        int    `json:"points"`
	Repeats       int    `json:"repeats"`
	Rating        int    `json:"rating"`
	Decay         int    `json:"decay"`
	StripPoints   []int  `json:"strip_points,omitempty"`
}

// PvPStandingsBest represents best standings
type PvPStandingsBest struct {
	TotalPoints int `json:"total_points"`
	Division    int `json:"division"`
	Tier        int `json:"tier"`
	Points      int `json:"points"`
	Repeats     int `json:"repeats"`
}

// PvPStats represents PvP statistics
type PvPStats struct {
	PvPRank          int                    `json:"pvp_rank"`
	PvPRankPoints    int                    `json:"pvp_rank_points"`
	PvPRankRollovers int                    `json:"pvp_rank_rollovers"`
	Aggregate        PvPStatsAggregate      `json:"aggregate"`
	Professions      map[string]PvPStatsProfession `json:"professions"`
	Ladders          map[string]PvPStatsLadder     `json:"ladders"`
}

// PvPStatsAggregate represents aggregate PvP stats
type PvPStatsAggregate struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Desertions int `json:"desertions"`
	Byes   int `json:"byes"`
	Forfeits int `json:"forfeits"`
}

// PvPStatsProfession represents profession-specific PvP stats
type PvPStatsProfession struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Desertions int `json:"desertions"`
	Byes   int `json:"byes"`
	Forfeits int `json:"forfeits"`
}

// PvPStatsLadder represents ladder-specific PvP stats
type PvPStatsLadder struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Desertions int `json:"desertions"`
	Byes   int `json:"byes"`
	Forfeits int `json:"forfeits"`
}

// PvPSeasonLeaderboardEntries represents the actual leaderboard data
type PvPSeasonLeaderboardEntries struct {
	Legendary []PvPLeaderboardEntry `json:"legendary"`
	Guild     []PvPLeaderboardEntry `json:"guild"`
}

// PvPLeaderboardEntry represents a single leaderboard entry
type PvPLeaderboardEntry struct {
	Name        string                `json:"name,omitempty"`
	ID          string                `json:"id,omitempty"`
	Team        string                `json:"team,omitempty"`
	TeamID      int                   `json:"team_id,omitempty"`
	Rank        int                   `json:"rank"`
	Date        string                `json:"date"`
	Scores      []PvPLeaderboardScore `json:"scores"`
}

// PvPLeaderboardScore represents a score in a leaderboard entry
type PvPLeaderboardScore struct {
	ID    string `json:"id"`
	Value int    `json:"value"`
}