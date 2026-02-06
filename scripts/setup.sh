#!/bin/bash
# scripts/setup.sh - Development environment setup

set -e

echo "🛡️  Setting up Sentinel-Remediator development environment..."

# Check prerequisites
command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed."; exit 1; }
command -v node >/dev/null 2>&1 || { echo "❌ Node.js is required but not installed."; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed."; exit 1; }

echo "✅ Prerequisites check passed"

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod download
go mod tidy

# Install dashboard dependencies
echo "📦 Installing dashboard dependencies..."
cd dashboard
npm install
cd ..

# Create .env if not exists
if [ ! -f .env ]; then
    echo "📝 Creating .env from template..."
    cp .env.example .env
    echo "⚠️  Please edit .env and add your API keys"
fi

# Create work directory
mkdir -p /tmp/sentinel

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Edit .env with your API keys (ANTHROPIC_API_KEY, GITHUB_TOKEN)"
echo "  2. Run 'make dev' to start the backend"
echo "  3. Run 'cd dashboard && npm run dev' for the frontend"
echo ""
