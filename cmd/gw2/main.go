package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"j5.nz/gw2/internal/gw2api"
)

// Global flags
var (
	outputFormat string
	language     string
	timeout      int
	apiKey       string
	verbose      bool
)

// Global client
var client *gw2api.Client

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gw2api",
	Short: "Guild Wars 2 API command-line client",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize client with global flags
		var opts []gw2api.ClientOption

		if timeout > 0 {
			opts = append(opts, gw2api.WithTimeout(time.Duration(timeout)*time.Second))
		}

		if apiKey != "" {
			opts = append(opts, gw2api.WithAPIKey(apiKey))
		}

		if language != "" {
			switch language {
			case "en":
				opts = append(opts, gw2api.WithLanguage(gw2api.LanguageEnglish))
			case "es":
				opts = append(opts, gw2api.WithLanguage(gw2api.LanguageSpanish))
			case "de":
				opts = append(opts, gw2api.WithLanguage(gw2api.LanguageGerman))
			case "fr":
				opts = append(opts, gw2api.WithLanguage(gw2api.LanguageFrench))
			case "zh":
				opts = append(opts, gw2api.WithLanguage(gw2api.LanguageChinese))
			default:
				fmt.Fprintf(os.Stderr, "Unsupported language: %s\n", language)
				os.Exit(1)
			}
		}

		opts = append(opts, gw2api.WithUserAgent("gw2api-cli/1.0"))
		client = gw2api.NewClient(opts...)
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (json, table, yaml)")
	rootCmd.PersistentFlags().StringVarP(&language, "lang", "l", "en", "Language (en, es, de, fr, zh)")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 30, "Request timeout in seconds")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authenticated endpoints")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Command-specific flags
	itemsSearchCmd.Flags().StringP("name", "n", "", "Search for items containing this name (case-insensitive)")
	itemsSearchCmd.Flags().StringP("rarity", "r", "", "Filter by rarity (Basic, Fine, Masterwork, Rare, Exotic, Ascended, Legendary)")
	itemsSearchCmd.Flags().IntP("limit", "", 50, "Maximum number of results to return (0 = no limit)")

	// Add all subcommands
	rootCmd.AddCommand(
		buildCmd,
		achievementsCmd,
		currenciesCmd,
		itemsCmd,
		worldsCmd,
		skillsCmd,
		commerceCmd,
		versionCmd,
	)

	// Add subcommands to their parents
	achievementsCmd.AddCommand(achievementsListCmd, achievementsGetCmd)
	currenciesCmd.AddCommand(currenciesListCmd, currenciesGetCmd, currenciesAllCmd)
	itemsCmd.AddCommand(itemsListCmd, itemsGetCmd, itemsSearchCmd)
	worldsCmd.AddCommand(worldsListCmd, worldsGetCmd, worldsAllCmd)
	skillsCmd.AddCommand(skillsListCmd, skillsGetCmd)
	commerceCmd.AddCommand(commercePricesCmd)
}

// Version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gw2api-cli v1.0.0")
		fmt.Println("Guild Wars 2 API Go Client")
	},
}

// Build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Get current game build",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		build, err := client.GetBuild(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		outputData(build)
	},
}

// Rest of the CLI commands would go here...
// For now, let me add placeholder commands to avoid compilation errors

var achievementsCmd = &cobra.Command{
	Use:     "achievements",
	Aliases: []string{"achievement", "ach"},
	Short:   "Achievement operations",
}

var achievementsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all achievement IDs",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ids, err := client.GetAchievementIDs(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		outputIDs(ids)
	},
}

