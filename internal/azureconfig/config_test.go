package azureconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test infrastructure defaults
	if cfg.Infrastructure.Location != "swedencentral" {
		t.Errorf("Expected location 'swedencentral', got '%s'", cfg.Infrastructure.Location)
	}

	if cfg.Infrastructure.VMSize != "Standard_B2s" {
		t.Errorf("Expected VM size 'Standard_B2s', got '%s'", cfg.Infrastructure.VMSize)
	}

	if cfg.Infrastructure.AdminUsername != "azureuser" {
		t.Errorf("Expected admin username 'azureuser', got '%s'", cfg.Infrastructure.AdminUsername)
	}

	// Test networking defaults
	if cfg.Networking.AppPort != 80 {
		t.Errorf("Expected app port 80, got %d", cfg.Networking.AppPort)
	}

	if cfg.Networking.SSHPort != 22 {
		t.Errorf("Expected SSH port 22, got %d", cfg.Networking.SSHPort)
	}

	// Test Coolify defaults
	if cfg.Coolify.DefaultAdminEmail != "admin@coolify.local" {
		t.Errorf("Expected admin email 'admin@coolify.local', got '%s'", cfg.Coolify.DefaultAdminEmail)
	}

	// Test Docker defaults
	if cfg.Docker.RegistryURL != "ghcr.io" {
		t.Errorf("Expected registry URL 'ghcr.io', got '%s'", cfg.Docker.RegistryURL)
	}
}

func TestLoadAndSave(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")

	// Create test config
	cfg := DefaultConfig()
	cfg.Infrastructure.Location = "westeurope"
	cfg.Infrastructure.VMSize = "Standard_B4ms"

	// Save config
	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config
	loadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if loadedCfg.Infrastructure.Location != "westeurope" {
		t.Errorf("Expected location 'westeurope', got '%s'", loadedCfg.Infrastructure.Location)
	}

	if loadedCfg.Infrastructure.VMSize != "Standard_B4ms" {
		t.Errorf("Expected VM size 'Standard_B4ms', got '%s'", loadedCfg.Infrastructure.VMSize)
	}
}

func TestLoadOrDefault(t *testing.T) {
	// Test with non-existent file (should return default)
	cfg, err := LoadOrDefault("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("LoadOrDefault should not fail for non-existent file: %v", err)
	}

	if cfg.Infrastructure.Location != "swedencentral" {
		t.Errorf("Expected default location, got '%s'", cfg.Infrastructure.Location)
	}

	// Test with existing file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")

	customCfg := DefaultConfig()
	customCfg.Infrastructure.Location = "northeurope"
	if err := Save(customCfg, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedCfg, err := LoadOrDefault(configPath)
	if err != nil {
		t.Fatalf("LoadOrDefault failed: %v", err)
	}

	if loadedCfg.Infrastructure.Location != "northeurope" {
		t.Errorf("Expected loaded location 'northeurope', got '%s'", loadedCfg.Infrastructure.Location)
	}
}

func TestGetAppURL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Coolify.AppURLTemplate = "http://{public_ip}"

	url := cfg.GetAppURL("1.2.3.4")
	expected := "http://1.2.3.4"

	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}

func TestGetPusherHost(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Coolify.PusherHostTemplate = "{public_ip}"

	host := cfg.GetPusherHost("1.2.3.4")
	expected := "1.2.3.4"

	if host != expected {
		t.Errorf("Expected host '%s', got '%s'", expected, host)
	}
}

func TestValidation(t *testing.T) {
	// Test valid config
	cfg := DefaultConfig()

	// Create temporary SSH key file for validation
	tmpDir := t.TempDir()
	sshKeyPath := filepath.Join(tmpDir, "id_rsa.pub")
	if err := os.WriteFile(sshKeyPath, []byte("ssh-rsa AAAAB3..."), 0644); err != nil {
		t.Fatalf("Failed to create temp SSH key: %v", err)
	}
	cfg.Infrastructure.SSHPublicKeyPath = sshKeyPath

	if err := cfg.Validate(); err != nil {
		t.Errorf("Valid config failed validation: %v", err)
	}

	// Test invalid configs
	tests := []struct {
		name     string
		modify   func(*Config)
		shouldFail bool
	}{
		{
			name: "empty location",
			modify: func(c *Config) {
				c.Infrastructure.Location = ""
			},
			shouldFail: true,
		},
		{
			name: "empty VM size",
			modify: func(c *Config) {
				c.Infrastructure.VMSize = ""
			},
			shouldFail: true,
		},
		{
			name: "invalid app port",
			modify: func(c *Config) {
				c.Networking.AppPort = 0
			},
			shouldFail: true,
		},
		{
			name: "port too high",
			modify: func(c *Config) {
				c.Networking.AppPort = 70000
			},
			shouldFail: true,
		},
		{
			name: "empty admin email",
			modify: func(c *Config) {
				c.Coolify.DefaultAdminEmail = ""
			},
			shouldFail: true,
		},
		{
			name: "empty docker image",
			modify: func(c *Config) {
				c.Docker.AppImage = ""
			},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCfg := DefaultConfig()
			testCfg.Infrastructure.SSHPublicKeyPath = sshKeyPath
			tt.modify(testCfg)

			err := testCfg.Validate()
			if tt.shouldFail && err == nil {
				t.Error("Expected validation to fail, but it passed")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Expected validation to pass, but it failed: %v", err)
			}
		})
	}
}
