package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"j5.nz/gw2/internal/gw2api"
)

// handleHome renders the main page
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "GW2 Items & Crafting",
	}
	
	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.Render(w, "index", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleItemSearch handles HTMX item search
func (s *Server) handleItemSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.FormValue("query"))
	if query == "" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<div id="search-results" class="mt-4"></div>`)
		return
	}

	// Search items using cache
	items, err := s.searchItems(r.Context(), query)
	if err != nil {
		http.Error(w, "Search error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get prices for the first 10 items only (for performance)
	itemsWithPrices := s.addPricesToItems(r.Context(), items, 10)

	data := ItemSearchData{
		Query: query,
		Items: itemsWithPrices,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.Render(w, "item_results", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleItemPage renders a full page for an item
func (s *Server) handleItemPage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	itemID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Get item details
	items, err := s.client.GetItems(r.Context(), []int{itemID})
	if err != nil || len(items) == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	item := items[0]

	// Get price
	price, hasPrice := s.getItemPrice(r.Context(), itemID)

	// Get recipes that create this item
	recipes, _ := s.getRecipesForItem(r.Context(), itemID)

	data := PageData{
		Title: item.Name + " - GW2 Items & Crafting",
		Content: ItemDetailData{
			Item:     item,
			Price:    price,
			HasPrice: hasPrice,
			Recipes:  recipes,
		},
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.Render(w, "item_page", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleItemDetail shows detailed item information
func (s *Server) handleItemDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	itemID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Get item details
	items, err := s.client.GetItems(r.Context(), []int{itemID})
	if err != nil || len(items) == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	item := items[0]

	// Get price
	price, hasPrice := s.getItemPrice(r.Context(), itemID)

	// Get recipes that create this item
	recipes, _ := s.getRecipesForItem(r.Context(), itemID)

	data := ItemDetailData{
		Item:     item,
		Price:    price,
		HasPrice: hasPrice,
		Recipes:  recipes,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.Render(w, "item_detail", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleRecipeDetail shows recipe breakdown with costs
func (s *Server) handleRecipeDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	recipeID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	// Get recipe details
	recipes, err := s.client.GetRecipes(r.Context(), []int{recipeID})
	if err != nil || len(recipes) == 0 {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}
	recipe := recipes[0]

	// Get output item
	outputItems, err := s.client.GetItems(r.Context(), []int{recipe.OutputItemID})
	if err != nil || len(outputItems) == 0 {
		http.Error(w, "Output item not found", http.StatusNotFound)
		return
	}
	outputItem := outputItems[0]

	// Get ingredient items and prices
	ingredients := s.buildIngredientsWithCosts(r.Context(), recipe.Ingredients)

	// Calculate total cost
	totalCost := 0
	for _, ing := range ingredients {
		totalCost += ing.Cost
	}

	data := struct {
		PageData
		RecipeDetailData
	}{
		PageData: PageData{
			Title: "Recipe: " + outputItem.Name,
		},
		RecipeDetailData: RecipeDetailData{
			Recipe:      recipe,
			OutputItem:  outputItem,
			Ingredients: ingredients,
			TotalCost:   totalCost,
		},
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.Render(w, "recipe_page", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleCraftingTree shows recursive crafting dependencies and costs
func (s *Server) handleCraftingTree(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	recipeID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	// Get recipe details
	recipes, err := s.client.GetRecipes(r.Context(), []int{recipeID})
	if err != nil || len(recipes) == 0 {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}
	recipe := recipes[0]

	// Get output item
	outputItems, err := s.client.GetItems(r.Context(), []int{recipe.OutputItemID})
	if err != nil || len(outputItems) == 0 {
		http.Error(w, "Output item not found", http.StatusNotFound)
		return
	}
	outputItem := outputItems[0]

	// Build the complete crafting tree
	craftingData := s.buildCraftingTree(r.Context(), recipe, outputItem, 1)

	data := PageData{
		Title:   "Crafting Tree: " + outputItem.Name,
		Content: craftingData,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.Render(w, "crafting_tree", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleCraftingNode returns a single crafting node with its immediate children (HTMX endpoint)
func (s *Server) handleCraftingNode(w http.ResponseWriter, r *http.Request) {
	recipeIDStr := r.PathValue("recipeId")
	itemIDStr := r.PathValue("itemId")
	quantityStr := r.PathValue("quantity")

	recipeID, err := strconv.Atoi(recipeIDStr)
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	// Get recipe and item details
	var recipe *gw2api.RecipeDetail
	if recipeID > 0 {
		recipes, err := s.client.GetRecipes(r.Context(), []int{recipeID})
		if err == nil && len(recipes) > 0 {
			recipe = recipes[0]
		}
	}

	items, err := s.client.GetItems(r.Context(), []int{itemID})
	if err != nil || len(items) == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	item := items[0]

	// Build single node with immediate children only (depth 1)
	cache := NewRequestCache()
	node := s.buildSingleCraftingNode(r.Context(), cache, recipe, item, quantity, 0, 1)

	data := struct {
		Node *CraftingNode
		Level int
	}{
		Node: node,
		Level: 0,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.Render(w, "crafting_node_partial", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleCraftingNodeExpand expands a specific node to show its children (HTMX endpoint)
func (s *Server) handleCraftingNodeExpand(w http.ResponseWriter, r *http.Request) {
	recipeIDStr := r.PathValue("recipeId")
	itemIDStr := r.PathValue("itemId")
	quantityStr := r.PathValue("quantity")
	levelStr := r.PathValue("level")

	recipeID, err := strconv.Atoi(recipeIDStr)
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	level, err := strconv.Atoi(levelStr)
	if err != nil {
		http.Error(w, "Invalid level", http.StatusBadRequest)
		return
	}

	// Get recipe details
	var recipe *gw2api.RecipeDetail
	if recipeID > 0 {
		recipes, err := s.client.GetRecipes(r.Context(), []int{recipeID})
		if err == nil && len(recipes) > 0 {
			recipe = recipes[0]
		}
	}

	items, err := s.client.GetItems(r.Context(), []int{itemID})
	if err != nil || len(items) == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	item := items[0]

	// Check if this is a collapse request (level=1 means collapse to just expand button)
	if level == 1 {
		// Build minimal node to get children count
		cache := NewRequestCache()
		node := s.buildSingleCraftingNode(r.Context(), cache, recipe, item, quantity, level, 1)
		
		data := struct {
			ParentItemID     int
			ParentRecipeID   int  
			ParentQuantity   int
			ChildrenCount    int
		}{
			ParentItemID:   itemID,
			ParentRecipeID: recipeID,
			ParentQuantity: quantity,
			ChildrenCount:  len(node.Children),
		}
		
		w.Header().Set("Content-Type", "text/html")
		if err := s.templates.Render(w, "crafting_expand_button", data); err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// Build children nodes for expansion (allow deeper recursion for proper recipe discovery)
	cache := NewRequestCache()
	node := s.buildSingleCraftingNode(r.Context(), cache, recipe, item, quantity, level, 4)

	data := struct {
		Children       []*CraftingNode
		Level          int
		ParentItemID   int
		ParentRecipeID int
		ParentQuantity int
	}{
		Children:       node.Children,
		Level:          level + 1,
		ParentItemID:   itemID,
		ParentRecipeID: recipeID,
		ParentQuantity: quantity,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.Render(w, "crafting_children_partial", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleCraftingSummary returns just the cost analysis and base materials (HTMX endpoint)
func (s *Server) handleCraftingSummary(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	recipeID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	// Get recipe details
	recipes, err := s.client.GetRecipes(r.Context(), []int{recipeID})
	if err != nil || len(recipes) == 0 {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}
	recipe := recipes[0]

	// Get output item
	outputItems, err := s.client.GetItems(r.Context(), []int{recipe.OutputItemID})
	if err != nil || len(outputItems) == 0 {
		http.Error(w, "Output item not found", http.StatusNotFound)
		return
	}
	outputItem := outputItems[0]

	// Build the complete crafting tree for cost analysis only
	craftingData := s.buildCraftingTree(r.Context(), recipe, outputItem, 1)

	data := struct {
		CraftingData *CraftingTreeData
	}{
		CraftingData: craftingData,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.Render(w, "crafting_summary_partial", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleSetAPIKey stores API key in session
func (s *Server) handleSetAPIKey(w http.ResponseWriter, r *http.Request) {
	apiKey := strings.TrimSpace(r.FormValue("api_key"))
	
	// TODO: Implement session storage for API key
	// For now, just validate the key works
	if apiKey != "" {
		client := gw2api.NewClient(gw2api.WithAPIKey(apiKey))
		_, err := client.GetAccount(r.Context())
		if err != nil {
			http.Error(w, "Invalid API key", http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<div class="text-green-600">API key set successfully!</div>`)
}

