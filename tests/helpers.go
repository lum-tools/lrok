package tests

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// buildLrokBinary builds the CLI binary to a temporary location and returns the path
func buildLrokBinary() (string, error) {
	// Create temporary file with proper extension for Windows
	var tmpFile *os.File
	var err error
	
	if runtime.GOOS == "windows" {
		tmpFile, err = os.CreateTemp("", "lrok-test-*.exe")
	} else {
		tmpFile, err = os.CreateTemp("", "lrok-test-*")
	}
	
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpFile.Close()

	binaryPath := tmpFile.Name()

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/lrok")
	cmd.Dir = ".." // Go up one level from tests/ to lrok/
	
	if err := cmd.Run(); err != nil {
		os.Remove(binaryPath)
		return "", fmt.Errorf("failed to build lrok binary: %w", err)
	}

	return binaryPath, nil
}

// startLrokTunnel executes the lrok CLI with given arguments and returns the command
func startLrokTunnel(binaryPath string, args ...string) *exec.Cmd {
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Start(); err != nil {
		panic(fmt.Sprintf("failed to start lrok tunnel: %v", err))
	}

	return cmd
}

// waitForTunnel polls the given URL until it's accessible or timeout is reached
func waitForTunnel(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	
	return fmt.Errorf("tunnel not ready after %v", timeout)
}

// generateTestName generates a random tunnel name with the given prefix
func generateTestName(prefix string) string {
	timestamp := time.Now().Unix()
	random := rand.Intn(10000)
	return fmt.Sprintf("%s-%d-%d", prefix, timestamp, random)
}

// getRandomPort finds an available random port
func getRandomPort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(fmt.Sprintf("failed to get random port: %v", err))
	}
	defer listener.Close()
	
	return listener.Addr().(*net.TCPAddr).Port
}

// cleanupTunnel gracefully kills the tunnel process
func cleanupTunnel(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	// Send SIGTERM first
	cmd.Process.Signal(os.Interrupt)
	
	// Wait for graceful shutdown with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	
	select {
	case <-done:
		// Process exited gracefully
		return
	case <-time.After(5 * time.Second):
		// Force kill if still running
		cmd.Process.Kill()
		cmd.Wait()
	}
}

// generateRandomSecret generates a random secret key for STCP/XTCP tunnels
func generateRandomSecret() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// debugTCPConnection tests raw TCP connectivity to a host:port
func debugTCPConnection(host string, port int, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return fmt.Errorf("TCP connection failed: %w", err)
	}
	defer conn.Close()
	
	// Test basic write/read
	testMsg := "debug-test"
	_, err = conn.Write([]byte(testMsg))
	if err != nil {
		return fmt.Errorf("TCP write failed: %w", err)
	}
	
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("TCP read failed: %w", err)
	}
	
	fmt.Printf("TCP debug: sent '%s', received '%s'\n", testMsg, string(buffer[:n]))
	return nil
}

// debugUDPConnection tests raw UDP connectivity to a host:port
func debugUDPConnection(host string, port int, timeout time.Duration) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("UDP address resolution failed: %w", err)
	}
	
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("UDP connection failed: %w", err)
	}
	defer conn.Close()
	
	// Test basic write/read
	testMsg := "debug-test"
	_, err = conn.Write([]byte(testMsg))
	if err != nil {
		return fmt.Errorf("UDP write failed: %w", err)
	}
	
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(timeout))
	n, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("UDP read failed: %w", err)
	}
	
	fmt.Printf("UDP debug: sent '%s', received '%s'\n", testMsg, string(buffer[:n]))
	return nil
}

// dumpFRPLogs captures FRP server logs during test execution
func dumpFRPLogs() {
	fmt.Println("=== FRP Server Logs ===")
	// This would require kubectl access, for now just print a placeholder
	fmt.Println("FRP logs would be captured here in a real implementation")
	fmt.Println("======================")
}
