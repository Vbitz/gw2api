package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
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
var client *Client

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gw2api",
	Short: "Guild Wars 2 API command-line client",
	Long: `A comprehensive command-line client for the Guild Wars 2 API.
	
Provides access to all public API endpoints with full type safety,
bulk operations, pagination, and multiple output formats.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize client with global flags
		var opts []ClientOption

		if timeout > 0 {
			opts = append(opts, WithTimeout(time.Duration(timeout)*time.Second))
		}

		if apiKey != "" {
			opts = append(opts, WithAPIKey(apiKey))
		}

		if language != "" {
			switch language {
			case "en":
				opts = append(opts, WithLanguage(LanguageEnglish))
			case "es":
				opts = append(opts, WithLanguage(LanguageSpanish))
			case "de":
				opts = append(opts, WithLanguage(LanguageGerman))
			case "fr":
				opts = append(opts, WithLanguage(LanguageFrench))
			case "zh":
				opts = append(opts, WithLanguage(LanguageChinese))
			default:
				fmt.Fprintf(os.Stderr, "Unsupported language: %s\n", language)
				os.Exit(1)
			}
		}

		opts = append(opts, WithUserAgent("gw2api-cli/1.0"))
		client = NewClient(opts...)
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format (json, table, yaml)")
	rootCmd.PersistentFlags().StringVarP(&language, "lang", "l", "en", "Language (en, es, de, fr, zh)")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 30, "Request timeout in seconds")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authenticated endpoints")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

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

// Achievements commands
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
	Use:   "get <id> [id...]",
	Short: "Get achievement(s) by ID",
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

// Currencies commands
var currenciesCmd = &cobra.Command{
	Use:     "currencies",
	Aliases: []string{"currency", "curr"},
	Short:   "Currency operations",
}

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
	Use:   "get <id> [id...]",
	Short: "Get currency/currencies by ID",
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

// Items commands
var itemsCmd = &cobra.Command{
	Use:     "items",
	Aliases: []string{"item"},
	Short:   "Item operations",
}

var itemsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all item IDs",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ids, err := client.GetItemIDs(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Total items: %d\n", len(ids))
		if outputFormat == "json" {
			outputIDs(ids)
		} else {
			fmt.Printf("First 10 IDs: %v\n", ids[:min(10, len(ids))])
			fmt.Println("Use --output json to see all IDs")
		}
	},
}

var itemsGetCmd = &cobra.Command{
	Use:   "get <id> [id...]",
	Short: "Get item(s) by ID",
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

// Worlds commands
var worldsCmd = &cobra.Command{
	Use:     "worlds",
	Aliases: []string{"world"},
	Short:   "World/server operations",
}

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
	Use:   "get <id> [id...]",
	Short: "Get world(s) by ID",
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

// Skills commands
var skillsCmd = &cobra.Command{
	Use:     "skills",
	Aliases: []string{"skill"},
	Short:   "Skill operations",
}

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

		fmt.Printf("Total skills: %d\n", len(ids))
		if outputFormat == "json" {
			outputIDs(ids)
		} else {
			fmt.Printf("First 10 IDs: %v\n", ids[:min(10, len(ids))])
			fmt.Println("Use --output json to see all IDs")
		}
	},
}

var skillsGetCmd = &cobra.Command{
	Use:   "get <id> [id...]",
	Short: "Get skill(s) by ID",
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
			// Note: GetSkills method would need to be implemented in the API
			fmt.Fprintf(os.Stderr, "Bulk skill retrieval not yet implemented\n")
			os.Exit(1)
		}
	},
}

// Commerce commands
var commerceCmd = &cobra.Command{
	Use:     "commerce",
	Aliases: []string{"tp", "trading"},
	Short:   "Trading post operations",
}

var commercePricesCmd = &cobra.Command{
	Use:   "prices <item_id> [item_id...]",
	Short: "Get trading post prices for items",
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

func init() {
	// Add subcommands to their parents
	achievementsCmd.AddCommand(achievementsListCmd, achievementsGetCmd)
	currenciesCmd.AddCommand(currenciesListCmd, currenciesGetCmd, currenciesAllCmd)
	itemsCmd.AddCommand(itemsListCmd, itemsGetCmd)
	worldsCmd.AddCommand(worldsListCmd, worldsGetCmd, worldsAllCmd)
	skillsCmd.AddCommand(skillsListCmd, skillsGetCmd)
	commerceCmd.AddCommand(commercePricesCmd)
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

func outputData(data interface{}) {
	switch outputFormat {
	case "json":
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	case "table":
		outputTable(data)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported output format: %s\n", outputFormat)
		os.Exit(1)
	}
}

func outputIDs(ids []int) {
	if outputFormat == "json" {
		jsonData, err := json.MarshalIndent(ids, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	} else {
		fmt.Printf("Total: %d IDs\n", len(ids))
		for i, id := range ids {
			if i > 0 && i%10 == 0 {
				fmt.Println()
			}
			fmt.Printf("%d ", id)
		}
		fmt.Println()
	}
}

func outputTable(data interface{}) {
	// Simple table formatting without external libraries
	switch v := data.(type) {
	case *Build:
		fmt.Printf("Build ID: %d\n", v.ID)

	case *Achievement:
		fmt.Printf("ID: %d\n", v.ID)
		fmt.Printf("Name: %s\n", v.Name)
		fmt.Printf("Description: %s\n", v.Description)
		fmt.Printf("Type: %s\n", v.Type)
		fmt.Printf("Tiers: %d\n", len(v.Tiers))
		fmt.Printf("Flags: %s\n", strings.Join(v.Flags, ", "))

	case []*Achievement:
		fmt.Printf("%-8s %-30s %-15s %-5s\n", "ID", "Name", "Type", "Tiers")
		fmt.Println(strings.Repeat("-", 60))
		for _, achievement := range v {
			name := achievement.Name
			if len(name) > 28 {
				name = name[:28] + ".."
			}
			fmt.Printf("%-8d %-30s %-15s %-5d\n",
				achievement.ID, name, achievement.Type, len(achievement.Tiers))
		}

	case *Currency:
		fmt.Printf("ID: %d\n", v.ID)
		fmt.Printf("Name: %s\n", v.Name)
		fmt.Printf("Description: %s\n", v.Description)
		fmt.Printf("Order: %d\n", v.Order)

	case []*Currency:
		fmt.Printf("%-5s %-20s %-50s %-5s\n", "ID", "Name", "Description", "Order")
		fmt.Println(strings.Repeat("-", 82))
		for _, currency := range v {
			desc := currency.Description
			if len(desc) > 48 {
				desc = desc[:48] + ".."
			}
			fmt.Printf("%-5d %-20s %-50s %-5d\n",
				currency.ID, currency.Name, desc, currency.Order)
		}

	case *World:
		fmt.Printf("ID: %d\n", v.ID)
		fmt.Printf("Name: %s\n", v.Name)
		fmt.Printf("Population: %s\n", v.Population)

	case []*World:
		fmt.Printf("%-6s %-25s %-12s\n", "ID", "Name", "Population")
		fmt.Println(strings.Repeat("-", 45))
		for _, world := range v {
			fmt.Printf("%-6d %-25s %-12s\n", world.ID, world.Name, world.Population)
		}

	case *Item:
		fmt.Printf("ID: %d\n", v.ID)
		fmt.Printf("Name: %s\n", v.Name)
		fmt.Printf("Type: %s\n", v.Type)
		fmt.Printf("Rarity: %s\n", v.Rarity)
		fmt.Printf("Level: %d\n", v.Level)
		fmt.Printf("Vendor Value: %d\n", v.VendorValue)

	case []*Item:
		fmt.Printf("%-8s %-30s %-15s %-10s %-5s\n", "ID", "Name", "Type", "Rarity", "Level")
		fmt.Println(strings.Repeat("-", 70))
		for _, item := range v {
			name := item.Name
			if len(name) > 28 {
				name = name[:28] + ".."
			}
			fmt.Printf("%-8d %-30s %-15s %-10s %-5d\n",
				item.ID, name, item.Type, item.Rarity, item.Level)
		}

	case *Price:
		fmt.Printf("Item ID: %d\n", v.ID)
		fmt.Printf("Buy Price: %d copper (%d available)\n", v.Buys.UnitPrice, v.Buys.Quantity)
		fmt.Printf("Sell Price: %d copper (%d available)\n", v.Sells.UnitPrice, v.Sells.Quantity)

	case []*Price:
		fmt.Printf("%-8s %-12s %-12s %-12s %-12s\n", "Item ID", "Buy Price", "Buy Qty", "Sell Price", "Sell Qty")
		fmt.Println(strings.Repeat("-", 62))
		for _, price := range v {
			fmt.Printf("%-8d %-12d %-12d %-12d %-12d\n",
				price.ID, price.Buys.UnitPrice, price.Buys.Quantity,
				price.Sells.UnitPrice, price.Sells.Quantity)
		}

	default:
		// Fallback to JSON for unknown types
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting data: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
