package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"j5.nz/gw2/internal/gw2api"
)

// Server represents the web server
type Server struct {
	client    *gw2api.Client
	templates *template.Template
}

// NewServer creates a new web server instance
func NewServer(gw2Client *gw2api.Client) *Server {
	return &Server{
		client: gw2Client,
	}
}

// SetupRoutes configures the HTTP routes
func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// API routes
	mux.HandleFunc("/", s.indexHandler)
	mux.HandleFunc("/api/build", s.buildHandler)
	mux.HandleFunc("/api/achievement", s.achievementHandler)
	mux.HandleFunc("/api/currency", s.currencyHandler)
	mux.HandleFunc("/api/item", s.itemHandler)
	mux.HandleFunc("/api/world", s.worldHandler)
	mux.HandleFunc("/api/skill", s.skillHandler)
	mux.HandleFunc("/api/prices", s.pricesHandler)

	// Web interface routes
	mux.HandleFunc("/web/", s.webInterfaceHandler)

	return mux
}

// indexHandler serves the main page
func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Guild Wars 2 API Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .method { background: #007acc; color: white; padding: 2px 8px; border-radius: 3px; font-size: 12px; }
        h1 { color: #333; }
        h2 { color: #666; }
        code { background: #f0f0f0; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Guild Wars 2 API Server</h1>
        <p>This server provides both a REST API and Discord bot interface for the Guild Wars 2 API.</p>
        
        <h2>Available API Endpoints</h2>
        
        <div class="endpoint">
            <span class="method">GET</span> <code>/api/build</code>
            <p>Get the current game build information</p>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span> <code>/api/achievement?id={id}</code>
            <p>Get information about a specific achievement</p>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span> <code>/api/currency?id={id}</code>
            <p>Get information about a specific currency</p>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span> <code>/api/item?id={id}</code>
            <p>Get information about a specific item</p>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span> <code>/api/world?id={id}</code>
            <p>Get information about a specific world/server</p>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span> <code>/api/skill?id={id}</code>
            <p>Get information about a specific skill</p>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span> <code>/api/prices?item_id={id}</code>
            <p>Get trading post prices for an item</p>
        </div>

        <h2>Discord Bot</h2>
        <p>The Discord bot provides slash commands for all the above functionality:</p>
        <ul>
            <li><code>/gw2-build</code> - Get current build</li>
            <li><code>/gw2-achievement &lt;id&gt;</code> - Get achievement info</li>
            <li><code>/gw2-currency &lt;id&gt;</code> - Get currency info</li>
            <li><code>/gw2-item &lt;id&gt;</code> - Get item info</li>
            <li><code>/gw2-world &lt;id&gt;</code> - Get world info</li>
            <li><code>/gw2-skill &lt;id&gt;</code> - Get skill info</li>
            <li><code>/gw2-prices &lt;item_id&gt;</code> - Get trading post prices</li>
        </ul>

        <h2>Web Interface</h2>
        <p><a href="/web/">Try the interactive web interface</a></p>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// buildHandler handles the build API endpoint
func (s *Server) buildHandler(w http.ResponseWriter, r *http.Request) {
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

// achievementHandler handles the achievement API endpoint
func (s *Server) achievementHandler(w http.ResponseWriter, r *http.Request) {
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

// currencyHandler handles the currency API endpoint
func (s *Server) currencyHandler(w http.ResponseWriter, r *http.Request) {
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

// itemHandler handles the item API endpoint
func (s *Server) itemHandler(w http.ResponseWriter, r *http.Request) {
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

// worldHandler handles the world API endpoint
func (s *Server) worldHandler(w http.ResponseWriter, r *http.Request) {
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

// skillHandler handles the skill API endpoint
func (s *Server) skillHandler(w http.ResponseWriter, r *http.Request) {
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

// pricesHandler handles the prices API endpoint
func (s *Server) pricesHandler(w http.ResponseWriter, r *http.Request) {
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

// webInterfaceHandler serves the interactive web interface
func (s *Server) webInterfaceHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>GW2 API Web Interface</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 1000px; margin: 0 auto; }
        .form-group { margin: 15px 0; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input, select { padding: 8px; width: 200px; border: 1px solid #ccc; border-radius: 4px; }
        button { background: #007acc; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #005c99; }
        .result { background: #f9f9f9; border: 1px solid #ddd; padding: 15px; margin-top: 20px; border-radius: 4px; }
        .error { background: #ffe6e6; border: 1px solid #ffcccc; color: #cc0000; }
        pre { background: #f5f5f5; padding: 10px; border-radius: 4px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Guild Wars 2 API Web Interface</h1>
        
        <div class="form-group">
            <label for="endpoint">Choose API Endpoint:</label>
            <select id="endpoint" onchange="updateForm()">
                <option value="build">Build</option>
                <option value="achievement">Achievement</option>
                <option value="currency">Currency</option>
                <option value="item">Item</option>
                <option value="world">World</option>
                <option value="skill">Skill</option>
                <option value="prices">Trading Post Prices</option>
            </select>
        </div>

        <div class="form-group" id="idGroup" style="display: none;">
            <label for="itemId">ID:</label>
            <input type="number" id="itemId" placeholder="Enter ID">
        </div>

        <div class="form-group">
            <button onclick="makeRequest()">Get Data</button>
        </div>

        <div id="result" class="result" style="display: none;"></div>
    </div>

    <script>
        function updateForm() {
            const endpoint = document.getElementById('endpoint').value;
            const idGroup = document.getElementById('idGroup');
            
            if (endpoint === 'build') {
                idGroup.style.display = 'none';
            } else {
                idGroup.style.display = 'block';
                const label = document.querySelector('#idGroup label');
                const input = document.getElementById('itemId');
                
                if (endpoint === 'prices') {
                    label.textContent = 'Item ID:';
                    input.placeholder = 'Enter Item ID';
                } else {
                    label.textContent = 'ID:';
                    input.placeholder = 'Enter ID';
                }
            }
        }

        function makeRequest() {
            const endpoint = document.getElementById('endpoint').value;
            const id = document.getElementById('itemId').value;
            const resultDiv = document.getElementById('result');
            
            let url = '/api/' + endpoint;
            
            if (endpoint !== 'build') {
                if (!id) {
                    showError('Please enter an ID');
                    return;
                }
                
                if (endpoint === 'prices') {
                    url += '?item_id=' + id;
                } else {
                    url += '?id=' + id;
                }
            }
            
            resultDiv.style.display = 'block';
            resultDiv.innerHTML = 'Loading...';
            resultDiv.className = 'result';
            
            fetch(url)
                .then(response => {
                    if (!response.ok) {
                        return response.text().then(text => {
                            throw new Error(text);
                        });
                    }
                    return response.json();
                })
                .then(data => {
                    resultDiv.innerHTML = '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
                })
                .catch(error => {
                    showError(error.message);
                });
        }

        function showError(message) {
            const resultDiv = document.getElementById('result');
            resultDiv.style.display = 'block';
            resultDiv.className = 'result error';
            resultDiv.innerHTML = 'Error: ' + message;
        }

        // Initialize form
        updateForm();
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}
