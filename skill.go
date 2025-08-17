package main

// Skill represents a Guild Wars 2 skill
type Skill struct {
	ID              int           `json:"id"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Icon            string        `json:"icon"`
	ChatLink        string        `json:"chat_link"`
	Type            string        `json:"type,omitempty"`
	WeaponType      string        `json:"weapon_type,omitempty"`
	Professions     []string      `json:"professions,omitempty"`
	Slot            string        `json:"slot,omitempty"`
	Facts           []SkillFact   `json:"facts,omitempty"`
	TraitedFacts    []TraitedFact `json:"traited_facts,omitempty"`
	Categories      []string      `json:"categories,omitempty"`
	Attunement      string        `json:"attunement,omitempty"`
	Cost            int           `json:"cost,omitempty"`
	DualWield       string        `json:"dual_wield,omitempty"`
	FlipSkill       int           `json:"flip_skill,omitempty"`
	Initiative      int           `json:"initiative,omitempty"`
	NextChain       int           `json:"next_chain,omitempty"`
	PrevChain       int           `json:"prev_chain,omitempty"`
	TransformSkills []int         `json:"transform_skills,omitempty"`
	BundleSkills    []int         `json:"bundle_skills,omitempty"`
	ToolbeltSkill   int           `json:"toolbelt_skill,omitempty"`
	Flags           []string      `json:"flags,omitempty"`
}

// SkillFact represents a fact about a skill
type SkillFact struct {
	Text          string  `json:"text,omitempty"`
	Type          string  `json:"type"`
	Icon          string  `json:"icon,omitempty"`
	Value         int     `json:"value,omitempty"`
	Target        string  `json:"target,omitempty"`
	Status        string  `json:"status,omitempty"`
	Description   string  `json:"description,omitempty"`
	ApplyCount    int     `json:"apply_count,omitempty"`
	Duration      int     `json:"duration,omitempty"`
	FieldType     string  `json:"field_type,omitempty"`
	FinisherType  string  `json:"finisher_type,omitempty"`
	Percent       float64 `json:"percent,omitempty"`
	HitCount      int     `json:"hit_count,omitempty"`
	DmgMultiplier float64 `json:"dmg_multiplier,omitempty"`
	Distance      int     `json:"distance,omitempty"`
	Prefix        string  `json:"prefix,omitempty"`
}

// TraitedFact represents a skill fact that is modified by traits
type TraitedFact struct {
	SkillFact
	RequiresTrait int `json:"requires_trait"`
	Overrides     int `json:"overrides,omitempty"`
}

// Specialization represents a trait line/specialization
type Specialization struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Profession  string `json:"profession"`
	Elite       bool   `json:"elite"`
	Icon        string `json:"icon"`
	Background  string `json:"background"`
	MinorTraits []int  `json:"minor_traits"`
	MajorTraits []int  `json:"major_traits"`
	WeaponTrait int    `json:"weapon_trait,omitempty"`
}

// Trait represents a trait
type Trait struct {
	ID             int           `json:"id"`
	Name           string        `json:"name"`
	Icon           string        `json:"icon"`
	Description    string        `json:"description"`
	Specialization int           `json:"specialization"`
	Tier           int           `json:"tier"`
	Order          int           `json:"order"`
	Slot           string        `json:"slot"`
	Facts          []SkillFact   `json:"facts,omitempty"`
	TraitedFacts   []TraitedFact `json:"traited_facts,omitempty"`
	Skills         []TraitSkill  `json:"skills,omitempty"`
}

// TraitSkill represents a skill granted by a trait
type TraitSkill struct {
	ID       int    `json:"id"`
	Category string `json:"category"`
}
