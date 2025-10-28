package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lum-tools/lrok/internal/config"
	"github.com/lum-tools/lrok/internal/dashboard"
	"github.com/lum-tools/lrok/internal/names"
	"github.com/lum-tools/lrok/internal/proxy"
	"github.com/lum-tools/lrok/internal/tunnel"
	"github.com/lum-tools/lrok/internal/version"
	"github.com/spf13/cobra"
)

var (
	versionInfo = "dev"
	commit      = "none"
	date        = "unknown"
)

var (
	port      int
	name      string
	subdomain string
	apiKey    string
	localIP   string
)

var rootCmd = &cobra.Command{
	Use:   "lrok [port]",
	Short: "Expose local services with readable tunnel names",
	Long: `lrok - Tunnel service powered by lum.tools

Expose your local services to the internet with HTTPS and readable URLs.

Examples:
  lrok 8000                    # Expose port 8000 with random name
  lrok 8000 --name my-app      # Expose with custom name
  lrok 3000 --subdomain api    # Use subdomain instead`,
	Args:    cobra.MaximumNArgs(1),
	Version: versionInfo,
	RunE:    runTunnel,
}

var httpCmd = &cobra.Command{
	Use:   "http [port]",
	Short: "Create HTTP tunnel (alias for default behavior)",
	Long:  `Create an HTTP tunnel to expose a local port`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTunnel,
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("lrok version %s\n", versionInfo)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
		
		// Check for updates (non-blocking)
		if hasUpdate, latest, method, err := version.CheckForUpdate(versionInfo); err == nil && hasUpdate {
			fmt.Println()
			version.ShowUpdateWarning(versionInfo, latest, method)
		}
	},
}

var loginCmd = &cobra.Command{
	Use:   "login [API_KEY]",
	Short: "Save API key to config file",
	Long: `Save your lum.tools platform API key to ~/.lrok/config.toml

This allows you to use lrok without setting environment variables.

Get your API key from: https://platform.lum.tools/keys`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey := args[0]
		
		// Validate API key format
		if !strings.HasPrefix(apiKey, "lum_") {
			return fmt.Errorf("invalid API key format (should start with 'lum_')")
		}
		
		// Save to config
		if err := config.SaveAPIKey(apiKey); err != nil {
			return fmt.Errorf("failed to save API key: %w", err)
		}
		
		configPath, _ := config.GetConfigPath()
		fmt.Println("✅ API key saved successfully!")
		fmt.Printf("   Config: %s\n", configPath)
		fmt.Println()
		fmt.Println("You can now run lrok without setting LUM_API_KEY:")
		fmt.Println("   lrok 8000")
		
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove saved API key",
	Long:  `Remove the API key from ~/.lrok/config.toml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.ClearConfig(); err != nil {
			return fmt.Errorf("failed to logout: %w", err)
		}
		
		fmt.Println("✅ Logged out successfully!")
		fmt.Println("   API key removed from config")
		
		return nil
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current API key configuration",
	Long:  `Display information about the currently configured API key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Try to get API key from different sources
		var apiKey string
		var source string
		
		// Check flag (if this was called with --api-key)
		if apiKey == "" {
			apiKey = os.Getenv("LUM_API_KEY")
			if apiKey != "" {
				source = "environment variable (LUM_API_KEY)"
			}
		}
		
		// Check config file
		if apiKey == "" {
			if key, err := config.GetAPIKey(); err == nil {
				apiKey = key
				configPath, _ := config.GetConfigPath()
				source = fmt.Sprintf("config file (%s)", configPath)
			}
		}
		
		if apiKey == "" {
			fmt.Println("❌ Not logged in")
			fmt.Println()
			fmt.Println("To login:")
			fmt.Println("   lrok login <your-api-key>")
			fmt.Println()
			fmt.Println("Or set environment variable:")
			fmt.Println("   export LUM_API_KEY='lum_your_key'")
			fmt.Println()
			fmt.Println("Get your API key: https://platform.lum.tools/keys")
			return nil
		}
		
		// Show prefix only for security
		prefix := apiKey
		if len(apiKey) > 16 {
			prefix = apiKey[:16] + "..." + apiKey[len(apiKey)-4:]
		}
		
		fmt.Println("✅ Logged in")
		fmt.Printf("   API Key: %s\n", prefix)
		fmt.Printf("   Source:  %s\n", source)
		
		return nil
	},
}

func init() {
	// Flags for root command
	rootCmd.Flags().IntVarP(&port, "port", "p", 0, "Local port to expose (optional if provided as argument)")
	rootCmd.Flags().StringVarP(&name, "name", "n", "", "Custom tunnel name (generates random if not provided)")
	rootCmd.Flags().StringVar(&subdomain, "subdomain", "", "Alias for --name")
	rootCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "lum.tools platform API key (or set LUM_API_KEY env var)")
	rootCmd.Flags().StringVar(&localIP, "ip", "127.0.0.1", "Local IP address to bind to")

	// Flags for http command (same as root)
	httpCmd.Flags().IntVarP(&port, "port", "p", 0, "Local port to expose (optional if provided as argument)")
	httpCmd.Flags().StringVarP(&name, "name", "n", "", "Custom tunnel name")
	httpCmd.Flags().StringVar(&subdomain, "subdomain", "", "Alias for --name")
	httpCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "API key")
	httpCmd.Flags().StringVar(&localIP, "ip", "127.0.0.1", "Local IP to bind to")

	rootCmd.AddCommand(httpCmd)
	rootCmd.AddCommand(tcpCmd)
	rootCmd.AddCommand(stcpCmd)
	rootCmd.AddCommand(xtcpCmd)
	rootCmd.AddCommand(visitorCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)
}

