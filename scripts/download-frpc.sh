#!/usr/bin/env bash
set -e

# Download frpc binaries for all platforms
VERSION="0.65.0"
BASE_URL="https://github.com/fatedier/frp/releases/download/v${VERSION}"
BIN_DIR="internal/embed/bins"

mkdir -p "$BIN_DIR"

echo "📦 Downloading frpc v${VERSION} for all platforms..."

# Function to download and extract a platform
download_platform() {
    local frp_platform=$1
    local output_name=$2
    local ext=$3
    
    echo "  → ${frp_platform}..."
    
    local archive="frp_${VERSION}_${frp_platform}.${ext}"
    local url="${BASE_URL}/${archive}"
    
    # Download
    curl -sL "$url" -o "/tmp/${archive}"
    
    # Extract frpc binary
    if [ "$ext" = "zip" ]; then
        unzip -q "/tmp/${archive}" -d /tmp/
        local binary_path="/tmp/frp_${VERSION}_${frp_platform}/frpc.exe"
    else
        tar -xzf "/tmp/${archive}" -C /tmp/
        local binary_path="/tmp/frp_${VERSION}_${frp_platform}/frpc"
    fi
    
    # Move to bins directory with our naming
    mv "$binary_path" "${BIN_DIR}/${output_name}"
    chmod +x "${BIN_DIR}/${output_name}" 2>/dev/null || true
    
    # Cleanup
    rm -rf "/tmp/${archive}" "/tmp/frp_${VERSION}_${frp_platform}"
    
    echo "  ✅ ${output_name}"
}

# Download all platforms
download_platform "windows_amd64" "frpc_windows_amd64.exe" "zip"
download_platform "linux_amd64" "frpc_linux_amd64" "tar.gz"
download_platform "darwin_amd64" "frpc_darwin_amd64" "tar.gz"
download_platform "darwin_arm64" "frpc_darwin_arm64" "tar.gz"
download_platform "linux_arm64" "frpc_linux_arm64" "tar.gz"

echo ""
echo "✅ All frpc binaries downloaded to ${BIN_DIR}/"
ls -lh "$BIN_DIR"

