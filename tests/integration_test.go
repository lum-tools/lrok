package tests

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPTunnel(t *testing.T) {
	// Build lrok binary
	binaryPath, err := buildLrokBinary()
	require.NoError(t, err)
	defer os.Remove(binaryPath)

	// Get random port and generate tunnel name
	localPort := getRandomPort()
	tunnelName := generateTestName("http")

	// Start HTTP test server
	server := startHTTPTestServer(localPort)
	defer server.Close()

	// Start lrok tunnel
	cmd := startLrokTunnel(binaryPath, "http",
		strconv.Itoa(localPort),
		"--name", tunnelName,
		"--api-key", TestAPIKey)
	defer cleanupTunnel(cmd)

	// Wait for tunnel to be ready
	publicURL := fmt.Sprintf("https://%s.%s", tunnelName, TunnelDomain)
	err = waitForTunnel(publicURL, 30*time.Second)
	require.NoError(t, err)

	// Test basic connectivity
	resp, err := http.Get(publicURL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	// Test health endpoint
	resp, err = http.Get(publicURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "healthy")

	// Test HTTPS (automatic SSL)
	assert.Equal(t, "https", resp.Request.URL.Scheme)

	// Test webhook endpoint
	webhookData := `{"test": "webhook", "timestamp": ` + strconv.FormatInt(time.Now().Unix(), 10) + `}`
	resp, err = http.Post(publicURL+"/webhook", "application/json", 
		strings.NewReader(webhookData))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	t.Logf("✅ HTTP tunnel test passed - URL: %s", publicURL)
}

func TestTCPTunnel(t *testing.T) {
	binaryPath, err := buildLrokBinary()
	require.NoError(t, err)
	defer os.Remove(binaryPath)

	localPort := getRandomPort()
	remotePort := 15000 // Use port 15000 which maps to NodePort 30150
	tunnelName := generateTestName("tcp")

	// Start HTTP server instead of TCP echo server
	localServer := startHTTPTestServer(localPort)
	defer localServer.Close()

	// Start TCP tunnel with encryption and compression
	cmd := startLrokTunnel(binaryPath, "tcp",
		strconv.Itoa(localPort),
		"--remote-port", strconv.Itoa(remotePort),
		"--name", tunnelName,
		"--api-key", TestAPIKey,
		"--encrypt",
		"--compress")
	defer cleanupTunnel(cmd)

	// Wait for tunnel to establish
	time.Sleep(10 * time.Second)

	// Test HTTP connectivity through TCP tunnel
	httpURL := fmt.Sprintf("http://%s:%d", FRPServerAddr, TCPTunnelPort1)
	err = waitForTunnel(httpURL, 15*time.Second)
	require.NoError(t, err, "Failed to connect to HTTP service through TCP tunnel")

	// Make HTTP request through tunnel
	resp, err := http.Get(httpURL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "HTTP Tunnel Test Server")

	t.Logf("✅ TCP tunnel test passed - Remote: %s:%d", FRPServerAddr, TCPTunnelPort1)
}

func TestUDPTunnel(t *testing.T) {
	// UDP tunnels have been removed from lrok due to infrastructure limitations
	// in Kubernetes networking. UDP tunnels work at the FRP level but cannot
	// be reliably exposed through LoadBalancer or NodePort services.
	t.Skip("UDP tunnels have been removed from lrok CLI")
}

func TestSTCPTunnel(t *testing.T) {
	binaryPath, err := buildLrokBinary()
	require.NoError(t, err)
	defer os.Remove(binaryPath)

	localPort := getRandomPort()
	visitorPort := getRandomPort()
	tunnelName := generateTestName("stcp")
	secretKey := generateRandomSecret()

	// Start local service
	server := startHTTPTestServer(localPort)
	defer server.Close()

	// Start STCP tunnel (server side)
	serverCmd := startLrokTunnel(binaryPath, "stcp",
		strconv.Itoa(localPort),
		"--name", tunnelName,
		"--secret-key", secretKey,
		"--api-key", TestAPIKey)
	defer cleanupTunnel(serverCmd)

	time.Sleep(5 * time.Second)

	// Start visitor (client side)
	visitorCmd := startLrokTunnel(binaryPath, "visitor",
		tunnelName,
		"--type", "stcp",
		"--secret-key", secretKey,
		"--bind-port", strconv.Itoa(visitorPort),
		"--api-key", TestAPIKey)
	defer cleanupTunnel(visitorCmd)

	// Wait for visitor connection
	time.Sleep(5 * time.Second)

	// Test connectivity through visitor
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", visitorPort))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	// Test health endpoint through visitor
	resp, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", visitorPort))
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "healthy")

	t.Logf("✅ STCP tunnel test passed - Visitor port: %d", visitorPort)
}

func TestXTCPTunnel(t *testing.T) {
	binaryPath, err := buildLrokBinary()
	require.NoError(t, err)
	defer os.Remove(binaryPath)

	localPort := getRandomPort()
	visitorPort := getRandomPort()
	tunnelName := generateTestName("xtcp")
	secretKey := generateRandomSecret()

	// Start local service
	server := startHTTPTestServer(localPort)
	defer server.Close()

	// Start XTCP tunnel (server side)
	serverCmd := startLrokTunnel(binaryPath, "xtcp",
		strconv.Itoa(localPort),
		"--name", tunnelName,
		"--secret-key", secretKey,
		"--api-key", TestAPIKey)
	defer cleanupTunnel(serverCmd)

	time.Sleep(5 * time.Second)

	// Start visitor (client side)
	visitorCmd := startLrokTunnel(binaryPath, "visitor",
		tunnelName,
		"--type", "xtcp",
		"--secret-key", secretKey,
		"--bind-port", strconv.Itoa(visitorPort),
		"--api-key", TestAPIKey)
	defer cleanupTunnel(visitorCmd)

	// Wait for P2P negotiation (XTCP takes longer)
	time.Sleep(10 * time.Second)

	// Test connectivity through visitor
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", visitorPort))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	// Test health endpoint through visitor
	resp, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", visitorPort))
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "healthy")

	t.Logf("✅ XTCP tunnel test passed - Visitor port: %d", visitorPort)
}