func runTunnel(cmd *cobra.Command, args []string) error {
	// Check for updates in background (non-blocking)
	go func() {
		if hasUpdate, latest, method, err := version.CheckForUpdate(versionInfo); err == nil && hasUpdate {
			version.ShowUpdateWarning(versionInfo, latest, method)
		}
	}()
	
	// Get port from args or flag
	if len(args) > 0 {
		// Port provided as argument (e.g., "lrok 8000")
		portArg := args[0]
		var err error
		port, err = strconv.Atoi(portArg)
		if err != nil {
			return fmt.Errorf("invalid port: %s", portArg)
		}
	}

	// Validate port
	if port == 0 {
		return fmt.Errorf(`❌ No port specified!

Usage:
  lrok 8000                    # Expose port 8000
  lrok 8000 --name my-app      # With custom name
  lrok http 3000               # Explicit HTTP tunnel

Run 'lrok --help' for more examples.`)
	}

	// Get API key with priority: flag > env var > config file
	var apiKeySource string
	
	if apiKey == "" {
		// Try environment variable
		apiKey = os.Getenv("LUM_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("FRP_API_KEY") // Legacy support
		}
		if apiKey != "" {
			apiKeySource = "environment variable"
		}
	} else {
		apiKeySource = "--api-key flag"
	}

	// Try config file if still not found
	if apiKey == "" {
		if key, err := config.GetAPIKey(); err == nil {
			apiKey = key
			apiKeySource = "config file (~/.lrok/config.toml)"
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
   lrok 8000 --api-key lum_your_key`)
	}

	if !strings.HasPrefix(apiKey, "lum_") {
		fmt.Println("⚠️  Warning: API key should start with 'lum_'")
		fmt.Println("   Make sure you're using a valid platform API key from https://platform.lum.tools/keys")
	}
	
	// Show API key source for transparency (debug mode or verbose)
	_ = apiKeySource // Used for debugging, currently unused in output

	// Determine subdomain
	tunnelName := name
	if subdomain != "" {
		tunnelName = subdomain
	}
	if tunnelName == "" {
		tunnelName = names.Generate()
	}

	tunnelURL := fmt.Sprintf("https://%s.t.lum.tools", tunnelName)

	// Start reverse proxy for request inspection
	fmt.Println("🔄 Starting request inspector proxy...")
	prox := proxy.New(port, 100)
	proxyPort, err := prox.Start()
	if err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}
	defer prox.Stop()
	
	fmt.Printf("✅ Proxy ready on port %d (forwarding to %d)\n", proxyPort, port)

	// Generate config with proxy port (frpc forwards to proxy, proxy forwards to user app)
	cfg := &config.TunnelConfig{
		APIKey:    apiKey,
		LocalPort: proxyPort,
		LocalIP:   localIP,
		Subdomain: tunnelName,
	}

	configPath, err := config.GenerateTOML(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// Start dashboard with proxy
	stats := &dashboard.Stats{
		TunnelName: tunnelName,
		PublicURL:  tunnelURL,
		LocalPort:  port,
		Status:     "Connected",
		StartTime:  time.Now(),
	}
	
	dash := dashboard.New(stats, prox)
	if err := dash.Start(4242); err != nil {
		// Dashboard failed to start, continue anyway
		fmt.Printf("⚠️  Dashboard failed to start: %v\n", err)
	} else {
		defer dash.Stop()
	}

	fmt.Println("\n🚀 Starting lrok tunnel...")
	fmt.Println("⏳ Connecting to frp.lum.tools...")
	
	// Start tunnel
	mgr := tunnel.New(configPath)
	defer mgr.Cleanup()
	
	// Start tunnel with graceful shutdown (this is blocking until Ctrl+C)
	// We'll verify in a separate goroutine
	go func() {
		// Wait a bit for tunnel to connect, then verify
		time.Sleep(3 * time.Second)
		
		fmt.Println("🔍 Verifying tunnel...")
		client := &http.Client{Timeout: 5 * time.Second}
		verified := false
		
		for i := 0; i < 10; i++ {
			resp, err := client.Get(tunnelURL + "/")
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					verified = true
					break
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
		
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("  📍 Local:      http://%s:%d\n", localIP, port)
		fmt.Printf("  🌐 Public URL: %s\n", tunnelURL)
		fmt.Printf("  🏷️  Name:       %s\n", tunnelName)
		if dash.Port() > 0 {
			fmt.Printf("  📊 Dashboard:  http://localhost:%d\n", dash.Port())
		}
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		
	if verified {
		fmt.Println("\n✅ Tunnel is ready and verified!")
	} else {
		fmt.Println("\n⏳ Tunnel is connecting... (may take a few more seconds)")
	}
	fmt.Println("   Open the dashboard to inspect requests in real-time!")
	}()

	return mgr.StartWithGracefulShutdown()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

