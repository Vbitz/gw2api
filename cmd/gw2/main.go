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
		
		// Enable comprehensive data cache if data directory exists
		if _, err := os.Stat("data"); err == nil {
			opts = append(opts, gw2api.WithDataCache("data"))
		}
		
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
var currenciesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all currency IDs",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ids, err := client.GetCurrencyIDs(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		outputIDs(ids)
	},
}
var currenciesGetCmd = &cobra.Command{
	Use:   "get [id...]",
	Short: "Get specific currencies",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ids := parseIDs(args)

		if len(ids) == 1 {
			currency, err := client.GetCurrency(ctx, ids[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			outputData(currency)
		} else {
			currencies, err := client.GetCurrencies(ctx, ids)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			outputData(currencies)
		}
	},
}
var currenciesAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Get all currencies",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		currencies, err := client.GetAllCurrencies(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		outputData(currencies)
	},
}

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
var worldsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all world IDs",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ids, err := client.GetWorldIDs(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		outputIDs(ids)
	},
}
var worldsGetCmd = &cobra.Command{
	Use:   "get [id...]",
	Short: "Get specific worlds",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ids := parseIDs(args)

		if len(ids) == 1 {
			world, err := client.GetWorld(ctx, ids[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			outputData(world)
		} else {
			worlds, err := client.GetWorlds(ctx, ids)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			outputData(worlds)
		}
	},
}
var worldsAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Get all worlds",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		worlds, err := client.GetAllWorlds(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		outputData(worlds)
	},
}

var skillsCmd = &cobra.Command{Use: "skills", Short: "Skill operations"}
var skillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all skill IDs",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ids, err := client.GetSkillIDs(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		outputIDs(ids)
	},
}
var skillsGetCmd = &cobra.Command{
	Use:   "get [id...]",
	Short: "Get specific skills",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ids := parseIDs(args)

		if len(ids) == 1 {
			skill, err := client.GetSkill(ctx, ids[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			outputData(skill)
		} else {
			skills, err := client.GetSkills(ctx, ids)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			outputData(skills)
		}
	},
}

var commerceCmd = &cobra.Command{Use: "commerce", Short: "Commerce operations"}
var commercePricesCmd = &cobra.Command{
	Use:   "prices [item_id...]",
	Short: "Get trading post prices",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ids := parseIDs(args)

		if len(ids) == 1 {
			price, err := client.GetCommercePrice(ctx, ids[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			outputData(price)
		} else {
			prices, err := client.GetCommercePrices(ctx, ids)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			outputData(prices)
		}
	},
}

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
	case *gw2api.Currency:
		outputCurrencyTable([]*gw2api.Currency{v})
	case []*gw2api.Currency:
		outputCurrencyTable(v)
	case *gw2api.World:
		outputWorldTable([]*gw2api.World{v})
	case []*gw2api.World:
		outputWorldTable(v)
	case *gw2api.Skill:
		outputSkillTable([]*gw2api.Skill{v})
	case []*gw2api.Skill:
		outputSkillTable(v)
	case *gw2api.Price:
		outputPriceTable([]*gw2api.Price{v})
	case []*gw2api.Price:
		outputPriceTable(v)
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

func outputCurrencyTable(currencies []*gw2api.Currency) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("ID", "Name", "Description", "Order")

	for _, currency := range currencies {
		name := currency.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		description := currency.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		table.Append(
			strconv.Itoa(currency.ID),
			name,
			description,
			strconv.Itoa(currency.Order),
		)
	}
	table.Render()
}

func outputWorldTable(worlds []*gw2api.World) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("ID", "Name", "Population")

	for _, world := range worlds {
		name := world.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		table.Append(
			strconv.Itoa(world.ID),
			name,
			world.Population,
		)
	}
	table.Render()
}

func outputSkillTable(skills []*gw2api.Skill) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("ID", "Name", "Type", "Professions")

	for _, skill := range skills {
		name := skill.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		skillType := skill.Type
		if skillType == "" {
			skillType = "N/A"
		}

		professions := "N/A"
		if len(skill.Professions) > 0 {
			professions = strings.Join(skill.Professions, ", ")
			if len(professions) > 20 {
				professions = professions[:17] + "..."
			}
		}

		table.Append(
			strconv.Itoa(skill.ID),
			name,
			skillType,
			professions,
		)
	}
	table.Render()
}

func outputPriceTable(prices []*gw2api.Price) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Item ID", "Buy Price", "Buy Qty", "Sell Price", "Sell Qty")

	for _, price := range prices {
		table.Append(
			strconv.Itoa(price.ID),
			strconv.Itoa(price.Buys.UnitPrice),
			strconv.Itoa(price.Buys.Quantity),
			strconv.Itoa(price.Sells.UnitPrice),
			strconv.Itoa(price.Sells.Quantity),
		)
	}
	table.Render()
}
