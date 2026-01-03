package azure

import (
	"fmt"
	"time"

	"github.com/entro314-labs/cool-kit/internal/ui"
)

// Installer handles Coolify installation on Azure VM
type Installer struct {
	ctx    *DeploymentContext
	client *SSHClient
}

// NewInstaller creates a new Coolify installer
func NewInstaller(ctx *DeploymentContext, client *SSHClient) *Installer {
	return &Installer{
		ctx:    ctx,
		client: client,
	}
}

// Install performs complete Coolify installation
func (i *Installer) Install() error {
	ui.Section("Installing Coolify")

	// Wait for SSH to become available
	if err := i.waitForSSH(); err != nil {
		return err
	}

	// Update system packages
	if err := i.updateSystem(); err != nil {
		return err
	}

	// Install Docker
	if err := i.installDocker(); err != nil {
		return err
	}

	// Create directory structure
	if err := i.createDirectories(); err != nil {
		return err
	}

	// Generate and upload configuration
	if err := i.uploadConfiguration(); err != nil {
		return err
	}

	// Deploy Coolify containers
	if err := i.deployCoolify(); err != nil {
		return err
	}

	// Wait for services to be healthy
	if err := i.waitForServices(); err != nil {
		return err
	}

	// Run initial setup
	if err := i.runInitialSetup(); err != nil {
		return err
	}

	ui.Success("Coolify installed successfully")
	return nil
}

// waitForSSH waits for SSH to become available
func (i *Installer) waitForSSH() error {
	ui.Info("Waiting for SSH to become available")

	if err := i.client.WaitForSSH(5 * time.Minute); err != nil {
		return fmt.Errorf("SSH did not become available: %w", err)
	}

	ui.Dim("SSH connection established")
	return nil
}

// updateSystem updates system packages
func (i *Installer) updateSystem() error {
	ui.Info("Updating system packages")

	script := `#!/bin/bash
set -e
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get upgrade -y -qq
apt-get install -y -qq curl wget git jq
`

	_, err := i.client.ExecuteWithRetry(script, 3, 10*time.Second)
	if err != nil {
		return fmt.Errorf("system update failed: %w", err)
	}

	ui.Dim("System packages updated")
	return nil
}

// installDocker installs Docker if not present
func (i *Installer) installDocker() error {
	ui.Info("Installing Docker")

	// Check if Docker is already installed
	hasDocker, err := i.client.CheckDocker()
	if err == nil && hasDocker {
		ui.Dim("Docker is already installed")
		return nil
	}

	// Install Docker
	if err := i.client.InstallDocker(); err != nil {
		return fmt.Errorf("Docker installation failed: %w", err)
	}

	ui.Dim("Docker installed successfully")
	return nil
}

// createDirectories creates required directory structure
func (i *Installer) createDirectories() error {
	ui.Info("Creating directory structure")

	script := `#!/bin/bash
set -e
mkdir -p /data/coolify/{source,ssh,applications,databases,backups,services,proxy}
mkdir -p /data/coolify/ssh/keys
mkdir -p /data/coolify/proxy/dynamic
chmod -R 755 /data/coolify
`

	_, err := i.client.Execute(script)
	if err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	ui.Dim("Directory structure created")
	return nil
}

// uploadConfiguration generates and uploads Coolify configuration
func (i *Installer) uploadConfiguration() error {
	ui.Info("Uploading configuration")

	// Generate environment file
	envContent := i.generateEnvFile()

	// Upload .env file
	if err := i.client.CopyFileContent(envContent, "/data/coolify/.env"); err != nil {
		return fmt.Errorf("failed to upload .env: %w", err)
	}

	// Generate docker-compose file
	composeContent := i.generateDockerCompose()

	// Upload docker-compose.yml
	if err := i.client.CopyFileContent(composeContent, "/data/coolify/docker-compose.yml"); err != nil {
		return fmt.Errorf("failed to upload docker-compose.yml: %w", err)
	}

	ui.Dim("Configuration uploaded")
	return nil
}

// generateEnvFile generates the .env file content
func (i *Installer) generateEnvFile() string {
	return fmt.Sprintf(`# Coolify Configuration
APP_ID=%s
APP_NAME=Coolify
APP_KEY=%s
APP_ENV=production
APP_DEBUG=false
APP_URL=http://%s
APP_PORT=8000

# Database
DB_CONNECTION=pgsql
DB_HOST=coolify-db
DB_PORT=5432
DB_DATABASE=coolify
DB_USERNAME=coolify
DB_PASSWORD=%s

# Redis
REDIS_HOST=coolify-redis
REDIS_PASSWORD=%s
REDIS_PORT=6379

# Pusher (for WebSocket)
PUSHER_APP_ID=%s
PUSHER_APP_KEY=%s
PUSHER_APP_SECRET=%s
PUSHER_HOST=coolify-realtime
PUSHER_PORT=6001
PUSHER_SCHEME=http

# Queue
QUEUE_CONNECTION=redis
HORIZON_BALANCE=auto
HORIZON_MAX_PROCESSES=10
HORIZON_BALANCE_MAX_SHIFT=1
HORIZON_BALANCE_COOLDOWN=3

# Session
SESSION_DRIVER=redis
SESSION_LIFETIME=120

# Cache
CACHE_DRIVER=redis

# Docker
DOCKER_HOST=unix:///var/run/docker.sock

# SSL
SSL_MODE=off
`,
		i.ctx.Config.Coolify.AppID,
		i.ctx.Config.Coolify.AppKey,
		i.ctx.PublicIP,
		i.ctx.Config.Coolify.DBPassword,
		i.ctx.Config.Coolify.RedisPassword,
		i.ctx.Config.Coolify.PusherAppID,
		i.ctx.Config.Coolify.PusherAppKey,
		i.ctx.Config.Coolify.PusherAppSecret,
	)
}

