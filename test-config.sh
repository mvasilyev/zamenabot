#!/bin/bash

# Test script to validate zamenabot configuration
# This script helps users test their multiple installations setup

echo "=== Zamenabot Configuration Test ==="
echo

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    exit 1
fi
echo "✅ Go is available"

# Build the application
echo "🔨 Building zamenabot..."
if go build -o zamenabot main.go; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi

# Check for configuration files
echo
echo "🔍 Checking configuration files..."

single_env_exists=false
multiple_env_exists=false

if [ -f ".env" ]; then
    echo "✅ Found .env (single installation)"
    single_env_exists=true
fi

env_files=($(ls .env.* 2>/dev/null | grep -v ".example"))
if [ ${#env_files[@]} -gt 0 ]; then
    echo "✅ Found multiple installation files:"
    for file in "${env_files[@]}"; do
        if [[ ! "$file" == *.example ]]; then
            echo "  - $file"
            multiple_env_exists=true
        fi
    done
fi

if [ "$single_env_exists" = false ] && [ "$multiple_env_exists" = false ]; then
    echo "❌ No configuration files found"
    echo "   Create .env for single installation or .env.{name} for multiple installations"
    echo "   Use the provided example files as templates"
    exit 1
fi

# Test the configuration by running the bot for 3 seconds
echo
echo "🧪 Testing configuration (3 second run)..."
echo "   (The bot will start and then be stopped - this is normal)"
echo

timeout 3 ./zamenabot 2>&1 | while IFS= read -r line; do
    if [[ "$line" == *"Missing required configuration"* ]]; then
        echo "❌ Configuration error: $line"
    elif [[ "$line" == *"Loaded installation"* ]]; then
        echo "✅ $line"
    elif [[ "$line" == *"Starting installation"* ]]; then
        echo "✅ $line"
    elif [[ "$line" == *"Error"* ]]; then
        echo "⚠️  $line"
    fi
done

echo
echo "🎉 Configuration test completed!"
echo "   If you see '✅ Loaded installation' messages above, your configuration is working."
echo "   The bot is ready to run with: ./zamenabot"

# Clean up
rm -f zamenabot