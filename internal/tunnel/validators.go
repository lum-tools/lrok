package tunnel

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ReservedPorts contains commonly reserved ports that should be avoided
var ReservedPorts = map[int]bool{
	22:   true, // SSH
	23:   true, // Telnet
	25:   true, // SMTP
	53:   true, // DNS
	80:   true, // HTTP
	110:  true, // POP3
	143:  true, // IMAP
	443:  true, // HTTPS
	993:  true, // IMAPS
	995:  true, // POP3S
	3389: true, // RDP
	5432: true, // PostgreSQL
	3306: true, // MySQL
	6379: true, // Redis
	27017: true, // MongoDB
}

// ValidatePort validates a port number
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	
	if ReservedPorts[port] {
		return fmt.Errorf("port %d is commonly reserved for system services, consider using a different port", port)
	}
	
	return nil
}

// ValidateSecretKey validates a secret key for STCP/XTCP tunnels
func ValidateSecretKey(secretKey string) error {
	if len(secretKey) < 8 {
		return fmt.Errorf("secret key must be at least 8 characters long")
	}
	
	if len(secretKey) > 64 {
		return fmt.Errorf("secret key must be no more than 64 characters long")
	}
	
	// Check for common weak patterns
	weakPatterns := []string{
		"password", "secret", "key", "123456", "abcdef",
		"qwerty", "admin", "test", "demo", "example",
	}
	
	lowerKey := strings.ToLower(secretKey)
	for _, pattern := range weakPatterns {
		if strings.Contains(lowerKey, pattern) {
			return fmt.Errorf("secret key contains common weak pattern '%s', please use a stronger key", pattern)
		}
	}
	
	return nil
}

// ValidateBandwidthLimit validates bandwidth limit format
func ValidateBandwidthLimit(bandwidth string) error {
	if bandwidth == "" {
		return nil // Empty is valid (no limit)
	}
	
	// Match patterns like "1MB", "500KB", "2.5MB", "1000KB"
	pattern := regexp.MustCompile(`^(\d+(?:\.\d+)?)(MB|KB)$`)
	matches := pattern.FindStringSubmatch(strings.ToUpper(bandwidth))
	
	if len(matches) != 3 {
		return fmt.Errorf("bandwidth limit must be in format 'XMB' or 'XKB' (e.g., '1MB', '500KB'), got '%s'", bandwidth)
	}
	
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return fmt.Errorf("invalid bandwidth value: %v", err)
	}
	
	if value <= 0 {
		return fmt.Errorf("bandwidth limit must be greater than 0, got %f", value)
	}
	
	unit := matches[2]
	if unit == "KB" && value > 1024 {
		return fmt.Errorf("bandwidth limit in KB cannot exceed 1024KB (1MB), got %fKB", value)
	}
	
	return nil
}

// ValidateProxyType validates the proxy type
func ValidateProxyType(proxyType string) error {
	validTypes := map[string]bool{
		"http":  true,
		"https": true,
		"tcp":   true,
		"udp":   true,
		"stcp":  true,
		"xtcp":  true,
	}
	
	if !validTypes[proxyType] {
		return fmt.Errorf("invalid proxy type '%s', must be one of: http, https, tcp, udp, stcp, xtcp", proxyType)
	}
	
	return nil
}

// ValidateHealthCheckType validates the health check type
func ValidateHealthCheckType(healthCheckType string) error {
	if healthCheckType == "" {
		return nil // Empty is valid (no health check)
	}
	
	validTypes := map[string]bool{
		"tcp":  true,
		"http": true,
	}
	
	if !validTypes[healthCheckType] {
		return fmt.Errorf("invalid health check type '%s', must be one of: tcp, http", healthCheckType)
	}
	
	return nil
}

// ValidateTunnelName validates tunnel name/subdomain
func ValidateTunnelName(name string) error {
	if len(name) < 3 {
		return fmt.Errorf("tunnel name must be at least 3 characters long")
	}
	
	if len(name) > 63 {
		return fmt.Errorf("tunnel name must be no more than 63 characters long")
	}
	
	// Check for valid characters (alphanumeric and hyphens)
	pattern := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !pattern.MatchString(name) {
		return fmt.Errorf("tunnel name can only contain letters, numbers, and hyphens")
	}
	
	// Cannot start or end with hyphen
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("tunnel name cannot start or end with a hyphen")
	}
	
	// Check for reserved names
	reservedNames := map[string]bool{
		"www":     true,
		"api":     true,
		"admin":   true,
		"mail":    true,
		"ftp":     true,
		"blog":    true,
		"shop":    true,
		"app":     true,
		"dev":     true,
		"test":    true,
		"staging": true,
		"prod":    true,
		"demo":    true,
	}
	
	if reservedNames[strings.ToLower(name)] {
		return fmt.Errorf("tunnel name '%s' is reserved, please choose a different name", name)
	}
	
	return nil
}
