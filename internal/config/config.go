package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultServerAddr = "142.132.245.5"
	DefaultServerPort = 7000
)

// TunnelConfig represents the configuration for a tunnel
type TunnelConfig struct {
	ServerAddr string
	ServerPort int
	APIKey     string
	LocalPort  int
	LocalIP    string
	Subdomain  string
}

// GenerateTOML creates a frpc TOML configuration file and returns the path
func GenerateTOML(cfg *TunnelConfig) (string, error) {
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = DefaultServerAddr
	}
	if cfg.ServerPort == 0 {
		cfg.ServerPort = DefaultServerPort
	}
	if cfg.LocalIP == "" {
		cfg.LocalIP = "127.0.0.1"
	}

	content := fmt.Sprintf(`# Auto-generated frpc configuration
# Powered by lum.tools platform
serverAddr = "%s"
serverPort = %d

log.level = "info"

# Pass API key and local port in metadata for plugin authentication and tracking
metadatas.api_key = "%s"
metadatas.local_port = "%d"

[[proxies]]
name = "tunnel-%s"
type = "http"
localIP = "%s"
localPort = %d
subdomain = "%s"
`,
		cfg.ServerAddr,
		cfg.ServerPort,
		cfg.APIKey,
		cfg.LocalPort,
		cfg.Subdomain,
		cfg.LocalIP,
		cfg.LocalPort,
		cfg.Subdomain,
	)

	// Create temp file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, fmt.Sprintf("lrok-%s.toml", cfg.Subdomain))

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return configPath, nil
}