// Helper functions

// searchItems searches for items using the client's cache
func (s *Server) searchItems(ctx context.Context, query string) ([]*gw2api.Item, error) {
	// Use the client's SearchItems method which uses the cache
	options := gw2api.ItemSearchOptions{
		Name:  query,
		Limit: 20, // Limit to 20 results for better performance
	}
	
	return s.client.SearchItems(ctx, options)
}

// addPricesToItems fetches prices for items with optional limit
func (s *Server) addPricesToItems(ctx context.Context, items []*gw2api.Item, limit int) []*ItemWithPrice {
	results := make([]*ItemWithPrice, len(items))
	
	// Determine how many items to fetch prices for
	priceLimit := len(items)
	if limit > 0 && limit < priceLimit {
		priceLimit = limit
	}
	
	// Collect item IDs for batch price fetching
	var itemIDs []int
	for i := 0; i < priceLimit; i++ {
		itemIDs = append(itemIDs, items[i].ID)
	}
	
	// Batch fetch prices
	priceMap := s.batchGetPrices(ctx, itemIDs)
	
	// Build results
	for i, item := range items {
		if price, hasPrice := priceMap[item.ID]; hasPrice && i < priceLimit {
			results[i] = &ItemWithPrice{
				Item:     item,
				Price:    price,
				HasPrice: true,
			}
		} else {
			results[i] = &ItemWithPrice{
				Item:     item,
				Price:    nil,
				HasPrice: false,
			}
		}
	}
	
	return results
}

