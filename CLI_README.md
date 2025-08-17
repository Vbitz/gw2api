# Guild Wars 2 API CLI

A comprehensive command-line interface for the Guild Wars 2 API with full type safety, multiple output formats, and extensive functionality.

## Installation

```bash
# Build from source
git clone <repository>
cd gw2api
go build -o gw2api .

# Make executable (optional)
chmod +x gw2api
sudo mv gw2api /usr/local/bin/
```

## Quick Start

```bash
# Get current game build
./gw2api build

# Get achievement with table output
./gw2api achievements get 1 --output table

# Get multiple currencies in German
./gw2api currencies get 1,2,3 --lang de

# Get trading post prices
./gw2api commerce prices 19684,19709 --output table
```

## Global Flags

All commands support these global flags:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `json` | Output format (`json`, `table`) |
| `--lang` | `-l` | `en` | Language (`en`, `es`, `de`, `fr`, `zh`) |
| `--timeout` | `-t` | `30` | Request timeout in seconds |
| `--api-key` | | | API key for authenticated endpoints |
| `--verbose` | `-v` | `false` | Verbose output |

## Commands Overview

### Core Information

#### `build` - Get Current Game Build
```bash
# Get current build ID
./gw2api build
./gw2api build --output table
```

### Achievements

#### `achievements` - Achievement Operations
```bash
# List all achievement IDs
./gw2api achievements list

# Get single achievement
./gw2api achievements get 1

# Get multiple achievements
./gw2api achievements get 1,2,3,4,5

# Get achievement in German
./gw2api achievements get 1 --lang de --output table
```

**Example Output (Table):**
```
ID: 1
Name: Centaur Slayer
Description: A few more centaur herds are thinned out.
Type: Default
Tiers: 4
Flags: Permanent
```

### Currencies

#### `currencies` - Currency Operations
```bash
# List all currency IDs
./gw2api currencies list

# Get single currency
./gw2api currencies get 1

# Get multiple currencies
./gw2api currencies get 1,2,3

# Get all currencies
./gw2api currencies all --output table
```

**Example Output (Table):**
```
ID    Name                 Description                                        Order
----------------------------------------------------------------------------------
1     Coin                 The primary currency of Tyria. Spent at vendors.. 101  
2     Karma                Earned through various activities. Spent at vend.. 102  
3     Laurel               Obtained from Wizard's Vault rewards. Used to pu.. 104
```

### Items

#### `items` - Item Operations
```bash
# List all item IDs (shows count and first 10)
./gw2api items list

# Get single item
./gw2api items get 100

# Get multiple items
./gw2api items get 100,200,300 --output table

# Search for items by name
./gw2api items search --name sword

# Search for items by rarity
./gw2api items search --rarity exotic

# Search with multiple filters
./gw2api items search --name "berserker" --rarity exotic --limit 10

# Search with table output
./gw2api items search --name sword --output table
```

**Example Output (Table):**
```
ID       Name                           Type            Rarity     Level
----------------------------------------------------------------------
100      Rampager's Seer Coat of Div... Armor           Exotic     72   
200      Rampager's Seer Pants of Di... Armor           Exotic     72   
300      Rampager's Seer Boots of Di... Armor           Exotic     72
```

**Search Examples:**
```bash
# Find all exotic swords
./gw2api items search --name sword --rarity exotic --output table

# Find items with "dragon" in the name
./gw2api items search --name dragon --limit 20

# Find all legendary items
./gw2api items search --rarity legendary

# Find berserker armor pieces
./gw2api items search --name berserker --output table
```

### Worlds

#### `worlds` - World/Server Operations
```bash
# List all world IDs
./gw2api worlds list

# Get single world
./gw2api worlds get 1001

# Get multiple worlds
./gw2api worlds get 1001,1002,1003 --output table

# Get all worlds
./gw2api worlds all
```

**Example Output (Table):**
```
ID     Name                      Population  
---------------------------------------------
1001   Anvil Rock                High        
1002   Borlis Pass               High        
1003   Yak's Bend                High
```

### Skills

#### `skills` - Skill Operations
```bash
# List all skill IDs
./gw2api skills list

# Get single skill
./gw2api skills get 1110
```

### Trading Post

#### `commerce` - Trading Post Operations
```bash
# Get trading post prices for items
./gw2api commerce prices 19684

# Get multiple item prices
./gw2api commerce prices 19684,19709,19698 --output table
```

**Example Output (Table):**
```
Item ID  Buy Price    Buy Qty      Sell Price   Sell Qty    
--------------------------------------------------------------
19684    167          861444       170          728849      
19709    172          296531       175          3856835     
19698    156          234567       159          987654
```

## Output Formats

### JSON Format (Default)
```bash
./gw2api build
# Output: {"id": 185315}

./gw2api achievements get 1
# Output: Full JSON object with all fields
```

