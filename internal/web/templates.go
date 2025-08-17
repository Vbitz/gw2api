package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"j5.nz/gw2/internal/gw2api"
)

// TemplateManager handles template loading and rendering
type TemplateManager struct {
	templates map[string]*template.Template
	funcs     template.FuncMap
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*template.Template),
		funcs:     make(template.FuncMap),
	}
	
	tm.setupFunctions()
	return tm
}

// setupFunctions configures template helper functions
func (tm *TemplateManager) setupFunctions() {
	tm.funcs = template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"formatDuration": func(d time.Duration) string {
			if d < time.Minute {
				return fmt.Sprintf("%.0fs", d.Seconds())
			}
			if d < time.Hour {
				return fmt.Sprintf("%.1fm", d.Minutes())
			}
			return fmt.Sprintf("%.1fh", d.Hours())
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"formatCoins": func(coins int) string {
			if coins == 0 {
				return "0 copper"
			}
			
			gold := coins / 10000
			silver := (coins % 10000) / 100
			copper := coins % 100
			
			var parts []string
			if gold > 0 {
				parts = append(parts, fmt.Sprintf("%dg", gold))
			}
			if silver > 0 {
				parts = append(parts, fmt.Sprintf("%ds", silver))
			}
			if copper > 0 {
				parts = append(parts, fmt.Sprintf("%dc", copper))
			}
			
			return strings.Join(parts, " ")
		},
		"rarityClass": func(rarity string) string {
			switch strings.ToLower(rarity) {
			case "basic", "fine":
				return "rarity-common"
			case "masterwork":
				return "rarity-uncommon"
			case "rare":
				return "rarity-rare"
			case "exotic":
				return "rarity-exotic"
			case "ascended":
				return "rarity-ascended"
			case "legendary":
				return "rarity-legendary"
			default:
				return "rarity-common"
			}
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"contains": func(s, substr string) bool {
			return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
		},
		"join": func(sep string, items []string) string {
			return strings.Join(items, sep)
		},
		"dict": func(values ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				if i+1 < len(values) {
					key := fmt.Sprintf("%v", values[i])
					dict[key] = values[i+1]
				}
			}
			return dict
		},
		"itemTypeIcon": func(itemType string) string {
			switch strings.ToLower(itemType) {
			case "armor":
				return "🛡️"
			case "weapon":
				return "⚔️"
			case "trinket":
				return "💍"
			case "consumable":
				return "🧪"
			case "tool":
				return "🔧"
			case "trophy":
				return "🏆"
			case "upgradcomponent":
				return "💎"
			case "bag":
				return "🎒"
			default:
				return "📦"
			}
		},
		"skillTypeIcon": func(skillType string) string {
			switch strings.ToLower(skillType) {
			case "weapon":
				return "⚔️"
			case "heal":
				return "❤️"
			case "utility":
				return "🔧"
			case "elite":
				return "⭐"
			case "profession":
				return "👤"
			default:
				return "✨"
			}
		},
	}
}

// LoadTemplates loads all templates from the embedded filesystem
func (tm *TemplateManager) LoadTemplates() error {
	templatesFS, err := fs.Sub(Assets, "assets/templates")
	if err != nil {
		return fmt.Errorf("failed to create templates sub-filesystem: %w", err)
	}
	
	// First, parse the base template
	baseContent, err := fs.ReadFile(templatesFS, "base.html")
	if err != nil {
		return fmt.Errorf("failed to read base template: %w", err)
	}
	
	// Load all page templates
	err = fs.WalkDir(templatesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		
		// Skip base template for page templates 
		if strings.Contains(path, "base.html") {
			return nil
		}
		
		// Generate template name from path
		name := strings.TrimSuffix(path, ".html")
		name = strings.ReplaceAll(name, string(filepath.Separator), "/")
		
		// Read page template content
		pageContent, err := fs.ReadFile(templatesFS, path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", path, err)
		}
		
		// Create a new template and parse both base and page templates
		tmpl := template.New(name).Funcs(tm.funcs)
		
		// Parse base template first, then page template
		_, err = tmpl.Parse(string(baseContent))
		if err != nil {
			return fmt.Errorf("failed to parse base template for %s: %w", name, err)
		}
		
		_, err = tmpl.Parse(string(pageContent))
		if err != nil {
			return fmt.Errorf("failed to parse page template %s: %w", name, err)
		}
		
		tm.templates[name] = tmpl
		return nil
	})
	
	return err
}


// GetTemplate returns a template by name
func (tm *TemplateManager) GetTemplate(name string) (*template.Template, bool) {
	tmpl, exists := tm.templates[name]
	return tmpl, exists
}

// PageData represents common data passed to all page templates
type PageData struct {
	Title       string
	Description string
	Keywords    string
	ActiveTab   string
	User        interface{} // For future authentication
	Cache       interface{} // Cache stats for debugging
	BuildInfo   *gw2api.Build
	Error       string
	Success     string
	Data        interface{} // Page-specific data
}

// SearchResult represents a generic search result
type SearchResult struct {
	Items   interface{} `json:"items"`
	Total   int         `json:"total"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
	Query   string      `json:"query"`
}

// ItemSearchData represents data for item search page
type ItemSearchData struct {
	Results     []*gw2api.Item `json:"results"`
	Query       string         `json:"query"`
	Rarity      string         `json:"rarity"`
	Type        string         `json:"type"`
	MinLevel    int            `json:"min_level"`
	MaxLevel    int            `json:"max_level"`
	Total       int            `json:"total"`
	Page        int            `json:"page"`
	PerPage     int            `json:"per_page"`
	HasMore     bool           `json:"has_more"`
	Rarities    []string       `json:"rarities"`
	Types       []string       `json:"types"`
}

// SkillSearchData represents data for skill search page
type SkillSearchData struct {
	Results     []*gw2api.Skill `json:"results"`
	Query       string          `json:"query"`
	Profession  string          `json:"profession"`
	Type        string          `json:"type"`
	Total       int             `json:"total"`
	Page        int             `json:"page"`
	PerPage     int             `json:"per_page"`
	HasMore     bool            `json:"has_more"`
	Professions []string        `json:"professions"`
	Types       []string        `json:"types"`
}