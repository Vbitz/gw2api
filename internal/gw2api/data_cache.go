package gw2api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// DataCache provides comprehensive caching for all GW2 API data types
type DataCache struct {
	items        *ItemCache
	skills       *SkillCache
	achievements *AchievementCache
	mutex        sync.RWMutex
	stats        DataCacheStats
}

// DataCacheStats tracks overall cache performance
type DataCacheStats struct {
	LoadTime       time.Duration
	LastLoadTime   time.Time
	TotalCacheHits int64
	ItemsLoaded    int
	SkillsLoaded   int
	AchievementsLoaded int
}

// NewDataCache creates a new comprehensive data cache
func NewDataCache() *DataCache {
	return &DataCache{
		items:        NewItemCache(),
		skills:       NewSkillCache(),
		achievements: NewAchievementCache(),
	}
}

// LoadFromDirectory loads all data files from the specified directory
func (dc *DataCache) LoadFromDirectory(dataDir string) error {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	startTime := time.Now()
	var errors []string

	// Load items
	itemsPath := fmt.Sprintf("%s/items.json", dataDir)
	if _, err := os.Stat(itemsPath); err == nil {
		if err := dc.items.LoadFromFile(itemsPath); err != nil {
			errors = append(errors, fmt.Sprintf("items: %v", err))
		} else {
			dc.stats.ItemsLoaded = dc.items.Size()
		}
	}

	// Load skills
	skillsPath := fmt.Sprintf("%s/skills.json", dataDir)
	if _, err := os.Stat(skillsPath); err == nil {
		if err := dc.skills.LoadFromFile(skillsPath); err != nil {
			errors = append(errors, fmt.Sprintf("skills: %v", err))
		} else {
			dc.stats.SkillsLoaded = dc.skills.Size()
		}
	}

	// Load achievements
	achievementsPath := fmt.Sprintf("%s/achievements.json", dataDir)
	if _, err := os.Stat(achievementsPath); err == nil {
		if err := dc.achievements.LoadFromFile(achievementsPath); err != nil {
			errors = append(errors, fmt.Sprintf("achievements: %v", err))
		} else {
			dc.stats.AchievementsLoaded = dc.achievements.Size()
		}
	}

	dc.stats.LoadTime = time.Since(startTime)
	dc.stats.LastLoadTime = time.Now()

	if len(errors) > 0 {
		return fmt.Errorf("cache loading errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

// GetItemCache returns the item cache
func (dc *DataCache) GetItemCache() *ItemCache {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	return dc.items
}

// GetSkillCache returns the skill cache
func (dc *DataCache) GetSkillCache() *SkillCache {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	return dc.skills
}

// GetAchievementCache returns the achievement cache
func (dc *DataCache) GetAchievementCache() *AchievementCache {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	return dc.achievements
}

// Stats returns overall cache statistics
func (dc *DataCache) Stats() DataCacheStats {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	
	// Aggregate cache hits from all sub-caches
	dc.stats.TotalCacheHits = dc.items.stats.CacheHits + 
	                         dc.skills.stats.CacheHits + 
	                         dc.achievements.stats.CacheHits
	
	return dc.stats
}

// Clear clears all caches
func (dc *DataCache) Clear() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	
	dc.items.Clear()
	dc.skills.Clear()
	dc.achievements.Clear()
	dc.stats = DataCacheStats{}
}

// SkillCache provides in-memory caching of skills
type SkillCache struct {
	skills    map[int]*Skill // ID -> Skill mapping
	skillsList []*Skill      // All skills as slice
	loaded    bool
	mutex     sync.RWMutex
	stats     SkillCacheStats
}

// SkillCacheStats tracks skill cache performance
type SkillCacheStats struct {
	LoadedSkills int
	LoadTime     time.Duration
	CacheHits    int64
	CacheMisses  int64
	LastLoadTime time.Time
}

// NewSkillCache creates a new skill cache
func NewSkillCache() *SkillCache {
	return &SkillCache{
		skills:     make(map[int]*Skill),
		skillsList: make([]*Skill, 0),
		loaded:     false,
	}
}

// LoadFromFile loads all skills from a JSONL file
func (sc *SkillCache) LoadFromFile(filePath string) error {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	startTime := time.Now()
	
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open skills file %s: %w", filePath, err)
	}
	defer file.Close()

	// Clear existing data
	sc.skills = make(map[int]*Skill)
	sc.skillsList = make([]*Skill, 0)

	scanner := bufio.NewScanner(file)
	skillCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		var skill Skill
		if err := json.Unmarshal([]byte(line), &skill); err != nil {
			// Skip invalid lines but continue processing
			continue
		}

		sc.skills[skill.ID] = &skill
		sc.skillsList = append(sc.skillsList, &skill)
		skillCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading skills file: %w", err)
	}

	sc.loaded = true
	sc.stats.LoadedSkills = skillCount
	sc.stats.LoadTime = time.Since(startTime)
	sc.stats.LastLoadTime = time.Now()

	return nil
}

// GetByID retrieves a skill by its ID
func (sc *SkillCache) GetByID(id int) (*Skill, bool) {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	if !sc.loaded {
		sc.stats.CacheMisses++
		return nil, false
	}

	skill, found := sc.skills[id]
	if found {
		sc.stats.CacheHits++
	} else {
		sc.stats.CacheMisses++
	}
	
	return skill, found
}

// GetByIDs retrieves multiple skills by their IDs
func (sc *SkillCache) GetByIDs(ids []int) []*Skill {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	if !sc.loaded {
		sc.stats.CacheMisses += int64(len(ids))
		return nil
	}

	results := make([]*Skill, 0, len(ids))
	for _, id := range ids {
		if skill, found := sc.skills[id]; found {
			results = append(results, skill)
			sc.stats.CacheHits++
		} else {
			sc.stats.CacheMisses++
		}
	}

	return results
}