var achievementsGetCmd = &cobra.Command{
	Use:   "get [id...]",
	Short: "Get specific achievements",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ids := parseIDs(args)

		if len(ids) == 1 {
			achievement, err := client.GetAchievement(ctx, ids[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			outputData(achievement)
		} else {
			achievements, err := client.GetAchievements(ctx, ids)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			outputData(achievements)
		}
	},
}

// Placeholder commands
var currenciesCmd = &cobra.Command{Use: "currencies", Short: "Currency operations"}
var currenciesListCmd = &cobra.Command{Use: "list", Short: "List currencies", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Not implemented yet") }}
var currenciesGetCmd = &cobra.Command{Use: "get", Short: "Get currencies", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Not implemented yet") }}
var currenciesAllCmd = &cobra.Command{Use: "all", Short: "Get all currencies", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Not implemented yet") }}

var itemsCmd = &cobra.Command{Use: "items", Short: "Item operations"}
var itemsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List item IDs",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ids, err := client.GetItemIDs(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		outputIDs(ids)
	},
}

var itemsGetCmd = &cobra.Command{
	Use:   "get [id...]",
	Short: "Get items by ID",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ids := parseIDs(args)

		if len(ids) == 1 {
			item, err := client.GetItem(ctx, ids[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			outputData(item)
		} else {
			items, err := client.GetItems(ctx, ids)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			outputData(items)
		}
	},
}

var itemsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search items by name and/or rarity",
	Long: `Search for items with optional filtering by name and rarity.
	
Examples:
  # Search for items with "sword" in the name
  gw2api items search --name sword
  
  # Search for exotic items
  gw2api items search --rarity exotic
  
  # Search for exotic swords (combining filters)
  gw2api items search --name sword --rarity exotic
  
  # Limit results to 10 items
  gw2api items search --name "berserker" --limit 10`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		name, _ := cmd.Flags().GetString("name")
		rarity, _ := cmd.Flags().GetString("rarity")
		limit, _ := cmd.Flags().GetInt("limit")

		if name == "" && rarity == "" {
			fmt.Fprintf(os.Stderr, "Error: At least one search criteria (--name or --rarity) must be provided\n")
			os.Exit(1)
		}

		options := gw2api.ItemSearchOptions{
			Name:  name,
			Limit: limit,
		}

		if rarity != "" {
			options.Rarities = []string{rarity}
		}

		items, err := client.SearchItems(ctx, options)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(items) == 0 {
			fmt.Println("No items found matching the search criteria")
			return
		}

		outputData(items)
	},
}

var worldsCmd = &cobra.Command{Use: "worlds", Short: "World operations"}
var worldsListCmd = &cobra.Command{Use: "list", Short: "List worlds", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Not implemented yet") }}
var worldsGetCmd = &cobra.Command{Use: "get", Short: "Get worlds", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Not implemented yet") }}
var worldsAllCmd = &cobra.Command{Use: "all", Short: "Get all worlds", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Not implemented yet") }}

var skillsCmd = &cobra.Command{Use: "skills", Short: "Skill operations"}
var skillsListCmd = &cobra.Command{Use: "list", Short: "List skills", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Not implemented yet") }}
var skillsGetCmd = &cobra.Command{Use: "get", Short: "Get skills", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Not implemented yet") }}

var commerceCmd = &cobra.Command{Use: "commerce", Short: "Commerce operations"}
var commercePricesCmd = &cobra.Command{Use: "prices", Short: "Get prices", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Not implemented yet") }}

// Helper functions
func parseIDs(args []string) []int {
	var ids []int
	for _, arg := range args {
		// Handle comma-separated IDs
		for _, idStr := range strings.Split(arg, ",") {
			id, err := strconv.Atoi(strings.TrimSpace(idStr))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid ID: %s\n", idStr)
				os.Exit(1)
			}
			ids = append(ids, id)
		}
	}
	return ids
}

func outputIDs(ids []int) {
	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(ids, "", "  ")
		fmt.Println(string(data))
	case "table":
		for _, id := range ids {
			fmt.Println(id)
		}
	default:
		data, _ := json.MarshalIndent(ids, "", "  ")
		fmt.Println(string(data))
	}
}

func outputData(data any) {
	switch outputFormat {
	case "json":
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(jsonData))
	case "table":
		outputTable(data)
	default:
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(jsonData))
	}
}

func outputTable(data any) {
	switch v := data.(type) {
	case *gw2api.Item:
		outputItemTable([]*gw2api.Item{v})
	case []*gw2api.Item:
		outputItemTable(v)
	case *gw2api.Achievement:
		outputAchievementTable([]*gw2api.Achievement{v})
	case []*gw2api.Achievement:
		outputAchievementTable(v)
	default:
		// Fallback to simple printing
		fmt.Printf("%+v\n", data)
	}
}

func outputItemTable(items []*gw2api.Item) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("ID", "Name", "Type", "Rarity", "Level")

	for _, item := range items {
		name := item.Name
		if len(name) > 50 {
			name = name[:47] + "..."
		}
		table.Append(
			strconv.Itoa(item.ID),
			name,
			item.Type,
			item.Rarity,
			strconv.Itoa(item.Level),
		)
	}
	table.Render()
}

func outputAchievementTable(achievements []*gw2api.Achievement) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("ID", "Name", "Type", "Points")

	for _, achievement := range achievements {
		name := achievement.Name
		if len(name) > 40 {
			name = name[:37] + "..."
		}

		points := "0"
		if len(achievement.Tiers) > 0 {
			totalPoints := 0
			for _, tier := range achievement.Tiers {
				totalPoints += tier.Points
			}
			points = strconv.Itoa(totalPoints)
		}

		table.Append(
			strconv.Itoa(achievement.ID),
			name,
			achievement.Type,
			points,
		)
	}
	table.Render()
}
