# Guild Wars 2 API Server

A comprehensive server providing both web and Discord bot interfaces for the Guild Wars 2 API with modern Discord slash commands.

## Features

- **Discord Bot**: Modern slash commands for all major GW2 API endpoints
- **Web Server**: REST API and interactive web interface
- **CLI Tool**: Command-line interface (preserved from original)
- **Full Type Safety**: Complete Go type definitions for all API responses
- **Organized Structure**: Clean separation of concerns with proper package organization

## Quick Start

### Web Server Only
```bash
go build -o server ./cmd/server/
./server -web-only
```
Visit http://localhost:8080 for the web interface.

### Discord Bot Only
```bash
export DISCORD_TOKEN="your_bot_token_here"
./server -discord-only -discord-token="$DISCORD_TOKEN"
```

### Both Web Server and Discord Bot
```bash
./server -discord-token="your_bot_token_here"
```

### CLI Tool
```bash
go build -o cli ./cmd/cli/
./cli build
./cli achievements get 1 --output table
```

## Project Structure

```
├── cmd/
│   ├── cli/          # Command-line interface
│   └── server/       # Main server (Discord bot + web server)
├── internal/
│   ├── gw2api/       # Core GW2 API client library
│   ├── discord/      # Discord bot handlers
│   └── web/          # Web server handlers
├── web/
│   └── static/       # Static web assets
└── README.md, go.mod, etc.
```

## Discord Bot Commands

The Discord bot provides the following slash commands:

- `/gw2-build` - Get current game build information
- `/gw2-achievement <id>` - Get achievement details
- `/gw2-currency <id>` - Get currency information  
- `/gw2-item <id>` - Get item details
- `/gw2-world <id>` - Get world/server information
- `/gw2-skill <id>` - Get skill details
- `/gw2-prices <item_id>` - Get trading post prices

### Setting up the Discord Bot

1. Create a Discord application at https://discord.com/developers/applications
2. Create a bot and copy the token
3. Invite the bot to your server with the `applications.commands` scope
4. Run the server with your bot token:
   ```bash
   ./server -discord-token="YOUR_BOT_TOKEN"
   ```

## Web API Endpoints

- `GET /api/build` - Current game build
- `GET /api/achievement?id={id}` - Achievement details
- `GET /api/currency?id={id}` - Currency information
- `GET /api/item?id={id}` - Item details
- `GET /api/world?id={id}` - World information
- `GET /api/skill?id={id}` - Skill details
- `GET /api/prices?item_id={id}` - Trading post prices

### Web Interface

Visit `/web/` for an interactive web interface to test all API endpoints.

## Command Line Options

### Server
```bash
./server [options]

Options:
  -discord-token string    Discord bot token
  -port string            Web server port (default "8080")
  -web-only               Run only the web server
  -discord-only           Run only the Discord bot
```

### CLI Tool
```bash
./cli [command] [options]

Global Flags:
  --api-key string   API key for authenticated endpoints
  -l, --lang string  Language (en, es, de, fr, zh) (default "en")
  -o, --output string Output format (json, table, yaml) (default "json")
  -t, --timeout int  Request timeout in seconds (default 30)
  -v, --verbose      Verbose output
```

## Building

### Build All
```bash
# Server (Discord bot + web server)
go build -o server ./cmd/server/

# CLI tool
go build -o cli ./cmd/cli/
```

### Development
```bash
# Run server in development mode
go run ./cmd/server/ -web-only

# Test CLI commands
go run ./cmd/cli/ build
go run ./cmd/cli/ achievements get 1
```

## API Coverage

The server supports all the same endpoints as the original CLI:

### Core Data
- ✅ **Build** - Current game build information
- ✅ **Currencies** - All in-game currencies  
- ✅ **Worlds** - Game world/server information

### Items & Equipment
- ✅ **Items** - All items with full details and stats

### Character Progression
- ✅ **Achievements** - All achievements with tiers and rewards
- ✅ **Skills** - All skills with facts and details

### Trading Post
- ✅ **Commerce Prices** - Buy/sell price information

## Examples

### Using the Web API
```bash
# Get current build
curl http://localhost:8080/api/build

# Get achievement details
curl http://localhost:8080/api/achievement?id=1

# Get item information
curl http://localhost:8080/api/item?id=100
```

### Using in Discord
```
/gw2-build
/gw2-achievement id:1
/gw2-item id:100
/gw2-prices item_id:19684
```

## Environment Variables

- `DISCORD_TOKEN` - Discord bot token (alternative to -discord-token flag)
- `PORT` - Web server port (alternative to -port flag)

## License

Same as original project license.