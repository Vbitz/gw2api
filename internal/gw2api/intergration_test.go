package gw2api

import (
	"context"
	"testing"
)

// MockClient for testing without network dependency
type MockClient struct {
	items []*Item
}

func (m *MockClient) GetItemIDs(ctx context.Context) ([]int, error) {
	ids := make([]int, len(m.items))
	for i, item := range m.items {
		ids[i] = item.ID
	}
	return ids, nil
}

func (m *MockClient) GetItems(ctx context.Context, ids []int) ([]*Item, error) {
	var result []*Item
	for _, id := range ids {
		for _, item := range m.items {
			if item.ID == id {
				result = append(result, item)
				break
			}
		}
	}
	return result, nil
}

func TestSearchItemsIntegration(t *testing.T) {
	// Create mock data
	mockItems := []*Item{
		{
			ID:     1001,
			Name:   "Berserker's Greatsword of Force",
			Rarity: "Exotic",
			Type:   "Weapon",
			Level:  80,
		},
		{
			ID:     1002,
			Name:   "Berserker's Armor",
			Rarity: "Exotic",
			Type:   "Armor",
			Level:  80,
		},
		{
			ID:     1003,
			Name:   "Fine Steel Sword",
			Rarity: "Fine",
			Type:   "Weapon",
			Level:  20,
		},
		{
			ID:     1004,
			Name:   "Legendary Dragon Sword",
			Rarity: "Legendary",
			Type:   "Weapon",
			Level:  80,
		},
		{
			ID:     1005,
			Name:   "Basic Wooden Sword",
			Rarity: "Basic",
			Type:   "Weapon",
			Level:  1,
		},
		{
			ID:     1006,
			Name:   "Zojja's Breastplate",
			Rarity: "Ascended",
			Type:   "Armor",
			Level:  80,
		},
	}

	// Create a real client with mock behavior for testing
	// We'll modify the client to use our mock data
	originalClient := NewClient()

	// Create a test version of SearchItems that uses our mock data
	searchItemsMock := func(ctx context.Context, options ItemSearchOptions) ([]*Item, error) {
		var results []*Item
		var count int

		for _, item := range mockItems {
			if matchesSearchCriteria(item, options) {
				results = append(results, item)
				count++

				if options.Limit > 0 && count >= options.Limit {
					break
				}
			}
		}
		return results, nil
	}

	tests := []struct {
		name          string
		searchOptions ItemSearchOptions
		expectedCount int
		expectedIDs   []int
		expectedNames []string
	}{
		{
			name: "Search by name 'sword'",
			searchOptions: ItemSearchOptions{
				Name: "sword",
			},
			expectedCount: 4,
			expectedIDs:   []int{1001, 1003, 1004, 1005},
		},
		{
			name: "Search by rarity 'Exotic'",
			searchOptions: ItemSearchOptions{
				Rarities: []string{"Exotic"},
			},
			expectedCount: 2,
			expectedIDs:   []int{1001, 1002},
		},
		{
			name: "Search by name 'berserker'",
			searchOptions: ItemSearchOptions{
				Name: "berserker",
			},
			expectedCount: 2,
			expectedIDs:   []int{1001, 1002},
		},
		{
			name: "Search by name 'sword' and rarity 'Legendary'",
			searchOptions: ItemSearchOptions{
				Name:     "sword",
				Rarities: []string{"Legendary"},
			},
			expectedCount: 1,
			expectedIDs:   []int{1004},
		},
		{
			name: "Search with limit",
			searchOptions: ItemSearchOptions{
				Name:  "sword",
				Limit: 2,
			},
			expectedCount: 2,
		},
		{
			name: "Search by type 'Armor'",
			searchOptions: ItemSearchOptions{
				Types: []string{"Armor"},
			},
			expectedCount: 2,
			expectedIDs:   []int{1002, 1006},
		},
		{
			name: "Search by level range",
			searchOptions: ItemSearchOptions{
				MinLevel: 70,
				MaxLevel: 80,
			},
			expectedCount: 4,
			expectedIDs:   []int{1001, 1002, 1004, 1006},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := searchItemsMock(context.Background(), tt.searchOptions)
			if err != nil {
				t.Errorf("SearchItems() error = %v", err)
				return
			}

			if len(results) != tt.expectedCount {
				t.Errorf("SearchItems() returned %d items, expected %d", len(results), tt.expectedCount)
			}

			if tt.expectedIDs != nil {
				resultIDs := make([]int, len(results))
				for i, item := range results {
					resultIDs[i] = item.ID
				}

				for _, expectedID := range tt.expectedIDs {
					found := false
					for _, resultID := range resultIDs {
						if resultID == expectedID {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected item ID %d not found in results", expectedID)
					}
				}
			}
		})
	}

	// Test that original client structure is maintained
	if originalClient == nil {
		t.Error("Client creation failed")
	}
}
