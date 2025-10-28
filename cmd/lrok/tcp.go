package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/lum-tools/lrok/internal/config"
	"github.com/lum-tools/lrok/internal/names"
	"github.com/lum-tools/lrok/internal/tunnel"
	"github.com/spf13/cobra"
)

var (
	tcpRemotePort    int
	tcpEncrypt       bool
	tcpCompress      bool
	tcpHealthCheck   bool
	tcpBandwidthLimit string
)

var tcpCmd = &cobra.Command{
	Use:   "tcp <local-port>",
	Short: "Create TCP tunnel for direct port forwarding",
	Long: `Create a TCP tunnel to expose a local port directly to the internet.

TCP tunnels provide direct port forwarding without HTTP/HTTPS overhead.
Perfect for databases, SSH, Redis, and other TCP-based services.

Examples:
  lrok tcp 5432 --remote-port 10001    # Expose PostgreSQL on port 10001
  lrok tcp 22 --remote-port 10002      # Expose SSH on port 10002
  lrok tcp 6379 --remote-port 10003    # Expose Redis on port 10003
  lrok tcp 3000 --remote-port 10004 --encrypt --compress  # With encryption and compression

Connection:
  Connect to: frp.lum.tools:<remote-port>
  Example:   psql -h frp.lum.tools -p 10001 -U myuser mydb`,
	Args: cobra.ExactArgs(1),
	RunE: runTCPTunnel,
}

func init() {
	tcpCmd.Flags().IntVar(&tcpRemotePort, "remote-port", 0, "Remote port on server (required)")
	tcpCmd.Flags().StringVarP(&name, "name", "n", "", "Custom tunnel name")
	tcpCmd.Flags().StringVar(&subdomain, "subdomain", "", "Alias for --name")
	tcpCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "lum.tools platform API key")
	tcpCmd.Flags().StringVar(&localIP, "ip", "127.0.0.1", "Local IP address to bind to")
	tcpCmd.Flags().BoolVar(&tcpEncrypt, "encrypt", false, "Enable encryption")
	tcpCmd.Flags().BoolVar(&tcpCompress, "compress", false, "Enable compression")
	tcpCmd.Flags().BoolVar(&tcpHealthCheck, "health-check", false, "Enable TCP health checks")
	tcpCmd.Flags().StringVar(&tcpBandwidthLimit, "bandwidth", "", "Bandwidth limit (e.g., 1MB, 500KB)")
	
	// Mark remote-port as required
	tcpCmd.MarkFlagRequired("remote-port")
}

func runTCPTunnel(cmd *cobra.Command, args []string) error {
	// Get port from args
	localPort, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid local port: %s", args[0])
	}

	// Validate local port
	if err := tunnel.ValidatePort(localPort); err != nil {
		return fmt.Errorf("invalid local port: %w", err)
	}

	// Validate remote port
	if err := tunnel.ValidatePort(tcpRemotePort); err != nil {
		return fmt.Errorf("invalid remote port: %w", err)
	}

	// Validate bandwidth limit if provided
	if tcpBandwidthLimit != "" {
		if err := tunnel.ValidateBandwidthLimit(tcpBandwidthLimit); err != nil {
			return fmt.Errorf("invalid bandwidth limit: %w", err)
		}
	}

	// Get API key with priority: flag > env var > config file
	if apiKey == "" {
		apiKey = os.Getenv("LUM_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("FRP_API_KEY") // Legacy support
		}
	}

	// Try config file if still not found
	if apiKey == "" {
		if key, err := config.GetAPIKey(); err == nil {
			apiKey = key
		}
	}

	if apiKey == "" {
		return fmt.Errorf(`❌ No API key configured!

You need a lum.tools platform API key to use lrok.

📝 Get your API key:
   1. Visit: https://platform.lum.tools/keys
   2. Login with your account
   3. Create a new API key
   4. Copy your API key (starts with 'lum_')

💡 Save it with login command (recommended):
   lrok login lum_your_api_key_here

Or use environment variable:
   export LUM_API_KEY='lum_your_api_key_here'

Or pass it directly:
   lrok tcp 5432 --remote-port 10001 --api-key lum_your_key`)
	}

	if !strings.HasPrefix(apiKey, "lum_") {
		fmt.Println("⚠️  Warning: API key should start with 'lum_'")
		fmt.Println("   Make sure you're using a valid platform API key from https://platform.lum.tools/keys")
	}

	// Determine tunnel name
	tunnelName := name
	if subdomain != "" {
		tunnelName = subdomain
	}
	if tunnelName == "" {
		tunnelName = names.Generate()
	}

	// Validate tunnel name
	if err := tunnel.ValidateTunnelName(tunnelName); err != nil {
		return fmt.Errorf("invalid tunnel name: %w", err)
	}

	// Determine health check type
	healthCheckType := ""
	if tcpHealthCheck {
		healthCheckType = "tcp"
	}

	// Generate config
	cfg := &config.TunnelConfig{
		APIKey:          apiKey,
		LocalPort:       localPort,
		LocalIP:         localIP,
		Subdomain:       tunnelName,
		ProxyType:       "tcp",
		RemotePort:      tcpRemotePort,
		BandwidthLimit:  tcpBandwidthLimit,
		UseEncryption:   tcpEncrypt,
		UseCompression:  tcpCompress,
		HealthCheckType: healthCheckType,
	}

	configPath, err := config.GenerateTOML(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	fmt.Println("🚀 Starting TCP tunnel...")
	fmt.Println("⏳ Connecting to frp.lum.tools...")
	fmt.Printf("📍 Local:      %s:%d\n", localIP, localPort)
	fmt.Printf("🌐 Remote:     frp.lum.tools:%d\n", tcpRemotePort)
	fmt.Printf("🏷️  Name:       %s\n", tunnelName)
	if tcpEncrypt {
		fmt.Println("🔒 Encryption: enabled")
	}
	if tcpCompress {
		fmt.Println("🗜️  Compression: enabled")
	}
	if tcpHealthCheck {
		fmt.Println("💓 Health check: enabled")
	}
	if tcpBandwidthLimit != "" {
		fmt.Printf("📊 Bandwidth: %s\n", tcpBandwidthLimit)
	}
	fmt.Println()

	// Start tunnel
	mgr := tunnel.New(configPath)
	defer mgr.Cleanup()

	return mgr.StartWithGracefulShutdown()
}
