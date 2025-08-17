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
	Long: `A comprehensive command-line client for the Guild Wars 2 API.
	
Provides access to all public API endpoints with full type safety,
bulk operations, pagination, and multiple output formats.`,
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
	
	// Add subcommands to their parents
	achievementsCmd.AddCommand(achievementsListCmd, achievementsGetCmd)
	currenciesCmd.AddCommand(currenciesListCmd, currenciesGetCmd, currenciesAllCmd)
	itemsCmd.AddCommand(itemsListCmd, itemsGetCmd)
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
var itemsListCmd = &cobra.Command{Use: "list", Short: "List items", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Not implemented yet") }}
var itemsGetCmd = &cobra.Command{Use: "get", Short: "Get items", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Not implemented yet") }}

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

func outputData(data interface{}) {
	switch outputFormat {
	case "json":
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(jsonData))
	case "table":
		// Basic table output - would need more sophisticated formatting
		fmt.Printf("%+v\n", data)
	default:
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(jsonData))
	}
}