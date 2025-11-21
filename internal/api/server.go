package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lum-tools/lrok/internal/config"
	"github.com/lum-tools/lrok/internal/tunnel"
)

// Server represents the HTTP API server
type Server struct {
	router   *mux.Router
	tunnels  map[string]*TunnelInstance
	mu       sync.RWMutex
	startAt  time.Time
	version  string
	apiKey   string
}

// TunnelInstance represents a running tunnel
type TunnelInstance struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	LocalPort int                    `json:"localPort"`
	PublicURL string                 `json:"publicUrl,omitempty"`
	RemotePort *int                  `json:"remotePort,omitempty"`
	Status    string                 `json:"status"`
	CreatedAt time.Time              `json:"createdAt"`
	Config    map[string]interface{} `json:"config"`
	Manager   *tunnel.Manager        `json:"-"`
	ctx       context.Context
	cancel    context.CancelFunc
}

// CreateTunnelRequest represents the request to create a tunnel
type CreateTunnelRequest struct {
	Type           string                 `json:"type"`
	LocalPort      int                    `json:"localPort"`
	Subdomain      *string                `json:"subdomain,omitempty"`
	Name           *string                `json:"name,omitempty"`
	RemotePort     *int                   `json:"remotePort,omitempty"`
	SecretKey      *string                `json:"secretKey,omitempty"`
	ServerName     *string                `json:"serverName,omitempty"`
	BindPort       *int                   `json:"bindPort,omitempty"`
	Encryption     bool                   `json:"encryption"`
	Compression    bool                   `json:"compression"`
	BandwidthLimit *string                `json:"bandwidthLimit,omitempty"`
	HealthCheck    map[string]interface{} `json:"healthCheck,omitempty"`
}

// TunnelStats represents tunnel statistics
type TunnelStats struct {
	BytesIn      int64 `json:"bytesIn"`
	BytesOut     int64 `json:"bytesOut"`
	Connections  int64 `json:"connections"`
	RequestCount int64 `json:"requestCount"`
	Uptime       int64 `json:"uptime"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewServer creates a new API server
func NewServer(version string) *Server {
	s := &Server{
		router:   mux.NewRouter(),
		tunnels:  make(map[string]*TunnelInstance),
		startAt:  time.Now(),
		version:  version,
	}

	// Load API key from config if available
	apiKey, _ := config.GetAPIKey()
	s.apiKey = apiKey

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// System routes (no auth required)
	s.router.HandleFunc("/api/v1/health", s.handleHealth).Methods("GET")
	s.router.HandleFunc("/api/v1/version", s.handleVersion).Methods("GET")

	// Auth routes
	s.router.HandleFunc("/api/v1/auth/login", s.handleLogin).Methods("POST")
	s.router.HandleFunc("/api/v1/auth/logout", s.authMiddleware(s.handleLogout)).Methods("POST")
	s.router.HandleFunc("/api/v1/auth/whoami", s.authMiddleware(s.handleWhoami)).Methods("GET")

	// Tunnel routes (auth required)
	s.router.HandleFunc("/api/v1/tunnels", s.authMiddleware(s.handleCreateTunnel)).Methods("POST")
	s.router.HandleFunc("/api/v1/tunnels", s.authMiddleware(s.handleListTunnels)).Methods("GET")
	s.router.HandleFunc("/api/v1/tunnels/{id}", s.authMiddleware(s.handleGetTunnel)).Methods("GET")
	s.router.HandleFunc("/api/v1/tunnels/{id}", s.authMiddleware(s.handleDeleteTunnel)).Methods("DELETE")
	s.router.HandleFunc("/api/v1/tunnels/{id}/stats", s.authMiddleware(s.handleGetStats)).Methods("GET")
	s.router.HandleFunc("/api/v1/tunnels/{id}/requests", s.authMiddleware(s.handleGetRequests)).Methods("GET")
	s.router.HandleFunc("/api/v1/tunnels/{id}/requests/stream", s.authMiddleware(s.handleStreamRequests)).Methods("GET")

	// Subdomain routes (auth required)
	s.router.HandleFunc("/api/v1/subdomains", s.authMiddleware(s.handleListSubdomains)).Methods("GET")
	s.router.HandleFunc("/api/v1/subdomains", s.authMiddleware(s.handleReserveSubdomain)).Methods("POST")
	s.router.HandleFunc("/api/v1/subdomains/{name}", s.authMiddleware(s.handleDeleteSubdomain)).Methods("DELETE")

	// CORS middleware
	s.router.Use(corsMiddleware)
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	log.Printf("Starting lrok daemon on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// authMiddleware checks for API key authentication
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if API key is required (daemon has one configured)
		if s.apiKey == "" {
			next(w, r)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			respondError(w, http.StatusUnauthorized, "API key required", "UNAUTHORIZED")
			return
		}

		// Support both "Bearer token" and "token" formats
		token := auth
		if len(auth) > 7 && auth[:7] == "Bearer " {
			token = auth[7:]
		}

		if token != s.apiKey {
			respondError(w, http.StatusUnauthorized, "Invalid API key", "UNAUTHORIZED")
			return
		}

		next(w, r)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Health check handler
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"version": s.version,
		"uptime":  int64(time.Since(s.startAt).Seconds()),
	})
}

// Version handler
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"version": s.version,
	})
}

// Login handler
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		APIKey string `json:"apiKey"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	if req.APIKey == "" || len(req.APIKey) < 4 || req.APIKey[:4] != "lum_" {
		respondError(w, http.StatusBadRequest, "Invalid API key format", "INVALID_API_KEY")
		return
	}

	// Save API key to config
	if err := config.SaveAPIKey(req.APIKey); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save API key", "SAVE_FAILED")
		return
	}

	s.apiKey = req.APIKey

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "API key saved successfully",
	})
}

