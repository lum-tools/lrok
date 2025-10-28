package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/lum-tools/lrok/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTunnelConfig(t *testing.T) {
	// Test configuration generation without network connectivity
	localPort := 8080
	remotePort := 15000
	tunnelName := "test-tunnel"

	// Test different tunnel types
	tunnelTypes := []string{"http", "tcp", "stcp", "xtcp"}
	
	for _, tunnelType := range tunnelTypes {
		t.Run(tunnelType, func(t *testing.T) {
			cfg := &config.TunnelConfig{
				APIKey:     TestAPIKey,
				LocalPort:  localPort,
				LocalIP:    "127.0.0.1",
				Subdomain:  tunnelName,
				ProxyType:  tunnelType,
				RemotePort: remotePort,
				SecretKey:  "test-secret-key",
			}

			// Generate TOML configuration
			configPath, err := config.GenerateTOML(cfg)
			require.NoError(t, err)
			defer os.Remove(configPath)

			// Verify config file exists
			_, err = os.Stat(configPath)
			require.NoError(t, err)

			// Read and verify content
			content, err := os.ReadFile(configPath)
			require.NoError(t, err)

			contentStr := string(content)

			// Basic assertions
			assert.Contains(t, contentStr, fmt.Sprintf(`name = "tunnel-%s"`, tunnelName))
			assert.Contains(t, contentStr, fmt.Sprintf(`type = "%s"`, tunnelType))
			assert.Contains(t, contentStr, fmt.Sprintf(`metadatas.api_key = "%s"`, TestAPIKey))
			assert.Contains(t, contentStr, fmt.Sprintf(`metadatas.local_port = "%d"`, localPort))
			assert.Contains(t, contentStr, fmt.Sprintf(`metadatas.proxy_type = "%s"`, tunnelType))

			// Type-specific assertions
			switch tunnelType {
			case "tcp":
				assert.Contains(t, contentStr, fmt.Sprintf(`remotePort = %d`, remotePort))
			case "stcp", "xtcp":
				assert.Contains(t, contentStr, fmt.Sprintf(`secretKey = "%s"`, cfg.SecretKey))
				assert.Contains(t, contentStr, fmt.Sprintf(`metadatas.secret_key = "%s"`, cfg.SecretKey))
			}

			t.Logf("✅ %s tunnel config test passed", tunnelType)
		})
	}
}

func TestVisitorConfig(t *testing.T) {
	// Test visitor configuration generation
	tunnelName := "test-visitor"
	localPort := 8080
	secretKey := "visitor-secret"

	cfg := &config.TunnelConfig{
		APIKey:     TestAPIKey,
		LocalPort:  localPort,
		LocalIP:    "127.0.0.1",
		Subdomain:  tunnelName,
		ProxyType:  "stcp",
		SecretKey:  secretKey,
	}

	// Generate visitor TOML configuration
	configPath, err := config.GenerateVisitorTOML(cfg)
	require.NoError(t, err)
	defer os.Remove(configPath)

	// Verify config file exists
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Read and verify content
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	contentStr := string(content)

	// Visitor-specific assertions
	assert.Contains(t, contentStr, `[[visitors]]`)
	assert.Contains(t, contentStr, fmt.Sprintf(`name = "visitor-%s"`, tunnelName))
	assert.Contains(t, contentStr, fmt.Sprintf(`serverName = "tunnel-%s"`, tunnelName))
	assert.Contains(t, contentStr, fmt.Sprintf(`secretKey = "%s"`, secretKey))
	assert.Contains(t, contentStr, fmt.Sprintf(`bindAddr = "%s"`, cfg.LocalIP))
	assert.Contains(t, contentStr, fmt.Sprintf(`bindPort = %d`, localPort))

	t.Logf("✅ Visitor config test passed")
}
