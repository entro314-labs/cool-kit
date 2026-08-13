package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var globalCfg *GlobalConfig

const (
	globalConfigDirName  = ".cool-kit"
	globalConfigFileName = "global.json"
)

// LoadGlobal loads the global configuration from disk
func LoadGlobal() (*GlobalConfig, error) {
	if globalCfg != nil {
		return globalCfg, nil
	}

	configPath, err := getGlobalConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			globalCfg = &GlobalConfig{}
			return globalCfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &GlobalConfig{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	globalCfg = cfg
	return cfg, nil
}

// SaveGlobal saves the global configuration to disk
func SaveGlobal(cfg *GlobalConfig) error {
	configPath, err := getGlobalConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	globalCfg = cfg
	return nil
}

// getGlobalConfigPath returns the path to the global config file
func getGlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, globalConfigDirName, globalConfigFileName), nil
}

// ClearGlobal removes the global configuration
func ClearGlobal() error {
	configPath, err := getGlobalConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config: %w", err)
	}
	_ = deleteToken("global-coolify-token")

	globalCfg = nil
	return nil
}
