package main

import (
	"fmt"
	"log"

	"github.com/lum-tools/lrok/internal/api"
	"github.com/spf13/cobra"
)

var (
	daemonPort int
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start lrok in daemon mode with HTTP API",
	Long: `Start lrok as a background daemon with HTTP API server.

The daemon provides a REST API for creating and managing tunnels programmatically.
This enables building SDKs and integrations in any programming language.

API Documentation:
  OpenAPI spec available at: api/openapi.yaml
  Default API endpoint: http://localhost:4243

Example:
  lrok daemon --port 4243

Once running, you can create tunnels via HTTP API:
  curl -X POST http://localhost:4243/api/v1/tunnels \
    -H "Content-Type: application/json" \
    -d '{"type":"http","localPort":8000}'
`,
	Run: runDaemon,
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().IntVarP(&daemonPort, "port", "p", 4243, "Port for API server")
}

func runDaemon(cmd *cobra.Command, args []string) {
	fmt.Printf("🚀 Starting lrok daemon v%s\n", versionInfo)
	fmt.Printf("📡 API server listening on http://localhost:%d\n", daemonPort)
	fmt.Printf("📖 API docs: /api/v1/health\n\n")

	server := api.NewServer(versionInfo)
	addr := fmt.Sprintf(":%d", daemonPort)

	if err := server.Start(addr); err != nil {
		log.Fatalf("Failed to start daemon: %v", err)
	}
}
