package azureconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Load reads and parses an Azure configuration file
func Load(path string) (*Config, error) {
	// Expand tilde in path
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &cfg, nil
}

// LoadOrDefault loads a configuration file or returns the default if not found
func LoadOrDefault(path string) (*Config, error) {
	cfg, err := Load(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	return cfg, nil
}

// Save writes a configuration to a file
func Save(cfg *Config, path string) error {
	// Expand tilde in path
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ExpandPaths expands tilde and environment variables in path configurations
func (c *Config) ExpandPaths() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Expand SSH key path
	if strings.HasPrefix(c.Infrastructure.SSHPublicKeyPath, "~/") {
		c.Infrastructure.SSHPublicKeyPath = filepath.Join(home, c.Infrastructure.SSHPublicKeyPath[2:])
	}

	return nil
}

// GetAppURL generates the app URL by replacing the public IP placeholder
func (c *Config) GetAppURL(publicIP string) string {
	return strings.ReplaceAll(c.Coolify.AppURLTemplate, "{public_ip}", publicIP)
}

// GetPusherHost generates the pusher host by replacing the public IP placeholder
func (c *Config) GetPusherHost(publicIP string) string {
	return strings.ReplaceAll(c.Coolify.PusherHostTemplate, "{public_ip}", publicIP)
}
