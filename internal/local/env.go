package local

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GenerateEnvFile creates the .env.local file with all configuration
func GenerateEnvFile(cfg *Config, workDir string) error {
	envPath := filepath.Join(workDir, ".env.local")

	content := fmt.Sprintf(`# Coolify Local Development Environment
# Generated on: %s

# Application
APP_NAME=Coolify
APP_ENV=local
APP_DEBUG=%t
APP_URL=http://localhost:%d
APP_ID=%s
APP_KEY=%s

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

# Cache & Session
CACHE_STORE=redis
SESSION_DRIVER=redis
QUEUE_CONNECTION=redis

# Broadcasting (WebSocket)
BROADCAST_DRIVER=pusher
PUSHER_APP_ID=%s
PUSHER_APP_KEY=%s
PUSHER_APP_SECRET=%s
PUSHER_HOST=localhost
PUSHER_PORT=%d
PUSHER_SCHEME=http
PUSHER_BACKEND_HOST=localhost
PUSHER_BACKEND_PORT=%d

# Logging
LOG_CHANNEL=stack
LOG_LEVEL=%s

# Mail (local development - use log driver)
MAIL_MAILER=log
MAIL_HOST=localhost
MAIL_PORT=1025
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null
MAIL_FROM_ADDRESS=noreply@coolify.local
MAIL_FROM_NAME=Coolify

# Default Admin Credentials
DEFAULT_ADMIN_EMAIL=%s
DEFAULT_ADMIN_PASSWORD=%s

# Additional Settings
QUEUE_WORKER_RESTART_AFTER_SECONDS=0
FORCE_HTTPS=false
SELF_HOSTED=true
SSH_MUX_ENABLED=true
`,
		time.Now().Format(time.RFC3339),
		cfg.AppDebug,
		cfg.AppPort,
		cfg.AppID,
		cfg.AppKey,
		cfg.DBPassword,
		cfg.RedisPassword,
		cfg.PusherAppID,
		cfg.PusherAppKey,
		cfg.PusherAppSecret,
		cfg.WebSocketPort,
		cfg.WebSocketPort,
		cfg.LogLevel,
		cfg.AdminEmail,
		cfg.AdminPassword,
	)

	// Ensure directory exists
	if err := os.MkdirAll(workDir, 0750); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// Write environment file
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write environment file: %w", err)
	}

	return nil
}

// CreateDirectoryStructure creates the necessary directory structure
func CreateDirectoryStructure(workDir string) error {
	directories := []string{
		"data/coolify",
		"data/ssh/keys",
		"data/ssh/mux",
		"data/applications",
		"data/databases",
		"data/backups",
		"data/services",
		"data/proxy/dynamic",
		"data/sentinel",
		"logs",
		"ssl",
	}

	for _, dir := range directories {
		path := filepath.Join(workDir, dir)
		if err := os.MkdirAll(path, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
