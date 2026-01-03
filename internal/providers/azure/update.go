package azure

import (
	"fmt"
	"time"

	"github.com/entro314-labs/cool-kit/internal/ui"
)

// Updater handles Coolify updates on Azure VM
type Updater struct {
	ctx    *DeploymentContext
	client *SSHClient
}

// NewUpdater creates a new Coolify updater
func NewUpdater(ctx *DeploymentContext, client *SSHClient) *Updater {
	return &Updater{
		ctx:    ctx,
		client: client,
	}
}

// Update performs complete Coolify update with backup and rollback support
func (u *Updater) Update(autoRollback bool) error {
	ui.Section("Updating Coolify")

	// Create backup before update
	backupID, err := u.createPreUpdateBackup()
	if err != nil {
		return err
	}

	ui.Dim(fmt.Sprintf("Backup created: %s", backupID))

	// Perform update
	if err := u.performUpdate(); err != nil {
		if autoRollback {
			ui.Warning("Update failed, initiating automatic rollback")
			if rollbackErr := u.rollbackToBackup(backupID); rollbackErr != nil {
				return fmt.Errorf("update failed and rollback failed: update error: %w, rollback error: %v", err, rollbackErr)
			}
			return fmt.Errorf("update failed, rolled back to backup %s: %w", backupID, err)
		}
		return fmt.Errorf("update failed: %w", err)
	}

	// Verify update
	if err := u.verifyUpdate(); err != nil {
		if autoRollback {
			ui.Warning("Update verification failed, initiating automatic rollback")
			if rollbackErr := u.rollbackToBackup(backupID); rollbackErr != nil {
				return fmt.Errorf("verification failed and rollback failed: verify error: %w, rollback error: %v", err, rollbackErr)
			}
			return fmt.Errorf("verification failed, rolled back to backup %s: %w", backupID, err)
		}
		return fmt.Errorf("update verification failed: %w", err)
	}

	ui.Success("Coolify updated successfully")
	ui.Dim(fmt.Sprintf("Backup %s is available for rollback if needed", backupID))

	return nil
}

// createPreUpdateBackup creates a backup before updating
func (u *Updater) createPreUpdateBackup() (string, error) {
	ui.Info("Creating pre-update backup")

	timestamp := time.Now().Format("20060102-150405")
	backupID := fmt.Sprintf("pre-update-%s", timestamp)
	backupPath := fmt.Sprintf("/data/coolify/backups/%s", backupID)

	script := fmt.Sprintf(`#!/bin/bash
set -e
cd /data/coolify

# Create backup directory
mkdir -p %s/{volumes,config}

# Backup database
docker compose exec -T coolify-db pg_dump -U coolify coolify | gzip > %s/database.sql.gz

# Backup Redis data
docker compose exec -T coolify-redis redis-cli --raw SAVE
docker cp coolify-redis:/data/dump.rdb %s/redis-dump.rdb

# Backup volumes
tar czf %s/volumes/coolify-data.tar.gz \
    -C /data/coolify \
    source ssh applications databases services proxy \
    2>/dev/null || true

# Backup configuration
cp -r /data/coolify/.env %s/config/
cp -r /data/coolify/docker-compose.yml %s/config/

# Get current Coolify version
COOLIFY_VERSION=$(docker compose exec -T coolify php artisan --version | awk '{print $NF}')
echo "$COOLIFY_VERSION" > %s/version.txt

# Create backup metadata
cat > %s/metadata.json << EOF
{
  "id": "%s",
  "timestamp": "%s",
  "type": "pre-update",
  "version": "$COOLIFY_VERSION"
}
EOF

echo "Backup completed: %s"
`, backupPath, backupPath, backupPath, backupPath, backupPath, backupPath, backupPath, backupPath, backupID, timestamp, backupID)

	_, err := u.client.Execute(script)
	if err != nil {
		return "", fmt.Errorf("backup creation failed: %w", err)
	}

	return backupID, nil
}

