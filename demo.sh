#!/bin/bash

# Guild Wars 2 API CLI Demo Script
# Demonstrates various CLI capabilities

echo "=== Guild Wars 2 API CLI Demo ==="
echo

echo "1. Current Game Build:"
./gw2api build --output table
echo

echo "2. Sample Achievement (English):"
./gw2api achievements get 1 --output table
echo

echo "3. Same Achievement in German:"
./gw2api achievements get 1 --lang de --output table
echo

echo "4. Multiple Currencies:"
./gw2api currencies get 1,2,3 --output table
echo

echo "5. Sample Worlds:"
./gw2api worlds get 1001,1002,1003 --output table
echo

echo "6. Trading Post Prices:"
./gw2api commerce prices 19684,19709 --output table
echo

echo "7. Sample Items:"
./gw2api items get 100,200 --output table
echo

echo "8. JSON Output Example (Achievement):"
./gw2api achievements get 1
echo

echo "=== Demo Complete ==="
echo "Try running commands yourself:"
echo "  ./gw2api --help"
echo "  ./gw2api achievements get 1 --output table"
echo "  ./gw2api currencies all --output table"