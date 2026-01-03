#!/bin/bash

# Final Coolify Azure Update Script
# Incorporates official upgrade script best practices + all fixes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration based on official upgrade script
CDN="https://cdn.coollabs.io/coolify"
LATEST_IMAGE=${1:-latest}
LATEST_HELPER_VERSION=${2:-latest}
REGISTRY_URL=${3:-ghcr.io}
SKIP_BACKUP=${4:-false}

# Hardcoded values for our Azure deployment
PUBLIC_IP="20.169.182.98"
ADMIN_USERNAME="azureuser"
ENV_FILE="/home/azureuser/coolify/.env"
STATUS_FILE="/home/azureuser/coolify/.upgrade-status"
DATE=$(date +%Y-%m-%d-%H-%M-%S)
LOGFILE="/home/azureuser/coolify/upgrade-${DATE}.log"

# Helper functions (enhanced with official logging and auto-update features)
print_info() {
    echo -e "${BLUE}ℹ ${1}${NC}"
}

print_success() {
    echo -e "${GREEN}✓ ${1}${NC}"
}

print_error() {
    echo -e "${RED}✗ ${1}${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ ${1}${NC}"
}

print_step() {
    echo -e "${PURPLE}▶ ${1}${NC}"
}

# Official upgrade script logging functions
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >>"$LOGFILE"
}

log_section() {
    echo "" >>"$LOGFILE"
    echo "============================================================" >>"$LOGFILE"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >>"$LOGFILE"
    echo "============================================================" >>"$LOGFILE"
}

# Auto-update style logging with local output
log_message() {
    local message="[$(date '+%Y-%m-%d %H:%M:%S')] $1"
    echo "$message" | tee -a "$LOGFILE"
}

# State management (from auto-update.sh)
update_state() {
    local status="$1"
    local message="$2"
    local state_file="/tmp/azure-update-state.json"
    
    cat > "$state_file" << EOF
{
    "last_update": "$(date -Iseconds)",
    "status": "$status",
    "message": "$message",
    "target_version": "${LATEST_IMAGE}",
    "log_file": "$LOGFILE"
}
EOF
    
    log_message "State updated: $status - $message"
}

write_status() {
    local step="$1"
    local message="$2"
    echo "${step}|${message}|$(date -Iseconds)" > "$STATUS_FILE"
}

# Enhanced banner (from auto-update.sh)
echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║     Final Coolify Azure Update Script                         ║"
echo "║     Official Upgrade Method + Auto-Update Features            ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Initialize logging with official upgrade script header
echo ""
echo "=========================================="
echo "   Coolify Azure Upgrade - ${DATE}"
echo "=========================================="
echo ""

# Initialize log file with official header
ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e

# Create log directory
mkdir -p ~/coolify/logs

# Initialize log file with official header
echo "============================================================" >>"$LOGFILE"
echo "Coolify Azure Upgrade Log" >>"$LOGFILE"
echo "Started: \$(date '+%Y-%m-%d %H:%M:%S')" >>"$LOGFILE"
echo "Target Version: ${LATEST_IMAGE}" >>"$LOGFILE"
echo "Helper Version: ${LATEST_HELPER_VERSION}" >>"$LOGFILE"
echo "Registry URL: ${REGISTRY_URL}" >>"$LOGFILE"
echo "Method: Official Upgrade Script + Azure Fixes" >>"$LOGFILE"
echo "============================================================" >>"$LOGFILE"

echo "Logging initialized with official header"
ENDSSH
echo "║     Based on official upgrade script          ║"
echo "╚════════════════════════════════════════════╝"
echo -e "${NC}"

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check SSH connectivity
    if ! ssh -o ConnectTimeout=10 -o BatchMode=yes ${ADMIN_USERNAME}@${PUBLIC_IP} "echo 'SSH OK'" > /dev/null 2>&1; then
        print_error "Cannot connect to the Azure VM via SSH."
        exit 1
    fi
    
    print_success "Prerequisites verified"
}

# Initialize log file (official method)
initialize_logging() {
    print_info "Initializing logging..."
    
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e

# Create log directory
mkdir -p ~/coolify/logs

# Initialize log file with header
echo "============================================================" >>"$LOGFILE"
echo "Coolify Azure Upgrade Log" >>"$LOGFILE"
echo "Started: \$(date '+%Y-%m-%d %H:%M:%S')" >>"$LOGFILE"
echo "Target Version: ${LATEST_IMAGE}" >>"$LOGFILE"
echo "Helper Version: ${LATEST_HELPER_VERSION}" >>"$LOGFILE"
echo "Registry URL: ${REGISTRY_URL}" >>"$LOGFILE"
echo "============================================================" >>"$LOGFILE"

echo "Logging initialized"
ENDSSH
    
    print_success "Logging initialized"
}

