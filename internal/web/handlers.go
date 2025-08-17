package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"j5.nz/gw2/internal/gw2api"
)

// EnhancedServer represents the enhanced web server with caching and templates
type EnhancedServer struct {
	client    *gw2api.Client
	cache     *WebCache
	templates *TemplateManager
}

// NewEnhancedServer creates a new enhanced web server instance
func NewEnhancedServer(gw2Client *gw2api.Client) *EnhancedServer {
	server := &EnhancedServer{
		client:    gw2Client,
		cache:     NewWebCache(),
		templates: NewTemplateManager(),
	}
	
	// Load templates
	if err := server.templates.LoadTemplates(); err != nil {
		log.Printf("Warning: Failed to load templates: %v", err)
	}
	
	return server
}

// SetupEnhancedRoutes configures the enhanced HTTP routes
func (s *EnhancedServer) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files from embedded filesystem
	staticFS, err := fs.Sub(Assets, "assets/static")
	if err != nil {
		log.Printf("Warning: Failed to create static filesystem: %v", err)
	} else {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}

	// Main pages
	mux.HandleFunc("/", s.cache.CacheMiddleware(s.indexHandler))
	
	// Search endpoints (HTMX)
	mux.HandleFunc("/search/items", s.cache.CacheMiddleware(s.searchItemsHandler))
	mux.HandleFunc("/search/skills", s.cache.CacheMiddleware(s.searchSkillsHandler))
	mux.HandleFunc("/search/worlds", s.cache.CacheMiddleware(s.searchWorldsHandler))
	mux.HandleFunc("/search/currencies", s.cache.CacheMiddleware(s.searchCurrenciesHandler))
	
	// API endpoints (with caching)
	mux.HandleFunc("/api/build", s.cache.CacheMiddleware(s.buildHandler))
	mux.HandleFunc("/api/achievement", s.cache.CacheMiddleware(s.achievementHandler))
	mux.HandleFunc("/api/currency", s.cache.CacheMiddleware(s.currencyHandler))
	mux.HandleFunc("/api/item", s.cache.CacheMiddleware(s.itemHandler))
	mux.HandleFunc("/api/world", s.cache.CacheMiddleware(s.worldHandler))
	mux.HandleFunc("/api/skill", s.cache.CacheMiddleware(s.skillHandler))
	mux.HandleFunc("/api/prices", s.cache.CacheMiddleware(s.pricesHandler))
	
	// Utility endpoints
	mux.HandleFunc("/api/cache/stats", s.cacheStatsHandler)
	mux.HandleFunc("/api/cache/clear", s.cacheClearHandler)
	mux.HandleFunc("/api/stats", s.statsHandler)

	return mux
}

// indexHandler serves the main page
func (s *EnhancedServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := context.Background()
	
	// Get current build info
	build, _ := s.client.GetBuild(ctx)
	
	// Get some basic stats (cached)
	statsData := map[string]interface{}{
		"Items":        "Loading...",
		"Skills":       "Loading...", 
		"Achievements": "Loading...",
		"Worlds":       "Loading...",
	}
	
	// Try to get actual counts (these will be cached)
	go func() {
		if ids, err := s.client.GetItemIDs(ctx); err == nil {
			statsData["Items"] = len(ids)
		}
		if ids, err := s.client.GetSkillIDs(ctx); err == nil {
			statsData["Skills"] = len(ids)
		}
		if ids, err := s.client.GetAchievementIDs(ctx); err == nil {
			statsData["Achievements"] = len(ids)
		}
		if ids, err := s.client.GetWorldIDs(ctx); err == nil {
			statsData["Worlds"] = len(ids)
		}
	}()

	data := PageData{
		Title:       "GW2 API Explorer",
		Description: "Explore Guild Wars 2 items, skills, achievements, and more",
		BuildInfo:   build,
		Data: map[string]interface{}{
			"Stats": statsData,
		},
	}
	
	// Make stats directly accessible for template
	data.Data = statsData

	tmpl, exists := s.templates.GetTemplate("index")
	if !exists {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template execution failed", http.StatusInternalServerError)
	}
}

// searchItemsHandler handles item search requests
func (s *EnhancedServer) searchItemsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	query := r.URL.Query()
	
	searchOptions := gw2api.ItemSearchOptions{
		Name:  strings.TrimSpace(query.Get("query")),
		Limit: 20, // Default limit
	}
	
	// Parse rarity filter
	if rarity := query.Get("rarity"); rarity != "" {
		searchOptions.Rarities = []string{rarity}
	}
	
	// Parse type filter  
	if itemType := query.Get("type"); itemType != "" {
		searchOptions.Types = []string{itemType}
	}
	
	// Parse level range
	if minLevel := query.Get("min_level"); minLevel != "" {
		if level, err := strconv.Atoi(minLevel); err == nil {
			searchOptions.MinLevel = level
		}
	}
	if maxLevel := query.Get("max_level"); maxLevel != "" {
		if level, err := strconv.Atoi(maxLevel); err == nil {
			searchOptions.MaxLevel = level
		}
	}
	
	// Parse pagination
	page := 0
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	// Adjust search options for pagination
	searchOptions.Limit = (page + 1) * 20

	// Perform search
	items, err := s.client.SearchItems(ctx, searchOptions)
	if err != nil {
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Paginate results
	startIdx := page * 20
	endIdx := startIdx + 20
	if endIdx > len(items) {
		endIdx = len(items)
	}
	
	var pageResults []*gw2api.Item
	if startIdx < len(items) {
		pageResults = items[startIdx:endIdx]
	}

	data := ItemSearchData{
		Results:  pageResults,
		Query:    searchOptions.Name,
		Rarity:   query.Get("rarity"),
		Type:     query.Get("type"),
		MinLevel: searchOptions.MinLevel,
		MaxLevel: searchOptions.MaxLevel,
		Total:    len(items),
		Page:     page,
		PerPage:  20,
		HasMore:  endIdx < len(items),
	}

	s.renderPartial(w, "partials/item_results", data)
}

