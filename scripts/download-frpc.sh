#!/bin/bash
set -e

# Download frpc binaries for all platforms
VERSION="0.65.0"
BASE_URL="https://github.com/fatedier/frp/releases/download/v${VERSION}"
BIN_DIR="internal/embed/bins"

mkdir -p "$BIN_DIR"

echo "📦 Downloading frpc v${VERSION} for all platforms..."

# Platform mappings: frp_format -> our_format
declare -A platforms=(
    ["darwin_amd64"]="frpc_darwin_amd64"
    ["darwin_arm64"]="frpc_darwin_arm64"
    ["linux_amd64"]="frpc_linux_amd64"
    ["linux_arm64"]="frpc_linux_arm64"
    ["windows_amd64"]="frpc_windows_amd64.exe"
)

for frp_platform in "${!platforms[@]}"; do
    output_name="${platforms[$frp_platform]}"
    
    echo "  → ${frp_platform}..."
    
    # Determine archive extension
    if [[ "$frp_platform" == windows* ]]; then
        ext="zip"
    else
        ext="tar.gz"
    fi
    
    archive="frp_${VERSION}_${frp_platform}.${ext}"
    url="${BASE_URL}/${archive}"
    
    # Download
    curl -sL "$url" -o "/tmp/${archive}"
    
    # Extract frpc binary
    if [[ "$ext" == "zip" ]]; then
        unzip -q "/tmp/${archive}" -d /tmp/
        binary_path="/tmp/frp_${VERSION}_${frp_platform}/frpc.exe"
    else
        tar -xzf "/tmp/${archive}" -C /tmp/
        binary_path="/tmp/frp_${VERSION}_${frp_platform}/frpc"
    fi
    
    # Move to bins directory with our naming
    mv "$binary_path" "${BIN_DIR}/${output_name}"
    chmod +x "${BIN_DIR}/${output_name}" 2>/dev/null || true
    
    # Cleanup
    rm -rf "/tmp/${archive}" "/tmp/frp_${VERSION}_${frp_platform}"
    
    echo "  ✅ ${output_name}"
done

echo ""
echo "✅ All frpc binaries downloaded to ${BIN_DIR}/"
ls -lh "$BIN_DIR"

