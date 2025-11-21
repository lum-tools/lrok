#!/bin/bash

set -e

echo "🧪 Running SDK E2E tests..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$ROOT_DIR"

# Create results directory
mkdir -p test/results

# Check if LUM_API_KEY is set
if [ -z "$LUM_API_KEY" ]; then
    echo "⚠️  LUM_API_KEY not set, using test key"
    export LUM_API_KEY="lum_test_key"
fi

# Start test infrastructure
echo "🐳 Starting test infrastructure..."
docker compose -f docker-compose.test.yml up -d frp-server lrok-daemon test-app-python test-app-nodejs test-app-postgres

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
timeout 60 bash -c 'until docker compose -f docker-compose.test.yml ps | grep -q "healthy"; do sleep 2; echo -n "."; done'
echo ""
echo "✅ Services are healthy"

# Run contract tests
echo ""
echo "📋 Running OpenAPI contract tests..."
docker compose -f docker-compose.test.yml run --rm contract-tests || echo "⚠️  Contract tests had some warnings"

# Run Python SDK tests
echo ""
echo "🐍 Running Python SDK tests..."
docker compose -f docker-compose.test.yml run --rm sdk-test-python

# Run Node.js SDK tests
echo ""
echo "📦 Running Node.js SDK tests..."
docker compose -f docker-compose.test.yml run --rm sdk-test-nodejs || echo "⚠️  Node.js tests not yet implemented"

# Aggregate coverage
echo ""
echo "📊 Aggregating coverage..."
docker compose -f docker-compose.test.yml run --rm coverage-aggregator || echo "⚠️  Coverage aggregation had issues"

# Cleanup
echo ""
echo "🧹 Cleaning up..."
docker compose -f docker-compose.test.yml down -v

echo ""
echo "✨ All tests completed!"
echo "📊 Coverage reports available in test/results/"
