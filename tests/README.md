# lrok Tunnel Testing Guide

This guide provides comprehensive testing procedures for all tunnel types implemented in the lrok platform using Go integration tests.

## Go Test Suite Overview

The lrok tests have been rewritten in Go for better reliability, maintainability, and integration with the Go ecosystem.

### Test Files

- `integration_test.go` - Main integration tests for all tunnel protocols
- `helpers.go` - Test utility functions (binary building, tunnel management, etc.)
- `servers.go` - Test server implementations (HTTP, TCP echo, UDP echo)
- `testenv.go` - Test environment configuration and constants
- `server_test.go` - Unit tests for test server implementations
- `config_test.go` - Configuration generation tests

## Individual Protocol Tests

### HTTP Tunnel Test (`TestHTTPTunnel`)
- **Purpose**: Tests HTTP/HTTPS tunnel functionality
- **Features Tested**:
  - Basic connectivity through public URL
  - Health endpoint (`/health`)
  - Webhook simulation (`/webhook`)
  - HTTPS automatic SSL
  - Custom subdomain support
- **Usage**: `go test -v -run TestHTTPTunnel`

### TCP Tunnel Test (`TestTCPTunnel`)
- **Purpose**: Tests TCP tunnel functionality
- **Features Tested**:
  - TCP echo server connectivity
  - Encryption (`--encrypt`)
  - Compression (`--compress`)
  - Remote port forwarding
- **Usage**: `go test -v -run TestTCPTunnel`
- **Note**: Requires external port exposure (see Infrastructure Limitations)

### UDP Tunnel Test (`TestUDPTunnel`)
- **Purpose**: Tests UDP tunnel functionality
- **Features Tested**:
  - UDP echo server connectivity
  - Packet transmission and echo
  - Remote port forwarding
- **Usage**: `go test -v -run TestUDPTunnel`
- **Note**: Requires external port exposure (see Infrastructure Limitations)

### STCP Tunnel Test (`TestSTCPTunnel`)
- **Purpose**: Tests STCP (Secret TCP) tunnel functionality
- **Features Tested**:
  - Server-side tunnel creation
  - Visitor mode connectivity
  - Secret key authentication
  - Local service access through visitor
- **Usage**: `go test -v -run TestSTCPTunnel`

### XTCP Tunnel Test (`TestXTCPTunnel`)
- **Purpose**: Tests XTCP (P2P) tunnel functionality
- **Features Tested**:
  - P2P tunnel creation
  - Visitor mode with P2P negotiation
  - Secret key authentication
  - Local service access through visitor
- **Usage**: `go test -v -run TestXTCPTunnel`

## Test Execution

### Prerequisites

1. **API Key**: Set the test API key in `testenv.go`
2. **Go Dependencies**: Run `go mod tidy` to install dependencies
3. **FRP Server**: Ensure the production FRP server is running and accessible

### Running Individual Tests

```bash
cd cli/lrok/tests

# Test HTTP tunnels (fully functional)
go test -v -run TestHTTPTunnel -timeout 60s

# Test STCP tunnels (fully functional)
go test -v -run TestSTCPTunnel -timeout 60s

# Test XTCP tunnels (fully functional)
go test -v -run TestXTCPTunnel -timeout 60s

# Test TCP tunnels (requires infrastructure changes)
go test -v -run TestTCPTunnel -timeout 120s

# Test UDP tunnels (requires infrastructure changes)
go test -v -run TestUDPTunnel -timeout 120s
```

### Running All Tests

```bash
cd cli/lrok/tests
go test -v -timeout 300s
```

### Running Test Servers Only

```bash
cd cli/lrok/tests
go test -v -run TestTestServers
```

## Test Features

### Automatic Cleanup
- All tests use `defer` statements for automatic cleanup
- Temporary binaries are removed after tests
- Tunnel processes are gracefully terminated
- Test servers are automatically stopped

### Random Port Generation
- Each test uses random ports to avoid conflicts
- Random tunnel names prevent naming collisions
- Tests are isolated and can run in parallel

### Real Integration Testing
- Tests use the actual built lrok CLI binary
- Tests connect to the production FRP server
- Tests validate real tunnel functionality
- Tests verify actual network connectivity

