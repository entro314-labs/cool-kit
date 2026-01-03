package azure

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/ui"
)

// BackupManager handles backup and restore operations
type BackupManager struct {
	ctx    *DeploymentContext
	client *SSHClient
}

// NewBackupManager creates a new backup manager
func NewBackupManager(ctx *DeploymentContext, client *SSHClient) *BackupManager {
	return &BackupManager{
		ctx:    ctx,
		client: client,
	}
}

// CreateBackup creates a new backup
func (b *BackupManager) CreateBackup(backupType string) (*BackupInfo, error) {
	ui.Section("Creating Backup")

	timestamp := time.Now().Format("20060102-150405")
	backupID := fmt.Sprintf("%s-%s", backupType, timestamp)
	backupPath := fmt.Sprintf("/data/coolify/backups/%s", backupID)

	ui.Info(fmt.Sprintf("Creating backup: %s", backupID))

	script := fmt.Sprintf(`#!/bin/bash
set -e
cd /data/coolify

# Create backup directory
mkdir -p %s/{volumes,config}

# Backup database
ui.Dim "Backing up database..."
docker compose exec -T coolify-db pg_dump -U coolify coolify | gzip > %s/database.sql.gz

# Backup Redis data
docker compose exec -T coolify-redis redis-cli --raw SAVE
docker cp coolify-redis:/data/dump.rdb %s/redis-dump.rdb

# Backup volumes
ui.Dim "Backing up volumes..."
tar czf %s/volumes/coolify-data.tar.gz \
    -C /data/coolify \
    source ssh applications databases services proxy \
    2>/dev/null || true

# Backup configuration
cp /data/coolify/.env %s/config/
cp /data/coolify/docker-compose.yml %s/config/

# Get current Coolify version
COOLIFY_VERSION=$(docker compose exec -T coolify php artisan --version 2>/dev/null | awk '{print $NF}' || echo "unknown")

# Get backup size
BACKUP_SIZE=$(du -sb %s | awk '{print $1}')

# Create backup metadata
cat > %s/metadata.json << EOF
{
  "id": "%s",
  "timestamp": "%s",
  "type": "%s",
  "version": "$COOLIFY_VERSION",
  "size": $BACKUP_SIZE
}
EOF

# Output metadata for parsing
cat %s/metadata.json
`, backupPath, backupPath, backupPath, backupPath, backupPath, backupPath, backupPath, backupPath, backupID, timestamp, backupType, backupPath)

	output, err := b.client.Execute(script)
	if err != nil {
		return nil, fmt.Errorf("backup creation failed: %w", err)
	}

	// Parse metadata from output
	var metadata BackupInfo
	// Find the JSON in the output
	jsonStart := strings.Index(output, "{")
	if jsonStart >= 0 {
		jsonData := output[jsonStart:]
		if err := json.Unmarshal([]byte(jsonData), &metadata); err != nil {
			// If parsing fails, create basic metadata
			metadata = BackupInfo{
				ID:        backupID,
				Timestamp: timestamp,
				Type:      backupType,
				Path:      backupPath,
			}
		}
	} else {
		metadata = BackupInfo{
			ID:        backupID,
			Timestamp: timestamp,
			Type:      backupType,
			Path:      backupPath,
		}
	}

	ui.Success(fmt.Sprintf("Backup created: %s", backupID))
	if metadata.Size > 0 {
		ui.Dim(fmt.Sprintf("Backup size: %d bytes", metadata.Size))
	}

	return &metadata, nil
}

// ListBackups lists all available backups
func (b *BackupManager) ListBackups() ([]BackupInfo, error) {
	ui.Info("Listing backups")

	script := `#!/bin/bash
cd /data/coolify/backups 2>/dev/null || exit 0

# List all backup directories
for dir in */; do
    if [ -d "$dir" ] && [ -f "$dir/metadata.json" ]; then
        cat "$dir/metadata.json"
        echo "---"
    fi
done
`

	output, err := b.client.Execute(script)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return []BackupInfo{}, nil
	}

	// Parse backups
	var backups []BackupInfo
	backupStrings := strings.Split(output, "---")

	for _, backupStr := range backupStrings {
		backupStr = strings.TrimSpace(backupStr)
		if backupStr == "" {
			continue
		}

		var backup BackupInfo
		if err := json.Unmarshal([]byte(backupStr), &backup); err != nil {
			continue // Skip invalid metadata
		}

		backups = append(backups, backup)
	}

	return backups, nil
}

