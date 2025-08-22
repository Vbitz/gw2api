package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
	"j5.nz/gw2/internal/gw2api"
)

func updateItemDb(client *gw2api.Client, f io.Writer, limit, groupSize int) error {
	itemIds, err := client.GetItemIDs(context.Background())
	if err != nil {
		return err
	}

	if len(itemIds) > limit {
		itemIds = itemIds[:limit]
	}

	pb := progressbar.Default(int64(len(itemIds)), "Fetching items")
	defer pb.Finish()

	for i := 0; i < len(itemIds); i += groupSize {
		end := min(i+groupSize, len(itemIds))

		items, err := client.GetItems(context.Background(), itemIds[i:end])
		if err != nil {
			return err
		}

		for _, item := range items {
			if err := json.NewEncoder(f).Encode(item); err != nil {
				return err
			}
			pb.Add(1)
		}

		time.Sleep(200 * time.Millisecond) // Rate limit to avoid hitting API too hard
	}

	return nil
}

func updateSkillsDb(client *gw2api.Client, f io.Writer, limit, groupSize int) error {
	skillIds, err := client.GetSkillIDs(context.Background())
	if err != nil {
		return err
	}

	if len(skillIds) > limit {
		skillIds = skillIds[:limit]
	}

	pb := progressbar.Default(int64(len(skillIds)), "Fetching skills")
	defer pb.Finish()

	for i := 0; i < len(skillIds); i += groupSize {
		end := min(i+groupSize, len(skillIds))

		skills, err := client.GetSkills(context.Background(), skillIds[i:end])
		if err != nil {
			return err
		}

		for _, skill := range skills {
			if err := json.NewEncoder(f).Encode(skill); err != nil {
				return err
			}
			pb.Add(1)
		}

		time.Sleep(200 * time.Millisecond) // Rate limit to avoid hitting API too hard
	}

	return nil
}

func updateRecipesDb(client *gw2api.Client, f io.Writer, limit, groupSize int) error {
	recipeIds, err := client.GetRecipeIDs(context.Background())
	if err != nil {
		return err
	}

	if len(recipeIds) > limit {
		recipeIds = recipeIds[:limit]
	}

	pb := progressbar.Default(int64(len(recipeIds)), "Fetching recipes")
	defer pb.Finish()

	for i := 0; i < len(recipeIds); i += groupSize {
		end := min(i+groupSize, len(recipeIds))

		recipes, err := client.GetRecipes(context.Background(), recipeIds[i:end])
		if err != nil {
			return err
		}

		for _, recipe := range recipes {
			if err := json.NewEncoder(f).Encode(recipe); err != nil {
				return err
			}
			pb.Add(1)
		}

		time.Sleep(200 * time.Millisecond) // Rate limit to avoid hitting API too hard
	}

	return nil
}

func updateAchievementsDb(client *gw2api.Client, f io.Writer, limit, groupSize int) error {
	achievementIds, err := client.GetAchievementIDs(context.Background())
	if err != nil {
		return err
	}

	if len(achievementIds) > limit {
		achievementIds = achievementIds[:limit]
	}

	pb := progressbar.Default(int64(len(achievementIds)), "Fetching achievements")
	defer pb.Finish()

	for i := 0; i < len(achievementIds); i += groupSize {
		end := min(i+groupSize, len(achievementIds))

		achievements, err := client.GetAchievements(context.Background(), achievementIds[i:end])
		if err != nil {
			return err
		}

		for _, achievement := range achievements {
			if err := json.NewEncoder(f).Encode(achievement); err != nil {
				return err
			}
			pb.Add(1)
		}

		time.Sleep(200 * time.Millisecond) // Rate limit to avoid hitting API too hard
	}

	return nil
}

func updateAchievementCategoryDb(client *gw2api.Client, f io.Writer, limit, groupSize int) error {
	categories, err := client.GetAchievementCategoryIDs(context.Background())
	if err != nil {
		return err
	}

	if len(categories) > limit {
		categories = categories[:limit]
	}

	pb := progressbar.Default(int64(len(categories)), "Fetching achievement categories")
	defer pb.Finish()

	for i := 0; i < len(categories); i += groupSize {
		end := min(i+groupSize, len(categories))

		cats, err := client.GetAchievementCategories(context.Background(), categories[i:end])
		if err != nil {
			return err
		}

		for _, cat := range cats {
			if err := json.NewEncoder(f).Encode(cat); err != nil {
				return err
			}
			pb.Add(1)
		}

		time.Sleep(200 * time.Millisecond) // Rate limit to avoid hitting API too hard
	}

	return nil
}

func updateSkinsDb(client *gw2api.Client, f io.Writer, limit, groupSize int) error {
	skinIds, err := client.GetSkinIDs(context.Background())
	if err != nil {
		return err
	}

	if len(skinIds) > limit {
		skinIds = skinIds[:limit]
	}

	pb := progressbar.Default(int64(len(skinIds)), "Fetching skins")
	defer pb.Finish()

	for i := 0; i < len(skinIds); i += groupSize {
		end := min(i+groupSize, len(skinIds))

		skins, err := client.GetSkins(context.Background(), skinIds[i:end])
		if err != nil {
			return err
		}

		for _, skin := range skins {
			if err := json.NewEncoder(f).Encode(skin); err != nil {
				return err
			}
			pb.Add(1)
		}

		time.Sleep(200 * time.Millisecond) // Rate limit to avoid hitting API too hard
	}

	return nil
}

func main() {
	var (
		kind      = flag.String("kind", "", "Kind of data to fetch (e.g., item, recipe, etc.)")
		groupSize = flag.Int("group-size", 200, "Number of items to fetch in each group")
		limit     = flag.Int("limit", 100000, "Maximum number of items to fetch")
	)

	flag.Parse()

	client := gw2api.NewClient()

	switch *kind {
	case "item":
		out, err := os.Create("data/items.json")
		if err != nil {
			panic(err)
		}
		defer out.Close()

		if err := updateItemDb(client, out, *limit, *groupSize); err != nil {
			panic(err)
		}
	case "skills":
		out, err := os.Create("data/skills.json")
		if err != nil {
			panic(err)
		}
		defer out.Close()

		if err := updateSkillsDb(client, out, *limit, *groupSize); err != nil {
			panic(err)
		}
	case "recipes":
		out, err := os.Create("data/recipes.json")
		if err != nil {
			panic(err)
		}
		defer out.Close()

		if err := updateRecipesDb(client, out, *limit, *groupSize); err != nil {
			panic(err)
		}
	case "achievements":
		out, err := os.Create("data/achievements.json")
		if err != nil {
			panic(err)
		}
		defer out.Close()

		if err := updateAchievementsDb(client, out, *limit, *groupSize); err != nil {
			panic(err)
		}
	case "achievement-categories":
		out, err := os.Create("data/achievement_categories.json")
		if err != nil {
			panic(err)
		}
		defer out.Close()

		if err := updateAchievementCategoryDb(client, out, *limit, *groupSize); err != nil {
			panic(err)
		}
	case "skins":
		out, err := os.Create("data/skins.json")
		if err != nil {
			panic(err)
		}
		defer out.Close()

		if err := updateSkinsDb(client, out, *limit, *groupSize); err != nil {
			panic(err)
		}
	default:
		panic("Unsupported kind: " + *kind)
	}
}
