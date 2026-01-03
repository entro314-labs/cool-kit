package local

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/entro314-labs/cool-kit/internal/ui"
)

// Setup performs complete local Coolify setup
func Setup() error {
	ui.Section("Local Coolify Setup")
	ui.Dim("Setting up local development environment")

	// Step 1: Check prerequisites
	ui.Info("Checking prerequisites...")
	if err := CheckPrerequisites(); err != nil {
		return fmt.Errorf("prerequisite check failed: %w", err)
	}
	ui.Success("Prerequisites verified")

	// Step 2: Collect configuration
	cfg, err := collectConfiguration()
	if err != nil {
		return fmt.Errorf("configuration collection failed: %w", err)
	}

	// Step 3: Generate credentials
	ui.Info("Generating secure credentials...")
	if err := cfg.GenerateCredentials(); err != nil {
		return fmt.Errorf("credential generation failed: %w", err)
	}
	ui.Success("Secure credentials generated")

	// Step 4: Create directory structure
	ui.Info("Creating directory structure...")
	if err := CreateDirectoryStructure(cfg.WorkDir); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}
	ui.Success("Directory structure created")

	// Step 5: Generate environment file
	ui.Info("Creating environment configuration...")
	if err := GenerateEnvFile(cfg, cfg.WorkDir); err != nil {
		return fmt.Errorf("failed to create environment file: %w", err)
	}
	ui.Success("Environment configuration created")

	// Step 6: Generate docker-compose file
	ui.Info("Creating Docker Compose configuration...")
	if err := GenerateDockerCompose(cfg, cfg.WorkDir); err != nil {
		return fmt.Errorf("failed to create docker-compose file: %w", err)
	}
	ui.Success("Docker Compose configuration created")

	// Step 7: Save configuration
	configPath := filepath.Join(cfg.WorkDir, ".coolify-config.json")
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Step 8: Start services
	ui.Info("Starting services (this may take a few minutes)...")
	if err := ComposeUp(cfg.WorkDir, true); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}
	ui.Success("Services started")

	// Step 9: Wait for services to be healthy
	ui.Info("Waiting for services to be ready...")
	if err := waitForServices(cfg.WorkDir); err != nil {
		ui.Warning(fmt.Sprintf("Some services may not be fully ready: %v", err))
		ui.Dim("You can check status with: cool-kit local logs")
	} else {
		ui.Success("All services are ready")
	}

	// Step 10: Run initial migrations
	ui.Info("Running database migrations...")
	if err := runMigrations(cfg.WorkDir); err != nil {
		ui.Warning(fmt.Sprintf("Migration setup incomplete: %v", err))
		ui.Dim("You may need to run migrations manually")
	} else {
		ui.Success("Database migrations completed")
	}

	// Step 11: Display summary
	displaySetupSummary(cfg)

	return nil
}

// collectConfiguration collects setup information from user
func collectConfiguration() (*Config, error) {
	cfg := DefaultConfig()

	// Ask for admin email
	email, err := ui.InputWithDefault("Admin Email", cfg.AdminEmail)
	if err != nil {
		return nil, err
	}
	if email != "" {
		cfg.AdminEmail = email
	}

	// Ask for admin password
	password, err := ui.InputWithDefault("Admin Password (leave blank to use default)", "")
	if err != nil {
		return nil, err
	}
	if password != "" {
		cfg.AdminPassword = password
	} else {
		ui.Dim(fmt.Sprintf("Using default password: %s", cfg.AdminPassword))
	}

	// Ask for app port
	portStr, err := ui.InputWithDefault("Application Port", fmt.Sprintf("%d", cfg.AppPort))
	if err != nil {
		return nil, err
	}
	if portStr != "" {
		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil && port > 0 && port < 65536 {
			cfg.AppPort = port
		}
	}

	// Ask for websocket port
	wsPortStr, err := ui.InputWithDefault("WebSocket Port", fmt.Sprintf("%d", cfg.WebSocketPort))
	if err != nil {
		return nil, err
	}
	if wsPortStr != "" {
		var port int
		if _, err := fmt.Sscanf(wsPortStr, "%d", &port); err == nil && port > 0 && port < 65536 {
			cfg.WebSocketPort = port
		}
	}

	// Ask for debug mode
	debugStr, err := ui.InputWithDefault("Enable debug mode? (Y/n)", "Y")
	if err != nil {
		return nil, err
	}
	cfg.AppDebug = debugStr == "" || debugStr == "Y" || debugStr == "y"
	if cfg.AppDebug {
		cfg.LogLevel = "debug"
		ui.Dim("Debug mode enabled")
	} else {
		cfg.LogLevel = "info"
		ui.Dim("Debug mode disabled")
	}

	// Ask for work directory
	workDir, err := ui.InputWithDefault("Work Directory", cfg.WorkDir)
	if err != nil {
		return nil, err
	}
	if workDir != "" {
		cfg.WorkDir = workDir
	}

	return cfg, nil
}

