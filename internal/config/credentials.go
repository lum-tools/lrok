package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Credentials stores user authentication config
type Credentials struct {
	APIKey string `toml:"api_key"`
}

// Config represents the lrok configuration file
type Config struct {
	Auth Credentials `toml:"auth"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".lrok")
	configFile := filepath.Join(configDir, "config.toml")

	return configFile, nil
}

// EnsureConfigDir ensures the config directory exists
func EnsureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".lrok")

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

// LoadConfig loads the configuration from file
func LoadConfig() (*Config, error) {
	configFile, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configFile, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveAPIKey saves an API key to the config file
func SaveAPIKey(apiKey string) error {
	config := &Config{
		Auth: Credentials{
			APIKey: apiKey,
		},
	}

	return SaveConfig(config)
}

// GetAPIKey retrieves the API key from the config file
func GetAPIKey() (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}

	if config.Auth.APIKey == "" {
		return "", fmt.Errorf("no API key found in config")
	}

	return config.Auth.APIKey, nil
}

// ClearConfig removes the config file
func ClearConfig() error {
	configFile, err := GetConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil // Already cleared
	}

	if err := os.Remove(configFile); err != nil {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	return nil
}