// SearchSkills performs in-memory search on cached skills
func (sc *SkillCache) SearchSkills(query string, profession string, skillType string, limit int) []*Skill {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	if !sc.loaded {
		return nil
	}

	var results []*Skill
	count := 0

	if limit == 0 {
		limit = 50
	}

	for _, skill := range sc.skillsList {
		match := true

		// Check name match
		if query != "" && !strings.Contains(strings.ToLower(skill.Name), strings.ToLower(query)) {
			match = false
		}

		// Check profession filter
		if profession != "" && match {
			professionMatch := false
			for _, prof := range skill.Professions {
				if strings.EqualFold(prof, profession) {
					professionMatch = true
					break
				}
			}
			if !professionMatch {
				match = false
			}
		}

		// Check skill type filter
		if skillType != "" && match && !strings.EqualFold(skill.Type, skillType) {
			match = false
		}

		if match {
			results = append(results, skill)
			count++
			if count >= limit {
				break
			}
		}
	}

	sc.stats.CacheHits++
	return results
}

// IsLoaded returns whether the cache has been loaded
func (sc *SkillCache) IsLoaded() bool {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	return sc.loaded
}

// Size returns the number of skills in the cache
func (sc *SkillCache) Size() int {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	return len(sc.skillsList)
}

// Clear clears the cache
func (sc *SkillCache) Clear() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	sc.skills = make(map[int]*Skill)
	sc.skillsList = make([]*Skill, 0)
	sc.loaded = false
	sc.stats = SkillCacheStats{}
}

// AchievementCache provides in-memory caching of achievements
type AchievementCache struct {
	achievements     map[int]*Achievement // ID -> Achievement mapping
	achievementsList []*Achievement       // All achievements as slice
	loaded           bool
	mutex            sync.RWMutex
	stats            AchievementCacheStats
}

// AchievementCacheStats tracks achievement cache performance
type AchievementCacheStats struct {
	LoadedAchievements int
	LoadTime           time.Duration
	CacheHits          int64
	CacheMisses        int64
	LastLoadTime       time.Time
}

// NewAchievementCache creates a new achievement cache
func NewAchievementCache() *AchievementCache {
	return &AchievementCache{
		achievements:     make(map[int]*Achievement),
		achievementsList: make([]*Achievement, 0),
		loaded:           false,
	}
}

// LoadFromFile loads all achievements from a JSONL file
func (ac *AchievementCache) LoadFromFile(filePath string) error {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	startTime := time.Now()
	
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open achievements file %s: %w", filePath, err)
	}
	defer file.Close()

	// Clear existing data
	ac.achievements = make(map[int]*Achievement)
	ac.achievementsList = make([]*Achievement, 0)

	scanner := bufio.NewScanner(file)
	achievementCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		var achievement Achievement
		if err := json.Unmarshal([]byte(line), &achievement); err != nil {
			// Skip invalid lines but continue processing
			continue
		}

		ac.achievements[achievement.ID] = &achievement
		ac.achievementsList = append(ac.achievementsList, &achievement)
		achievementCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading achievements file: %w", err)
	}

	ac.loaded = true
	ac.stats.LoadedAchievements = achievementCount
	ac.stats.LoadTime = time.Since(startTime)
	ac.stats.LastLoadTime = time.Now()

	return nil
}

// GetByID retrieves an achievement by its ID
func (ac *AchievementCache) GetByID(id int) (*Achievement, bool) {
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()

	if !ac.loaded {
		ac.stats.CacheMisses++
		return nil, false
	}

	achievement, found := ac.achievements[id]
	if found {
		ac.stats.CacheHits++
	} else {
		ac.stats.CacheMisses++
	}
	
	return achievement, found
}

// GetByIDs retrieves multiple achievements by their IDs
func (ac *AchievementCache) GetByIDs(ids []int) []*Achievement {
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()

	if !ac.loaded {
		ac.stats.CacheMisses += int64(len(ids))
		return nil
	}

	results := make([]*Achievement, 0, len(ids))
	for _, id := range ids {
		if achievement, found := ac.achievements[id]; found {
			results = append(results, achievement)
			ac.stats.CacheHits++
		} else {
			ac.stats.CacheMisses++
		}
	}

	return results
}

// SearchAchievements performs in-memory search on cached achievements
func (ac *AchievementCache) SearchAchievements(query string, category string, limit int) []*Achievement {
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()

	if !ac.loaded {
		return nil
	}

	var results []*Achievement
	count := 0

	if limit == 0 {
		limit = 50
	}

	for _, achievement := range ac.achievementsList {
		match := true

		// Check name match
		if query != "" && !strings.Contains(strings.ToLower(achievement.Name), strings.ToLower(query)) {
			match = false
		}

		// Check category filter (if Achievement has Category field)
		// Note: Achievement structure may vary, adapt as needed

		if match {
			results = append(results, achievement)
			count++
			if count >= limit {
				break
			}
		}
	}

	ac.stats.CacheHits++
	return results
}

// IsLoaded returns whether the cache has been loaded
func (ac *AchievementCache) IsLoaded() bool {
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()
	return ac.loaded
}

// Size returns the number of achievements in the cache
func (ac *AchievementCache) Size() int {
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()
	return len(ac.achievementsList)
}

// Clear clears the cache
func (ac *AchievementCache) Clear() {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	ac.achievements = make(map[int]*Achievement)
	ac.achievementsList = make([]*Achievement, 0)
	ac.loaded = false
	ac.stats = AchievementCacheStats{}
}