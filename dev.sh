#!/bin/bash

# Development script for Maukemana backend
# Uses Air for live reloading

echo "🚀 Starting Maukemana backend with live reloading..."
echo "📝 Air configuration: .air.toml"
echo "🔄 Will auto-reload on .go file changes"
echo ""

# Check if Air is installed
if ! command -v air &> /dev/null && ! command -v ~/go/bin/air &> /dev/null; then
    echo "❌ Air not found. Installing..."
    go install github.com/air-verse/air@latest
    echo "✅ Air installed!"
    echo ""
fi

# Start Air
if command -v air &> /dev/null; then
    air
else
    ~/go/bin/air
fi