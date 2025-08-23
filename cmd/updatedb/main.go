package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"
	"sync"

	"github.com/schollz/progressbar/v3"
	"j5.nz/gw2/internal/gw2api"
)

type workJob[T any] struct {
	ids []int
}

type workResult[T any] struct {
	data []T
	err  error
}

func genericUpdate[T any](
	f io.Writer,
	limit, groupSize, concurrency int,
	getIDs func(context.Context) ([]int, error),
	getData func(context.Context, []int) ([]T, error),
	label string,
) error {
	ids, err := getIDs(context.Background())
	if err != nil {
		return err
	}

	if len(ids) > limit {
		ids = ids[:limit]
	}

	pb := progressbar.Default(int64(len(ids)), label)
	defer pb.Finish()

	// Calculate number of batches
	batches := make([][]int, 0)
	for i := 0; i < len(ids); i += groupSize {
		end := min(i+groupSize, len(ids))
		batches = append(batches, ids[i:end])
	}

	// Use a worker pool with controlled concurrency
	// Since rate limiting is handled at the client level, we can be more aggressive
	maxWorkers := min(20, concurrency) // Cap at 20 to be reasonable
	numWorkers := min(maxWorkers, len(batches))
	jobs := make(chan workJob[T], len(batches))
	results := make(chan workResult[T], len(batches))

	// Start workers
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				data, err := getData(context.Background(), job.ids)
				results <- workResult[T]{
					data: data,
					err:  err,
				}
			}
		}()
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for _, batch := range batches {
			jobs <- workJob[T]{ids: batch}
		}
	}()

	// Close workers
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results as they come in
	var writeMu sync.Mutex
	for result := range results {
		if result.err != nil {
			return result.err
		}
		
		// Write results immediately as they arrive
		for _, item := range result.data {
			writeMu.Lock()
			if err := json.NewEncoder(f).Encode(item); err != nil {
				writeMu.Unlock()
				return err
			}
			writeMu.Unlock()
			pb.Add(1)
		}
	}

	return nil
}

func main() {
	var (
		kind        = flag.String("kind", "", "Kind of data to fetch (e.g., item, recipe, etc.)")
		groupSize   = flag.Int("group-size", 200, "Number of items to fetch in each group")
		limit       = flag.Int("limit", 100000, "Maximum number of items to fetch")
		concurrency = flag.Int("concurrency", 10, "Number of concurrent requests (max 20)")
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

		if err := genericUpdate(out, *limit, *groupSize, *concurrency,
			func(ctx context.Context) ([]int, error) { return client.GetItemIDs(ctx) },
			func(ctx context.Context, ids []int) ([]*gw2api.Item, error) { return client.GetItems(ctx, ids) },
			"Fetching items"); err != nil {
			panic(err)
		}
	case "skills":
		out, err := os.Create("data/skills.json")
		if err != nil {
			panic(err)
		}
		defer out.Close()

		if err := genericUpdate(out, *limit, *groupSize, *concurrency,
			func(ctx context.Context) ([]int, error) { return client.GetSkillIDs(ctx) },
			func(ctx context.Context, ids []int) ([]*gw2api.Skill, error) { return client.GetSkills(ctx, ids) },
			"Fetching skills"); err != nil {
			panic(err)
		}
	case "recipes":
		out, err := os.Create("data/recipes.json")
		if err != nil {
			panic(err)
		}
		defer out.Close()

		if err := genericUpdate(out, *limit, *groupSize, *concurrency,
			func(ctx context.Context) ([]int, error) { return client.GetRecipeIDs(ctx) },
			func(ctx context.Context, ids []int) ([]*gw2api.RecipeDetail, error) {
				return client.GetRecipes(ctx, ids)
			},
			"Fetching recipes"); err != nil {
			panic(err)
		}
	case "achievements":
		out, err := os.Create("data/achievements.json")
		if err != nil {
			panic(err)
		}
		defer out.Close()

		if err := genericUpdate(out, *limit, *groupSize, *concurrency,
			func(ctx context.Context) ([]int, error) { return client.GetAchievementIDs(ctx) },
			func(ctx context.Context, ids []int) ([]*gw2api.Achievement, error) {
				return client.GetAchievements(ctx, ids)
			},
			"Fetching achievements"); err != nil {
			panic(err)
		}
	case "achievement-categories":
		out, err := os.Create("data/achievement_categories.json")
		if err != nil {
			panic(err)
		}
		defer out.Close()

		if err := genericUpdate(out, *limit, *groupSize, *concurrency,
			func(ctx context.Context) ([]int, error) { return client.GetAchievementCategoryIDs(ctx) },
			func(ctx context.Context, ids []int) ([]*gw2api.AchievementCategory, error) {
				return client.GetAchievementCategories(ctx, ids)
			},
			"Fetching achievement categories"); err != nil {
			panic(err)
		}
	case "skins":
		out, err := os.Create("data/skins.json")
		if err != nil {
			panic(err)
		}
		defer out.Close()

		if err := genericUpdate(out, *limit, *groupSize, *concurrency,
			func(ctx context.Context) ([]int, error) { return client.GetSkinIDs(ctx) },
			func(ctx context.Context, ids []int) ([]*gw2api.SkinDetail, error) { return client.GetSkins(ctx, ids) },
			"Fetching skins"); err != nil {
			panic(err)
		}
	default:
		panic("Unsupported kind: " + *kind)
	}
}
