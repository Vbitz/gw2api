# Guild Wars 2 API Go Client and CLI

A Go-based client library and CLI application for the Guild Wars 2 API v2 with full type safety and comprehensive endpoint coverage.

**ALWAYS follow these instructions first. Only fallback to additional search or context gathering if the information here is incomplete or found to be in error.**

## Working Effectively

### Essential Setup and Build Commands
Run these commands in sequence for a fresh clone:

```bash
# Install Go dependencies
go mod tidy

# Build the CLI application 
go build -o gw2api .
```

**Build Timing:** Build takes approximately 0.4 seconds after dependencies are downloaded, 0.1 seconds for incremental builds. NEVER CANCEL builds - they complete quickly.

### Code Quality and Validation
ALWAYS run these before committing changes:

```bash
# Format Go code (fixes most style issues)
go fmt ./...

# Lint and check for issues
go vet ./...

# Build to ensure no compilation errors
go build -o gw2api .
```

**Critical:** `go vet` must pass without errors. Fix any reported issues before proceeding.

### Testing and Validation
This project uses a demo-based testing approach rather than formal unit tests:

```bash
# Test CLI basic functionality (no network required)
./gw2api --help
./gw2api version
./gw2api achievements --help

# Test API functionality (requires internet access to api.guildwars2.com)
./gw2api build
./gw2api achievements get 1 --output table
./demo.sh
```

**Network Dependency:** Most CLI commands require internet access to api.guildwars2.com. If network requests fail, it's expected in restricted environments.

## Validation Scenarios

ALWAYS test these scenarios after making changes:

### CLI Validation
```bash
# 1. Help system works
./gw2api --help
./gw2api achievements --help
./gw2api currencies --help

# 2. Version and basic commands
./gw2api version

# 3. Output format options
./gw2api build --output json
./gw2api build --output table

# 4. Language options  
./gw2api achievements get 1 --lang en
./gw2api achievements get 1 --lang de
```

### API Client Library Validation
If modifying the Go library (gw2api.go), test with:

```bash
# Run the test API function (requires network)
go run *.go
```

This executes test_api.go which validates:
- Build information retrieval
- Achievement data with type safety
- Multiple currencies handling
- World information
- Bulk operations

## Project Structure

### Core Files
- **gw2api.go**: Main API client library with typed endpoints
- **cmd.go**: CLI application using Cobra framework
- **common.go**: Shared types and utilities
- **test_api.go**: Demo/testing functionality

### Endpoint-Specific Files
- **achievement.go**: Achievement data types
- **commerce.go**: Trading post types
- **currency.go**: Currency definitions
- **item.go**: Item data structures
- **skill.go**: Skill information
- **world.go**: World/server data

### Documentation
- **README.md**: Library documentation and examples
- **CLI_README.md**: Comprehensive CLI usage guide
- **demo.sh**: Interactive CLI demonstration script

## Key Development Patterns

### Adding New CLI Commands
1. Add API method to gw2api.go with proper types
2. Add CLI command structure to cmd.go
3. Follow existing pattern with global flags support
4. Test both JSON and table output formats

### Type Safety Requirements
- ALL API responses must be strongly typed (no `interface{}`)
- Use proper Go struct tags for JSON marshaling
- Include proper error handling for all API calls
- Support bulk operations where the API allows

### CLI Command Patterns
```go
// Standard command structure
var newCmd = &cobra.Command{
    Use:   "command-name",
    Short: "Brief description",
    Run: func(cmd *cobra.Command, args []string) {
        ctx := context.Background()
        // API call with error handling
        // Output using outputData() or outputIDs()
    },
}
```

## Timeout and Performance Guidelines

### Build and Test Timeouts
- **Build**: Set timeout to 30+ seconds (typically completes in <1 second)
- **API Tests**: Set timeout to 60+ seconds for network requests
- **NEVER CANCEL**: All build and test operations complete quickly

### API Request Timeouts
- Default timeout: 30 seconds (configurable via --timeout flag)
- Large bulk operations: Consider 60+ second timeouts
- Network issues are common in restricted environments

## Common Issues and Solutions

### Build Issues
```bash
# Missing dependencies
go mod tidy

# Compilation errors
go build -o gw2api .

# Format issues  
go fmt ./...
```

### Network Issues
```bash
# API requests fail with "dial tcp: lookup api.guildwars2.com"
# This is expected in environments without internet access
# Use --timeout flag to adjust request timeouts
./gw2api build --timeout 60
```

### Linting Issues
```bash
# Fix common Go vet issues
go vet ./...

# Common fix: Remove redundant newlines in fmt.Println
# Example: fmt.Println("text\n") -> fmt.Println("text")
```

## Integration Testing

When the environment has internet access, ALWAYS validate end-to-end scenarios:

```bash
# Complete CLI workflow
./gw2api build --output table
./gw2api achievements get 1 --output table
./gw2api currencies get 1,2,3 --output table
./gw2api commerce prices 19684 --output table

# Bulk operations
./gw2api achievements get 1,2,3,4,5
./gw2api items get 100,200,300

# Different languages
./gw2api achievements get 1 --lang de --output table
./gw2api achievements get 1 --lang fr --output table

# Run full demo
./demo.sh
```

## Dependencies and Tools

### Required
- **Go 1.24+**: Language runtime
- **github.com/spf13/cobra**: CLI framework
- **Internet access**: For API functionality

### Development Tools
- `go fmt`: Code formatting
- `go vet`: Static analysis
- `go build`: Compilation
- `go mod tidy`: Dependency management

Always ensure dependencies are properly managed through go.mod and avoid manual dependency installation.