// performUpdate executes the update process
func (u *Updater) performUpdate() error {
	ui.Info("Pulling latest Coolify images")

	pullScript := `#!/bin/bash
set -e
cd /data/coolify
docker compose pull
`

	if _, err := u.client.Execute(pullScript); err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}

	ui.Dim("Latest images pulled")

	ui.Info("Recreating containers")

	recreateScript := `#!/bin/bash
set -e
cd /data/coolify
docker compose up -d --force-recreate --remove-orphans
`

	if _, err := u.client.Execute(recreateScript); err != nil {
		return fmt.Errorf("failed to recreate containers: %w", err)
	}

	ui.Dim("Containers recreated")

	// Wait for services to be healthy
	if err := u.waitForServices(); err != nil {
		return err
	}

	ui.Info("Running database migrations")

	migrateScript := `#!/bin/bash
set -e
cd /data/coolify
docker compose exec -T coolify php artisan migrate --force
docker compose exec -T coolify php artisan config:cache
docker compose exec -T coolify php artisan route:cache
docker compose exec -T coolify php artisan view:cache
`

	if _, err := u.client.Execute(migrateScript); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	ui.Dim("Migrations completed")

	return nil
}

// waitForServices waits for all services to be healthy after update
func (u *Updater) waitForServices() error {
	ui.Info("Waiting for services to be healthy")

	services := []string{"coolify-db", "coolify-redis", "coolify-realtime", "coolify"}
	deadline := time.Now().Add(3 * time.Minute)

	for _, service := range services {
		for time.Now().Before(deadline) {
			status, err := u.client.GetDockerContainerStatus(service)
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

// verifyUpdate verifies the update was successful
func (u *Updater) verifyUpdate() error {
	ui.Info("Verifying update")

	// Check all containers are running
	checkScript := `#!/bin/bash
set -e
cd /data/coolify

# Check all services are up
SERVICES="coolify-db coolify-redis coolify-realtime coolify"
for service in $SERVICES; do
    STATUS=$(docker inspect --format='{{.State.Status}}' $service 2>/dev/null || echo "not found")
    if [ "$STATUS" != "running" ]; then
        echo "Service $service is not running: $STATUS"
        exit 1
    fi
done

# Verify Coolify is responding
HEALTH=$(docker compose exec -T coolify curl -f http://localhost:80/api/health 2>&1 || echo "failed")
if [[ "$HEALTH" == *"failed"* ]]; then
    echo "Coolify health check failed"
    exit 1
fi

echo "All checks passed"
`

	output, err := u.client.Execute(checkScript)
	if err != nil {
		return fmt.Errorf("verification failed: %w\nOutput: %s", err, output)
	}

	ui.Dim("Update verification passed")
	return nil
}

// rollbackToBackup rolls back to a specific backup
func (u *Updater) rollbackToBackup(backupID string) error {
	ui.Section(fmt.Sprintf("Rolling back to backup: %s", backupID))

	backupPath := fmt.Sprintf("/data/coolify/backups/%s", backupID)

	// Verify backup exists
	verifyScript := fmt.Sprintf(`#!/bin/bash
if [ ! -d "%s" ]; then
    echo "Backup not found: %s"
    exit 1
fi
`, backupPath, backupID)

	if _, err := u.client.Execute(verifyScript); err != nil {
		return fmt.Errorf("backup %s not found: %w", backupID, err)
	}

	ui.Info("Stopping services")

	stopScript := `#!/bin/bash
set -e
cd /data/coolify
docker compose down
`

	if _, err := u.client.Execute(stopScript); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	ui.Info("Restoring from backup")

	restoreScript := fmt.Sprintf(`#!/bin/bash
set -e
cd /data/coolify

# Restore configuration
cp %s/config/.env /data/coolify/.env
cp %s/config/docker-compose.yml /data/coolify/docker-compose.yml

# Restore volumes
tar xzf %s/volumes/coolify-data.tar.gz -C /data/coolify/ 2>/dev/null || true

# Start services
docker compose up -d

# Wait for database to be ready
sleep 15

# Restore database
zcat %s/database.sql.gz | docker compose exec -T coolify-db psql -U coolify coolify

# Restore Redis data
docker cp %s/redis-dump.rdb coolify-redis:/data/dump.rdb
docker compose restart coolify-redis

# Wait for services
sleep 10

# Restart Coolify to ensure all changes are applied
docker compose restart coolify

echo "Rollback completed"
`, backupPath, backupPath, backupPath, backupPath, backupPath)

	if _, err := u.client.Execute(restoreScript); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	ui.Info("Verifying rollback")

	// Wait for services
	if err := u.waitForServices(); err != nil {
		return fmt.Errorf("rollback verification failed: %w", err)
	}

	ui.Success(fmt.Sprintf("Successfully rolled back to backup: %s", backupID))
	return nil
}