// searchSkillsHandler handles skill search requests
func (s *EnhancedServer) searchSkillsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	query := r.URL.Query()
	searchQuery := strings.TrimSpace(query.Get("query"))
	profession := query.Get("profession")
	skillType := query.Get("type")
	
	// Use the new cached search functionality
	searchOptions := gw2api.SkillSearchOptions{
		Name:       searchQuery,
		Profession: profession,
		Type:       skillType,
		Limit:      20, // Default limit for web display
	}
	
	skills, err := s.client.SearchSkills(ctx, searchOptions)
	if err != nil {
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	data := SkillSearchData{
		Results:    skills,
		Query:      searchQuery,
		Profession: profession,
		Type:       skillType,
		Total:      len(skills),
		Page:       0,
		PerPage:    20,
		HasMore:    false,
	}

	s.renderPartial(w, "partials/skill_results", data)
}

// renderPartial renders a partial template
func (s *EnhancedServer) renderPartial(w http.ResponseWriter, templateName string, data interface{}) {
	tmpl, exists := s.templates.GetTemplate(templateName)
	if !exists {
		// Fallback to base template
		tmpl, exists = s.templates.GetTemplate("base")
		if !exists {
			http.Error(w, "Template not found", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Partial template execution error: %v", err)
		http.Error(w, "Template execution failed", http.StatusInternalServerError)
	}
}

// cacheStatsHandler returns cache statistics
func (s *EnhancedServer) cacheStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := s.cache.Stats()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Failed to encode stats", http.StatusInternalServerError)
	}
}

// cacheClearHandler clears the cache
func (s *EnhancedServer) cacheClearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	s.cache.Clear()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// statsHandler returns general API stats
func (s *EnhancedServer) statsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	stats := map[string]interface{}{
		"cache": s.cache.Stats(),
		"timestamp": time.Now().Unix(),
	}
	
	// Get current build
	if build, err := s.client.GetBuild(ctx); err == nil {
		stats["build"] = build
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Failed to encode stats", http.StatusInternalServerError)
	}
}

// Placeholder handlers for missing search methods
func (s *EnhancedServer) searchWorldsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	// Get all world IDs
	ids, err := s.client.GetWorldIDs(ctx)
	if err != nil {
		http.Error(w, "Failed to get world IDs", http.StatusInternalServerError)
		return
	}
	
	// Get world details
	worlds, err := s.client.GetWorlds(ctx, ids[:20]) // Limit to first 20 for performance
	if err != nil {
		http.Error(w, "Failed to get world details", http.StatusInternalServerError)
		return
	}
	
	// Create HTML response
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">`)
	
	for _, world := range worlds {
		fmt.Fprintf(w, `
			<div class="gw2-card p-4">
				<h3 class="font-bold text-lg mb-2">%s</h3>
				<div class="text-sm text-gray-600 dark:text-gray-400">
					<p>ID: %d</p>
					<p>Population: %s</p>
				</div>
			</div>
		`, world.Name, world.ID, world.Population)
	}
	
	fmt.Fprint(w, `</div>`)
}

func (s *EnhancedServer) searchCurrenciesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	// Get all currency IDs
	ids, err := s.client.GetCurrencyIDs(ctx)
	if err != nil {
		http.Error(w, "Failed to get currency IDs", http.StatusInternalServerError)
		return
	}
	
	// Get currency details (limit to first 30 for performance)
	limit := 30
	if len(ids) < limit {
		limit = len(ids)
	}
	currencies, err := s.client.GetCurrencies(ctx, ids[:limit])
	if err != nil {
		http.Error(w, "Failed to get currency details", http.StatusInternalServerError)
		return
	}
	
	// Create HTML response
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">`)
	
	for _, currency := range currencies {
		icon := "💰"
		if currency.Icon != "" {
			icon = fmt.Sprintf(`<img src="%s" alt="%s" class="w-8 h-8 inline-block">`, currency.Icon, currency.Name)
		}
		
		fmt.Fprintf(w, `
			<div class="gw2-card p-4">
				<div class="flex items-center mb-2">
					%s
					<h3 class="font-bold text-lg ml-2">%s</h3>
				</div>
				<div class="text-sm text-gray-600 dark:text-gray-400">
					<p class="mb-1">%s</p>
					<p class="text-xs">Order: %d</p>
				</div>
			</div>
		`, icon, currency.Name, currency.Description, currency.Order)
	}
	
	fmt.Fprint(w, `</div>`)
}

// API handlers reused from original server
func (s *EnhancedServer) buildHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	build, err := s.client.GetBuild(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(build)
}

func (s *EnhancedServer) achievementHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	achievement, err := s.client.GetAchievement(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(achievement)
}

func (s *EnhancedServer) currencyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	currency, err := s.client.GetCurrency(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currency)
}

func (s *EnhancedServer) itemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	item, err := s.client.GetItem(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (s *EnhancedServer) worldHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	world, err := s.client.GetWorld(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(world)
}

func (s *EnhancedServer) skillHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	skill, err := s.client.GetSkill(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(skill)
}

func (s *EnhancedServer) pricesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("item_id")
	if idStr == "" {
		http.Error(w, "Missing item_id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item_id parameter", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	price, err := s.client.GetCommercePrice(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(price)
}