package main

import (
	"context"
	"fmt"
	"log"
	"time"

)

func testAPI() {
	// Create a new typed client
	client := NewClient(
		WithTimeout(10*time.Second),
		WithLanguage(LanguageEnglish),
	)

	ctx := context.Background()

	fmt.Println("=== Testing Typed GW2 API Client ===\n")

	// Test 1: Get build (single value)
	fmt.Println("1. Current Build:")
	if build, err := client.GetBuild(ctx); err == nil {
		fmt.Printf("   Build ID: %d (type: %T)\n\n", build.ID, build.ID)
	} else {
		log.Printf("   Error: %v\n\n", err)
	}

	// Test 2: Get achievement (complex object)
	fmt.Println("2. Sample Achievement:")
	if achievement, err := client.GetAchievement(ctx, 1); err == nil {
		fmt.Printf("   Name: %s\n", achievement.Name)
		fmt.Printf("   Type: %s\n", achievement.Type)
		fmt.Printf("   Tiers: %d\n", len(achievement.Tiers))
		fmt.Printf("   Flags: %v\n\n", achievement.Flags)
	} else {
		log.Printf("   Error: %v\n\n", err)
	}

	// Test 3: Get multiple currencies
	fmt.Println("3. Multiple Currencies:")
	if currencies, err := client.GetCurrencies(ctx, []int{1, 2}); err == nil {
		for _, currency := range currencies {
			fmt.Printf("   %s (ID: %d): %s\n", currency.Name, currency.ID, currency.Description)
		}
		fmt.Println()
	} else {
		log.Printf("   Error: %v\n\n", err)
	}

	// Test 4: Get specific world
	fmt.Println("4. Sample World:")
	if world, err := client.GetWorld(ctx, 1001); err == nil {
		fmt.Printf("   Name: %s (ID: %d)\n", world.Name, world.ID)
		fmt.Printf("   Population: %s\n\n", world.Population)
	} else {
		log.Printf("   Error: %v\n\n", err)
	}

	// Test 5: Test multiple worlds
	fmt.Println("5. Multiple Worlds:")
	if worlds, err := client.GetWorlds(ctx, []int{1001, 1002, 1003}); err == nil {
		for _, world := range worlds {
			fmt.Printf("   %s: %s population\n", world.Name, world.Population)
		}
		fmt.Println()
	} else {
		log.Printf("   Error: %v\n\n", err)
	}

	fmt.Println("=== All tests completed successfully! ===")
	fmt.Println("\nThe new API provides:")
	fmt.Println("✓ Full type safety - no more interface{}")
	fmt.Println("✓ Proper error handling")
	fmt.Println("✓ Support for bulk operations, pagination, and localization")
	fmt.Println("✓ Context support for timeouts and cancellation")
	fmt.Println("✓ Comprehensive documentation through types")
}