// getItemPrice gets item price (cached or fresh)
func (s *Server) getItemPrice(ctx context.Context, itemID int) (*gw2api.Price, bool) {
	// Check cache first
	priceCache := &PriceCache{cache: s.priceCache}
	if price, found := priceCache.GetPrice(itemID); found {
		return price, true
	}

	// Fetch from API
	prices, err := s.client.GetCommercePrices(ctx, []int{itemID})
	if err != nil || len(prices) == 0 {
		return nil, false
	}

	price := prices[0]
	priceCache.SetPrice(itemID, price)
	return price, true
}

// batchGetPrices fetches multiple item prices with caching
func (s *Server) batchGetPrices(ctx context.Context, itemIDs []int) map[int]*gw2api.Price {
	priceCache := &PriceCache{cache: s.priceCache}
	result := make(map[int]*gw2api.Price)
	var uncachedIDs []int
	
	// Check cache first
	for _, itemID := range itemIDs {
		if price, found := priceCache.GetPrice(itemID); found {
			result[itemID] = price
		} else {
			uncachedIDs = append(uncachedIDs, itemID)
		}
	}
	
	// Fetch uncached prices in batch
	if len(uncachedIDs) > 0 {
		prices, err := s.client.GetCommercePrices(ctx, uncachedIDs)
		if err == nil {
			for _, price := range prices {
				if price != nil {
					result[price.ID] = price
					priceCache.SetPrice(price.ID, price)
				}
			}
		}
	}
	
	return result
}

// getRecipesForItem finds both recipes that create the item and recipes that use the item
func (s *Server) getRecipesForItem(ctx context.Context, itemID int) (*ItemRecipes, error) {
	result := &ItemRecipes{
		CreatesItem: []*RecipeWithOutput{},
		UsesItem:    []*RecipeWithOutput{},
	}

	// Get recipes that create this item
	createRecipeIDs, err := s.searchRecipesByOutput(ctx, itemID)
	if err == nil && len(createRecipeIDs) > 0 {
		// Limit to first 5 for performance
		if len(createRecipeIDs) > 5 {
			createRecipeIDs = createRecipeIDs[:5]
		}
		if createRecipes, err := s.client.GetRecipes(ctx, createRecipeIDs); err == nil {
			result.CreatesItem = s.enrichRecipesWithOutputItems(ctx, createRecipes)
		}
	}

	// Get recipes that use this item as ingredient
	useRecipeIDs, err := s.searchRecipesByInput(ctx, itemID)
	if err == nil && len(useRecipeIDs) > 0 {
		// Limit to first 10 for performance
		if len(useRecipeIDs) > 10 {
			useRecipeIDs = useRecipeIDs[:10]
		}
		if useRecipes, err := s.client.GetRecipes(ctx, useRecipeIDs); err == nil {
			result.UsesItem = s.enrichRecipesWithOutputItems(ctx, useRecipes)
		}
	}

	return result, nil
}

// enrichRecipesWithOutputItems adds output item details to recipes
func (s *Server) enrichRecipesWithOutputItems(ctx context.Context, recipes []*gw2api.RecipeDetail) []*RecipeWithOutput {
	if len(recipes) == 0 {
		return []*RecipeWithOutput{}
	}

	// Get all unique output item IDs
	itemIDs := make([]int, 0, len(recipes))
	itemIDSet := make(map[int]bool)
	for _, recipe := range recipes {
		if !itemIDSet[recipe.OutputItemID] {
			itemIDs = append(itemIDs, recipe.OutputItemID)
			itemIDSet[recipe.OutputItemID] = true
		}
	}

	// Fetch all output items in one API call
	items, err := s.client.GetItems(ctx, itemIDs)
	if err != nil {
		// If we can't get items, return recipes without output items
		result := make([]*RecipeWithOutput, len(recipes))
		for i, recipe := range recipes {
			result[i] = &RecipeWithOutput{
				Recipe:     recipe,
				OutputItem: nil,
			}
		}
		return result
	}

	// Create a map for quick item lookup
	itemMap := make(map[int]*gw2api.Item)
	for _, item := range items {
		itemMap[item.ID] = item
	}

	// Combine recipes with their output items
	result := make([]*RecipeWithOutput, len(recipes))
	for i, recipe := range recipes {
		result[i] = &RecipeWithOutput{
			Recipe:     recipe,
			OutputItem: itemMap[recipe.OutputItemID],
		}
	}

	return result
}

