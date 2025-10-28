package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// HTTPTestServer wraps an HTTP server for testing
type HTTPTestServer struct {
	server *http.Server
	port   int
}

// Close stops the HTTP test server
func (s *HTTPTestServer) Close() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// Port returns the port the server is listening on
func (s *HTTPTestServer) Port() int {
	return s.port
}

// startHTTPTestServer starts an HTTP server with test endpoints
func startHTTPTestServer(port int) *HTTPTestServer {
	mux := http.NewServeMux()
	
	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<h1>HTTP Tunnel Test Server</h1><p>Server is running on port %d!</p>", port)
	})
	
	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"port":      port,
		}
		json.NewEncoder(w).Encode(response)
	})
	
	// Webhook endpoint
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		body, _ := io.ReadAll(r.Body)
		
		response := map[string]interface{}{
			"webhook":   "received",
			"data":      string(body),
			"timestamp": time.Now().Unix(),
			"method":    r.Method,
			"headers":   r.Header,
		}
		json.NewEncoder(w).Encode(response)
	})
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	
	// Start server in goroutine
	go func() {
		server.ListenAndServe()
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	return &HTTPTestServer{
		server: server,
		port:   port,
	}
}

// TCPEchoServer wraps a TCP server that echoes messages
type TCPEchoServer struct {
	listener net.Listener
	port     int
	done     chan bool
}

// Close stops the TCP echo server
func (s *TCPEchoServer) Close() error {
	close(s.done)
	return s.listener.Close()
}

// Port returns the port the server is listening on
func (s *TCPEchoServer) Port() int {
	return s.port
}

// startTCPEchoServer starts a TCP server that echoes messages
func startTCPEchoServer(port int) (*TCPEchoServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to start TCP server: %w", err)
	}
	
	server := &TCPEchoServer{
		listener: listener,
		port:     port,
		done:     make(chan bool),
	}
	
	// Start server in goroutine
	go func() {
		for {
			select {
			case <-server.done:
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					continue
				}
				
				go func(c net.Conn) {
					defer c.Close()
					
					// Set timeouts to prevent hanging connections
					c.SetDeadline(time.Now().Add(30 * time.Second))
					
					// Use explicit read/write loop instead of io.Copy
					buffer := make([]byte, 1024)
					for {
						// Read data from client
						n, err := c.Read(buffer)
						if err != nil {
							// Connection closed or error - exit gracefully
							return
						}
						
						// Echo the data back to client
						_, err = c.Write(buffer[:n])
						if err != nil {
							// Write error - exit gracefully
							return
						}
						
						// Reset deadline for next read
						c.SetDeadline(time.Now().Add(30 * time.Second))
					}
				}(conn)
			}
		}
	}()
	
	return server, nil
}

// UDPEchoServer wraps a UDP server that echoes packets
type UDPEchoServer struct {
	conn *net.UDPConn
	port int
	done chan bool
}

// Close stops the UDP echo server
func (s *UDPEchoServer) Close() error {
	close(s.done)
	return s.conn.Close()
}

// Port returns the port the server is listening on
func (s *UDPEchoServer) Port() int {
	return s.port
}

// startUDPEchoServer starts a UDP server that echoes packets
// Note: This is kept for potential future use but UDP tunnels are not supported
func startUDPEchoServer(port int) (*UDPEchoServer, error) {
	// Bind to 0.0.0.0 instead of 127.0.0.1 to ensure FRP can reach the server
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP addr: %w", err)
	}
	
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to start UDP server: %w", err)
	}
	
	server := &UDPEchoServer{
		conn: conn,
		port: port,
		done: make(chan bool),
	}
	
	// Start server in goroutine
	go func() {
		buffer := make([]byte, 1024)
		for {
			select {
			case <-server.done:
				return
			default:
				n, clientAddr, err := conn.ReadFromUDP(buffer)
				if err != nil {
					continue
				}
				
				// Echo the packet back
				conn.WriteToUDP(buffer[:n], clientAddr)
			}
		}
	}()
	
	return server, nil
}