### Table Format
```bash
./gw2api build --output table
# Output: Build ID: 185315

./gw2api currencies get 1,2 --output table
# Output: Formatted table with columns
```

## Language Support

The CLI supports all Guild Wars 2 API languages:

```bash
# English (default)
./gw2api achievements get 1 --lang en

# Spanish
./gw2api achievements get 1 --lang es

# German
./gw2api achievements get 1 --lang de

# French
./gw2api achievements get 1 --lang fr

# Chinese
./gw2api achievements get 1 --lang zh
```

## Advanced Usage

### Bulk Operations
```bash
# Get multiple items efficiently (single API call)
./gw2api items get 100,200,300,400,500

# Comma-separated or space-separated IDs
./gw2api achievements get 1 2 3 4 5
./gw2api achievements get 1,2,3,4,5
```

### Error Handling
```bash
# Invalid ID
./gw2api achievements get 999999
# Output: Error: {"text":"no such id"}

# Network timeout
./gw2api build --timeout 1
# Output: Error: context deadline exceeded
```

### Verbose Output
```bash
# Enable verbose logging
./gw2api achievements get 1 --verbose
```

## Common Use Cases

### 1. Price Checking Items
```bash
# Check current trading post prices
./gw2api commerce prices 19684,19709 --output table

# Get item details for context
./gw2api items get 19684,19709 --output table

# Find items first, then check prices
./gw2api items search --name "eternity" --rarity legendary
./gw2api commerce prices <item_id_from_search>
```

### 2. Achievement Hunting
```bash
# Get achievement details
./gw2api achievements get 1 --output table

# Check multiple achievements
./gw2api achievements get 1,2,3,4,5 --output table
```

### 3. Server Information
```bash
# Check world populations
./gw2api worlds all --output table

# Check specific servers
./gw2api worlds get 1001,1002,1003 --output table
```

### 4. Currency Information
```bash
# See all available currencies
./gw2api currencies all --output table

# Check specific currencies
./gw2api currencies get 1,2,3 --output table
```

### 5. Item Discovery and Search
```bash
# Find gear for your profession
./gw2api items search --name "berserker" --output table

# Browse by item rarity
./gw2api items search --rarity ascended --limit 20 --output table

# Look for specific weapon types
./gw2api items search --name "greatsword" --output table

# Discover exotic armor pieces
./gw2api items search --rarity exotic --name "armor" --limit 15
```

# Check specific currencies
./gw2api currencies get 1,2,3 --output table
```

## Integration Examples

### Shell Scripting
```bash
#!/bin/bash

# Get current build and store in variable
BUILD=$(./gw2api build | jq -r '.id')
echo "Current GW2 build: $BUILD"

# Check if specific items are profitable
PRICE_DATA=$(./gw2api commerce prices 19684)
BUY_PRICE=$(echo $PRICE_DATA | jq -r '.buys.unit_price')
SELL_PRICE=$(echo $PRICE_DATA | jq -r '.sells.unit_price')

if [ $SELL_PRICE -gt $BUY_PRICE ]; then
    echo "Item 19684 is profitable!"
fi
```

### Python Integration
```python
import subprocess
import json

def get_item_price(item_id):
    result = subprocess.run(['./gw2api', 'commerce', 'prices', str(item_id)], 
                          capture_output=True, text=True)
    return json.loads(result.stdout)

# Get current prices
price_data = get_item_price(19684)
print(f"Buy: {price_data['buys']['unit_price']}, Sell: {price_data['sells']['unit_price']}")
```

## Performance Notes

- **Bulk Operations**: Always prefer getting multiple items in one command rather than multiple single-item commands
- **Caching**: The CLI doesn't cache responses - consider implementing caching in your scripts for frequently accessed data
- **Rate Limiting**: Be respectful of API rate limits when scripting

## Troubleshooting

### Common Issues

1. **"no such id" Error**
   ```bash
   # Check if ID exists in the list first
   ./gw2api achievements list | grep 999999
   ```

2. **Timeout Errors**
   ```bash
   # Increase timeout for slow connections
   ./gw2api items get 100 --timeout 60
   ```

3. **Invalid Language Code**
   ```bash
   # Use supported language codes only
   ./gw2api achievements get 1 --lang en  # ✓ Correct
   ./gw2api achievements get 1 --lang jp  # ✗ Not supported
   ```

## Future Enhancements

- Authentication support for account-specific endpoints
- Configuration file support
- Response caching
- Pagination support for large datasets
- Export to CSV/Excel formats
- Interactive mode

## Contributing

This CLI is built on top of the typed GW2 API Go library. To add new commands:

1. Add the API method to `gw2api.go`
2. Add the CLI command to `cmd.go`
3. Update the help documentation
4. Test with both JSON and table output formats