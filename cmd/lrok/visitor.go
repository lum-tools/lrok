package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/lum-tools/lrok/internal/config"
	"github.com/lum-tools/lrok/internal/tunnel"
	"github.com/spf13/cobra"
)

var (
	visitorType      string
	visitorSecretKey string
	visitorBindPort  int
	visitorBindAddr  string
)

var visitorCmd = &cobra.Command{
	Use:   "visitor <tunnel-name>",
	Short: "Connect to STCP/XTCP tunnel as visitor",
	Long: `Connect to an STCP or XTCP tunnel as a visitor.

This command allows you to connect to secure tunnels created by others.
You need the tunnel name and the pre-shared secret key.

Examples:
  lrok visitor my-db-tunnel --type stcp --secret-key my-secret --bind-port 5432
  lrok visitor my-app-tunnel --type xtcp --secret-key p2p-key --bind-port 8080

Connection:
  Local service will be accessible on 127.0.0.1:<bind-port>
  Example: psql -h 127.0.0.1 -p 5432 -U myuser mydb`,
	Args: cobra.ExactArgs(1),
	RunE: runVisitor,
}

func init() {
	visitorCmd.Flags().StringVar(&visitorType, "type", "", "Tunnel type: stcp or xtcp (required)")
	visitorCmd.Flags().StringVar(&visitorSecretKey, "secret-key", "", "Pre-shared secret key (required)")
	visitorCmd.Flags().IntVar(&visitorBindPort, "bind-port", 0, "Local port to bind to (required)")
	visitorCmd.Flags().StringVar(&visitorBindAddr, "bind-addr", "127.0.0.1", "Local address to bind to")
	visitorCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "lum.tools platform API key")
	
	// Mark required flags
	visitorCmd.MarkFlagRequired("type")
	visitorCmd.MarkFlagRequired("secret-key")
	visitorCmd.MarkFlagRequired("bind-port")
}

func runVisitor(cmd *cobra.Command, args []string) error {
	tunnelName := args[0]

	// Validate tunnel type
	if err := tunnel.ValidateProxyType(visitorType); err != nil {
		return fmt.Errorf("invalid tunnel type: %w", err)
	}

	// Only allow stcp and xtcp for visitors
	if visitorType != "stcp" && visitorType != "xtcp" {
		return fmt.Errorf("visitor only supports stcp and xtcp tunnel types, got %s", visitorType)
	}

	// Validate secret key
	if err := tunnel.ValidateSecretKey(visitorSecretKey); err != nil {
		return fmt.Errorf("invalid secret key: %w", err)
	}

	// Validate bind port
	if err := tunnel.ValidatePort(visitorBindPort); err != nil {
		return fmt.Errorf("invalid bind port: %w", err)
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
   lrok visitor my-tunnel --type stcp --secret-key my-secret --bind-port 5432 --api-key lum_your_key`)
	}

	if !strings.HasPrefix(apiKey, "lum_") {
		fmt.Println("⚠️  Warning: API key should start with 'lum_'")
		fmt.Println("   Make sure you're using a valid platform API key from https://platform.lum.tools/keys")
	}

	// Validate tunnel name
	if err := tunnel.ValidateTunnelName(tunnelName); err != nil {
		return fmt.Errorf("invalid tunnel name: %w", err)
	}

	// Generate visitor config
	cfg := &config.TunnelConfig{
		APIKey:     apiKey,
		Subdomain:  tunnelName,
		ProxyType:  visitorType,
		SecretKey:  visitorSecretKey,
		LocalPort:  visitorBindPort,
		LocalIP:    visitorBindAddr,
	}

	configPath, err := config.GenerateVisitorTOML(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate visitor config: %w", err)
	}

	fmt.Printf("🚀 Starting %s visitor...\n", strings.ToUpper(visitorType))
	fmt.Println("⏳ Connecting to frp.lum.tools...")
	fmt.Printf("🔗 Tunnel:     %s\n", tunnelName)
	fmt.Printf("📍 Local:      %s:%d\n", visitorBindAddr, visitorBindPort)
	secretDisplay := visitorSecretKey
	if len(secretDisplay) > 8 {
		secretDisplay = secretDisplay[:8]
	}
	fmt.Printf("🔐 Secret:     %s...\n", secretDisplay)
	fmt.Println()
	
	if visitorType == "xtcp" {
		fmt.Println("⚡ P2P Mode: Attempting direct connection")
		fmt.Println("ℹ️  If P2P fails, connection will fall back to server relay")
	} else {
		fmt.Println("🔒 Secure Mode: Connection encrypted with secret key")
	}
	fmt.Println()

	// Start tunnel
	mgr := tunnel.New(configPath)
	defer mgr.Cleanup()

	return mgr.StartWithGracefulShutdown()
}

