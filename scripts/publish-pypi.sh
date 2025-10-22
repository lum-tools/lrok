#!/bin/bash
set -e

# This script publishes the PyPI package after a GitHub release

if [ -z "$PYPI_TOKEN" ]; then
    echo "❌ PYPI_TOKEN not set"
    exit 1
fi

# Get version from git tag
VERSION=${GITHUB_REF#refs/tags/v}
if [ -z "$VERSION" ]; then
    echo "❌ No version tag found"
    exit 1
fi

echo "📦 Publishing lrok v${VERSION} to PyPI..."

cd packaging/pypi

# Update setup.py version
sed -i "s/VERSION = \".*\"/VERSION = \"${VERSION}\"/" setup.py

# Copy LICENSE from root
cp ../../LICENSE .

# Install build dependencies
pip install build twine

# Build package
python -m build

# Configure PyPI authentication
cat > ~/.pypirc << EOF
[pypi]
username = __token__
password = ${PYPI_TOKEN}
EOF

# Publish
python -m twine upload dist/*

echo "✅ Published to PyPI: lrok==${VERSION}"

