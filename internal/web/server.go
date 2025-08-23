package web

import (
	"fmt"
	"net/http"
	"time"

	"j5.nz/gw2/internal/cache"
	"j5.nz/gw2/internal/gw2api"
)

// Server represents the web server
type Server struct {
	client     *gw2api.Client
	priceCache cache.Cache
	templates  *Templates
	*http.ServeMux
}

// NewServer creates a new web server
func NewServer(client *gw2api.Client, priceCache cache.Cache) *Server {
	s := &Server{
		client:     client,
		priceCache: priceCache,
		ServeMux:   http.NewServeMux(),
	}

	// Initialize templates
	s.templates = NewTemplates()

	// Setup routes
	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	// Main pages
	s.HandleFunc("GET /", s.handleHome)
	s.HandleFunc("GET /items/{id}", s.handleItemPage)
	s.HandleFunc("GET /inventory", s.handleInventoryPage)
	
	// HTMX endpoints
	s.HandleFunc("POST /search/items", s.handleItemSearch)
	s.HandleFunc("GET /item/{id}", s.handleItemDetail)
	s.HandleFunc("GET /recipe/{id}", s.handleRecipeDetail)
	s.HandleFunc("GET /crafting/{id}", s.handleCraftingTree)
	s.HandleFunc("GET /crafting/node/{recipeId}/{itemId}/{quantity}", s.handleCraftingNode)
	s.HandleFunc("GET /crafting/expand/{recipeId}/{itemId}/{quantity}/{level}", s.handleCraftingNodeExpand)
	s.HandleFunc("GET /crafting/summary/{id}", s.handleCraftingSummary)
	s.HandleFunc("GET /characters", s.handleCharacters)
	s.HandleFunc("GET /inventory/{character}", s.handleCharacterInventory)
	s.HandleFunc("GET /account", s.handleAccountPage)
	s.HandleFunc("GET /bank", s.handleBankPage)
	s.HandleFunc("GET /shared", s.handleSharedInventoryPage)
	
	// API key handling
	s.HandleFunc("POST /api-key", s.handleSetAPIKey)
	
	// Static files
	s.Handle("GET /static/", s.staticFileHandler())
}

// staticFileHandler serves static files
func (s *Server) staticFileHandler() http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.Dir("internal/web/assets/static")))
}

// PriceCache wraps the trading post price cache with proper TTL
type PriceCache struct {
	cache cache.Cache
}

// GetPrice gets a cached price or returns nil if not found/expired
func (pc *PriceCache) GetPrice(itemID int) (*gw2api.Price, bool) {
	key := fmt.Sprintf("price_%d", itemID)
	if value, found := pc.cache.Get(key); found {
		if price, ok := value.(*gw2api.Price); ok {
			return price, true
		}
	}
	return nil, false
}

// SetPrice caches a price for 3 hours
func (pc *PriceCache) SetPrice(itemID int, price *gw2api.Price) {
	key := fmt.Sprintf("price_%d", itemID)
	pc.cache.Set(key, price, 3*time.Hour)
}