// Logout handler
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := config.RemoveAPIKey(); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to remove API key", "REMOVE_FAILED")
		return
	}

	s.apiKey = ""

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "API key removed successfully",
	})
}

// Whoami handler
func (s *Server) handleWhoami(w http.ResponseWriter, r *http.Request) {
	apiKey, _ := config.GetAPIKey()

	masked := ""
	if apiKey != "" {
		if len(apiKey) > 8 {
			masked = apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
		} else {
			masked = "****"
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"authenticated": apiKey != "",
		"apiKey":        masked,
	})
}

// Create tunnel handler
func (s *Server) handleCreateTunnel(w http.ResponseWriter, r *http.Request) {
	var req CreateTunnelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	// Validate request
	if err := s.validateCreateTunnelRequest(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	// Generate tunnel ID
	tunnelID := uuid.New().String()

	// Build tunnel config
	configPath, err := s.buildTunnelConfig(tunnelID, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), "CONFIG_ERROR")
		return
	}

	// Create tunnel manager
	mgr := tunnel.New(configPath)

	// Create context for tunnel
	ctx, cancel := context.WithCancel(context.Background())

	instance := &TunnelInstance{
		ID:        tunnelID,
		Type:      req.Type,
		Name:      s.getTunnelName(&req),
		LocalPort: req.LocalPort,
		RemotePort: req.RemotePort,
		Status:    "starting",
		CreatedAt: time.Now(),
		Config:    s.requestToConfig(&req),
		Manager:   mgr,
		ctx:       ctx,
		cancel:    cancel,
	}

	// Store tunnel
	s.mu.Lock()
	s.tunnels[tunnelID] = instance
	s.mu.Unlock()

	// Start tunnel in background
	go func() {
		if err := mgr.Start(ctx); err != nil {
			log.Printf("Tunnel %s failed: %v", tunnelID, err)
			s.mu.Lock()
			instance.Status = "error"
			s.mu.Unlock()
		}
	}()

	// Wait for tunnel to connect or timeout
	publicURL, err := s.waitForTunnelURL(mgr, instance, 10*time.Second)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error(), "TUNNEL_START_FAILED")
		return
	}

	instance.PublicURL = publicURL
	instance.Status = "connected"

	respondJSON(w, http.StatusCreated, instance)
}

