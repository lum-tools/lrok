#!/bin/bash
set -e

# Debug script to test npm publication locally

echo "🔍 Debugging npm publication..."

# Check if we're in the right directory
if [ ! -f "packaging/npm/package.json" ]; then
    echo "❌ Not in lrok root directory"
    exit 1
fi

cd packaging/npm

echo "📦 Testing npm package..."

# Check package.json
echo "Package.json contents:"
cat package.json | jq '.'

# Test npm pack (dry run)
echo "Testing npm pack..."
npm pack --dry-run

# Check if LICENSE exists
if [ ! -f "LICENSE" ]; then
    echo "❌ LICENSE file not found"
    exit 1
fi

echo "✅ npm package structure looks good"

# Test npm publish (dry run)
echo "Testing npm publish (dry run)..."
npm publish --dry-run

echo "✅ npm publication test passed"

