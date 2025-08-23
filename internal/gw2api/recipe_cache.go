package gw2api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// RecipeCache provides in-memory caching of recipes loaded from a local JSON file
type RecipeCache struct {
	recipes         map[int]*RecipeDetail // ID -> Recipe mapping for fast lookups
	recipesByOutput map[int][]int         // OutputItemID -> []RecipeID mapping for recipe search
	recipesByInput  map[int][]int         // IngredientItemID -> []RecipeID mapping for recipe search
	recipesList     []*RecipeDetail       // All recipes as slice for iteration
	loaded          bool
	mutex           sync.RWMutex
	stats           RecipeCacheStats
}

// RecipeCacheStats tracks cache performance
type RecipeCacheStats struct {
	LoadedRecipes int
	LoadTime      time.Duration
	CacheHits     int64
	CacheMisses   int64
	LastLoadTime  time.Time
}

// NewRecipeCache creates a new recipe cache
func NewRecipeCache() *RecipeCache {
	return &RecipeCache{
		recipes:         make(map[int]*RecipeDetail),
		recipesByOutput: make(map[int][]int),
		recipesByInput:  make(map[int][]int),
		recipesList:     make([]*RecipeDetail, 0),
		loaded:          false,
	}
}

// LoadFromFile loads all recipes from a JSONL file (one JSON object per line)
func (rc *RecipeCache) LoadFromFile(filePath string) error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	startTime := time.Now()
	
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open recipes file %s: %w", filePath, err)
	}
	defer file.Close()

	// Clear existing data
	rc.recipes = make(map[int]*RecipeDetail)
	rc.recipesByOutput = make(map[int][]int)
	rc.recipesByInput = make(map[int][]int)
	rc.recipesList = make([]*RecipeDetail, 0)

	scanner := bufio.NewScanner(file)
	recipeCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		var recipe RecipeDetail
		if err := json.Unmarshal([]byte(line), &recipe); err != nil {
			// Skip invalid lines but continue processing
			continue
		}

		// Store in map and slice
		rc.recipes[recipe.ID] = &recipe
		rc.recipesList = append(rc.recipesList, &recipe)
		
		// Build output item mapping for recipe search
		if recipe.OutputItemID > 0 {
			rc.recipesByOutput[recipe.OutputItemID] = append(rc.recipesByOutput[recipe.OutputItemID], recipe.ID)
		}
		
		// Build input item mapping for recipe search
		for _, ingredient := range recipe.Ingredients {
			if ingredient.ItemID > 0 {
				rc.recipesByInput[ingredient.ItemID] = append(rc.recipesByInput[ingredient.ItemID], recipe.ID)
			}
		}
		
		recipeCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading recipes file: %w", err)
	}

	rc.loaded = true
	rc.stats.LoadedRecipes = recipeCount
	rc.stats.LoadTime = time.Since(startTime)
	rc.stats.LastLoadTime = time.Now()

	return nil
}

// GetByID retrieves a recipe by its ID from the cache
func (rc *RecipeCache) GetByID(id int) (*RecipeDetail, bool) {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	if !rc.loaded {
		rc.stats.CacheMisses++
		return nil, false
	}

	recipe, found := rc.recipes[id]
	if found {
		rc.stats.CacheHits++
	} else {
		rc.stats.CacheMisses++
	}
	
	return recipe, found
}

// GetByIDs retrieves multiple recipes by their IDs
func (rc *RecipeCache) GetByIDs(ids []int) []*RecipeDetail {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	if !rc.loaded {
		rc.stats.CacheMisses += int64(len(ids))
		return make([]*RecipeDetail, 0)
	}

	results := make([]*RecipeDetail, 0, len(ids))
	for _, id := range ids {
		if recipe, found := rc.recipes[id]; found {
			results = append(results, recipe)
			rc.stats.CacheHits++
		} else {
			rc.stats.CacheMisses++
		}
	}

	return results
}

// SearchByOutput finds recipes that create a specific item
func (rc *RecipeCache) SearchByOutput(itemID int) []int {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	if !rc.loaded {
		rc.stats.CacheMisses++
		return []int{}
	}

	if recipeIDs, found := rc.recipesByOutput[itemID]; found {
		rc.stats.CacheHits++
		// Return a copy to avoid race conditions
		result := make([]int, len(recipeIDs))
		copy(result, recipeIDs)
		return result
	}

	rc.stats.CacheMisses++
	return []int{}
}

// SearchByInput finds recipes that use a specific item as ingredient
func (rc *RecipeCache) SearchByInput(itemID int) []int {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	if !rc.loaded {
		rc.stats.CacheMisses++
		return []int{}
	}

	if recipeIDs, found := rc.recipesByInput[itemID]; found {
		rc.stats.CacheHits++
		// Return a copy to avoid race conditions
		result := make([]int, len(recipeIDs))
		copy(result, recipeIDs)
		return result
	}

	rc.stats.CacheMisses++
	return []int{}
}

// IsLoaded returns true if the cache has been loaded with data
func (rc *RecipeCache) IsLoaded() bool {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	return rc.loaded
}

// Size returns the number of recipes in the cache
func (rc *RecipeCache) Size() int {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	return len(rc.recipes)
}

// Clear clears the cache
func (rc *RecipeCache) Clear() {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	
	rc.recipes = make(map[int]*RecipeDetail)
	rc.recipesByOutput = make(map[int][]int)
	rc.recipesByInput = make(map[int][]int)
	rc.recipesList = make([]*RecipeDetail, 0)
	rc.loaded = false
	rc.stats = RecipeCacheStats{}
}

// GetStats returns cache performance statistics
func (rc *RecipeCache) GetStats() RecipeCacheStats {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	return rc.stats
}

// GetAll returns all cached recipes (use with caution for large datasets)
func (rc *RecipeCache) GetAll() []*RecipeDetail {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	
	if !rc.loaded {
		return []*RecipeDetail{}
	}
	
	// Return a copy to avoid race conditions
	result := make([]*RecipeDetail, len(rc.recipesList))
	copy(result, rc.recipesList)
	return result
}