package local

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents local Coolify instance configuration
type Config struct {
	AdminEmail    string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`
	AppPort       int    `json:"app_port"`
	WebSocketPort int    `json:"websocket_port"`
	AppDebug      bool   `json:"app_debug"`
	LogLevel      string `json:"log_level"`
	WorkDir       string `json:"work_dir"`

	// Generated credentials
	AppID           string `json:"app_id"`
	AppKey          string `json:"app_key"`
	DBPassword      string `json:"db_password"`
	RedisPassword   string `json:"redis_password"`
	PusherAppID     string `json:"pusher_app_id"`
	PusherAppKey    string `json:"pusher_app_key"`
	PusherAppSecret string `json:"pusher_app_secret"`
}

// DefaultConfig returns the default local configuration
func DefaultConfig() *Config {
	return &Config{
		AdminEmail:    "admin@coolify.local",
		AdminPassword: "admin123",
		AppPort:       8000,
		WebSocketPort: 6001,
		AppDebug:      true,
		LogLevel:      "debug",
		WorkDir:       "./coolify-local",
	}
}

// GenerateCredentials generates secure random credentials
func (c *Config) GenerateCredentials() error {
	var err error

	// Generate App ID (32 hex chars)
	c.AppID, err = generateHex(16)
	if err != nil {
		return fmt.Errorf("failed to generate app ID: %w", err)
	}

	// Generate App Key (base64:32 bytes)
	appKeyBytes, err := generateBytes(32)
	if err != nil {
		return fmt.Errorf("failed to generate app key: %w", err)
	}
	c.AppKey = "base64:" + base64.StdEncoding.EncodeToString(appKeyBytes)

	// Generate DB Password (32 alphanumeric chars)
	c.DBPassword, err = generatePassword(32)
	if err != nil {
		return fmt.Errorf("failed to generate DB password: %w", err)
	}

	// Generate Redis Password (32 alphanumeric chars)
	c.RedisPassword, err = generatePassword(32)
	if err != nil {
		return fmt.Errorf("failed to generate Redis password: %w", err)
	}

	// Generate Pusher credentials (64 hex chars each)
	c.PusherAppID, err = generateHex(32)
	if err != nil {
		return fmt.Errorf("failed to generate Pusher app ID: %w", err)
	}

	c.PusherAppKey, err = generateHex(32)
	if err != nil {
		return fmt.Errorf("failed to generate Pusher app key: %w", err)
	}

	c.PusherAppSecret, err = generateHex(32)
	if err != nil {
		return fmt.Errorf("failed to generate Pusher app secret: %w", err)
	}

	return nil
}

// Save writes the configuration to a file
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Load reads configuration from a file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &cfg, nil
}

// LoadOrDefault loads configuration or returns default if not found
func LoadOrDefault(path string) (*Config, error) {
	cfg, err := Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	return cfg, nil
}

// Helper functions for credential generation

func generateBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func generateHex(n int) (string, error) {
	bytes, err := generateBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generatePassword(length int) (string, error) {
	bytes, err := generateBytes(length)
	if err != nil {
		return "", err
	}

	// Convert to base64 and clean up special characters
	password := base64.StdEncoding.EncodeToString(bytes)
	// Remove special characters
	password = cleanPassword(password, length)

	return password, nil
}

func cleanPassword(s string, maxLen int) string {
	result := make([]byte, 0, maxLen)
	for i := 0; i < len(s) && len(result) < maxLen; i++ {
		c := s[i]
		// Keep only alphanumeric characters
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result = append(result, c)
		}
	}
	return string(result)
}
