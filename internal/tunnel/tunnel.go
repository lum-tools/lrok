package tunnel

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
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

	// Stream output
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	// Wait for completion
	return m.cmd.Wait()
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