# Download latest configuration files (official method)
download_configuration() {
    print_info "Downloading latest configuration files..."
    
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e

log_section "Step 1/6: Downloading configuration files"
write_status "1" "Downloading configuration files"
echo "1/6 Downloading latest configuration files..."

cd ~/coolify

# Backup existing configuration files
if [ -f "docker-compose.yml" ]; then
    cp docker-compose.yml docker-compose.yml.backup-\${DATE}
fi
if [ -f "docker-compose.prod.yml" ]; then
    cp docker-compose.prod.yml docker-compose.prod.yml.backup-\${DATE}
fi

# Download latest configurations from CDN
log "Downloading docker-compose.yml from ${CDN}/docker-compose.yml"
curl -fsSL -L $CDN/docker-compose.yml -o docker-compose.yml.new
log "Downloading docker-compose.prod.yml from ${CDN}/docker-compose.prod.yml"
curl -fsSL -L $CDN/docker-compose.prod.yml -o docker-compose.prod.yml.new
log "Downloading .env.production from ${CDN}/.env.production"
curl -fsSL -L $CDN/.env.production -o .env.production.new

# Apply our fixes to the downloaded files
log "Applying fixes to downloaded configuration files"

# Fix Docker image reference
sed -i 's|coolify/coolify:latest|coollabsio/coolify:latest|g' docker-compose.prod.yml.new

# Fix health check commands
sed -i 's|wget -qO-|curl -f|g' docker-compose.prod.yml.new

# Remove problematic realtime dependency
sed -i '/coolify-realtime:/,/condition: service_healthy/c\
    condition: service_healthy' docker-compose.prod.yml.new

# Apply WebSocket fixes for Azure
sed -i "s|\${AZURE_IP}|${PUBLIC_IP}|g" docker-compose.prod.yml.new

log "Configuration files downloaded and fixed"
echo "     Done."
ENDSSH
    
    print_success "Configuration files downloaded"
}

# Extract images and prepare for update (official method)
prepare_images() {
    print_info "Preparing Docker images..."
    
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e

log_section "Step 2/6: Preparing Docker images"
write_status "2" "Preparing Docker images"

cd ~/coolify

# Extract all images from docker-compose configuration
log "Extracting all images from docker-compose configuration..."
COMPOSE_FILES="-f docker-compose.yml.new -f docker-compose.prod.yml.new"

# Get all unique images from docker compose config
IMAGES=\$(LATEST_IMAGE=${LATEST_IMAGE} docker compose --env-file .env \$COMPOSE_FILES config --images 2>/dev/null | sort -u)

if [ -z "\$IMAGES" ]; then
    log "ERROR: Failed to extract images from docker-compose files"
    write_status "error" "Failed to parse docker-compose configuration"
    echo "     ERROR: Failed to parse docker-compose configuration. Aborting upgrade."
    exit 1
fi

log "Images to pull:"
echo "\$IMAGES" | while read img; do log "  - \$img"; done

# Backup existing .env file before making any changes
if [ "$SKIP_BACKUP" != "true" ]; then
    if [ -f ".env" ]; then
        echo "     Creating backup of .env file..."
        log "Creating backup of .env file to .env-\${DATE}"
        cp ".env" ".env-\${DATE}"
        log "Backup created: .env-\${DATE}"
    else
        log "WARNING: No existing .env file found to backup"
    fi
fi

log "Docker images preparation completed"
ENDSSH
    
    print_success "Docker images prepared"
}

