package tests

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestServers(t *testing.T) {
	// Test HTTP server
	port := getRandomPort()
	server := startHTTPTestServer(port)
	defer server.Close()
	
	// Test HTTP server locally
	time.Sleep(100 * time.Millisecond)
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	require.NoError(t, err)
	conn.Close()
	
	// Test TCP echo server
	tcpPort := getRandomPort()
	tcpServer, err := startTCPEchoServer(tcpPort)
	require.NoError(t, err)
	defer tcpServer.Close()
	
	// Test TCP server locally
	time.Sleep(100 * time.Millisecond)
	tcpConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", tcpPort))
	require.NoError(t, err)
	defer tcpConn.Close()
	
	// Send test message
	testMsg := "Hello Test"
	_, err = tcpConn.Write([]byte(testMsg + "\n"))
	require.NoError(t, err)
	
	// Read echo response
	buffer := make([]byte, 1024)
	n, err := tcpConn.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, testMsg+"\n", string(buffer[:n]))
	
	// Test UDP echo server
	udpPort := getRandomPort()
	udpServer, err := startUDPEchoServer(udpPort)
	require.NoError(t, err)
	defer udpServer.Close()
	
	// Test UDP server locally
	time.Sleep(100 * time.Millisecond)
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", udpPort))
	require.NoError(t, err)
	
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	require.NoError(t, err)
	defer udpConn.Close()
	
	// Send test packet
	_, err = udpConn.Write([]byte(testMsg))
	require.NoError(t, err)
	
	// Read echo response
	n, err = udpConn.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, testMsg, string(buffer[:n]))
	
	t.Log("✅ All test servers working correctly")
}
