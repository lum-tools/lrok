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
	stcpSecretKey    string
	stcpEncrypt      bool
	stcpCompress     bool
	stcpBandwidthLimit string
)

var stcpCmd = &cobra.Command{
	Use:   "stcp <local-port>",
	Short: "Create Secret TCP tunnel (requires visitor on other end)",
	Long: `Create a Secret TCP (STCP) tunnel for secure access.

STCP tunnels require a pre-shared secret key and a visitor on the other end.
This provides secure access without exposing the service publicly.

The tunnel is only accessible to clients that know the secret key.
Use 'lrok visitor' command on the client side to connect.

Examples:
  lrok stcp 5432 --secret-key my-secret-key    # Expose PostgreSQL securely
  lrok stcp 22 --secret-key ssh-secret-123     # Expose SSH securely
  lrok stcp 3000 --secret-key api-key-456 --encrypt --compress  # With encryption

On client side:
  lrok visitor tunnel-name --type stcp --secret-key my-secret-key --bind-port 5432`,
	Args: cobra.ExactArgs(1),
	RunE: runSTCPTunnel,
}

func init() {
	stcpCmd.Flags().StringVar(&stcpSecretKey, "secret-key", "", "Pre-shared secret key (required)")
	stcpCmd.Flags().StringVarP(&name, "name", "n", "", "Custom tunnel name")
	stcpCmd.Flags().StringVar(&subdomain, "subdomain", "", "Alias for --name")
	stcpCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "lum.tools platform API key")
	stcpCmd.Flags().StringVar(&localIP, "ip", "127.0.0.1", "Local IP address to bind to")
	stcpCmd.Flags().BoolVar(&stcpEncrypt, "encrypt", false, "Enable encryption")
	stcpCmd.Flags().BoolVar(&stcpCompress, "compress", false, "Enable compression")
	stcpCmd.Flags().StringVar(&stcpBandwidthLimit, "bandwidth", "", "Bandwidth limit (e.g., 1MB, 500KB)")
	
	// Mark secret-key as required
	stcpCmd.MarkFlagRequired("secret-key")
}

func runSTCPTunnel(cmd *cobra.Command, args []string) error {
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
	if err := tunnel.ValidateSecretKey(stcpSecretKey); err != nil {
		return fmt.Errorf("invalid secret key: %w", err)
	}

	// Validate bandwidth limit if provided
	if stcpBandwidthLimit != "" {
		if err := tunnel.ValidateBandwidthLimit(stcpBandwidthLimit); err != nil {
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
   lrok stcp 5432 --secret-key my-secret --api-key lum_your_key`)
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

	// Generate config
	cfg := &config.TunnelConfig{
		APIKey:         apiKey,
		LocalPort:      localPort,
		LocalIP:        localIP,
		Subdomain:      tunnelName,
		ProxyType:      "stcp",
		SecretKey:      stcpSecretKey,
		BandwidthLimit: stcpBandwidthLimit,
		UseEncryption:  stcpEncrypt,
		UseCompression: stcpCompress,
	}

	configPath, err := config.GenerateTOML(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	fmt.Println("🚀 Starting Secret TCP tunnel...")
	fmt.Println("⏳ Connecting to frp.lum.tools...")
	fmt.Printf("📍 Local:      %s:%d\n", localIP, localPort)
	fmt.Printf("🏷️  Name:       %s\n", tunnelName)
	secretDisplay := stcpSecretKey
	if len(secretDisplay) > 8 {
		secretDisplay = secretDisplay[:8]
	}
	fmt.Printf("🔐 Secret:     %s...\n", secretDisplay)
	if stcpEncrypt {
		fmt.Println("🔒 Encryption: enabled")
	}
	if stcpCompress {
		fmt.Println("🗜️  Compression: enabled")
	}
	if stcpBandwidthLimit != "" {
		fmt.Printf("📊 Bandwidth: %s\n", stcpBandwidthLimit)
	}
	fmt.Println()
	fmt.Println("ℹ️  This tunnel requires a visitor with the secret key to access")
	fmt.Println("   Use 'lrok visitor' command on the client side")
	fmt.Println()

	// Start tunnel
	mgr := tunnel.New(configPath)
	defer mgr.Cleanup()

	return mgr.StartWithGracefulShutdown()
}