// waitForServices waits for all services to be healthy
func waitForServices(workDir string) error {
	services := []string{"coolify-db", "coolify-redis", "coolify-soketi"}
	maxWait := 120 * time.Second
	checkInterval := 5 * time.Second

	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		allHealthy := true

		for _, service := range services {
			healthy, err := CheckServiceHealth(workDir, service)
			if err != nil || !healthy {
				allHealthy = false
				break
			}
		}

		if allHealthy {
			return nil
		}

		time.Sleep(checkInterval)
	}

	return fmt.Errorf("services did not become healthy within %v", maxWait)
}

// runMigrations runs database migrations
func runMigrations(workDir string) error {
	// Wait a bit for the application to fully start
	time.Sleep(10 * time.Second)

	// Run migrations
	// Use a long timeout for migrations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	_, err := ComposeExecOutput(ctx, workDir, "coolify", []string{"php", "artisan", "migrate", "--force"})
	return err
}

// displaySetupSummary shows setup completion summary
func displaySetupSummary(cfg *Config) {
	ui.Section("Setup Completed Successfully! 🎉")

	ui.KeyValue("Access URL", fmt.Sprintf("http://localhost:%d", cfg.AppPort))
	ui.KeyValue("Admin Email", cfg.AdminEmail)
	ui.KeyValue("Admin Password", cfg.AdminPassword)
	ui.KeyValue("WebSocket URL", fmt.Sprintf("http://localhost:%d", cfg.WebSocketPort))
	ui.KeyValue("Work Directory", cfg.WorkDir)

	ui.Dim("")
	ui.Info("Useful commands:")
	ui.Dim("  cool-kit local start    - Start services")
	ui.Dim("  cool-kit local stop     - Stop services")
	ui.Dim("  cool-kit local logs     - View logs")
	ui.Dim("  cool-kit local update   - Update to latest version")
	ui.Dim("  cool-kit local reset    - Reset to clean state")
}

// Start starts the local Coolify instance
func Start() error {
	ui.Section("Starting Local Coolify")

	// Load configuration
	cfg, err := LoadOrDefault("./coolify-local/.coolify-config.json")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w (run 'cool-kit local setup' first)", err)
	}

	ui.Info("Starting services...")
	if err := ComposeUp(cfg.WorkDir, true); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	ui.Success("Services started")
	ui.Dim(fmt.Sprintf("Access your instance at: http://localhost:%d", cfg.AppPort))

	return nil
}

// Stop stops the local Coolify instance
func Stop() error {
	ui.Section("Stopping Local Coolify")

	// Try to find work directory
	workDir := "./coolify-local"

	ui.Info("Stopping services...")
	if err := ComposeDown(workDir, false); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	ui.Success("Services stopped")
	ui.Dim("Data and configuration preserved")

	return nil
}

// Logs displays logs from local services
func Logs(service string, follow bool, tail int) error {
	ui.Section("Viewing Logs")

	// Try to find work directory
	workDir := "./coolify-local"

	if service != "" {
		ui.Info(fmt.Sprintf("Service: %s", service))
	} else {
		ui.Info("All services")
	}

	return ComposeLogs(workDir, service, follow, tail)
}

// Reset resets the local instance to clean state
func Reset(force bool) error {
	ui.Section("Resetting Local Coolify")

	if !force {
		ui.Warning("This will delete all data and configuration")
		confirm, err := ui.InputWithDefault("Are you sure? (yes/no)", "no")
		if err != nil {
			return err
		}
		if confirm != "yes" {
			ui.Info("Reset cancelled")
			return nil
		}
	}

	// Try to find work directory
	workDir := "./coolify-local"

	ui.Info("Stopping services...")
	if err := ComposeDown(workDir, true); err != nil {
		ui.Warning(fmt.Sprintf("Failed to stop services: %v", err))
	}

	ui.Success("Local Coolify has been reset")
	ui.Dim("Run 'cool-kit local setup' to set up again")

	return nil
}

// Update updates the local instance to latest version
func Update() error {
	ui.Section("Updating Local Coolify")

	// Try to find work directory
	workDir := "./coolify-local"

	ui.Info("Pulling latest images...")
	if err := ComposePull(workDir); err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}

	ui.Info("Recreating containers...")
	if err := ComposeDown(workDir, false); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	if err := ComposeUp(workDir, true); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	ui.Info("Running migrations...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for services to be ready
	time.Sleep(10 * time.Second)

	// Run migrations
	_, err := ComposeExecOutput(ctx, workDir, "coolify", []string{"php", "artisan", "migrate", "--force"})
	if err != nil {
		ui.Warning(fmt.Sprintf("Migration failed: %v", err))
		ui.Dim("You may need to run migrations manually")
	}

	ui.Success("Update completed")
	ui.Dim("Services have been updated to the latest version")

	return nil
}
