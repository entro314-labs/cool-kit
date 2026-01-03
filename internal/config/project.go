package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const projectConfigFile = ".coolify-deployer/config.json"

// LoadProject loads the project configuration from the current directory
func LoadProject() (*ProjectConfig, error) {
	configPath, err := getProjectConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	cfg := &ProjectConfig{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	return cfg, nil
}

// SaveProject saves the project configuration to the current directory
func SaveProject(cfg *ProjectConfig) error {
	configPath, err := getProjectConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// getProjectConfigPath returns the path to the project config file
func getProjectConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	return filepath.Join(cwd, projectConfigFile), nil
}

// DeleteProject removes the project configuration
func DeleteProject() error {
	configPath, err := getProjectConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove project config: %w", err)
	}

	// Try to remove the directory if it's empty
	configDir := filepath.Dir(configPath)
	os.Remove(configDir) // Ignore error - directory might not be empty

	return nil
}

// HasProject checks if the current directory has a project configuration
func HasProject() bool {
	configPath, err := getProjectConfigPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(configPath)
	return err == nil
}

// ProjectExists is an alias for HasProject for backwards compatibility
func ProjectExists() bool {
	return HasProject()
}
