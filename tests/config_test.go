package tests

import (
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTunnelConfig(t *testing.T) {
	// Build lrok binary
	binaryPath, err := buildLrokBinary()
	require.NoError(t, err)
	defer os.Remove(binaryPath)

	// Test TCP tunnel config generation
	localPort := getRandomPort()
	remotePort := 10000 + rand.Intn(50000)
	tunnelName := generateTestName("tcp")

	// Start TCP echo server
	server, err := startTCPEchoServer(localPort)
	require.NoError(t, err)
	defer server.Close()

	// Start TCP tunnel
	cmd := startLrokTunnel(binaryPath, "tcp",
		strconv.Itoa(localPort),
		"--remote-port", strconv.Itoa(remotePort),
		"--name", tunnelName,
		"--api-key", TestAPIKey,
		"--encrypt",
		"--compress")
	defer cleanupTunnel(cmd)

	// Wait a bit for tunnel to start
	time.Sleep(3 * time.Second)

	// Check if the tunnel process is still running
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		t.Errorf("Tunnel process exited unexpectedly")
	}

	t.Logf("✅ TCP tunnel config test - Local: %d, Remote: %d, Name: %s", localPort, remotePort, tunnelName)
}