// List tunnels handler
func (s *Server) handleListTunnels(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tunnels := make([]*TunnelInstance, 0, len(s.tunnels))
	for _, t := range s.tunnels {
		tunnels = append(tunnels, t)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"tunnels": tunnels,
	})
}

// Get tunnel handler
func (s *Server) handleGetTunnel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tunnelID := vars["id"]

	s.mu.RLock()
	tunnel, exists := s.tunnels[tunnelID]
	s.mu.RUnlock()

	if !exists {
		respondError(w, http.StatusNotFound, "Tunnel not found", "NOT_FOUND")
		return
	}

	respondJSON(w, http.StatusOK, tunnel)
}

// Delete tunnel handler
func (s *Server) handleDeleteTunnel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tunnelID := vars["id"]

	s.mu.Lock()
	tunnel, exists := s.tunnels[tunnelID]
	if !exists {
		s.mu.Unlock()
		respondError(w, http.StatusNotFound, "Tunnel not found", "NOT_FOUND")
		return
	}
	delete(s.tunnels, tunnelID)
	s.mu.Unlock()

	// Stop tunnel
	tunnel.cancel()
	tunnel.Manager.Stop()
	tunnel.Manager.Cleanup()

	w.WriteHeader(http.StatusNoContent)
}

// Get stats handler
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tunnelID := vars["id"]

	s.mu.RLock()
	tunnel, exists := s.tunnels[tunnelID]
	s.mu.RUnlock()

	if !exists {
		respondError(w, http.StatusNotFound, "Tunnel not found", "NOT_FOUND")
		return
	}

	stats := TunnelStats{
		BytesIn:     0, // TODO: Implement actual stats tracking
		BytesOut:    0,
		Connections: 0,
		Uptime:      int64(time.Since(tunnel.CreatedAt).Seconds()),
	}

	respondJSON(w, http.StatusOK, stats)
}

// Get requests handler (HTTP tunnels only)
func (s *Server) handleGetRequests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tunnelID := vars["id"]

	s.mu.RLock()
	tunnel, exists := s.tunnels[tunnelID]
	s.mu.RUnlock()

	if !exists {
		respondError(w, http.StatusNotFound, "Tunnel not found", "NOT_FOUND")
		return
	}

	if tunnel.Type != "http" {
		respondError(w, http.StatusBadRequest, "Only HTTP tunnels support request inspection", "INVALID_TUNNEL_TYPE")
		return
	}

	// TODO: Implement request history fetching from dashboard
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"requests": []interface{}{},
	})
}

// Stream requests handler (HTTP tunnels only)
func (s *Server) handleStreamRequests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tunnelID := vars["id"]

	s.mu.RLock()
	tunnel, exists := s.tunnels[tunnelID]
	s.mu.RUnlock()

	if !exists {
		respondError(w, http.StatusNotFound, "Tunnel not found", "NOT_FOUND")
		return
	}

	if tunnel.Type != "http" {
		respondError(w, http.StatusBadRequest, "Only HTTP tunnels support request streaming", "INVALID_TUNNEL_TYPE")
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// TODO: Implement SSE streaming from dashboard
	fmt.Fprintf(w, "data: {\"message\": \"Stream connected\"}\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// List subdomains handler
func (s *Server) handleListSubdomains(w http.ResponseWriter, r *http.Request) {
	// TODO: Call platform API
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"subdomains": []interface{}{},
	})
}

// Reserve subdomain handler
func (s *Server) handleReserveSubdomain(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	// TODO: Call platform API to reserve subdomain
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"name":      req.Name,
		"createdAt": time.Now(),
	})
}

