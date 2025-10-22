#!/usr/bin/env bash
set -e

# Download frpc binaries for all platforms
VERSION="0.65.0"
BASE_URL="https://github.com/fatedier/frp/releases/download/v${VERSION}"
BIN_DIR="internal/embed/bins"

mkdir -p "$BIN_DIR"

echo "📦 Downloading frpc v${VERSION} for Linux platforms only..."

# Function to download and extract a platform
download_platform() {
    local frp_platform=$1
    local output_name=$2
    
    echo "  → ${frp_platform}..."
    
    local archive="frp_${VERSION}_${frp_platform}.tar.gz"
    local url="${BASE_URL}/${archive}"
    
    # Download
    curl -sL "$url" -o "/tmp/${archive}"
    
    # Extract frpc binary
    tar -xzf "/tmp/${archive}" -C /tmp/
    local binary_path="/tmp/frp_${VERSION}_${frp_platform}/frpc"
    
    # Move to bins directory with our naming
    mv "$binary_path" "${BIN_DIR}/${output_name}"
    chmod +x "${BIN_DIR}/${output_name}"
    
    # Cleanup
    rm -rf "/tmp/${archive}" "/tmp/frp_${VERSION}_${frp_platform}"
    
    echo "  ✅ ${output_name}"
}

# Download Linux platforms only (free GitHub runners)
download_platform "linux_amd64" "frpc_linux_amd64"
download_platform "linux_arm64" "frpc_linux_arm64"

echo ""
echo "✅ All frpc binaries downloaded to ${BIN_DIR}/"
ls -lh "$BIN_DIR"