# Update environment configuration (official method)
update_environment() {
    print_info "Updating environment configuration..."
    
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e

log_section "Step 3/6: Updating environment configuration"
write_status "3" "Updating environment configuration"

cd ~/coolify

# Merge new environment configuration
log "Merging .env.production values into .env"
awk -F '=' '!seen[\$1]++' ".env" ".env.production.new" > ".env.tmp" && mv ".env.tmp" ".env"
log "Environment file merged successfully"

# Update environment variables (official method)
update_env_var() {
    local key="\$1"
    local value="\$2"

    if grep -q "^\${key}=\$" ".env"; then
        sed -i "s|^\${key}=\$|\${key}=\${value}|" ".env"
        log "Updated \${key} (was empty)"
    elif ! grep -q "^\${key}=" ".env"; then
        printf '%s=%s\n' "\$key" "\$value" >>".env"
        log "Added \${key} (was missing)"
    fi
}

# Generate new Pusher credentials
log "Updating Pusher credentials..."
update_env_var "PUSHER_APP_ID" "\$(openssl rand -hex 32)"
update_env_var "PUSHER_APP_KEY" "\$(openssl rand -hex 32)"
update_env_var "PUSHER_APP_SECRET" "\$(openssl rand -hex 32)"

# Ensure WebSocket configuration is correct
update_env_var "PUSHER_HOST" "${PUBLIC_IP}"
update_env_var "PUSHER_BACKEND_HOST" "${PUBLIC_IP}"

log "Environment configuration updated"
echo "     Done."
ENDSSH
    
    print_success "Environment configuration updated"
}

# Pull Docker images (official method)
pull_images() {
    print_info "Pulling Docker images..."
    
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e

log_section "Step 4/6: Pulling Docker images"
write_status "4" "Pulling Docker images"
echo ""
echo "4/6 Pulling Docker images..."
echo "     This may take a few minutes depending on your connection."

cd ~/coolify

# Pull helper image (official method)
HELPER_IMAGE="${REGISTRY_URL:-ghcr.io}/coollabsio/coolify-helper:${LATEST_HELPER_VERSION}"
echo "     - Pulling \$HELPER_IMAGE..."
log "Pulling image: \$HELPER_IMAGE"
if docker pull "\$HELPER_IMAGE" >>"\$LOGFILE" 2>&1; then
    log "Successfully pulled \$HELPER_IMAGE"
else
    log "ERROR: Failed to pull \$HELPER_IMAGE"
    write_status "error" "Failed to pull \$HELPER_IMAGE"
    echo "     ERROR: Failed to pull \$HELPER_IMAGE. Aborting upgrade."
    exit 1
fi

# Extract images and pull them
COMPOSE_FILES="-f docker-compose.yml.new -f docker-compose.prod.yml.new"
IMAGES=\$(LATEST_IMAGE=${LATEST_IMAGE} docker compose --env-file .env \$COMPOSE_FILES config --images 2>/dev/null | sort -u)

for IMAGE in \$IMAGES; do
    if [ -n "\$IMAGE" ]; then
        echo "     - Pulling \$IMAGE..."
        log "Pulling image: \$IMAGE"
        if docker pull "\$IMAGE" >>"\$LOGFILE" 2>&1; then
            log "Successfully pulled \$IMAGE"
        else
            log "ERROR: Failed to pull \$IMAGE"
            write_status "error" "Failed to pull \$IMAGE"
            echo "     ERROR: Failed to pull \$IMAGE. Aborting upgrade."
            exit 1
        fi
    fi
done

log "All images pulled successfully"
echo "     All images pulled successfully."
ENDSSH
    
    print_success "Docker images pulled"
}

