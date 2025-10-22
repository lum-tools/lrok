package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// Request represents a captured HTTP request/response
type Request struct {
	ID              string              `json:"id"`
	Timestamp       time.Time           `json:"timestamp"`
	Method          string              `json:"method"`
	Path            string              `json:"path"`
	StatusCode      int                 `json:"status_code"`
	Duration        time.Duration       `json:"duration"`
	RequestHeaders  map[string]string   `json:"request_headers"`
	ResponseHeaders map[string]string   `json:"response_headers"`
	RequestBody     string              `json:"request_body"`
	ResponseBody    string              `json:"response_body"`
	BytesIn         int64               `json:"bytes_in"`
	BytesOut        int64               `json:"bytes_out"`
}

// Proxy captures and forwards HTTP requests
type Proxy struct {
	targetURL    *url.URL
	server       *http.Server
	port         int
	requests     []*Request
	requestsMu   sync.RWMutex
	maxRequests  int
	listeners    []chan *Request
	listenersMu  sync.RWMutex
	totalBytesIn  int64
	totalBytesOut int64
	totalConns    int64
	statsMu       sync.RWMutex
}

// New creates a new proxy to the target port
func New(targetPort int, maxRequests int) *Proxy {
	if maxRequests == 0 {
		maxRequests = 100
	}
	
	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", targetPort))
	
	return &Proxy{
		targetURL:   target,
		requests:    make([]*Request, 0, maxRequests),
		maxRequests: maxRequests,
		listeners:   make([]chan *Request, 0),
	}
}

// Start starts the proxy on an available port and waits for it to be ready
func (p *Proxy) Start() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	
	p.port = listener.Addr().(*net.TCPAddr).Port
	
	proxy := httputil.NewSingleHostReverseProxy(p.targetURL)
	
	// Customize the director to capture requests
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
	}
	
	// Custom transport to capture response
	proxy.Transport = &captureTransport{
		base:  http.DefaultTransport,
		proxy: p,
	}
	
	// Add health check handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
	mux.HandleFunc("/__lrok_health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	p.server = &http.Server{
		Handler: mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	// Start server in background
	ready := make(chan bool)
	go func() {
		ready <- true
		p.server.Serve(listener)
	}()
	
	// Wait for server to be ready
	<-ready
	time.Sleep(200 * time.Millisecond) // Longer wait for full initialization
	
	// Verify proxy is responding to health checks
	if err := p.healthCheck(); err != nil {
		p.server.Close()
		return 0, fmt.Errorf("proxy health check failed: %w", err)
	}
	
	// CRITICAL: Warm up the reverse proxy by making a test request to target
	// This initializes the connection pool and ensures proxy is fully ready
	if err := p.warmUp(); err != nil {
		p.server.Close()
		return 0, fmt.Errorf("proxy warm-up failed: %w", err)
	}
	
	return p.port, nil
}

// healthCheck verifies the proxy is responding
func (p *Proxy) healthCheck() error {
	url := fmt.Sprintf("http://127.0.0.1:%d/__lrok_health", p.port)
	client := &http.Client{Timeout: 2 * time.Second}
	
	for i := 0; i < 10; i++ {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if err == nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	return fmt.Errorf("proxy not responding after 10 attempts")
}

// warmUp makes test requests through the proxy to ensure it's fully ready
func (p *Proxy) warmUp() error {
	// Make a few test requests to warm up the reverse proxy
	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
			DisableCompression:  false,
		},
	}
	
	proxyURL := fmt.Sprintf("http://127.0.0.1:%d/", p.port)
	
	// Make 3 warm-up requests to ensure connection pool is ready
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", proxyURL, nil)
		req.Header.Set("X-Lrok-Warmup", "true")
		
		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		io.ReadAll(resp.Body) // Consume body
		resp.Body.Close()
		
		// After first successful request, proxy is warmed up
		if i == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	return nil
}

// Stop stops the proxy
func (p *Proxy) Stop() error {
	if p.server != nil {
		return p.server.Close()
	}
	return nil
}

// GetRequests returns all captured requests
func (p *Proxy) GetRequests() []*Request {
	p.requestsMu.RLock()
	defer p.requestsMu.RUnlock()
	
	result := make([]*Request, len(p.requests))
	copy(result, p.requests)
	return result
}

// Subscribe subscribes to new requests
func (p *Proxy) Subscribe() chan *Request {
	p.listenersMu.Lock()
	defer p.listenersMu.Unlock()
	
	ch := make(chan *Request, 10)
	p.listeners = append(p.listeners, ch)
	return ch
}

// Unsubscribe unsubscribes from new requests
func (p *Proxy) Unsubscribe(ch chan *Request) {
	p.listenersMu.Lock()
	defer p.listenersMu.Unlock()
	
	for i, listener := range p.listeners {
		if listener == ch {
			p.listeners = append(p.listeners[:i], p.listeners[i+1:]...)
			close(ch)
			break
		}
	}
}

// GetStats returns current traffic stats
func (p *Proxy) GetStats() (int64, int64, int64) {
	p.statsMu.RLock()
	defer p.statsMu.RUnlock()
	return p.totalBytesIn, p.totalBytesOut, p.totalConns
}

// addRequest adds a request to the buffer
func (p *Proxy) addRequest(req *Request) {
	p.requestsMu.Lock()
	p.requests = append(p.requests, req)
	if len(p.requests) > p.maxRequests {
		p.requests = p.requests[1:]
	}
	p.requestsMu.Unlock()
	
	// Update stats
	p.statsMu.Lock()
	p.totalBytesIn += req.BytesIn
	p.totalBytesOut += req.BytesOut
	p.totalConns++
	p.statsMu.Unlock()
	
	// Notify listeners
	p.listenersMu.RLock()
	for _, listener := range p.listeners {
		select {
		case listener <- req:
		default:
		}
	}
	p.listenersMu.RUnlock()
}

// captureTransport wraps http.Transport to capture responses
type captureTransport struct {
	base  http.RoundTripper
	proxy *Proxy
}

func (t *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	reqID := fmt.Sprintf("%d", time.Now().UnixNano())
	
	// Capture request
	reqHeaders := make(map[string]string)
	for k, v := range req.Header {
		if len(v) > 0 {
			reqHeaders[k] = v[0]
		}
	}
	
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}
	
	// Forward request
	resp, err := t.base.RoundTrip(req)
	duration := time.Since(start)
	
	if err != nil {
		return nil, err
	}
	
	// Capture response
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}
	
	var respBody []byte
	if resp.Body != nil {
		respBody, _ = io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
	}
	
	// Store request
	captured := &Request{
		ID:              reqID,
		Timestamp:       start,
		Method:          req.Method,
		Path:            req.URL.Path,
		StatusCode:      resp.StatusCode,
		Duration:        duration,
		RequestHeaders:  reqHeaders,
		ResponseHeaders: respHeaders,
		RequestBody:     string(reqBody),
		ResponseBody:    string(respBody),
		BytesIn:         int64(len(reqBody)),
		BytesOut:        int64(len(respBody)),
	}
	
	t.proxy.addRequest(captured)
	
	return resp, nil
}