// getRecipesForItemLegacy finds recipes that create the given item (legacy compatibility)
func (s *Server) getRecipesForItemLegacy(ctx context.Context, itemID int) ([]*gw2api.RecipeDetail, error) {
	recipes, err := s.getRecipesForItem(ctx, itemID)
	if err != nil {
		return []*gw2api.RecipeDetail{}, err
	}
	
	// Extract just the recipe details for legacy compatibility
	result := make([]*gw2api.RecipeDetail, len(recipes.CreatesItem))
	for i, recipeWithOutput := range recipes.CreatesItem {
		result[i] = recipeWithOutput.Recipe
	}
	return result, nil
}

// searchRecipesByOutput searches for recipes that create a specific item using cache or API  
func (s *Server) searchRecipesByOutput(ctx context.Context, itemID int) ([]int, error) {
	// Try to search cached recipes first
	if s.client.DataCache() != nil && s.client.DataCache().GetRecipeCache().IsLoaded() {
		recipeIDs := s.client.DataCache().GetRecipeCache().SearchByOutput(itemID)
		if len(recipeIDs) > 0 {
			// Limit to first 5 for performance
			if len(recipeIDs) > 5 {
				recipeIDs = recipeIDs[:5]
			}
			return recipeIDs, nil
		}
	}
	
	// Fallback to API search if cache not available or no results
	return s.searchAPIRecipesByOutput(ctx, itemID)
}

// searchRecipesByInput searches for recipes that use a specific item as ingredient
func (s *Server) searchRecipesByInput(ctx context.Context, itemID int) ([]int, error) {
	// Try to search cached recipes first
	if s.client.DataCache() != nil && s.client.DataCache().GetRecipeCache().IsLoaded() {
		recipeIDs := s.client.DataCache().GetRecipeCache().SearchByInput(itemID)
		if len(recipeIDs) > 0 {
			// Limit to first 10 for performance
			if len(recipeIDs) > 10 {
				recipeIDs = recipeIDs[:10]
			}
			return recipeIDs, nil
		}
	}
	
	// Fallback to API search if cache not available or no results
	return s.searchAPIRecipesByInput(ctx, itemID)
}

// searchAPIRecipesByInput searches for recipes using the API (fallback)
func (s *Server) searchAPIRecipesByInput(ctx context.Context, itemID int) ([]int, error) {
	url := fmt.Sprintf("https://api.guildwars2.com/v2/recipes/search?input=%d", itemID)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		return []int{}, nil
	}
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var recipeIDs []int
	if err := json.Unmarshal(body, &recipeIDs); err != nil {
		return nil, err
	}

	return recipeIDs, nil
}

// searchAPIRecipesByOutput searches for recipes using the API (fallback)
func (s *Server) searchAPIRecipesByOutput(ctx context.Context, itemID int) ([]int, error) {
	url := fmt.Sprintf("https://api.guildwars2.com/v2/recipes/search?output=%d", itemID)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		return []int{}, nil
	}
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var recipeIDs []int
	if err := json.Unmarshal(body, &recipeIDs); err != nil {
		return nil, err
	}

	return recipeIDs, nil
}

// buildIngredientsWithCosts builds ingredient list with items and costs
func (s *Server) buildIngredientsWithCosts(ctx context.Context, ingredients []gw2api.RecipeIngredient) []*IngredientWithItem {
	results := make([]*IngredientWithItem, len(ingredients))
	
	// Get all ingredient item IDs
	itemIDs := make([]int, len(ingredients))
	for i, ing := range ingredients {
		itemIDs[i] = ing.ItemID
	}

	// Fetch items
	items, err := s.client.GetItems(ctx, itemIDs)
	if err != nil {
		return results
	}

	// Build results with prices
	for i, ing := range ingredients {
		item := items[i]
		price, hasPrice := s.getItemPrice(ctx, ing.ItemID)
		
		cost := 0
		if hasPrice && price.Sells.UnitPrice > 0 {
			cost = ing.Count * price.Sells.UnitPrice
		}

		results[i] = &IngredientWithItem{
			RecipeIngredient: &ing,
			Item:             item,
			Price:            price,
			HasPrice:         hasPrice,
			Cost:             cost,
		}
	}

	return results
}

// CharacterWithDetails represents a character with core information
type CharacterWithDetails struct {
	Name       string
	Level      int
	Profession string
	Race       string
	Guild      string
}

