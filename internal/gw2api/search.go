package gw2api

import (
	"context"
	"fmt"
	"strings"
)

// ItemSearchOptions represents search options for items
type ItemSearchOptions struct {
	Name     string   // Partial name to search for
	Rarities []string // Filter by rarity (e.g., "Basic", "Fine", "Masterwork", "Rare", "Exotic", "Ascended", "Legendary")
	Types    []string // Filter by type (e.g., "Armor", "Weapon", "Trinket", "Consumable", etc.)
	MinLevel int      // Minimum level requirement
	MaxLevel int      // Maximum level requirement
	Limit    int      // Maximum number of results to return (0 = no limit)
}

// SearchItems searches for items based on the provided criteria
// This function uses cached data if available, otherwise falls back to API
func (c *Client) SearchItems(ctx context.Context, options ItemSearchOptions) ([]*Item, error) {
	// Try cache first if available
	if c.dataCache != nil && c.dataCache.GetItemCache().IsLoaded() {
		return c.dataCache.GetItemCache().SearchItems(options), nil
	}
	
	// Fallback to API-based search (limited for performance)
	// Get all item IDs first
	allIDs, err := c.GetItemIDs(ctx)
	if err != nil {
		return nil, err
	}

	var results []*Item
	var count int
	
	// Set reasonable limits for demo purposes when using API
	maxItemsToCheck := 1000  // Reduced from 2000 for better performance
	if options.Limit == 0 {
		options.Limit = 50   // Default limit if none specified
	}
	
	if len(allIDs) > maxItemsToCheck {
		allIDs = allIDs[:maxItemsToCheck]
	}

	// Process items in smaller batches
	batchSize := 100
	for i := 0; i < len(allIDs); i += batchSize {
		end := i + batchSize
		if end > len(allIDs) {
			end = len(allIDs)
		}

		batch := allIDs[i:end]
		items, err := c.GetItems(ctx, batch)
		if err != nil {
			// Continue with next batch on error to be resilient
			continue
		}

		for _, item := range items {
			if matchesSearchCriteria(item, options) {
				results = append(results, item)
				count++

				// Check if we've reached the limit
				if options.Limit > 0 && count >= options.Limit {
					return results, nil
				}
			}
		}
	}

	return results, nil
}

// matchesSearchCriteria checks if an item matches the search criteria
func matchesSearchCriteria(item *Item, options ItemSearchOptions) bool {
	// Check name match (case-insensitive partial match)
	if options.Name != "" {
		if !strings.Contains(strings.ToLower(item.Name), strings.ToLower(options.Name)) {
			return false
		}
	}

	// Check rarity filter
	if len(options.Rarities) > 0 {
		found := false
		for _, rarity := range options.Rarities {
			if strings.EqualFold(item.Rarity, rarity) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check type filter
	if len(options.Types) > 0 {
		found := false
		for _, itemType := range options.Types {
			if strings.EqualFold(item.Type, itemType) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check level range
	if options.MinLevel > 0 && item.Level < options.MinLevel {
		return false
	}
	if options.MaxLevel > 0 && item.Level > options.MaxLevel {
		return false
	}

	return true
}

// GetItemsByName finds items by partial name match
func (c *Client) GetItemsByName(ctx context.Context, name string, limit int) ([]*Item, error) {
	return c.SearchItems(ctx, ItemSearchOptions{
		Name:  name,
		Limit: limit,
	})
}

// GetItemsByRarity finds items by rarity
func (c *Client) GetItemsByRarity(ctx context.Context, rarity string, limit int) ([]*Item, error) {
	return c.SearchItems(ctx, ItemSearchOptions{
		Rarities: []string{rarity},
		Limit:    limit,
	})
}

// GetItemsByNameAndRarity finds items by both name and rarity
func (c *Client) GetItemsByNameAndRarity(ctx context.Context, name, rarity string, limit int) ([]*Item, error) {
	return c.SearchItems(ctx, ItemSearchOptions{
		Name:     name,
		Rarities: []string{rarity},
		Limit:    limit,
	})
}

// SkillSearchOptions represents search options for skills
type SkillSearchOptions struct {
	Name       string // Partial name to search for
	Profession string // Filter by profession
	Type       string // Filter by skill type
	Limit      int    // Maximum number of results to return (0 = no limit)
}

// SearchSkills searches for skills based on the provided criteria
// This function uses cached data if available, otherwise falls back to API
func (c *Client) SearchSkills(ctx context.Context, options SkillSearchOptions) ([]*Skill, error) {
	// Try cache first if available
	if c.dataCache != nil && c.dataCache.GetSkillCache().IsLoaded() {
		return c.dataCache.GetSkillCache().SearchSkills(options.Name, options.Profession, options.Type, options.Limit), nil
	}
	
	// Fallback to API-based search (limited for performance)
	// Note: This is a simplified fallback - in practice you might want to implement
	// a more sophisticated API-based search or just return an error
	return nil, fmt.Errorf("skill search requires data cache to be loaded")
}

// GetSkillsByName finds skills by partial name match
func (c *Client) GetSkillsByName(ctx context.Context, name string, limit int) ([]*Skill, error) {
	return c.SearchSkills(ctx, SkillSearchOptions{
		Name:  name,
		Limit: limit,
	})
}

// GetSkillsByProfession finds skills by profession
func (c *Client) GetSkillsByProfession(ctx context.Context, profession string, limit int) ([]*Skill, error) {
	return c.SearchSkills(ctx, SkillSearchOptions{
		Profession: profession,
		Limit:      limit,
	})
}