// DeleteBackup deletes a specific backup
func (b *BackupManager) DeleteBackup(backupID string) error {
	ui.Info(fmt.Sprintf("Deleting backup: %s", backupID))

	script := fmt.Sprintf(`#!/bin/bash
set -e

BACKUP_PATH="/data/coolify/backups/%s"

if [ ! -d "$BACKUP_PATH" ]; then
    echo "Backup not found: %s"
    exit 1
fi

rm -rf "$BACKUP_PATH"
echo "Backup deleted: %s"
`, backupID, backupID, backupID)

	_, err := b.client.Execute(script)
	if err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	ui.Success(fmt.Sprintf("Backup deleted: %s", backupID))
	return nil
}

// RestoreBackup restores from a specific backup
func (b *BackupManager) RestoreBackup(backupID string) error {
	ui.Section(fmt.Sprintf("Restoring from backup: %s", backupID))

	backupPath := fmt.Sprintf("/data/coolify/backups/%s", backupID)

	// Verify backup exists
	ui.Info("Verifying backup")

	verifyScript := fmt.Sprintf(`#!/bin/bash
if [ ! -d "%s" ]; then
    echo "Backup not found: %s"
    exit 1
fi

if [ ! -f "%s/metadata.json" ]; then
    echo "Backup metadata not found"
    exit 1
fi

cat %s/metadata.json
`, backupPath, backupID, backupPath, backupPath)

	output, err := b.client.Execute(verifyScript)
	if err != nil {
		return fmt.Errorf("backup verification failed: %w", err)
	}

	// Parse and display backup info
	var metadata BackupInfo
	if err := json.Unmarshal([]byte(output), &metadata); err == nil {
		ui.Dim(fmt.Sprintf("Backup type: %s", metadata.Type))
		ui.Dim(fmt.Sprintf("Created: %s", metadata.Timestamp))
		if metadata.Version != "" {
			ui.Dim(fmt.Sprintf("Coolify version: %s", metadata.Version))
		}
	}

	ui.Info("Stopping services")

	stopScript := `#!/bin/bash
set -e
cd /data/coolify
docker compose down
`

	if _, err := b.client.Execute(stopScript); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	ui.Info("Restoring data")

	restoreScript := fmt.Sprintf(`#!/bin/bash
set -e
cd /data/coolify

# Restore configuration
cp %s/config/.env /data/coolify/.env
cp %s/config/docker-compose.yml /data/coolify/docker-compose.yml

# Restore volumes
tar xzf %s/volumes/coolify-data.tar.gz -C /data/coolify/ 2>/dev/null || true

echo "Data restored from backup"
`, backupPath, backupPath, backupPath)

	if _, err := b.client.Execute(restoreScript); err != nil {
		return fmt.Errorf("data restore failed: %w", err)
	}

	ui.Info("Starting services")

	startScript := `#!/bin/bash
set -e
cd /data/coolify
docker compose up -d
sleep 15
`

	if _, err := b.client.Execute(startScript); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	ui.Info("Restoring database")

	dbRestoreScript := fmt.Sprintf(`#!/bin/bash
set -e
cd /data/coolify

# Wait for database to be ready
sleep 10

# Drop and recreate database
docker compose exec -T coolify-db psql -U coolify -c "DROP DATABASE IF EXISTS coolify;" postgres
docker compose exec -T coolify-db psql -U coolify -c "CREATE DATABASE coolify;" postgres

# Restore database
zcat %s/database.sql.gz | docker compose exec -T coolify-db psql -U coolify coolify

echo "Database restored"
`, backupPath)

	if _, err := b.client.Execute(dbRestoreScript); err != nil {
		return fmt.Errorf("database restore failed: %w", err)
	}

	ui.Info("Restoring Redis data")

	redisRestoreScript := fmt.Sprintf(`#!/bin/bash
set -e
cd /data/coolify

# Restore Redis data
docker cp %s/redis-dump.rdb coolify-redis:/data/dump.rdb
docker compose restart coolify-redis

# Wait for Redis
sleep 5

# Restart Coolify to ensure all changes are applied
docker compose restart coolify

echo "Redis restored"
`, backupPath)

	if _, err := b.client.Execute(redisRestoreScript); err != nil {
		return fmt.Errorf("Redis restore failed: %w", err)
	}

	ui.Info("Verifying restore")

	// Wait for services
	services := []string{"coolify-db", "coolify-redis", "coolify-realtime", "coolify"}
	deadline := time.Now().Add(3 * time.Minute)

	for _, service := range services {
		for time.Now().Before(deadline) {
			status, err := b.client.GetDockerContainerStatus(service)
			if err == nil && status == "running" {
				break
			}
			time.Sleep(5 * time.Second)
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("service %s did not become healthy", service)
		}
	}

	ui.Success(fmt.Sprintf("Successfully restored from backup: %s", backupID))
	return nil
}

// BackupInfo represents backup metadata (defined in types.go but repeated here for clarity)
type BackupInfoInternal struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Version   string `json:"version"`
	Size      int64  `json:"size"`
	Path      string `json:"path,omitempty"`
}
