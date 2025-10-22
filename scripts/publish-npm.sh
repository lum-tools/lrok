#!/bin/bash
set -e

# This script publishes the npm package after a GitHub release

if [ -z "$NPM_TOKEN" ]; then
    echo "❌ NPM_TOKEN not set"
    exit 1
fi

# Get version from git tag
VERSION=${GITHUB_REF#refs/tags/v}
if [ -z "$VERSION" ]; then
    echo "❌ No version tag found"
    exit 1
fi

echo "📦 Publishing lrok v${VERSION} to npm..."

cd packaging/npm

# Update package.json version
sed -i "s/\"version\": \".*\"/\"version\": \"${VERSION}\"/" package.json

# Configure npm authentication
echo "//registry.npmjs.org/:_authToken=${NPM_TOKEN}" > .npmrc

# Copy LICENSE from root
cp ../../LICENSE .

# Publish
npm publish --access public

echo "✅ Published to npm: lrok@${VERSION}"