// Delete subdomain handler
func (s *Server) handleDeleteSubdomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	// TODO: Call platform API to delete subdomain
	_ = name

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func (s *Server) validateCreateTunnelRequest(req *CreateTunnelRequest) error {
	if req.LocalPort < 1 || req.LocalPort > 65535 {
		return fmt.Errorf("invalid local port: must be between 1 and 65535")
	}

	switch req.Type {
	case "http":
		// HTTP tunnels are always valid
	case "tcp":
		if req.RemotePort == nil {
			return fmt.Errorf("remote port required for TCP tunnels")
		}
		if *req.RemotePort < 1024 || *req.RemotePort > 65535 {
			return fmt.Errorf("invalid remote port: must be between 1024 and 65535")
		}
	case "stcp", "xtcp":
		if req.SecretKey == nil || len(*req.SecretKey) < 8 {
			return fmt.Errorf("secret key required and must be at least 8 characters")
		}
	case "visitor":
		if req.ServerName == nil {
			return fmt.Errorf("server name required for visitor mode")
		}
		if req.SecretKey == nil {
			return fmt.Errorf("secret key required for visitor mode")
		}
	default:
		return fmt.Errorf("invalid tunnel type: %s", req.Type)
	}

	return nil
}

func (s *Server) getTunnelName(req *CreateTunnelRequest) string {
	if req.Name != nil {
		return *req.Name
	}
	if req.Subdomain != nil {
		return *req.Subdomain
	}
	// TODO: Generate random name
	return fmt.Sprintf("tunnel-%d", time.Now().Unix())
}

func (s *Server) buildTunnelConfig(tunnelID string, req *CreateTunnelRequest) (string, error) {
	// Get API key
	apiKey, _ := config.GetAPIKey()
	if apiKey == "" {
		apiKey = s.apiKey
	}

	// Build tunnel name/subdomain
	tunnelName := s.getTunnelName(req)

	// Create TunnelConfig
	cfg := &config.TunnelConfig{
		ServerAddr:        config.DefaultServerAddr,
		ServerPort:        config.DefaultServerPort,
		APIKey:            apiKey,
		LocalPort:         req.LocalPort,
		LocalIP:           "127.0.0.1",
		Subdomain:         tunnelName,
		ExplicitSubdomain: req.Subdomain != nil,
		ProxyType:         req.Type,
		UseEncryption:     req.Encryption,
		UseCompression:    req.Compression,
	}

	// Set type-specific fields
	if req.RemotePort != nil {
		cfg.RemotePort = *req.RemotePort
	}
	if req.SecretKey != nil {
		cfg.SecretKey = *req.SecretKey
	}
	if req.BandwidthLimit != nil {
		cfg.BandwidthLimit = *req.BandwidthLimit
	}

	// Generate TOML config file
	if req.Type == "visitor" {
		return config.GenerateVisitorTOML(cfg)
	}

	return config.GenerateTOML(cfg)
}

func (s *Server) requestToConfig(req *CreateTunnelRequest) map[string]interface{} {
	cfg := map[string]interface{}{
		"type":      req.Type,
		"localPort": req.LocalPort,
	}

	if req.Subdomain != nil {
		cfg["subdomain"] = *req.Subdomain
	}
	if req.RemotePort != nil {
		cfg["remotePort"] = *req.RemotePort
	}
	if req.Encryption {
		cfg["encryption"] = true
	}
	if req.Compression {
		cfg["compression"] = true
	}

	return cfg
}

func (s *Server) waitForTunnelURL(mgr *tunnel.Manager, instance *TunnelInstance, timeout time.Duration) (string, error) {
	// TODO: Actually wait for tunnel to connect and extract URL from output
	// For now, generate a mock URL
	time.Sleep(1 * time.Second)

	if instance.Type == "http" {
		subdomain := instance.Name
		return fmt.Sprintf("https://%s.lum.tools", subdomain), nil
	}

	if instance.Type == "tcp" && instance.RemotePort != nil {
		return fmt.Sprintf("frp.lum.tools:%d", *instance.RemotePort), nil
	}

	return "", nil
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message, code string) {
	respondJSON(w, status, ErrorResponse{
		Error: message,
		Code:  code,
	})
}
