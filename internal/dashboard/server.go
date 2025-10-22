package dashboard

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/lum-tools/lrok/internal/proxy"
)

// Stats holds tunnel statistics
type Stats struct {
	TunnelName   string    `json:"tunnel_name"`
	PublicURL    string    `json:"public_url"`
	LocalPort    int       `json:"local_port"`
	Status       string    `json:"status"`
	StartTime    time.Time `json:"start_time"`
	BytesIn      int64     `json:"bytes_in"`
	BytesOut     int64     `json:"bytes_out"`
	Connections  int64     `json:"connections"`
	mu           sync.RWMutex
}

// UpdateStats updates the tunnel statistics
func (s *Stats) UpdateStats(bytesIn, bytesOut, connections int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BytesIn = bytesIn
	s.BytesOut = bytesOut
	s.Connections = connections
}

// GetStats returns a copy of current stats
func (s *Stats) GetStats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Stats{
		TunnelName:  s.TunnelName,
		PublicURL:   s.PublicURL,
		LocalPort:   s.LocalPort,
		Status:      s.Status,
		StartTime:   s.StartTime,
		BytesIn:     s.BytesIn,
		BytesOut:    s.BytesOut,
		Connections: s.Connections,
	}
}

// Server is a minimal HTTP server for tunnel dashboard
type Server struct {
	stats  *Stats
	proxy  *proxy.Proxy
	server *http.Server
	port   int
}

// New creates a new dashboard server
func New(stats *Stats, prox *proxy.Proxy) *Server {
	return &Server{
		stats: stats,
		proxy: prox,
	}
}

// Start starts the dashboard server on the specified port (or finds available port)
func (s *Server) Start(preferredPort int) error {
	// Try preferred port first, then find available
	port := preferredPort
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// Port occupied, find random available port
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			return fmt.Errorf("failed to start dashboard: %w", err)
		}
		port = listener.Addr().(*net.TCPAddr).Port
	}
	
	s.port = port
	
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/requests", s.handleRequests)
	mux.HandleFunc("/api/requests/stream", s.handleRequestsStream)
	
	s.server = &http.Server{
		Handler: mux,
	}
	
	go s.server.Serve(listener)
	
	return nil
}

// Stop stops the dashboard server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// Port returns the port the dashboard is running on
func (s *Server) Port() int {
	return s.port
}

// handleStats serves the stats API endpoint
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	stats := s.stats.GetStats()
	
	// Add proxy stats if available
	if s.proxy != nil {
		bytesIn, bytesOut, conns := s.proxy.GetStats()
		stats.BytesIn = bytesIn
		stats.BytesOut = bytesOut
		stats.Connections = conns
	}
	
	json.NewEncoder(w).Encode(stats)
}

