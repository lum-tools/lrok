package tunnel

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/lum-tools/lrok/internal/embed"
)

// Manager handles tunnel lifecycle
type Manager struct {
	configPath string
	cmd        *exec.Cmd
}

// New creates a new tunnel manager
func New(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// ErrSubdomainLimit is returned when the user has reached the maximum number of reserved subdomains
var ErrSubdomainLimit = fmt.Errorf("subdomain limit reached")

// Start starts the frpc tunnel and streams output (blocking)
func (m *Manager) Start(ctx context.Context) error {
	frpcPath, err := embed.GetFrpcPath()
	if err != nil {
		return fmt.Errorf("failed to locate frpc binary: %w", err)
	}

	m.cmd = exec.CommandContext(ctx, frpcPath, "-c", m.configPath)
	
	// Stream stdout and stderr
	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := m.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start frpc: %w", err)
	}

	// Stream stdout normally
	go io.Copy(os.Stdout, stdout)

	// Parse stderr line by line to detect subdomain limit errors
	errorChan := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			// Output the line for debugging
			fmt.Fprintln(os.Stderr, line)
			
			// Check for subdomain limit error pattern
			// Pattern: "Subdomain '...' is not available: You have reached the maximum of 5 reserved subdomains"
			if strings.Contains(line, "reached the maximum of 5 reserved subdomains") ||
				strings.Contains(line, "maximum of 5 reserved subdomains") ||
				(strings.Contains(line, "Subdomain") && strings.Contains(line, "is not available") && strings.Contains(line, "maximum")) {
				// Kill the process immediately
				if m.cmd != nil && m.cmd.Process != nil {
					m.cmd.Process.Kill()
				}
				errorChan <- ErrSubdomainLimit
				return
			}
		}
		// If scanner finished without error, check for scan errors
		if err := scanner.Err(); err != nil {
			errorChan <- err
		}
	}()

	// Wait for either process completion or subdomain limit error
	done := make(chan error, 1)
	go func() {
		done <- m.cmd.Wait()
	}()

	select {
	case err := <-errorChan:
		// Subdomain limit error detected
		return err
	case err := <-done:
		// Process completed normally
		return err
	}
}

// StartWithGracefulShutdown starts the tunnel and handles Ctrl+C gracefully
func (m *Manager) StartWithGracefulShutdown() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start tunnel (only once!)
	errChan := make(chan error, 1)
	go func() {
		errChan <- m.Start(ctx)
	}()

	select {
	case <-sigChan:
		fmt.Println("\n\n🛑 Shutting down tunnel gracefully...")
		cancel()
		if m.cmd != nil && m.cmd.Process != nil {
			m.cmd.Process.Signal(os.Interrupt)
		}
		return nil
	case err := <-errChan:
		// Check if it's a subdomain limit error
		if err == ErrSubdomainLimit {
			return err
		}
		return err
	}
}

// Stop stops the tunnel
func (m *Manager) Stop() error {
	if m.cmd != nil && m.cmd.Process != nil {
		return m.cmd.Process.Kill()
	}
	return nil
}

// Cleanup removes the config file
func (m *Manager) Cleanup() error {
	if m.configPath != "" {
		return os.Remove(m.configPath)
	}
	return nil
}

