package embed

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//go:embed bins/frpc_*
var binaries embed.FS

// GetFrpcPath extracts the embedded frpc binary to a temp directory and returns the path
func GetFrpcPath() (string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	
	// Only Linux binaries are embedded to reduce size and CI costs
	var binaryName string
	switch {
	case goos == "linux" && goarch == "amd64":
		binaryName = "frpc_linux_amd64"
	case goos == "linux" && goarch == "arm64":
		binaryName = "frpc_linux_arm64"
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s - only Linux (amd64/arm64) is currently supported", goos, goarch)
	}
	
	// Read embedded binary
	embeddedPath := filepath.Join("bins", binaryName)
	data, err := binaries.ReadFile(embeddedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded binary %s: %w", binaryName, err)
	}
	
	// Extract to temp directory
	tmpDir := os.TempDir()
	extractPath := filepath.Join(tmpDir, "lrok-frpc", binaryName)
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(extractPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	// Check if binary already exists and is valid
	if _, err := os.Stat(extractPath); err == nil {
		return extractPath, nil
	}
	
	// Write binary to temp location
	if err := os.WriteFile(extractPath, data, 0755); err != nil {
		return "", fmt.Errorf("failed to write binary to temp: %w", err)
	}
	
	return extractPath, nil
}

// CheckFrpcVersion checks if frpc is available and returns its version
func CheckFrpcVersion() (string, error) {
	frpcPath, err := GetFrpcPath()
	if err != nil {
		return "", err
	}
	
	cmd := exec.Command(frpcPath, "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get frpc version: %w", err)
	}
	
	return string(output), nil
}

