package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/lum-tools/lrok/internal/config"
	"github.com/lum-tools/lrok/internal/names"
	"github.com/lum-tools/lrok/internal/tunnel"
	"github.com/lum-tools/lrok/internal/version"
	"github.com/spf13/cobra"
)

var (
	xtcpSecretKey    string
	xtcpStunServer   string
	xtcpBandwidthLimit string
)

var xtcpCmd = &cobra.Command{
	Use:   "xtcp <local-port>",
	Short: "Create P2P tunnel (XTCP) for direct client-to-client connection",
	Long: `Create an XTCP (P2P) tunnel for direct client-to-client connection.

XTCP tunnels establish direct P2P connections when possible, bypassing the server
for data transmission. This provides better performance for high-bandwidth applications.

The tunnel requires a pre-shared secret key and a visitor on the other end.
If P2P connection fails, it falls back to server relay.

Examples:
  lrok xtcp 8080 --secret-key p2p-secret-123     # Expose web server via P2P
  lrok xtcp 22 --secret-key ssh-p2p-key         # Expose SSH via P2P
  lrok xtcp 3000 --secret-key api-p2p --stun-server stun.l.google.com:19302

On client side:
  lrok visitor tunnel-name --type xtcp --secret-key p2p-secret-123 --bind-port 8080`,
	Args: cobra.ExactArgs(1),
	RunE: runXTCPTunnel,
}

func init() {
	xtcpCmd.Flags().StringVar(&xtcpSecretKey, "secret-key", "", "Pre-shared secret key (required)")
	xtcpCmd.Flags().StringVar(&xtcpStunServer, "stun-server", "", "Custom STUN server (optional)")
	xtcpCmd.Flags().StringVarP(&name, "name", "n", "", "Custom tunnel name")
	xtcpCmd.Flags().StringVar(&subdomain, "subdomain", "", "Alias for --name")
	xtcpCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "lum.tools platform API key")
	xtcpCmd.Flags().StringVar(&localIP, "ip", "127.0.0.1", "Local IP address to bind to")
	xtcpCmd.Flags().StringVar(&xtcpBandwidthLimit, "bandwidth", "", "Bandwidth limit (e.g., 1MB, 500KB)")
	
	// Mark secret-key as required
	xtcpCmd.MarkFlagRequired("secret-key")
}

func runXTCPTunnel(cmd *cobra.Command, args []string) error {
	// Check for updates and prompt user before starting tunnel
	if err := version.PromptForUpdate(versionInfo); err != nil {
		return err
	}

	// Get port from args
	localPort, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid local port: %s", args[0])
	}

	// Validate local port
	if err := tunnel.ValidatePort(localPort); err != nil {
		return fmt.Errorf("invalid local port: %w", err)
	}

	// Validate secret key
	if err := tunnel.ValidateSecretKey(xtcpSecretKey); err != nil {
		return fmt.Errorf("invalid secret key: %w", err)
	}

	// Validate bandwidth limit if provided
	if xtcpBandwidthLimit != "" {
		if err := tunnel.ValidateBandwidthLimit(xtcpBandwidthLimit); err != nil {
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
   lrok xtcp 8080 --secret-key p2p-secret --api-key lum_your_key`)
	}

	if !strings.HasPrefix(apiKey, "lum_") {
		fmt.Println("⚠️  Warning: API key should start with 'lum_'")
		fmt.Println("   Make sure you're using a valid platform API key from https://platform.lum.tools/keys")
	}

	// Determine tunnel name and track if explicitly specified
	tunnelName := name
	explicitSubdomain := false
	if subdomain != "" {
		tunnelName = subdomain
		explicitSubdomain = true
	} else if name != "" {
		explicitSubdomain = true
	}
	if tunnelName == "" {
		tunnelName = names.Generate()
		// explicitSubdomain remains false for randomly generated names
	}

	// Validate tunnel name
	if err := tunnel.ValidateTunnelName(tunnelName); err != nil {
		return fmt.Errorf("invalid tunnel name: %w", err)
	}

	// Generate config
	cfg := &config.TunnelConfig{
		APIKey:            apiKey,
		LocalPort:         localPort,
		LocalIP:           localIP,
		Subdomain:         tunnelName,
		ExplicitSubdomain: explicitSubdomain,
		ProxyType:         "xtcp",
		SecretKey:         xtcpSecretKey,
		BandwidthLimit:    xtcpBandwidthLimit,
		// Note: XTCP doesn't support encryption/compression due to P2P nature
	}

	configPath, err := config.GenerateTOML(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	fmt.Println("🚀 Starting P2P tunnel (XTCP)...")
	fmt.Println("⏳ Connecting to frp.lum.tools...")
	fmt.Printf("📍 Local:      %s:%d\n", localIP, localPort)
	fmt.Printf("🏷️  Name:       %s\n", tunnelName)
	secretDisplay := xtcpSecretKey
	if len(secretDisplay) > 8 {
		secretDisplay = secretDisplay[:8]
	}
	fmt.Printf("🔐 Secret:     %s...\n", secretDisplay)
	if xtcpStunServer != "" {
		fmt.Printf("🌐 STUN:       %s\n", xtcpStunServer)
	}
	if xtcpBandwidthLimit != "" {
		fmt.Printf("📊 Bandwidth: %s\n", xtcpBandwidthLimit)
	}
	fmt.Println()
	fmt.Println("⚡ P2P Mode: Direct client-to-client connection")
	fmt.Println("ℹ️  This tunnel requires a visitor with the secret key to access")
	fmt.Println("   Use 'lrok visitor' command on the client side")
	fmt.Println("   If P2P fails, connection will fall back to server relay")
	fmt.Println()

	// Start tunnel
	mgr := tunnel.New(configPath)
	defer mgr.Cleanup()

	if err := mgr.StartWithGracefulShutdown(); err != nil {
		if err == tunnel.ErrSubdomainLimit {
			fmt.Println()
			fmt.Println("❌ Subdomain limit reached: You have reached the maximum of 5 reserved subdomains")
			fmt.Println()
			fmt.Println("To manage your subdomains:")
			fmt.Println("  • Web UI: https://lrok.lum.tools/subdomains")
			fmt.Println("  • CLI: lrok subdomains list")
			fmt.Println("  • Release one: lrok subdomains delete <subdomain>")
			fmt.Println()
		}
		return err
	}
	return nil
}