// handleInventoryPage renders the inventory page with character names only
func (s *Server) handleInventoryPage(w http.ResponseWriter, r *http.Request) {
	// Get character names only (single API call)
	characterNames, err := s.client.GetCharacterNames(r.Context())
	if err != nil {
		// If API fails, show page with error message
		data := struct {
			PageData
			Error      string
			Characters []string
		}{
			PageData:   PageData{Title: "Character Inventory"},
			Error:      "Failed to load characters: " + err.Error(),
			Characters: []string{},
		}
		
		if err := s.templates.Render(w, "inventory", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	data := struct {
		PageData
		Characters []string
		Error      string
	}{
		PageData:   PageData{Title: "Character Inventory"},
		Characters: characterNames,
		Error:      "",
	}
	
	if err := s.templates.Render(w, "inventory", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleCharacters returns character list as HTMX response
func (s *Server) handleCharacters(w http.ResponseWriter, r *http.Request) {
	characterNames, err := s.client.GetCharacterNames(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch characters: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := struct {
		Characters []string
	}{
		Characters: characterNames,
	}
	
	if err := s.templates.Render(w, "character_list", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// InventoryItem represents an inventory item with details
type InventoryItem struct {
	Item     *gw2api.Item
	Count    int
	Binding  string
	BoundTo  string
	BagIndex int
	SlotIndex int
}

// handleCharacterInventory renders character details and inventory page  
func (s *Server) handleCharacterInventory(w http.ResponseWriter, r *http.Request) {
	characterName := r.PathValue("character")
	if characterName == "" {
		http.Error(w, "Character name required", http.StatusBadRequest)
		return
	}
	
	// Get character inventory
	inventory, err := s.client.GetCharacterInventory(r.Context(), characterName)
	if err != nil {
		http.Error(w, "Failed to fetch character inventory: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Collect all unique item IDs for batch fetching
	var itemIDs []int
	itemIDSet := make(map[int]bool)
	
	for _, bag := range inventory.Bags {
		for _, slot := range bag.Inventory {
			if slot.ID != 0 && !itemIDSet[slot.ID] {
				itemIDs = append(itemIDs, slot.ID)
				itemIDSet[slot.ID] = true
			}
		}
	}
	
	// Fetch item details in batch
	var itemDetails []*gw2api.Item
	if len(itemIDs) > 0 {
		itemDetails, err = s.client.GetItems(r.Context(), itemIDs)
		if err != nil {
			http.Error(w, "Failed to fetch item details: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	// Create item lookup map
	itemMap := make(map[int]*gw2api.Item)
	for _, item := range itemDetails {
		itemMap[item.ID] = item
	}
	
	// Build inventory items list
	var inventoryItems []InventoryItem
	
	for bagIndex, bag := range inventory.Bags {
		for slotIndex, slot := range bag.Inventory {
			if slot.ID != 0 {
				if item, found := itemMap[slot.ID]; found {
					inventoryItems = append(inventoryItems, InventoryItem{
						Item:      item,
						Count:     slot.Count,
						Binding:   slot.Binding,
						BoundTo:   slot.BoundTo,
						BagIndex:  bagIndex,
						SlotIndex: slotIndex,
					})
				}
			}
		}
	}
	
	// Get character details
	core, err := s.client.GetCharacterCore(r.Context(), characterName)
	if err != nil {
		http.Error(w, "Failed to fetch character details: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Render character detail page
	data := struct {
		PageData
		Character CharacterWithDetails
		Items     []InventoryItem
	}{
		PageData: PageData{Title: characterName + " - Character Details"},
		Character: CharacterWithDetails{
			Name:       core.Name,
			Level:      core.Level,
			Profession: core.Profession,
			Race:       core.Race,
			Guild:      core.Guild,
		},
		Items: inventoryItems,
	}
	
	if err := s.templates.Render(w, "character_detail", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleAccountPage shows account overview with characters and navigation links
func (s *Server) handleAccountPage(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		data := PageData{
			Title:   "My Account",
			Content: map[string]interface{}{"Error": "API key not configured"},
		}
		w.Header().Set("Content-Type", "text/html")
		if err := s.templates.Render(w, "account", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get character names
	characterNames, err := s.client.GetCharacterNames(r.Context())
	if err != nil {
		data := PageData{
			Title:   "My Account",
			Content: map[string]interface{}{"Error": err.Error()},
		}
		w.Header().Set("Content-Type", "text/html")
		if err := s.templates.Render(w, "account", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	data := PageData{
		Title: "My Account",
		Content: map[string]interface{}{
			"Characters": characterNames,
		},
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.Render(w, "account", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleBankPage shows account bank
func (s *Server) handleBankPage(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		data := PageData{
			Title:   "Bank",
			Content: map[string]interface{}{"Error": "API key not configured"},
		}
		w.Header().Set("Content-Type", "text/html")
		if err := s.templates.Render(w, "bank", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get bank items
	bankItems, err := s.client.GetAccountBank(r.Context())
	if err != nil {
		data := PageData{
			Title:   "Bank",
			Content: map[string]interface{}{"Error": err.Error()},
		}
		w.Header().Set("Content-Type", "text/html")
		if err := s.templates.Render(w, "bank", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Convert to inventory items and get item details
	var inventoryItems []InventoryItem
	var itemIDs []int
	
	for slot, bankItem := range bankItems {
		if bankItem.ID != 0 {
			itemIDs = append(itemIDs, bankItem.ID)
			inventoryItems = append(inventoryItems, InventoryItem{
				Count:     bankItem.Count,
				BagIndex:  0,
				SlotIndex: slot,
			})
		}
	}

	// Get item details
	if len(itemIDs) > 0 {
		items, err := s.client.GetItems(r.Context(), itemIDs)
		if err == nil {
			for i, item := range items {
				inventoryItems[i].Item = item
			}
		}
	}

	data := PageData{
		Title: "Bank",
		Content: map[string]interface{}{
			"Items": inventoryItems,
		},
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.Render(w, "bank", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleSharedInventoryPage shows shared inventory slots
func (s *Server) handleSharedInventoryPage(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		data := PageData{
			Title:   "Shared Inventory",
			Content: map[string]interface{}{"Error": "API key not configured"},
		}
		w.Header().Set("Content-Type", "text/html")
		if err := s.templates.Render(w, "shared", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get shared inventory items  
	sharedItems, err := s.client.GetAccountInventory(r.Context())
	if err != nil {
		data := PageData{
			Title:   "Shared Inventory",
			Content: map[string]interface{}{"Error": err.Error()},
		}
		w.Header().Set("Content-Type", "text/html")
		if err := s.templates.Render(w, "shared", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Convert to inventory items and get item details
	var inventoryItems []InventoryItem
	var itemIDs []int
	
	for slot, sharedItem := range sharedItems {
		if sharedItem.ID != 0 {
			itemIDs = append(itemIDs, sharedItem.ID)
			inventoryItems = append(inventoryItems, InventoryItem{
				Count:     sharedItem.Count,
				Binding:   sharedItem.Binding,
				BoundTo:   sharedItem.BoundTo,
				BagIndex:  0,
				SlotIndex: slot,
			})
		}
	}

	// Get item details
	if len(itemIDs) > 0 {
		items, err := s.client.GetItems(r.Context(), itemIDs)
		if err == nil {
			for i, item := range items {
				inventoryItems[i].Item = item
			}
		}
	}

	data := PageData{
		Title: "Shared Inventory",
		Content: map[string]interface{}{
			"Items": inventoryItems,
		},
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.Render(w, "shared", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// buildCraftingTree builds a crafting dependency tree with optimized batched requests
func (s *Server) buildCraftingTree(ctx context.Context, recipe *gw2api.RecipeDetail, item *gw2api.Item, quantity int) *CraftingTreeData {
	cache := NewRequestCache()
	const maxDepth = 8 // Allow deeper recursion to reach base materials
	
	// Phase 1: Collect all required data IDs through tree traversal
	requiredItems := make(map[int]bool)
	requiredRecipes := make(map[int]bool) 
	requiredPrices := make(map[int]bool)
	
	s.collectRequiredData(recipe, item, quantity, 0, maxDepth, requiredItems, requiredRecipes, requiredPrices, make(map[int]bool))
	
	// Phase 2: Batch fetch all required data
	s.batchFetchData(ctx, cache, requiredItems, requiredRecipes, requiredPrices)
	
	// Phase 3: Build the complete tree from cached data
	rootNode := s.buildOptimizedCraftingNode(cache, recipe, item, quantity, 0, maxDepth, make(map[int]bool))
	
	// Collect base materials
	baseMaterials := make(map[int]*MaterialSummary)
	s.collectBaseMaterials(rootNode, baseMaterials)
	
	// Convert map to slice
	materialsList := make([]*MaterialSummary, 0, len(baseMaterials))
	for _, material := range baseMaterials {
		materialsList = append(materialsList, material)
	}
	
	// Calculate totals
	totalCraftCost := rootNode.TotalCost
	totalBuyCost := 0
	if price, ok := cache.prices[item.ID]; ok && price != nil {
		totalBuyCost = quantity * price.Sells.UnitPrice
	}
	savings := totalBuyCost - totalCraftCost
	extraCost := 0
	isCraftingCheaper := savings >= 0
	if savings < 0 {
		extraCost = -savings
	}
	
	savingsPercent := 0.0
	if totalBuyCost > 0 {
		savingsPercent = float64(savings) / float64(totalBuyCost) * 100
	}
	
	return &CraftingTreeData{
		RootItem:          item,
		Recipe:            recipe,
		Tree:              rootNode,
		BaseMaterials:     materialsList,
		TotalCraftCost:    totalCraftCost,
		TotalBuyCost:      totalBuyCost,
		Savings:           savings,
		SavingsPercent:    savingsPercent,
		ExtraCost:         extraCost,
		IsCraftingCheaper: isCraftingCheaper,
	}
}

// Legacy buildCraftingNode - replaced with optimized version
// func (s *Server) buildCraftingNode(...) - REMOVED for performance

// collectBaseMaterials recursively collects all base materials needed
func (s *Server) collectBaseMaterials(node *CraftingNode, materials map[int]*MaterialSummary) {
	if !node.HasRecipe || !node.CanCraft {
		// This is a base material
		if existing, exists := materials[node.Item.ID]; exists {
			existing.TotalRequired += node.RequiredCount
			existing.TotalCost = existing.TotalRequired * existing.UnitPrice
		} else {
			materials[node.Item.ID] = &MaterialSummary{
				Item:          node.Item,
				TotalRequired: node.RequiredCount,
				UnitPrice:     node.BuyPrice,
				TotalCost:     node.RequiredCount * node.BuyPrice,
			}
		}
		return
	}
	
	// Recurse into children
	for _, child := range node.Children {
		s.collectBaseMaterials(child, materials)
	}
}

// collectRequiredData performs first pass to collect all IDs needed for the tree
func (s *Server) collectRequiredData(recipe *gw2api.RecipeDetail, item *gw2api.Item, quantity int, level int, maxDepth int, 
	requiredItems, requiredRecipes, requiredPrices map[int]bool, visited map[int]bool) {
	
	// Prevent infinite recursion and respect depth limits
	if visited[item.ID] || level >= maxDepth {
		requiredItems[item.ID] = true
		requiredPrices[item.ID] = true
		return
	}
	
	visited[item.ID] = true
	defer delete(visited, item.ID)
	
	// Always need the item and its price
	requiredItems[item.ID] = true
	requiredPrices[item.ID] = true
	
	if recipe == nil {
		return
	}
	
	// Need the recipe
	requiredRecipes[recipe.ID] = true
	
	// Collect ingredient requirements
	for _, ingredient := range recipe.Ingredients {
		requiredItems[ingredient.ItemID] = true
		requiredPrices[ingredient.ItemID] = true
		
		// Look for ALL recipes that create this ingredient (using cache first)
		if s.client.DataCache() != nil && s.client.DataCache().GetRecipeCache().IsLoaded() {
			recipeIDs := s.client.DataCache().GetRecipeCache().SearchByOutput(ingredient.ItemID)
			for _, recipeID := range recipeIDs {
				// Collect ALL recipes for this ingredient
				requiredRecipes[recipeID] = true
				if cachedRecipe, found := s.client.DataCache().GetRecipeCache().GetByID(recipeID); found && cachedRecipe != nil {
					// Recursively collect for each recipe found (use first one for dependency collection to avoid exponential explosion)
					if len(recipeIDs) == 1 || recipeID == recipeIDs[0] {
						s.collectRequiredData(cachedRecipe, &gw2api.Item{ID: ingredient.ItemID}, 
							ingredient.Count*quantity, level+1, maxDepth, requiredItems, requiredRecipes, requiredPrices, visited)
					}
				}
			}
		}
	}
}

// batchFetchData performs batch API calls to fetch all required data
func (s *Server) batchFetchData(ctx context.Context, cache *RequestCache, requiredItems, requiredRecipes, requiredPrices map[int]bool) {
	// Convert maps to slices
	itemIDs := make([]int, 0, len(requiredItems))
	for id := range requiredItems {
		itemIDs = append(itemIDs, id)
	}
	
	recipeIDs := make([]int, 0, len(requiredRecipes))
	for id := range requiredRecipes {
		recipeIDs = append(recipeIDs, id)
	}
	
	priceIDs := make([]int, 0, len(requiredPrices))
	for id := range requiredPrices {
		priceIDs = append(priceIDs, id)
	}
	
	// Batch fetch items
	if len(itemIDs) > 0 {
		if items, err := s.client.GetItems(ctx, itemIDs); err == nil {
			for _, item := range items {
				cache.items[item.ID] = item
			}
		}
	}
	
	// Batch fetch recipes
	if len(recipeIDs) > 0 {
		if recipes, err := s.client.GetRecipes(ctx, recipeIDs); err == nil {
			for _, recipe := range recipes {
				cache.recipes[recipe.ID] = recipe
			}
		}
	}
	
	// Batch fetch prices (in chunks due to API limits)
	if len(priceIDs) > 0 {
		chunkSize := 100 // API typically limits to ~200 IDs
		for i := 0; i < len(priceIDs); i += chunkSize {
			end := i + chunkSize
			if end > len(priceIDs) {
				end = len(priceIDs)
			}
			
			chunk := priceIDs[i:end]
			if prices, err := s.client.GetCommercePrices(ctx, chunk); err == nil {
				for _, price := range prices {
					cache.prices[price.ID] = price
				}
			}
		}
	}
}

// buildOptimizedCraftingNode builds nodes using cached data (no API calls)
func (s *Server) buildOptimizedCraftingNode(cache *RequestCache, recipe *gw2api.RecipeDetail, item *gw2api.Item, 
	quantity int, level int, maxDepth int, visited map[int]bool) *CraftingNode {
	
	// Prevent infinite recursion and respect depth limits
	if visited[item.ID] || level >= maxDepth {
		price := cache.prices[item.ID]
		unitCost := 0
		if price != nil {
			unitCost = price.Sells.UnitPrice
		}
		return &CraftingNode{
			Item:          item,
			Recipe:        nil,
			RequiredCount: quantity,
			UnitCost:      unitCost,
			TotalCost:     quantity * unitCost,
			HasRecipe:     false,
			CanCraft:      false,
			BuyPrice:      unitCost,
			TotalBuyCost:  quantity * unitCost,
			Children:      []*CraftingNode{},
			Level:         level,
		}
	}
	
	visited[item.ID] = true
	defer delete(visited, item.ID)
	
	node := &CraftingNode{
		Item:          item,
		Recipe:        recipe,
		RequiredCount: quantity,
		Level:         level,
		Children:      []*CraftingNode{},
	}
	
	// Get cached market price
	price := cache.prices[item.ID]
	buyPrice := 0
	if price != nil {
		buyPrice = price.Sells.UnitPrice
	}
	node.BuyPrice = buyPrice
	node.TotalBuyCost = quantity * buyPrice
	
	if recipe == nil {
		// No recipe - must buy
		node.HasRecipe = false
		node.CanCraft = false
		node.UnitCost = buyPrice
		node.TotalCost = quantity * buyPrice
		return node
	}
	
	node.HasRecipe = true
	totalCraftCost := 0
	
	// Build children for each ingredient using cached data
	for _, ingredient := range recipe.Ingredients {
		ingredientItem := cache.items[ingredient.ItemID]
		if ingredientItem == nil {
			// Fallback if item not cached
			ingredientItem = &gw2api.Item{ID: ingredient.ItemID, Name: fmt.Sprintf("Item %d", ingredient.ItemID)}
		}
		
		requiredCount := ingredient.Count * quantity
		
		// Find best recipe for this ingredient (cheapest to craft)
		var ingredientRecipe *gw2api.RecipeDetail
		if s.client.DataCache() != nil && s.client.DataCache().GetRecipeCache().IsLoaded() {
			recipeIDs := s.client.DataCache().GetRecipeCache().SearchByOutput(ingredient.ItemID)
			var bestRecipe *gw2api.RecipeDetail
			var bestCost int = -1
			
			for _, recipeID := range recipeIDs {
				if candidateRecipe := cache.recipes[recipeID]; candidateRecipe != nil {
					// Quick cost estimate for this recipe
					recipeCost := 0
					for _, subIngredient := range candidateRecipe.Ingredients {
						if subPrice := cache.prices[subIngredient.ItemID]; subPrice != nil {
							recipeCost += subIngredient.Count * subPrice.Sells.UnitPrice
						}
					}
					
					// Choose the cheapest recipe, or first one if costs are equal
					if bestCost == -1 || recipeCost < bestCost {
						bestRecipe = candidateRecipe
						bestCost = recipeCost
					}
				}
			}
			ingredientRecipe = bestRecipe
		}
		
		// Recursively build child node
		childNode := s.buildOptimizedCraftingNode(cache, ingredientRecipe, ingredientItem, requiredCount, level+1, maxDepth, visited)
		node.Children = append(node.Children, childNode)
		
		// For cost calculation, use the cheaper option (buy vs craft)
		childCostForTotal := childNode.TotalCost
		if childNode.HasRecipe && !childNode.CanCraft && childNode.BuyPrice > 0 {
			// If buying is cheaper, use buy cost for the total
			childCostForTotal = requiredCount * childNode.BuyPrice
		}
		totalCraftCost += childCostForTotal
	}
	
	// Always show crafting cost, but determine if it's economical
	craftCostPerItem := 0
	if quantity > 0 {
		craftCostPerItem = totalCraftCost / quantity
	}
	
	// Determine cost effectiveness (with 10% margin for cost stability)
	isCraftingCheaper := totalCraftCost < int(float64(quantity*buyPrice)*0.9) || buyPrice == 0
	
	// Always set crafting costs, but flag whether it's economical
	node.CanCraft = isCraftingCheaper
	node.UnitCost = craftCostPerItem
	node.TotalCost = totalCraftCost
	
	return node
}

// buildSingleCraftingNode builds a single node with limited depth for HTMX loading
func (s *Server) buildSingleCraftingNode(ctx context.Context, cache *RequestCache, recipe *gw2api.RecipeDetail, item *gw2api.Item, quantity int, level int, maxDepth int) *CraftingNode {
	// Collect required data for this single node
	requiredItems := map[int]bool{item.ID: true}
	requiredRecipes := make(map[int]bool)
	requiredPrices := map[int]bool{item.ID: true}

	if recipe != nil {
		requiredRecipes[recipe.ID] = true
		for _, ingredient := range recipe.Ingredients {
			requiredItems[ingredient.ItemID] = true
			requiredPrices[ingredient.ItemID] = true

			// Always collect child recipe data so we know if items are craftable
			if s.client.DataCache() != nil && s.client.DataCache().GetRecipeCache().IsLoaded() {
				recipeIDs := s.client.DataCache().GetRecipeCache().SearchByOutput(ingredient.ItemID)
				for _, recipeID := range recipeIDs {
					requiredRecipes[recipeID] = true
				}
			}
		}
	}

	// Batch fetch the required data
	s.batchFetchData(ctx, cache, requiredItems, requiredRecipes, requiredPrices)

	// Build the node using the cached data
	return s.buildOptimizedCraftingNode(cache, recipe, item, quantity, level, level+maxDepth, make(map[int]bool))
}