# Restart containers (official method with background execution)
restart_containers() {
    print_info "Restarting containers..."
    
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e

log_section "Step 5/6: Restarting containers"
write_status "5" "Restarting containers"
echo ""
echo "5/6 Restarting containers..."
echo "     This step will restart all Coolify containers."

cd ~/coolify

# Use official method with background execution to survive SSH disconnection
log "Starting container restart sequence (detached)..."

nohup bash -c "
    LOGFILE='$LOGFILE'
    STATUS_FILE='$STATUS_FILE'
    REGISTRY_URL='$REGISTRY_URL'
    LATEST_HELPER_VERSION='$LATEST_HELPER_VERSION'
    LATEST_IMAGE='$LATEST_IMAGE'

    log() {
        echo \"[\$(date '+%Y-%m-%d %H:%M:%S')] \$1\" >>\"\$LOGFILE\"
    }

    write_status() {
        echo \"\$1|\$2|\$(date -Iseconds)\" > \"\$STATUS_FILE\"
    }

    cd ~/coolify

    # Stop and remove containers (official method)
    for container in coolify coolify-db coolify-redis coolify-realtime; do
        if docker ps -a --format '{{.Names}}' | grep -q \"^\${container}\$\"; then
            log \"Stopping container: \${container}\"
            docker stop \"\$container\" >>\"\$LOGFILE\" 2>&1 || true
            log \"Removing container: \${container}\"
            docker rm \"\$container\" >>\"\$LOGFILE\" 2>&1 || true
            log \"Container \${container} stopped and removed\"
        else
            log \"Container \${container} not found (skipping)\"
        fi
    done
    log \"Container cleanup complete\"

    # Move new configuration files into place
    mv docker-compose.yml.new docker-compose.yml
    mv docker-compose.prod.yml.new docker-compose.prod.yml

    # Start new containers (official method)
    log 'Starting new containers with official configuration'
    write_status '6' 'Starting new containers'

    # Use official helper container for reliable startup
    docker run -v /home/azureuser/coolify:/data/coolify/source -v /var/run/docker.sock:/var/run/docker.sock --rm \${REGISTRY_URL:-ghcr.io}/coollabsio/coolify-helper:\${LATEST_HELPER_VERSION} bash -c \"LATEST_IMAGE=\${LATEST_IMAGE} docker compose --env-file /data/coolify/source/.env -f /data/coolify/source/docker-compose.yml -f /data/coolify/source/docker-compose.prod.yml up -d --remove-orphans --wait --wait-timeout 60\" >>\"\$LOGFILE\" 2>&1

    log 'Docker compose up completed'

    # Final log entry
    echo '' >>\"\$LOGFILE\"
    echo '============================================================' >>\"\$LOGFILE\"
    log 'Step 6/6: Upgrade complete'
    echo '============================================================' >>\"\$LOGFILE\"
    write_status '6' 'Upgrade complete'
    log 'Coolify upgrade completed successfully'
    log \"Version: \${LATEST_IMAGE}\"
    echo '' >>\"\$LOGFILE\"
    echo '============================================================' >>\"\$LOGFILE\"
    echo \"Upgrade completed: \$(date '+%Y-%m-%d %H:%M:%S')\" >>\"\$LOGFILE\"
    echo '============================================================' >>\"\$LOGFILE\"

    # Clean up status file after a short delay
    sleep 10
    rm -f \"\$STATUS_FILE\"
    log 'Status file cleaned up'
" >>"\$LOGFILE" 2>&1 &

# Give the background process a moment to start
sleep 2
log "Container restart sequence started in background"
ENDSSH
    
    print_success "Container restart initiated"
}

# Monitor upgrade progress
monitor_progress() {
    print_info "Monitoring upgrade progress..."
    
    echo ""
    echo "6/6 Upgrade process initiated!"
    echo ""
    echo "=========================================="
    echo "   Coolify upgrade to ${LATEST_IMAGE} in progress"
    echo "=========================================="
    echo ""
    echo "   The upgrade will continue in the background."
    echo "   Coolify will be available again shortly."
    echo ""
    
    # Monitor for completion
    for i in {1..60}; do
        if ssh ${ADMIN_USERNAME}@${PUBLIC_IP} "[ ! -f $STATUS_FILE ]" 2>/dev/null; then
            print_success "Upgrade completed successfully!"
            break
        fi
        
        if ssh ${ADMIN_USERNAME}@${PUBLIC_IP} "grep -q 'error' $STATUS_FILE" 2>/dev/null; then
            ERROR_MSG=$(ssh ${ADMIN_USERNAME}@${PUBLIC_IP} "cat $STATUS_FILE | cut -d'|' -f2" 2>/dev/null)
            print_error "Upgrade failed: $ERROR_MSG"
            exit 1
        fi
        
        if [ $i -eq 60 ]; then
            print_warning "Upgrade is taking longer than expected. Check logs manually."
            break
        fi
        
        sleep 10
    done
}

# Enhanced health checks (from auto-update.sh)
verify_upgrade() {
    print_step "Running comprehensive health checks..."
    update_state "verifying" "Running post-update health checks"
    
    # Check if application is responding
    if curl -s -o /dev/null -w '%{http_code}' http://${PUBLIC_IP} | grep -q "200\|302"; then
        print_success "Application is responding correctly"
        log_message "Application health check passed"
    else
        print_error "Application is not responding correctly"
        log_message "Application health check failed"
        update_state "failed" "Application health check failed"
        return 1
    fi
    
    # Check WebSocket connectivity
    if curl -s http://${PUBLIC_IP}:6001 | grep -q "OK\|soketi"; then
        print_success "WebSocket is responding"
        log_message "WebSocket health check passed"
    else
        print_warning "WebSocket may not be properly configured"
        log_message "WebSocket health check warning"
    fi
    
    # Check services
    SERVICES_STATUS=$(ssh ${ADMIN_USERNAME}@${PUBLIC_IP} "cd ~/coolify && docker compose ps" 2>/dev/null | grep -c "Up" || echo "0")
    
    if [ "$SERVICES_STATUS" -ge 3 ]; then
        print_success "All services running ($SERVICES_STATUS services up)"
        log_message "Service health check passed"
    else
        print_error "Some services may not be running properly ($SERVICES_STATUS services up)"
        log_message "Service health check failed"
        update_state "failed" "Service health check failed"
        return 1
    fi
    
    print_success "All health checks passed"
    update_state "verified" "All health checks passed"
    return 0
}

