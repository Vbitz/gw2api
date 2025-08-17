# Guild Wars 2 API Go Client

A fully typed, comprehensive Go client library for the Guild Wars 2 API v2.

## Features

✅ **Full Type Safety** - No more `interface{}`, all responses are strongly typed  
✅ **Comprehensive Coverage** - Support for all major public API endpoints  
✅ **Bulk Operations** - Efficient fetching of multiple items at once  
✅ **Pagination Support** - Built-in pagination with metadata  
✅ **Localization** - Support for all 5 languages (EN, ES, DE, FR, ZH)  
✅ **Context Support** - Timeout and cancellation support  
✅ **Generic Architecture** - Leverages Go generics for clean, reusable code  
✅ **Error Handling** - Proper error types and handling  

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    // Create client with options
    client := NewClient(
        WithTimeout(10*time.Second),
        WithLanguage(LanguageEnglish),
    )
    
    ctx := context.Background()
    
    // Get current build
    build, err := client.GetBuild(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Current build: %d\n", build.ID)
    
    // Get achievement with full type safety
    achievement, err := client.GetAchievement(ctx, 1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Achievement: %s (%d tiers)\n", achievement.Name, len(achievement.Tiers))
    
    // Get multiple items efficiently
    items, err := client.GetItems(ctx, []int{100, 200, 300})
    if err != nil {
        panic(err)
    }
    for _, item := range items {
        fmt.Printf("Item: %s (Level %d %s)\n", item.Name, item.Level, item.Rarity)
    }
}
```

## API Coverage

### Core Data
- ✅ **Build** - Current game build information
- ✅ **Currencies** - All in-game currencies
- ✅ **Worlds** - Game world/server information
- ✅ **Maps** - Map details and metadata
- ✅ **Files** - Game asset references

### Items & Equipment
- ✅ **Items** - All items with full details and stats
- ✅ **Item Stats** - Stat combination definitions
- ✅ **Colors** - Dye colors with material information
- ✅ **Materials** - Crafting material categories

### Character Progression
- ✅ **Achievements** - All achievements with tiers and rewards
- ✅ **Achievement Categories** - Achievement organization
- ✅ **Achievement Groups** - Achievement groupings
- ✅ **Skills** - All skills with facts and details
- ✅ **Specializations** - Elite and core specializations
- ✅ **Traits** - All traits with effects

### Trading Post
- ✅ **Commerce Prices** - Buy/sell price information
- ✅ **Commerce Listings** - Detailed trading post listings
- ✅ **Commerce Exchange** - Gem/gold exchange rates

## Request Patterns

### Single Item
```go
achievement, err := client.GetAchievement(ctx, 1)
item, err := client.GetItem(ctx, 12345)
world, err := client.GetWorld(ctx, 1001)
```

### Multiple Items (Bulk)
```go
// Efficient bulk fetching
achievements, err := client.GetAchievements(ctx, []int{1, 2, 3, 4, 5})
items, err := client.GetItems(ctx, []int{100, 200, 300})
worlds, err := client.GetWorlds(ctx, []int{1001, 1002, 1003})
```

### ID Lists
```go
// Get all available IDs
achievementIDs, err := client.GetAchievementIDs(ctx)
itemIDs, err := client.GetItemIDs(ctx)
worldIDs, err := client.GetWorldIDs(ctx)
```

### Pagination
```go
// Get first page of items (50 per page)
items, pagination, err := client.GetItemsPage(ctx, 
    WithPage(0), 
    WithPageSize(50))

fmt.Printf("Page %d of %d (Total: %d items)\n", 
    pagination.Page+1, pagination.PageTotal, pagination.Total)
```

### Localization
```go
// Get achievement in German
achievement, err := client.GetAchievement(ctx, 1, WithLang(LanguageGerman))

// Set default language for client
client := NewClient(WithLanguage(LanguageFrench))
```

## Type Safety Examples

### Before (interface{})
```go
// Old way - no type safety
response, err := client.GetAchievement(1)
data := response.(map[string]interface{})
name := data["name"].(string)  // Runtime panic risk!
tiers := data["tiers"].([]interface{})  // More casting needed
```

### After (Typed)
```go
// New way - full type safety
achievement, err := client.GetAchievement(ctx, 1)
name := achievement.Name  // String guaranteed by compiler
for _, tier := range achievement.Tiers {
    fmt.Printf("Tier: %d points for %d count\n", tier.Points, tier.Count)
}
```

## Error Handling

The client provides comprehensive error handling:

```go
achievement, err := client.GetAchievement(ctx, 99999)
if err != nil {
    // Check for API errors
    if apiErr, ok := err.(APIError); ok {
        fmt.Printf("API Error: %s\n", apiErr.Text)
    } else {
        fmt.Printf("Network/other error: %v\n", err)
    }
}
```

## Performance Features

### Bulk Operations
```go
// Instead of 100 individual requests
items := make([]*Item, 100)
for i, id := range itemIDs[:100] {
    item, err := client.GetItem(ctx, id)  // DON'T DO THIS
    // ...
}

// Do this - single request for all items
items, err := client.GetItems(ctx, itemIDs[:100])  // Much faster!
```

### Context Support
```go
// Timeout support
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

items, err := client.GetItems(ctx, []int{1, 2, 3})
// Request will be cancelled after 5 seconds
```

## Complete Type Definitions

All API responses are fully typed with comprehensive struct definitions:

- **Achievement** - ID, Name, Description, Type, Flags, Tiers, Rewards, Prerequisites
- **Item** - ID, Name, Type, Rarity, Level, Details (polymorphic based on item type)
- **Currency** - ID, Name, Description, Icon, Order
- **World** - ID, Name, Population
- **Map** - ID, Name, Levels, Type, Regions, Continent information
- **Skill** - ID, Name, Type, Professions, Facts, Traited Facts
- **Price** - Item ID, Buy/Sell prices and quantities
- And many more...

## Error Types

- **APIError** - Server-side API errors with descriptive messages
- **Network errors** - Connection, timeout, and other network issues
- **JSON parsing errors** - Malformed response handling

## Language Support

```go
const (
    LanguageEnglish Language = "en"
    LanguageSpanish Language = "es"
    LanguageGerman  Language = "de"
    LanguageFrench  Language = "fr"
    LanguageChinese Language = "zh"
)
```

## Running the Example

```bash
go run *.go
```

This will run a comprehensive test of all major functionality, demonstrating:
- Type safety in action
- Bulk operations
- Error handling
- Multiple data types

## Future Enhancements

- Authentication support for account-specific endpoints
- Response caching for better performance
- Rate limiting awareness
- Additional endpoints (WvW, PvP, Guild details)
- Webhook/websocket support for real-time data

## License

This library is designed for Guild Wars 2 API access and follows ArenaNet's API Terms of Service.