package web

import (
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"

	"j5.nz/gw2/internal/gw2api"
)

// Templates handles HTML template rendering
type Templates struct {
	templates map[string]*template.Template
}

// NewTemplates creates a new template handler
func NewTemplates() *Templates {
	t := &Templates{
		templates: make(map[string]*template.Template),
	}
	
	// Load templates
	t.loadTemplates()
	
	return t
}

func (t *Templates) loadTemplates() {
	// Template functions
	funcMap := template.FuncMap{
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"join": func(slice []string, sep string) string {
			return strings.Join(slice, sep)
		},
		"divide": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"multiply": func(a int, b float64) float64 {
			return float64(a) * b
		},
		"subtract": func(a, b interface{}) float64 {
			var aVal, bVal float64
			switch v := a.(type) {
			case int:
				aVal = float64(v)
			case float64:
				aVal = v
			}
			switch v := b.(type) {
			case int:
				bVal = float64(v)
			case float64:
				bVal = v
			}
			return aVal - bVal
		},
		"formatCurrency": func(copper int) string {
			if copper == 0 {
				return "0c"
			}
			
			gold := copper / 10000
			remaining := copper % 10000
			silver := remaining / 100
			copperLeft := remaining % 100
			
			var parts []string
			if gold > 0 {
				parts = append(parts, fmt.Sprintf("%dg", gold))
			}
			if silver > 0 {
				parts = append(parts, fmt.Sprintf("%ds", silver))
			}
			if copperLeft > 0 || len(parts) == 0 {
				parts = append(parts, fmt.Sprintf("%dc", copperLeft))
			}
			
			return strings.Join(parts, " ")
		},
		"atoi": func(s string) int {
			i, _ := strconv.Atoi(s)
			return i
		},
		"add": func(a, b int) int {
			return a + b
		},
		"substr": func(s string, start, length int) string {
			if start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
	}

	// Base template with common layout
	base := template.Must(template.New("base").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/base.html",
	))
	t.templates["base"] = base

	// Home page
	home := template.Must(template.New("index").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/base.html",
		"internal/web/assets/templates/index.html",
	))
	t.templates["index"] = home

	// Item page
	itemPage := template.Must(template.New("item_page").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/base.html",
		"internal/web/assets/templates/item_page.html",
	))
	t.templates["item_page"] = itemPage

	// Partials for HTMX
	itemResults := template.Must(template.New("item_results").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/partials/item_results.html",
	))
	t.templates["item_results"] = itemResults

	itemDetail := template.Must(template.New("item_detail").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/partials/item_detail.html",
	))
	t.templates["item_detail"] = itemDetail

	recipeTree := template.Must(template.New("recipe_tree").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/partials/recipe_tree.html",
	))
	t.templates["recipe_tree"] = recipeTree

	// Inventory page
	inventory := template.Must(template.New("inventory").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/base.html",
		"internal/web/assets/templates/inventory.html",
	))
	t.templates["inventory"] = inventory

	// Character detail page
	characterDetail := template.Must(template.New("character_detail").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/base.html",
		"internal/web/assets/templates/character_detail.html",
	))
	t.templates["character_detail"] = characterDetail

	// Character list partial
	characterList := template.Must(template.New("character_list").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/partials/character_list.html",
	))
	t.templates["character_list"] = characterList

	// Character inventory partial
	characterInventory := template.Must(template.New("character_inventory").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/partials/character_inventory.html",
	))
	t.templates["character_inventory"] = characterInventory

	// Account page
	account := template.Must(template.New("account").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/base.html",
		"internal/web/assets/templates/account.html",
	))
	t.templates["account"] = account

	// Bank page
	bank := template.Must(template.New("bank").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/base.html",
		"internal/web/assets/templates/bank.html",
	))
	t.templates["bank"] = bank

	// Shared inventory page
	shared := template.Must(template.New("shared").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/base.html",
		"internal/web/assets/templates/shared.html",
	))
	t.templates["shared"] = shared

	// Recipe page
	recipePage := template.Must(template.New("recipe_page").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/base.html",
		"internal/web/assets/templates/recipe_page.html",
	))
	t.templates["recipe_page"] = recipePage

	// Crafting tree page
	craftingTree := template.Must(template.New("crafting_tree").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/base.html",
		"internal/web/assets/templates/crafting_tree.html",
	))
	t.templates["crafting_tree"] = craftingTree

	// Partial templates for HTMX
	craftingSummaryPartial := template.Must(template.New("").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/partials/crafting_summary_partial.html",
	))
	t.templates["crafting_summary_partial"] = craftingSummaryPartial

	craftingNodePartial := template.Must(template.New("").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/partials/crafting_node_partial.html",
	))
	t.templates["crafting_node_partial"] = craftingNodePartial

	craftingChildrenPartial := template.Must(template.New("").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/partials/crafting_children_partial.html",
	))
	t.templates["crafting_children_partial"] = craftingChildrenPartial

	craftingExpandButton := template.Must(template.New("").Funcs(funcMap).ParseFiles(
		"internal/web/assets/templates/partials/crafting_expand_button.html",
	))
	t.templates["crafting_expand_button"] = craftingExpandButton

}

