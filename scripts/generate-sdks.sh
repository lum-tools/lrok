#!/bin/bash

set -e

echo "🚀 Generating lrok SDKs from OpenAPI spec..."

SPEC_FILE="api/openapi.yaml"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$ROOT_DIR"

# Check if OpenAPI Generator CLI is installed
if ! command -v openapi-generator-cli &> /dev/null; then
    echo "📦 OpenAPI Generator CLI not found. Installing via npm..."
    npm install -g @openapitools/openapi-generator-cli
fi

# Validate OpenAPI spec
echo "✅ Validating OpenAPI spec..."
openapi-generator-cli validate -i "$SPEC_FILE"

# Clean previous SDK directories
echo "🧹 Cleaning previous SDK directories..."
rm -rf sdk/python sdk/nodejs sdk/ruby sdk/go

# Generate Python SDK
echo "🐍 Generating Python SDK..."
openapi-generator-cli generate \
  -i "$SPEC_FILE" \
  -g python \
  -o sdk/python \
  --additional-properties=packageName=lrok,projectName=lrok-python,packageVersion=1.0.0

# Generate Node.js/TypeScript SDK
echo "📦 Generating Node.js/TypeScript SDK..."
openapi-generator-cli generate \
  -i "$SPEC_FILE" \
  -g typescript-axios \
  -o sdk/nodejs \
  --additional-properties=npmName=lrok,npmVersion=1.0.0,supportsES6=true

# Generate Ruby SDK
echo "💎 Generating Ruby SDK..."
openapi-generator-cli generate \
  -i "$SPEC_FILE" \
  -g ruby \
  -o sdk/ruby \
  --additional-properties=gemName=lrok,gemVersion=1.0.0

# Generate Go SDK
echo "🐹 Generating Go SDK..."
openapi-generator-cli generate \
  -i "$SPEC_FILE" \
  -g go \
  -o sdk/go \
  --additional-properties=packageName=lrok,packageVersion=1.0.0

echo ""
echo "✨ All SDKs generated successfully!"
echo ""
echo "Generated SDKs:"
echo "  - Python:     sdk/python"
echo "  - Node.js:    sdk/nodejs"
echo "  - Ruby:       sdk/ruby"
echo "  - Go:         sdk/go"
echo ""
echo "Next steps:"
echo "  1. Install dependencies for each SDK"
echo "  2. Run tests: ./scripts/test-sdks.sh"
echo "  3. Build packages: ./scripts/build-sdks.sh"
