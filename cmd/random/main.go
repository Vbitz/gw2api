package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"time"

	"j5.nz/gw2/internal/gw2api"
)

const BLACK_LION_COLLECTIONS_ID = 76

type GW2Price int

func (p GW2Price) String() string {
	// price is in copper, 100 copper = 1 silver, 100 silver = 1 gold
	gold := p / 10000
	silver := (p % 10000) / 100
	copper := p % 100
	return fmt.Sprintf("%dg %ds %dc", gold, silver, copper)
}

type BlackLionSkin struct {
	SkinID    int      `json:"skin_id"`
	UnlockID  int      `json:"unlock_id"`
	Name      string   `json:"name"`
	SellPrice GW2Price `json:"sell_price"`
}

type BlackLionCollection struct {
	ID    int             `json:"id"`
	Name  string          `json:"name"`
	Skins []BlackLionSkin `json:"skins"`
}

func queryBlackLionCollections(client *gw2api.Client) error {
	cat, err := client.GetAchievementCategory(context.Background(), BLACK_LION_COLLECTIONS_ID)
	if err != nil {
		return fmt.Errorf("failed to query black lion collections: %w", err)
	}

	var ret []BlackLionCollection

	achievementInfo, err := client.GetAchievements(context.Background(), cat.Achievements)
	if err != nil {
		return fmt.Errorf("failed to get achievements for black lion collections: %w", err)
	}
	for _, achievement := range achievementInfo {
		slog.Info("black lion collection achievement", "id", achievement.ID, "name", achievement.Name)

		allItems := make(map[int]*gw2api.Item)
		for _, bit := range achievement.Bits {
			if bit.Type != "Skin" {
				continue
			}
			// slog.Info("  bit", "kind", bit.Type, "id", bit.ID)

			// find a item that grants the skin
			items, err := client.SearchItems(context.Background(), gw2api.ItemSearchOptions{
				UnlocksSkin: bit.ID,
			})
			if err != nil {
				slog.Error("failed to search items for skin", "skin_id", bit.ID, "error", err)
				continue
			}

			for _, item := range items {
				// slog.Info("    item", "id", item.ID, "name", item.Name, "type", item.Type, "rarity", item.Rarity)

				allItems[item.ID] = item
			}
		}

		itemIds := make([]int, 0, len(allItems))
		for id := range allItems {
			itemIds = append(itemIds, id)
		}

		if len(itemIds) > 0 {
			prices, err := client.GetCommercePrices(context.Background(), itemIds)
			if err != nil {
				slog.Error("failed to get commerce prices for items", "error", err)
				continue
			}

			collection := BlackLionCollection{
				ID:   achievement.ID,
				Name: achievement.Name,
			}

			for _, price := range prices {
				item, ok := allItems[price.ID]
				if !ok {
					slog.Error("item not found for price", "id", price.ID)
					continue
				}
				// slog.Info("    price", "id", item.ID, "name", item.Name, "price", GW2Price(price.Sells.UnitPrice))
				collection.Skins = append(collection.Skins, BlackLionSkin{
					SkinID:    item.Details.Skins[0], // assuming the first unlock
					UnlockID:  item.ID,
					Name:      item.Name,
					SellPrice: GW2Price(price.Sells.UnitPrice),
				})
			}

			ret = append(ret, collection)

			time.Sleep(200 * time.Millisecond) // Rate limit to avoid hitting API too hard
		} else {
			slog.Info("no items found for achievement", "id", achievement.ID, "name", achievement.Name)
		}
	}

	out, err := os.Create("data/black_lion_collections.json")
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	if err := json.NewEncoder(out).Encode(ret); err != nil {
		return fmt.Errorf("failed to write black lion collections to file: %w", err)
	}

	return nil
}

func checkBlackLionCollections(client *gw2api.Client) error {
	// load the saved JSON file
	file, err := os.Open("data/black_lion_collections.json")
	if err != nil {
		return fmt.Errorf("failed to open black lion collections file: %w", err)
	}
	defer file.Close()

	var collections []BlackLionCollection
	if err := json.NewDecoder(file).Decode(&collections); err != nil {
		return fmt.Errorf("failed to decode black lion collections JSON: %w", err)
	}

	accountSkins, err := client.GetAccountSkins(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get account skins: %w", err)
	}

	totalCollectionPrices := make(map[int]GW2Price)

	// Find how expensive each collection is to complete
	for _, collection := range collections {
		// slog.Info("Collection", "id", collection.ID, "skins_count", len(collection.Skins))

		remainingSkins := len(collection.Skins)

		totalPrice := GW2Price(0)
		for _, skin := range collection.Skins {
			if slices.Contains(accountSkins, skin.SkinID) {
				// slog.Info("Skin already collected", "skin_id", skin.SkinID, "name", skin.Name)
				remainingSkins--
			} else {
				// slog.Info("Skin not collected", "skin_id", skin.SkinID, "name", skin.Name, "sell_price", skin.SellPrice)
				totalPrice += skin.SellPrice
			}
		}

		totalCollectionPrices[collection.ID] = totalPrice
	}

	// Sort collections by total price
	slices.SortFunc(collections, func(a, b BlackLionCollection) int {
		return int(totalCollectionPrices[a.ID]) - int(totalCollectionPrices[b.ID])
	})

	// Print the collections with their total prices
	for _, collection := range collections {
		totalPrice := totalCollectionPrices[collection.ID]
		if totalPrice > 0 {
			slog.Info("Collection", "id", collection.ID, "name", collection.Name,
				"skins_count", len(collection.Skins),
				"total_price", totalPrice.String(),
			)
		}
	}

	return nil
}

func appMain() error {
	apiKey := os.Getenv("GW2_API_KEY")

	client := gw2api.NewClient(
		gw2api.WithDataCache("data"),
		gw2api.WithAPIKey(apiKey),
	)

	switch os.Args[1] {
	case "black-lion":
		return queryBlackLionCollections(client)
	case "check-black-lion":
		return checkBlackLionCollections(client)
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func main() {
	if err := appMain(); err != nil {
		slog.Error("error running app", "error", err)
		os.Exit(1)
	}
}