// Render executes a template with the given data
func (t *Templates) Render(w io.Writer, name string, data interface{}) error {
	tmpl, exists := t.templates[name]
	if !exists {
		return fmt.Errorf("template %s not found", name)
	}
	
	// For pages that inherit from base, execute the base template
	if name == "index" || name == "item_page" || name == "inventory" || name == "character_detail" || name == "account" || name == "bank" || name == "shared" || name == "recipe_page" || name == "crafting_tree" {
		return tmpl.ExecuteTemplate(w, "base.html", data)
	}
	
	// For partials, execute using the filename
	switch name {
	case "item_results":
		return tmpl.ExecuteTemplate(w, "item_results.html", data)
	case "item_detail":
		return tmpl.ExecuteTemplate(w, "item_detail.html", data)
	case "recipe_tree":
		return tmpl.ExecuteTemplate(w, "recipe_tree.html", data)
	case "character_list":
		return tmpl.ExecuteTemplate(w, "character_list.html", data)
	case "character_inventory":
		return tmpl.ExecuteTemplate(w, "character_inventory.html", data)
	case "crafting_summary_partial":
		return tmpl.ExecuteTemplate(w, "crafting_summary_partial.html", data)
	case "crafting_node_partial":
		return tmpl.ExecuteTemplate(w, "crafting_node_partial.html", data)
	case "crafting_children_partial":
		return tmpl.ExecuteTemplate(w, "crafting_children_partial.html", data)
	case "crafting_expand_button":
		return tmpl.ExecuteTemplate(w, "crafting_expand_button.html", data)
	default:
		return tmpl.Execute(w, data)
	}
}

// Template data structures
type PageData struct {
	Title   string
	Content interface{}
}

type ItemSearchData struct {
	Query string
	Items []*ItemWithPrice
}

type ItemWithPrice struct {
	*gw2api.Item
	Price    *gw2api.Price
	HasPrice bool
}

// RecipeWithOutput represents a recipe with its output item details
type RecipeWithOutput struct {
	Recipe     *gw2api.RecipeDetail
	OutputItem *gw2api.Item
}

// ItemRecipes represents both types of recipes for an item
type ItemRecipes struct {
	CreatesItem []*RecipeWithOutput // Recipes that create this item
	UsesItem    []*RecipeWithOutput // Recipes that use this item as ingredient
}

type ItemDetailData struct {
	Item     *gw2api.Item
	Price    *gw2api.Price
	HasPrice bool
	Recipes  *ItemRecipes
}

type RecipeDetailData struct {
	Recipe      *gw2api.RecipeDetail
	OutputItem  *gw2api.Item
	Ingredients []*IngredientWithItem
	TotalCost   int
}

type IngredientWithItem struct {
	*gw2api.RecipeIngredient
	Item     *gw2api.Item
	Price    *gw2api.Price
	HasPrice bool
	Cost     int // count * unit_price
}

// CraftingNode represents a node in the recursive crafting tree
type CraftingNode struct {
	Item          *gw2api.Item
	Recipe        *gw2api.RecipeDetail
	RequiredCount int
	UnitCost      int  // Cost per item (buy price if no recipe, craft cost if has recipe)
	TotalCost     int  // RequiredCount * UnitCost
	HasRecipe     bool
	CanCraft      bool // true if crafting is cheaper than buying
	BuyPrice      int  // Market buy price for comparison
	TotalBuyCost  int  // RequiredCount * BuyPrice for easy template access
	Children      []*CraftingNode // Ingredients needed if crafting
	Level         int  // Tree depth level
}

// MaterialSummary represents aggregated base materials needed
type MaterialSummary struct {
	Item          *gw2api.Item
	TotalRequired int
	UnitPrice     int
	TotalCost     int
}

// CraftingTreeData contains the complete crafting analysis
type CraftingTreeData struct {
	RootItem        *gw2api.Item
	Recipe          *gw2api.RecipeDetail
	Tree            *CraftingNode
	BaseMaterials   []*MaterialSummary
	TotalCraftCost  int
	TotalBuyCost    int
	Savings         int  // TotalBuyCost - TotalCraftCost (positive = savings, negative = extra cost)
	SavingsPercent  float64
	ExtraCost       int  // Absolute value when crafting costs more than buying
	IsCraftingCheaper bool // True if crafting is cheaper than buying
}

// RequestCache provides memoization for a single crafting tree request
type RequestCache struct {
	recipes     map[int]*gw2api.RecipeDetail  // recipeID -> recipe
	items       map[int]*gw2api.Item          // itemID -> item
	recipeSearch map[int]*ItemRecipes         // itemID -> recipes that create/use it
	prices      map[int]*gw2api.Price         // itemID -> price
	craftNodes  map[string]*CraftingNode      // "itemID:quantity:level" -> node
}

// NewRequestCache creates a new request-scoped cache
func NewRequestCache() *RequestCache {
	return &RequestCache{
		recipes:     make(map[int]*gw2api.RecipeDetail),
		items:       make(map[int]*gw2api.Item),
		recipeSearch: make(map[int]*ItemRecipes),
		prices:      make(map[int]*gw2api.Price),
		craftNodes:  make(map[string]*CraftingNode),
	}
}