// generateDockerCompose generates docker-compose.yml content
func (i *Installer) generateDockerCompose() string {
	return `version: '3.8'

services:
  coolify-db:
    image: postgres:15-alpine
    container_name: coolify-db
    restart: unless-stopped
    env_file: .env
    environment:
      POSTGRES_USER: coolify
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: coolify
    volumes:
      - coolify-db:/var/lib/postgresql/data
    networks:
      - coolify
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U coolify"]
      interval: 10s
      timeout: 5s
      retries: 5

  coolify-redis:
    image: redis:7-alpine
    container_name: coolify-redis
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - coolify-redis:/data
    networks:
      - coolify
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  coolify-realtime:
    image: quay.io/soketi/soketi:latest-16-alpine
    container_name: coolify-realtime
    restart: unless-stopped
    env_file: .env
    ports:
      - "6001:6001"
    environment:
      SOKETI_DEFAULT_APP_ID: ${PUSHER_APP_ID}
      SOKETI_DEFAULT_APP_KEY: ${PUSHER_APP_KEY}
      SOKETI_DEFAULT_APP_SECRET: ${PUSHER_APP_SECRET}
    networks:
      - coolify
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:6001/ready"]
      interval: 10s
      timeout: 5s
      retries: 5

  coolify:
    image: ghcr.io/coollabsio/coolify:latest
    container_name: coolify
    restart: unless-stopped
    env_file: .env
    ports:
      - "8000:80"
      - "6001:6001"
      - "6002:6002"
    volumes:
      - /data/coolify/source:/data/coolify/source
      - /data/coolify/ssh:/data/coolify/ssh
      - /data/coolify/applications:/data/coolify/applications
      - /data/coolify/databases:/data/coolify/databases
      - /data/coolify/backups:/data/coolify/backups
      - /data/coolify/services:/data/coolify/services
      - /data/coolify/proxy:/data/coolify/proxy
      - /var/run/docker.sock:/var/run/docker.sock
    networks:
      - coolify
    depends_on:
      coolify-db:
        condition: service_healthy
      coolify-redis:
        condition: service_healthy
      coolify-realtime:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:80/api/health"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  coolify-db:
  coolify-redis:

networks:
  coolify:
    driver: bridge
`
}

// deployCoolify starts Coolify containers
func (i *Installer) deployCoolify() error {
	ui.Info("Deploying Coolify containers")

	script := `#!/bin/bash
set -e
cd /data/coolify
docker compose pull
docker compose up -d
`

	_, err := i.client.Execute(script)
	if err != nil {
		return fmt.Errorf("failed to deploy Coolify: %w", err)
	}

	ui.Dim("Coolify containers started")
	return nil
}

// waitForServices waits for all services to be healthy
func (i *Installer) waitForServices() error {
	ui.Info("Waiting for services to be healthy")

	services := []string{"coolify-db", "coolify-redis", "coolify-realtime", "coolify"}
	deadline := time.Now().Add(5 * time.Minute)

	for _, service := range services {
		ui.Dim(fmt.Sprintf("Checking %s...", service))

		for time.Now().Before(deadline) {
			status, err := i.client.GetDockerContainerStatus(service)
			if err == nil && status == "running" {
				break
			}
			time.Sleep(5 * time.Second)
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("service %s did not become healthy", service)
		}
	}

	ui.Dim("All services are healthy")
	return nil
}

// runInitialSetup runs initial Coolify setup commands
func (i *Installer) runInitialSetup() error {
	ui.Info("Running initial setup")

	script := `#!/bin/bash
set -e
cd /data/coolify

# Wait for database to be fully ready
sleep 10

# Run migrations
docker compose exec -T coolify php artisan migrate --force

# Generate application key if not set
docker compose exec -T coolify php artisan key:generate --force

# Clear and cache config
docker compose exec -T coolify php artisan config:cache
docker compose exec -T coolify php artisan route:cache
docker compose exec -T coolify php artisan view:cache

# Create initial admin user if specified
if [ -n "${ADMIN_EMAIL}" ] && [ -n "${ADMIN_PASSWORD}" ]; then
    docker compose exec -T coolify php artisan coolify:user:create \
        --email="${ADMIN_EMAIL}" \
        --password="${ADMIN_PASSWORD}" \
        --name="Administrator" \
        --is-admin
fi
`

	// Set environment variables for the script
	script = fmt.Sprintf("export ADMIN_EMAIL='%s'\nexport ADMIN_PASSWORD='%s'\n%s",
		i.ctx.AdminEmail,
		i.ctx.AdminPassword,
		script)

	_, err := i.client.Execute(script)
	if err != nil {
		return fmt.Errorf("initial setup failed: %w", err)
	}

	ui.Dim("Initial setup completed")
	return nil
}

// GetInstallationURL returns the Coolify installation URL
func (i *Installer) GetInstallationURL() string {
	return fmt.Sprintf("http://%s:8000", i.ctx.PublicIP)
}
