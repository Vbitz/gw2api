# Guild Wars 2 API Go Client

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Bootstrap and Build (REQUIRED FIRST STEPS)
- Go 1.24.2+ is required (confirmed working with Go 1.24.6)
- Install dependencies and build the project:
  - `go mod tidy` -- takes <1 second
  - `go build -o gw2 ./cmd/gw2/` -- takes 1-2 seconds. Builds the CLI tool
  - `go build -o server ./cmd/server/` -- takes 1-2 seconds. Builds the web server + Discord bot
  - `go build ./...` -- takes 1-2 seconds. Builds all components at once

### Development Mode (Alternative to Building)
- CLI development: `go run ./cmd/gw2/ [command]` -- takes 1-2 seconds to start
- Server development: `go run ./cmd/server/ [flags]` -- takes 1-2 seconds to start

### Testing
- `go test ./...` -- takes 4-5 seconds. No test files exist currently
- `go vet ./...` -- takes <1 second. Validates code for issues
- `gofmt -w .` -- takes <1 second. Auto-formats all Go code (ALWAYS run before committing)

### Running the Applications

#### CLI Tool
- ALWAYS build first: `go build -o gw2 ./cmd/gw2/`
- Basic usage: `./gw2 --help`
- Version check: `./gw2 version`
- Available commands: `build`, `achievements`, `currencies`, `items`, `worlds`, `skills`, `commerce`
- **NETWORK DEPENDENCY**: All API commands require internet access to api.guildwars2.com
- **EXPECTED OFFLINE ERROR**: `dial tcp: lookup api.guildwars2.com: server misbehaving` when no internet
- Test offline functionality: `./gw2 --help`, `./gw2 version`, `./gw2 achievements --help`

#### Web Server + Discord Bot
- ALWAYS build first: `go build -o server ./cmd/server/`
- Web server only: `./server -web-only`
- Discord bot only: `./server -discord-only -discord-token="YOUR_TOKEN"`
- Both (default): `./server -discord-token="YOUR_TOKEN"`
- Custom port: `./server -web-only -port 8081`
- **WEB INTERFACE**: Visit http://localhost:8080 after starting server
- **INTERACTIVE API TESTING**: Visit http://localhost:8080/web/ for interactive interface

### Key Development Workflows

#### Making Changes to CLI
1. Edit files in `cmd/gw2/` or `internal/`
2. `go build -o gw2 ./cmd/gw2/`
3. Test with `./gw2 [command]`
4. Run `gofmt -w .` before committing

#### Making Changes to Server  
1. Edit files in `cmd/server/` or `internal/`
2. `go build -o server ./cmd/server/`
3. Test with `./server -web-only`
4. Verify web interface at http://localhost:8080
5. Run `gofmt -w .` before committing

#### Adding New API Endpoints
- Edit `internal/gw2api/client.go` for new methods
- Edit `internal/gw2api/types.go` for new data structures
- Update CLI commands in `cmd/gw2/main.go`
- Update web handlers in `internal/web/server.go`
- Update Discord handlers in `internal/discord/bot.go`

## Validation Steps

### ALWAYS run these before completing changes:
1. `gofmt -w .` -- Format all code
2. `go vet ./...` -- Check for issues  
3. `go build ./...` -- Ensure everything builds
4. Test CLI: `./gw2 --help` and `./gw2 version`
5. Test server: Start with `./server -web-only` and visit http://localhost:8080

### Manual Testing Scenarios
- **CLI Help System**: Test `./gw2 --help`, `./gw2 achievements --help`, etc.
- **Web Interface**: Start server and test interactive interface at /web/
- **Network Error Handling**: Test commands with `--timeout 1` to verify error messages
- **Output Formats**: Test with `--output table` and `--output json`
- **Language Support**: Test with `--lang de`, `--lang fr`, etc.

## Project Structure

### Key Directories
- `cmd/gw2/` - Command-line interface main package
- `cmd/server/` - Web server + Discord bot main package  
- `internal/gw2api/` - Core API client library
- `internal/web/` - Web server and HTTP handlers
- `internal/discord/` - Discord bot implementation
- `web/static/` - Static web assets (CSS, etc.)

### Important Files
- `go.mod` - Go module definition (module name: j5.nz/gw2)
- `demo.sh` - Demonstration script showing CLI usage
- `CLI_README.md` - Comprehensive CLI documentation
- `SERVER_README.md` - Server and Discord bot documentation
- `README.md` - Main project documentation

### Generated Binaries
- `gw2` - Pre-built CLI tool (functional)
- Build outputs: `gw2_new`, `server_new`, etc. (your builds)

## API Overview

### Supported Endpoints
- **Build** - Current game build information
- **Achievements** - Achievement data with tiers and rewards
- **Currencies** - In-game currency information
- **Items** - Item details with stats and metadata
- **Worlds** - Game server/world information
- **Skills** - Skill details with facts and descriptions
- **Commerce** - Trading post price data

### CLI Command Examples
```bash
# Get current build
./gw2 build

# Get achievement with table output
./gw2 achievements get 1 --output table

# Get multiple items in German
./gw2 items get 100,200,300 --lang de

# Search for items (placeholder implementation)
./gw2 items search --name sword --limit 10
```

### Web API Endpoints
- `GET /api/build` - Current game build
- `GET /api/achievement?id={id}` - Achievement details
- `GET /api/currency?id={id}` - Currency information
- `GET /api/item?id={id}` - Item details
- `GET /api/world?id={id}` - World information
- `GET /api/skill?id={id}` - Skill details
- `GET /api/prices?item_id={id}` - Trading post prices

## Common Issues

### Build Problems
- **Missing package error**: Ensure `internal/gw2api/` package exists with `client.go` and `types.go`
- **Import errors**: Module name is `j5.nz/gw2`, not `github.com/...`
- **Format issues**: Always run `gofmt -w .` before building

### Runtime Problems  
- **Network errors**: Expected when Guild Wars 2 API is unreachable
- **Timeout errors**: Use `--timeout 60` for slow connections
- **Invalid IDs**: Check if ID exists first with list commands

### Discord Bot Setup
- Requires Discord application and bot token
- Use `-discord-token="YOUR_TOKEN"` flag
- Bot provides slash commands: `/gw2-build`, `/gw2-achievement`, etc.
- Will fail gracefully if token not provided

## No Test Infrastructure
- Project currently has no test files
- Use manual testing and validation steps above
- Focus on building, running, and exercising functionality
- Verify output formats and error handling manually