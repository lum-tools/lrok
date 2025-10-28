package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultServerAddr = "142.132.245.5"
	DefaultServerPort = 7000
)

// TunnelConfig represents the configuration for a tunnel
type TunnelConfig struct {
	ServerAddr      string
	ServerPort      int
	APIKey          string
	LocalPort       int
	LocalIP         string
	Subdomain       string
	ProxyType       string // http, tcp, stcp, xtcp
	RemotePort      int    // For TCP tunnels
	SecretKey       string // For STCP/XTCP tunnels
	BandwidthLimit  string // e.g., "1MB", "500KB"
	UseEncryption   bool
	UseCompression  bool
	HealthCheckType string // tcp, http
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
	if cfg.ProxyType == "" {
		cfg.ProxyType = "http" // Default to HTTP for backward compatibility
	}

	// Build metadata section
	metadataLines := []string{
		fmt.Sprintf(`metadatas.api_key = "%s"`, cfg.APIKey),
		fmt.Sprintf(`metadatas.local_port = "%d"`, cfg.LocalPort),
		fmt.Sprintf(`metadatas.proxy_type = "%s"`, cfg.ProxyType),
	}
	
	if cfg.RemotePort > 0 {
		metadataLines = append(metadataLines, fmt.Sprintf(`metadatas.remote_port = "%d"`, cfg.RemotePort))
	}
	if cfg.SecretKey != "" {
		metadataLines = append(metadataLines, fmt.Sprintf(`metadatas.secret_key = "%s"`, cfg.SecretKey))
	}
	if cfg.BandwidthLimit != "" {
		metadataLines = append(metadataLines, fmt.Sprintf(`metadatas.bandwidth_limit = "%s"`, cfg.BandwidthLimit))
	}
	if cfg.UseEncryption {
		metadataLines = append(metadataLines, `metadatas.use_encryption = "true"`)
	}
	if cfg.UseCompression {
		metadataLines = append(metadataLines, `metadatas.use_compression = "true"`)
	}
	if cfg.HealthCheckType != "" {
		metadataLines = append(metadataLines, fmt.Sprintf(`metadatas.health_check_type = "%s"`, cfg.HealthCheckType))
	}

	// Build proxy configuration based on type
	var proxyConfig string
	switch cfg.ProxyType {
	case "http", "https":
		proxyConfig = fmt.Sprintf(`[[proxies]]
name = "tunnel-%s"
type = "%s"
localIP = "%s"
localPort = %d
subdomain = "%s"`,
			cfg.Subdomain, cfg.ProxyType, cfg.LocalIP, cfg.LocalPort, cfg.Subdomain)
		
	case "tcp":
		if cfg.RemotePort == 0 {
			return "", fmt.Errorf("remote port is required for %s tunnels", cfg.ProxyType)
		}
		proxyConfig = fmt.Sprintf(`[[proxies]]
name = "tunnel-%s"
type = "%s"
localIP = "%s"
localPort = %d
remotePort = %d`,
			cfg.Subdomain, cfg.ProxyType, cfg.LocalIP, cfg.LocalPort, cfg.RemotePort)
		
	case "stcp", "xtcp":
		if cfg.SecretKey == "" {
			return "", fmt.Errorf("secret key is required for %s tunnels", cfg.ProxyType)
		}
		proxyConfig = fmt.Sprintf(`[[proxies]]
name = "tunnel-%s"
type = "%s"
secretKey = "%s"
localIP = "%s"
localPort = %d`,
			cfg.Subdomain, cfg.ProxyType, cfg.SecretKey, cfg.LocalIP, cfg.LocalPort)
		
	default:
		return "", fmt.Errorf("unsupported proxy type: %s", cfg.ProxyType)
	}

	// Add transport options if specified
	if cfg.BandwidthLimit != "" || cfg.UseEncryption || cfg.UseCompression {
		proxyConfig += "\n\n[proxies.transport]"
		if cfg.BandwidthLimit != "" {
			proxyConfig += fmt.Sprintf("\nbandwidthLimit = \"%s\"", cfg.BandwidthLimit)
		}
		if cfg.UseEncryption {
			proxyConfig += "\nuseEncryption = true"
		}
		if cfg.UseCompression {
			proxyConfig += "\nuseCompression = true"
		}
	}

	// Add health check if specified
	if cfg.HealthCheckType != "" {
		proxyConfig += fmt.Sprintf("\n\n[proxies.healthCheck]\ntype = \"%s\"", cfg.HealthCheckType)
		if cfg.HealthCheckType == "tcp" {
			proxyConfig += "\ntimeoutSeconds = 3\nmaxFailed = 3\nintervalSeconds = 10"
		} else if cfg.HealthCheckType == "http" {
			proxyConfig += "\npath = \"/health\"\ntimeoutSeconds = 3\nmaxFailed = 3\nintervalSeconds = 10"
		}
	}

	content := fmt.Sprintf(`# Auto-generated frpc configuration
# Powered by lum.tools platform
serverAddr = "%s"
serverPort = %d

log.level = "info"

# Pass configuration in metadata for plugin authentication and tracking
%s

%s
`,
		cfg.ServerAddr,
		cfg.ServerPort,
		strings.Join(metadataLines, "\n"),
		proxyConfig,
	)

	// Create temp file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, fmt.Sprintf("lrok-%s.toml", cfg.Subdomain))

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return configPath, nil
}

// GenerateVisitorTOML creates a frpc visitor TOML configuration file and returns the path
func GenerateVisitorTOML(cfg *TunnelConfig) (string, error) {
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = DefaultServerAddr
	}
	if cfg.ServerPort == 0 {
		cfg.ServerPort = DefaultServerPort
	}
	if cfg.LocalIP == "" {
		cfg.LocalIP = "127.0.0.1"
	}

	// Build metadata section for visitor
	metadataLines := []string{
		fmt.Sprintf(`metadatas.api_key = "%s"`, cfg.APIKey),
		fmt.Sprintf(`metadatas.proxy_type = "%s"`, cfg.ProxyType),
		fmt.Sprintf(`metadatas.secret_key = "%s"`, cfg.SecretKey),
		fmt.Sprintf(`metadatas.bind_port = "%d"`, cfg.LocalPort),
		fmt.Sprintf(`metadatas.bind_addr = "%s"`, cfg.LocalIP),
	}

	// Build visitor configuration
	visitorConfig := fmt.Sprintf(`[[visitors]]
name = "visitor-%s"
type = "%s"
serverName = "tunnel-%s"
secretKey = "%s"
bindAddr = "%s"
bindPort = %d`,
		cfg.Subdomain, cfg.ProxyType, cfg.Subdomain, cfg.SecretKey, cfg.LocalIP, cfg.LocalPort)

	content := fmt.Sprintf(`# Auto-generated frpc visitor configuration
# Powered by lum.tools platform
serverAddr = "%s"
serverPort = %d

log.level = "info"

# Pass configuration in metadata for plugin authentication and tracking
%s

%s
`,
		cfg.ServerAddr,
		cfg.ServerPort,
		strings.Join(metadataLines, "\n"),
		visitorConfig,
	)

	// Create temp file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, fmt.Sprintf("lrok-visitor-%s.toml", cfg.Subdomain))

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("failed to write visitor config file: %w", err)
	}

	return configPath, nil
}