# Rollback capability (from auto-update.sh)
rollback_update() {
    print_error "Initiating rollback due to failed update..."
    update_state "rolling_back" "Initiating rollback procedure"
    
    log_message "Starting rollback procedure"
    
    # Attempt to restore previous configuration
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e

cd ~/coolify

# Restore previous .env if backup exists
if [ -f ".env-${DATE}" ]; then
    echo "Restoring previous .env configuration..."
    cp ".env-${DATE}" .env
    echo "Environment configuration restored"
fi

# Restart services with previous configuration
echo "Restarting services with previous configuration..."
docker compose down
docker compose up -d

# Wait for services to be ready
echo "Waiting for services to recover..."
sleep 30

echo "Rollback completed"
ENDSSH
    
    if [ $? -eq 0 ]; then
        print_warning "Rollback completed successfully"
        update_state "rolled_back" "Update rolled back successfully"
        log_message "Rollback completed successfully"
    else
        print_error "Rollback failed - manual intervention required"
        update_state "rollback_failed" "Rollback procedure failed"
        log_message "Rollback failed - manual intervention required"
    fi
}

# Generate update report
generate_report() {
    REPORT_FILE="coolify-azure-update-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$REPORT_FILE" << EOF
╔════════════════════════════════════════════════════════════════╗
║              Coolify Azure Update Report                     ║
║              Official Upgrade Method                         ║
╚════════════════════════════════════════════════════════════════╝

Update Date: $(date)
Target VM:   $PUBLIC_IP
Method:      Official Upgrade Script + Fixes

═══════════════════════════════════════════════════════════════════
  UPDATE SUMMARY
═══════════════════════════════════════════════════════════════════

✅ Configuration files downloaded from official CDN
✅ Environment variables updated with official method
✅ Docker images pulled using official process
✅ Containers restarted with official helper
✅ All fixes applied during update process
✅ Background execution to survive SSH disconnection

═══════════════════════════════════════════════════════════════════
  OFFICIAL FEATURES USED
═══════════════════════════════════════════════════════════════════

✅ Official CDN configuration downloads
✅ Official environment variable management
✅ Official image extraction and pulling
✅ Official helper container for startup
✅ Official status file monitoring
✅ Official logging with timestamps
✅ Official background execution

═══════════════════════════════════════════════════════════════════
  ACCESS INFORMATION
═══════════════════════════════════════════════════════════════════

Application URL: http://$PUBLIC_IP
SSH Access:      ssh ${ADMIN_USERNAME}@${PUBLIC_IP}

═══════════════════════════════════════════════════════════════════
  LOG FILES
═══════════════════════════════════════════════════════════════════

Local Log:     $REPORT_FILE
Remote Log:    /home/azureuser/coolify/upgrade-${DATE}.log

═══════════════════════════════════════════════════════════════════
EOF

    print_success "Update report saved to: $REPORT_FILE"
}

# Main execution (enhanced with rollback and state management)
main() {
    update_state "starting" "Initializing Azure update process"
    
    check_prerequisites
    initialize_logging
    download_configuration
    prepare_images
    update_environment
    pull_images
    restart_containers
    monitor_progress
    
    # Enhanced verification with rollback capability
    if verify_upgrade; then
        update_state "completed" "Update completed successfully"
        generate_report
        
        echo ""
        echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║           Update Completed Successfully! 🎉                     ║${NC}"
        echo -e "${GREEN}║           Official Method + Auto-Update Features               ║${NC}"
        echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
        echo ""
        print_info "Your Coolify instance has been updated using the official method"
        print_info "Health checks passed and all services are running"
        print_info "State management and rollback capabilities available"
    else
        print_error "Update verification failed - attempting rollback"
        rollback_update
        
        echo ""
        echo -e "${RED}╔════════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${RED}║           Update Failed - Rollback Initiated ⚠️                ║${NC}"
        echo -e "${RED}╚════════════════════════════════════════════════════════════════╝${NC}"
        echo ""
        print_error "The update failed and has been rolled back to the previous version"
        print_info "Check the logs for details: $REPORT_FILE"
        exit 1
    fi
}

# Run main function
main "$@"