## Infrastructure Limitations

### TCP/UDP Tunnel Port Exposure

**Issue**: TCP and UDP tunnels require external port exposure that is not currently configured in the Kubernetes infrastructure.

**Current Status**: 
- HTTP tunnels work perfectly (use subdomain routing)
- STCP/XTCP tunnels work perfectly (use visitor mode)
- TCP/UDP tunnels fail external connectivity tests

**Root Cause**: The Kubernetes service configuration only exposes the FRP control port (7000) externally. Tunnel ports (10000-60000) are not exposed through LoadBalancer or Ingress.

**Workaround**: 
- TCP/UDP tunnel creation and configuration works correctly
- Tunnels are successfully created on the FRP server
- External connectivity fails due to infrastructure limitations
- Tests validate tunnel creation and configuration

### Required Infrastructure Changes

To enable TCP/UDP tunnel testing:

1. **Expose Tunnel Ports**: Configure Kubernetes to expose ports 10000-60000 externally
2. **LoadBalancer Service**: Create external service for tunnel ports
3. **Firewall Rules**: Ensure external access to tunnel ports
4. **DNS Configuration**: Configure DNS for tunnel port access

## Test Results Interpretation

### Successful Tests
- ✅ **HTTP Tunnels**: Full functionality, external connectivity works
- ✅ **STCP Tunnels**: Full functionality, visitor mode works
- ✅ **XTCP Tunnels**: Full functionality, P2P mode works
- ⚠️ **TCP Tunnels**: Tunnel creation works, external connectivity fails
- ⚠️ **UDP Tunnels**: Tunnel creation works, external connectivity fails

### Test Output
- Tests show tunnel creation success
- Tests validate configuration generation
- Tests verify local server functionality
- External connectivity tests fail for TCP/UDP (expected due to infrastructure)

## Development and Debugging

### Test Server Validation
```bash
go test -v -run TestTestServers
```

### Configuration Testing
```bash
go test -v -run TestTunnelConfig
```

### Verbose Output
```bash
go test -v -run TestHTTPTunnel
```

### Test Timeout
```bash
go test -v -run TestHTTPTunnel -timeout 120s
```

## Continuous Integration

### CI/CD Integration
```bash
# In CI pipeline
cd cli/lrok/tests
go test -v -timeout 300s
```

### Test Automation
```bash
# Automated test script
#!/bin/bash
cd cli/lrok/tests
go test -v -run TestHTTPTunnel
go test -v -run TestSTCPTunnel
go test -v -run TestXTCPTunnel
```

## Troubleshooting

### Common Issues

#### 1. API Key Issues
- Verify API key is set in `testenv.go`
- Check API key format (must start with 'lum_')
- Ensure API key is valid on the platform

#### 2. Build Issues
- Run `go mod tidy` to update dependencies
- Ensure Go 1.22+ is installed
- Check that lrok CLI builds successfully

#### 3. Network Connectivity
- Verify FRP server is accessible on port 7000
- Check that platform API is accessible
- Ensure no firewall blocking connections

#### 4. Test Failures
- Check test logs for specific error messages
- Verify test servers are starting correctly
- Ensure tunnel processes are not conflicting

### Debug Mode
```bash
# Run with verbose output
go test -v -run TestHTTPTunnel

# Run with timeout
go test -v -run TestHTTPTunnel -timeout 120s

# Run specific test
go test -v -run TestHTTPTunnel -timeout 60s
```

## Security Considerations

### Test Security
- All tests use temporary, isolated environments
- No sensitive data is logged
- API keys are masked in output
- Test servers are automatically cleaned up

### Production Safety
- Tests use non-production ports
- No interference with production services
- Automatic cleanup on completion
- Isolated test data

## Support and Maintenance

### Test Maintenance
- Update test scripts when new features are added
- Verify test data and expected outputs
- Maintain compatibility with API changes
- Update documentation as needed

### Reporting Issues
When reporting test failures:
1. Include full test output
2. Provide system information
3. Include API key format (masked)
4. Specify which test failed
5. Include relevant logs

### Test Updates
- Regular validation of test scripts
- Compatibility with new lrok versions
- Platform API changes
- Infrastructure updates