package main

// Achievement represents a Guild Wars 2 achievement
type Achievement struct {
	ID           int                `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Requirement  string             `json:"requirement"`
	LockedText   string             `json:"locked_text"`
	Type         string             `json:"type"`
	Flags        []string           `json:"flags"`
	Tiers        []AchievementTier  `json:"tiers"`
	Prerequisites []int             `json:"prerequisites,omitempty"`
	Rewards      []AchievementReward `json:"rewards,omitempty"`
	Bits         []AchievementBit   `json:"bits,omitempty"`
	PointCap     int                `json:"point_cap,omitempty"`
}

// AchievementTier represents a tier within an achievement
type AchievementTier struct {
	Count  int `json:"count"`
	Points int `json:"points"`
}

// AchievementReward represents a reward for completing an achievement
type AchievementReward struct {
	Type   string `json:"type"`
	ID     int    `json:"id,omitempty"`
	Count  int    `json:"count,omitempty"`
	Region string `json:"region,omitempty"`
}

// AchievementBit represents a bit/objective within an achievement
type AchievementBit struct {
	Type string `json:"type"`
	ID   int    `json:"id,omitempty"`
	Text string `json:"text,omitempty"`
}

// AchievementCategory represents an achievement category
type AchievementCategory struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Order        int    `json:"order"`
	Icon         string `json:"icon"`
	Achievements []int  `json:"achievements"`
}

// AchievementGroup represents an achievement group
type AchievementGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	Categories  []int  `json:"categories"`
}