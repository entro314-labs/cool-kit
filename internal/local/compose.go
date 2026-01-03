package local

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateDockerCompose creates the docker-compose.yml file
func GenerateDockerCompose(cfg *Config, workDir string) error {
	composePath := filepath.Join(workDir, "docker-compose.yml")

	content := fmt.Sprintf(`version: '3.8'

services:
  # PostgreSQL Database
  coolify-db:
    image: postgres:15-alpine
    container_name: coolify-db
    restart: unless-stopped
    environment:
      POSTGRES_DB: coolify
      POSTGRES_USER: coolify
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - ./data/databases/postgres:/var/lib/postgresql/data
    networks:
      - coolify-local
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U coolify"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis
  coolify-redis:
    image: redis:7-alpine
    container_name: coolify-redis
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - ./data/databases/redis:/data
    networks:
      - coolify-local
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Soketi (WebSocket Server)
  coolify-soketi:
    image: quay.io/soketi/soketi:latest-16-alpine
    container_name: coolify-soketi
    restart: unless-stopped
    environment:
      SOKETI_DEBUG: '${APP_DEBUG}'
      SOKETI_DEFAULT_APP_ID: '${PUSHER_APP_ID}'
      SOKETI_DEFAULT_APP_KEY: '${PUSHER_APP_KEY}'
      SOKETI_DEFAULT_APP_SECRET: '${PUSHER_APP_SECRET}'
    ports:
      - "%d:6001"
    networks:
      - coolify-local
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:6001/ready"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Coolify Application
  coolify:
    image: ghcr.io/coollabsio/coolify:latest
    container_name: coolify
    restart: unless-stopped
    environment:
      APP_NAME: '${APP_NAME}'
      APP_ENV: '${APP_ENV}'
      APP_DEBUG: '${APP_DEBUG}'
      APP_URL: '${APP_URL}'
      APP_ID: '${APP_ID}'
      APP_KEY: '${APP_KEY}'
      DB_CONNECTION: '${DB_CONNECTION}'
      DB_HOST: '${DB_HOST}'
      DB_PORT: '${DB_PORT}'
      DB_DATABASE: '${DB_DATABASE}'
      DB_USERNAME: '${DB_USERNAME}'
      DB_PASSWORD: '${DB_PASSWORD}'
      REDIS_HOST: '${REDIS_HOST}'
      REDIS_PASSWORD: '${REDIS_PASSWORD}'
      REDIS_PORT: '${REDIS_PORT}'
      CACHE_STORE: '${CACHE_STORE}'
      SESSION_DRIVER: '${SESSION_DRIVER}'
      QUEUE_CONNECTION: '${QUEUE_CONNECTION}'
      BROADCAST_DRIVER: '${BROADCAST_DRIVER}'
      PUSHER_APP_ID: '${PUSHER_APP_ID}'
      PUSHER_APP_KEY: '${PUSHER_APP_KEY}'
      PUSHER_APP_SECRET: '${PUSHER_APP_SECRET}'
      PUSHER_HOST: '${PUSHER_HOST}'
      PUSHER_PORT: '${PUSHER_PORT}'
      PUSHER_SCHEME: '${PUSHER_SCHEME}'
      LOG_CHANNEL: '${LOG_CHANNEL}'
      LOG_LEVEL: '${LOG_LEVEL}'
      MAIL_MAILER: '${MAIL_MAILER}'
      SELF_HOSTED: '${SELF_HOSTED}'
    ports:
      - "%d:80"
    volumes:
      - ./data/coolify:/data/coolify
      - ./data/ssh:/data/coolify/ssh
      - ./data/applications:/data/coolify/applications
      - ./data/databases:/data/coolify/databases
      - ./data/backups:/data/coolify/backups
      - ./data/services:/data/coolify/services
      - ./data/proxy:/data/coolify/proxy
      - ./logs:/var/log
      - /var/run/docker.sock:/var/run/docker.sock
    networks:
      - coolify-local
    depends_on:
      coolify-db:
        condition: service_healthy
      coolify-redis:
        condition: service_healthy
      coolify-soketi:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:80/api/health"]
      interval: 30s
      timeout: 10s
      retries: 5
      start_period: 60s

networks:
  coolify-local:
    name: coolify-local
    driver: bridge
`, cfg.WebSocketPort, cfg.AppPort)

	// Write docker-compose file
	if err := os.WriteFile(composePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose file: %w", err)
	}

	return nil
}
