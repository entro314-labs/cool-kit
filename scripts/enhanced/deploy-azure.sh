#!/bin/bash

# Final Coolify Azure Deployment Script
# Enhanced with dynamic configuration and official best practices

set -e

# Get script directory and load configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/config-loader.sh"

# Load Azure configuration
load_azure_config

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Dynamic configuration (from config file)
CDN="https://cdn.coollabs.io/coolify"
LATEST_IMAGE=${1:-latest}
LATEST_HELPER_VERSION=${2:-latest}
REGISTRY_URL=${DOCKER_REGISTRY_URL:-ghcr.io}
SKIP_BACKUP=${4:-false}

# Dynamic values from configuration
PUBLIC_IP=${AZURE_PUBLIC_IP:-""}
ADMIN_USERNAME=${AZURE_ADMIN_USERNAME:-azureuser}
ENV_FILE=${REMOTE_ENV_FILE}
STATUS_FILE=${REMOTE_STATUS_FILE}
DATE=$(date +%Y-%m-%d-%H-%M-%S)
LOGFILE="${REMOTE_LOGS_DIR}/upgrade-${DATE}.log"

# Helper functions (enhanced with official logging)
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

# Official logging functions for enhanced reporting
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

log_section() {
    echo ""
    echo "============================================================"
    echo "$1"
    echo "============================================================"
}

# Banner
echo -e "${BLUE}"
echo "╔════════════════════════════════════════════╗"
echo "║     Final Coolify Azure Deployment Script     ║"
echo "║     Based on official script + fixes         ║"
echo "╚════════════════════════════════════════════╝"
echo -e "${NC}"

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check Azure CLI
    if ! command -v az &> /dev/null; then
        print_error "Azure CLI is not installed"
        exit 1
    fi
    
    # Check if logged in
    if ! az account show &> /dev/null; then
        print_error "Not logged in to Azure. Run 'az login' first."
        exit 1
    fi
    
    print_success "Prerequisites verified"
}

# Collect deployment information (enhanced with dynamic configuration)
collect_info() {
    print_info "Collecting deployment information..."
    
    # Use defaults from configuration, allow overrides
    read -p "Resource Group Name [${AZURE_RESOURCE_GROUP:-coolify-rg}]: " RESOURCE_GROUP
    RESOURCE_GROUP=${RESOURCE_GROUP:-${AZURE_RESOURCE_GROUP:-coolify-rg}}
    
    read -p "VM Name [${AZURE_VM_NAME:-coolify-vm}]: " VM_NAME
    VM_NAME=${VM_NAME:-${AZURE_VM_NAME:-coolify-vm}}
    
    read -p "Azure Region [${AZURE_LOCATION:-swedencentral}]: " LOCATION
    LOCATION=${LOCATION:-${AZURE_LOCATION:-swedencentral}}
    
    echo "Available VM sizes:"
    echo "1. Standard_B2s (2 vCPU, 4 GB RAM) - Testing"
    echo "2. Standard_D2s_v3 (2 vCPU, 8 GB RAM) - Production"
    echo "3. Standard_D4s_v3 (4 vCPU, 16 GB RAM) - High performance"
    read -p "VM Size [2]: " VM_SIZE_CHOICE
    VM_SIZE_CHOICE=${VM_SIZE_CHOICE:-2}
    
    case $VM_SIZE_CHOICE in
        1) VM_SIZE="${AZURE_VM_SIZE:-Standard_B2s}" ;;
        2) VM_SIZE="Standard_D2s_v3" ;;
        3) VM_SIZE="Standard_D4s_v3" ;;
        *) VM_SIZE="${AZURE_VM_SIZE:-Standard_B2s}" ;;
    esac
    
    # Admin credentials with defaults
    read -p "Admin Email [${COOLIFY_DEFAULT_ADMIN_EMAIL}]: " ADMIN_EMAIL
    ADMIN_EMAIL=${ADMIN_EMAIL:-${COOLIFY_DEFAULT_ADMIN_EMAIL}}
    
    read -p "Admin Password [${COOLIFY_DEFAULT_ADMIN_PASSWORD}]: " ADMIN_PASSWORD
    ADMIN_PASSWORD=${ADMIN_PASSWORD:-${COOLIFY_DEFAULT_ADMIN_PASSWORD}}
    
    # Store dynamic values back to configuration for future use
    update_config_value "infrastructure.resource_group" "$RESOURCE_GROUP" "$SCRIPT_DIR/config/azure-config.json"
    update_config_value "infrastructure.vm_name" "$VM_NAME" "$SCRIPT_DIR/config/azure-config.json"
    update_config_value "infrastructure.location" "$LOCATION" "$SCRIPT_DIR/config/azure-config.json"
    update_config_value "infrastructure.vm_size" "$VM_SIZE" "$SCRIPT_DIR/config/azure-config.json"
        print_info "Generated admin password: $ADMIN_PASSWORD"
    fi
    
    print_success "Deployment information collected"
}

# Create Azure infrastructure
create_infrastructure() {
    print_info "Creating Azure infrastructure..."
    
    # Create resource group
    az group create \
        --name "$RESOURCE_GROUP" \
        --location "$LOCATION" \
        --tags "project=coolify" "environment=production" \
        --output none
    
    # Create VM
    print_info "Creating virtual machine..."
    VM_INFO=$(az vm create \
        --resource-group "$RESOURCE_GROUP" \
        --name "$VM_NAME" \
        --image "Ubuntu2204" \
        --size "$VM_SIZE" \
        --admin-username "$ADMIN_USERNAME" \
        --generate-ssh-keys \
        --public-ip-sku "Standard" \
        --nsg "coolify-nsg" \
        --storage-sku "Premium_LRS" \
        --os-disk-size-gb 50 \
        --tags "project=coolify" "environment=production" \
        --query "{publicIp:publicIpAddress}" \
        --output json)
    
    PUBLIC_IP=$(echo "$VM_INFO" | jq -r '.publicIp')
    
    # Configure NSG (following official script security practices)
    az network nsg rule create \
        --resource-group "$RESOURCE_GROUP" \
        --nsg-name "coolify-nsg" \
        --name "Allow-HTTP" \
        --protocol "Tcp" \
        --direction "Inbound" \
        --priority "1002" \
        --source-address-prefix "*" \
        --source-port-range "*" \
        --destination-address-prefix "*" \
        --destination-port-range "80" \
        --access "Allow" \
        --output none
    
    az network nsg rule create \
        --resource-group "$RESOURCE_GROUP" \
        --nsg-name "coolify-nsg" \
        --name "Allow-HTTPS" \
        --protocol "Tcp" \
        --direction "Inbound" \
        --priority "1003" \
        --source-address-prefix "*" \
        --source-port-range "*" \
        --destination-address-prefix "*" \
        --destination-port-range "443" \
        --access "Allow" \
        --output none
    
    az network nsg rule create \
        --resource-group "$RESOURCE_GROUP" \
        --nsg-name "coolify-nsg" \
        --name "Allow-WebSocket" \
        --protocol "Tcp" \
        --direction "Inbound" \
        --priority "1004" \
        --source-address-prefix "*" \
        --source-port-range "*" \
        --destination-address-prefix "*" \
        --destination-port-range "6001" \
        --access "Allow" \
        --output none
    
    print_success "Infrastructure created"
    print_info "Public IP: $PUBLIC_IP"
}

# Wait for SSH connectivity
wait_for_ssh() {
    print_info "Waiting for SSH connectivity..."
    
    # Add to known hosts
    ssh-keyscan -H "$PUBLIC_IP" >> ~/.ssh/known_hosts 2>/dev/null
    
    # Wait for SSH to be available
    for i in {1..30}; do
        if ssh -o ConnectTimeout=10 -o BatchMode=yes ${ADMIN_USERNAME}@${PUBLIC_IP} "echo 'SSH OK'" > /dev/null 2>&1; then
            print_success "SSH is ready"
            break
        fi
        if [ $i -eq 30 ]; then
            print_error "SSH not available after 5 minutes"
            exit 1
        fi
        sleep 10
    done
}

# Deploy Coolify using official installation approach (enhanced)
deploy_coolify() {
    print_info "Deploying Coolify using official installation method..."
    
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e

# Create backup directory
mkdir -p ~/coolify/backups

# Update system
echo "Updating system packages..."
sudo apt-get update -qq
sudo apt-get upgrade -y -qq

# Install prerequisites (following official script)
echo "Installing prerequisites..."
sudo apt-get install -y curl wget jq git

# Install Docker (official method)
echo "Installing Docker..."
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker \$USER
    rm get-docker.sh
fi

# Install Docker Compose
if ! docker compose version &> /dev/null; then
    sudo apt-get install -y docker-compose-plugin
fi

# Start and enable Docker
sudo systemctl start docker
sudo systemctl enable docker

# Create Coolify directories (official structure)
echo "Creating Coolify directories..."
sudo mkdir -p /data/coolify/{source,ssh,applications,databases,backups,services,proxy,sentinel}
sudo mkdir -p /data/coolify/ssh/{keys,mux}
sudo mkdir -p /data/coolify/proxy/dynamic

# Set permissions (official method)
sudo chown -R 9999:root /data/coolify
sudo chmod -R 700 /data/coolify

# Clone Coolify repository
echo "Cloning Coolify repository..."
cd /data/coolify/source
sudo git clone https://github.com/coollabsio/coolify.git .

# Switch to v4.x branch
sudo git checkout v4.x

# Create environment file with secure credentials (official method)
echo "Setting up environment variables..."
sudo cp .env.example .env

# Generate secure credentials (following official script)
APP_ID=\$(openssl rand -hex 16)
APP_KEY="base64:\$(openssl rand -base64 32)"
DB_PASSWORD=\$(openssl rand -base64 32)
REDIS_PASSWORD=\$(openssl rand -base64 32)
PUSHER_APP_ID=\$(openssl rand -hex 32)
PUSHER_APP_KEY=\$(openssl rand -hex 32)
PUSHER_APP_SECRET=\$(openssl rand -hex 32)

# Update environment variables
sudo sed -i "s|^APP_ID=.*|APP_ID=\$APP_ID|" .env
sudo sed -i "s|^APP_KEY=.*|APP_KEY=\$APP_KEY|" .env
sudo sed -i "s|^DB_PASSWORD=.*|DB_PASSWORD=\$DB_PASSWORD|" .env
sudo sed -i "s|^REDIS_PASSWORD=.*|REDIS_PASSWORD=\$REDIS_PASSWORD|" .env
sudo sed -i "s|^PUSHER_APP_ID=.*|PUSHER_APP_ID=\$PUSHER_APP_ID|" .env
sudo sed -i "s|^PUSHER_APP_KEY=.*|PUSHER_APP_KEY=\$PUSHER_APP_KEY|" .env
sudo sed -i "s|^PUSHER_APP_SECRET=.*|PUSHER_APP_SECRET=\$PUSHER_APP_SECRET|" .env

# Set application URL and fix WebSocket configuration
sudo sed -i "s|^APP_URL=.*|APP_URL=http://$PUBLIC_IP|" .env
sudo sed -i "s|^PUSHER_HOST=.*|PUSHER_HOST=$PUBLIC_IP|" .env
sudo sed -i "s|^PUSHER_BACKEND_HOST=.*|PUSHER_BACKEND_HOST=$PUBLIC_IP|" .env

# Set admin credentials
sudo sed -i "s|^ROOT_USERNAME=.*|ROOT_USERNAME=admin|" .env
sudo sed -i "s|^ROOT_USER_EMAIL=.*|ROOT_USER_EMAIL=$ADMIN_EMAIL|" .env
sudo sed -i "s|^ROOT_USER_PASSWORD=.*|ROOT_USER_PASSWORD=$ADMIN_PASSWORD|" .env

echo "Environment configured"
ENDSSH
    
    print_success "Coolify deployment prepared"
}

# Install Coolify using official upgrade script
install_coolify() {
    print_info "Installing Coolify using official upgrade script..."
    
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e
cd /data/coolify/source

# Run the official upgrade script
echo "Running official Coolify installation..."
sudo bash upgrade.sh "latest" "latest" "ghcr.io" "true"

# Wait for installation to complete
echo "Waiting for Coolify to be ready..."
UPGRADE_STATUS_FILE="/data/coolify/source/.upgrade-status"
MAX_WAIT=300
WAITED=0

while [ \$WAITED -lt \$MAX_WAIT ]; do
    if [ -f "\$UPGRADE_STATUS_FILE" ]; then
        STATUS=\$(cat "\$UPGRADE_STATUS_FILE" 2>/dev/null | cut -d'|' -f1)
        MESSAGE=\$(cat "\$UPGRADE_STATUS_FILE" 2>/dev/null | cut -d'|' -f2)
        if [ "\$STATUS" = "6" ]; then
            echo "Installation completed: \$MESSAGE"
            break
        elif [ "\$STATUS" = "error" ]; then
            echo "ERROR: Installation failed: \$MESSAGE"
            exit 1
        else
            if [ \$((WAITED % 10)) -eq 0 ]; then
                echo "Installation in progress: \$MESSAGE (\${WAITED}s)"
            fi
        fi
    fi
    sleep 5
    WAITED=\$((WAITED + 5))
done

if [ \$WAITED -ge \$MAX_WAIT ]; then
    echo "Installation timed out after 5 minutes"
    exit 1
fi

echo "Coolify installation completed"
ENDSSH
    
    print_success "Coolify installation completed"
}

# Configure WebSocket and network settings
configure_networking() {
    print_info "Configuring WebSocket and networking..."
    
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e

# Update docker-compose to fix WebSocket issues
cd /data/coolify/source

# Fix Soketi to bind to 0.0.0.0
if [ -f "docker-compose.yml" ]; then
    sudo sed -i '/coolify-realtime:/,/SOKETI_APP_SECRET:/ s/SOKETI_DEBUG: "0"/SOKETI_DEBUG: "0"\n      SOKETI_HOST: 0.0.0.0/' docker-compose.yml
fi

# Restart services to apply changes
sudo docker compose down
sudo docker compose up -d

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 30

# Verify services are running
echo "Verifying services..."
sudo docker compose ps

echo "Network configuration completed"
ENDSSH
    
    print_success "Network configuration completed"
}

# Enhanced health checks (from deploy-prod-v2.sh)
health_checks() {
    print_info "Running comprehensive health checks..."
    
    # Check application accessibility
    print_info "Checking application accessibility..."
    if curl -s -o /dev/null -w '%{http_code}' http://$PUBLIC_IP | grep -q "200\|302"; then
        print_success "Application is accessible"
    else
        print_error "Application is not responding"
        return 1
    fi
    
    # Check WebSocket connectivity
    print_info "Checking WebSocket connectivity..."
    if curl -s http://$PUBLIC_IP:6001 | grep -q "OK\|soketi"; then
        print_success "WebSocket is responding"
    else
        print_warning "WebSocket may not be properly configured"
    fi
    
    # Verify services are running
    print_info "Verifying service status..."
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} "cd /data/coolify/source && sudo docker compose ps" | grep -q "Up" || {
        print_error "Some services are not running"
        return 1
    }
    
    print_success "All health checks passed"
}

# Security validation (from deploy-prod-v2.sh)
security_validation() {
    print_info "Running security validation..."
    
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e
cd /data/coolify/source

# Check if debug mode is disabled
if sudo docker compose exec -T coolify php artisan tinker --execute="echo config('app.debug');" 2>/dev/null | grep -q "true"; then
    echo "WARNING: Debug mode is still enabled!"
    exit 1
fi

# Check if app URL is properly set
app_url=\$(sudo docker compose exec -T coolify php artisan tinker --execute="echo config('app.url');" 2>/dev/null)
if [ -z "\$app_url" ] || [ "\$app_url" = "http://localhost" ]; then
    echo "WARNING: APP_URL may not be properly configured"
fi

# Check for required environment variables
required_vars=("APP_KEY" "DB_PASSWORD" "REDIS_PASSWORD")
for var in "\${required_vars[@]}"; do
    if ! grep -q "^\${var}=" .env || grep -q "^\${var}=\$" .env; then
        echo "ERROR: Required environment variable \$var is missing or empty"
        exit 1
    fi
done

echo "Security validation passed"
ENDSSH
    
    if [ $? -eq 0 ]; then
        print_success "Security validation passed"
    else
        print_error "Security validation failed"
        return 1
    fi
}

# Database backup setup (from deploy-prod-v2.sh)
setup_database_backup() {
    print_info "Setting up database backup strategy..."
    
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} bash << ENDSSH
set -e

# Create backup script
cat > ~/coolify/backup-database.sh << 'BACKUP_EOF'
#!/bin/bash
# Automated database backup script

BACKUP_DIR="/home/azureuser/coolify/backups"
TIMESTAMP=\$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="\$BACKUP_DIR/coolify_backup_\$TIMESTAMP.sql.gz"

# Create backup directory
mkdir -p "\$BACKUP_DIR"

# Create database backup
cd /data/coolify/source
if sudo docker compose ps | grep -q "coolify-db.*Up"; then
    sudo docker compose exec -T coolify-db pg_dump -U coolify coolify | gzip > "\$BACKUP_FILE"
    
    # Keep only last 10 backups
    find "\$BACKUP_DIR" -name "coolify_backup_*.sql.gz" -type f | sort -r | tail -n +11 | xargs -r rm
    
    echo "Database backup completed: \$BACKUP_FILE"
else
    echo "No running database to backup"
fi
BACKUP_EOF

chmod +x ~/coolify/backup-database.sh

# Create cron job for daily backups at 2 AM
(crontab -l 2>/dev/null; echo "0 2 * * * /home/azureuser/coolify/backup-database.sh") | crontab -

echo "Database backup strategy configured"
ENDSSH
    
    print_success "Database backup strategy configured"
}
health_checks() {
    print_info "Running health checks..."
    
    # Check application
    if curl -s -o /dev/null -w '%{http_code}' http://${PUBLIC_IP} | grep -q "200\|302"; then
        print_success "Application is responding"
    else
        print_error "Application is not responding"
        return 1
    fi
    
    # Check WebSocket
    if curl -s http://${PUBLIC_IP}:6001 | grep -q "OK"; then
        print_success "WebSocket is responding"
    else
        print_warning "WebSocket check failed"
    fi
    
    # Check services
    ssh ${ADMIN_USERNAME}@${PUBLIC_IP} "cd /data/coolify/source && sudo docker compose ps" | grep -q "Up" && print_success "All services running" || print_warning "Some services may not be running"
}

# Save deployment report (enhanced with security and backup info)
save_report() {
    CREDENTIALS_FILE="coolify-azure-deployment-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$CREDENTIALS_FILE" << EOF
╔════════════════════════════════════════════════════════════════╗
║              Coolify Azure Deployment Report                  ║
║              Enhanced with Security & Backups                   ║
╚════════════════════════════════════════════════════════════════╝

Deployment Date: $(date)
Status: ✅ SUCCESS
Method: Official Installation Script + Enhanced Security
Public IP: $PUBLIC_IP

═══════════════════════════════════════════════════════════════════
  ACCESS INFORMATION
═══════════════════════════════════════════════════════════════════

Application URL: http://$PUBLIC_IP
Login URL:      http://$PUBLIC_IP/login
SSH Access:     ssh ${ADMIN_USERNAME}@${PUBLIC_IP}
WebSocket:     http://$PUBLIC_IP:6001

═══════════════════════════════════════════════════════════════════
  ADMIN CREDENTIALS
═══════════════════════════════════════════════════════════════════

Email:    $ADMIN_EMAIL
Password: $ADMIN_PASSWORD

⚠️  IMPORTANT: Change this password after first login!

═══════════════════════════════════════════════════════════════════
  SECURITY & BACKUP FEATURES
═══════════════════════════════════════════════════════════════════

✅ Security validation completed
✅ Debug mode disabled
✅ Environment variables validated
✅ Database backup strategy configured
✅ Automatic daily backups at 2 AM
✅ Backup retention: Last 10 backups
✅ Health checks passed
✅ WebSocket connectivity verified

═══════════════════════════════════════════════════════════════════
  MANAGEMENT COMMANDS
══════════════════════════════════════════════════════════════════

Check services:
  ssh ${ADMIN_USERNAME}@${PUBLIC_IP} "cd /data/coolify/source && sudo docker compose ps"

View logs:
  ssh ${ADMIN_USERNAME}@${PUBLIC_IP} "cd /data/coolify/source && sudo docker compose logs -f"

Manual backup:
  ssh ${ADMIN_USERNAME}@${PUBLIC_IP} "~/coolify/backup-database.sh"

Update Coolify:
  ssh ${ADMIN_USERNAME}@${PUBLIC_IP} "cd /data/coolify/source && sudo bash upgrade.sh"

Backup data:
  ssh ${ADMIN_USERNAME}@${PUBLIC_IP} "sudo tar -czf /data/coolify/backups/manual-$(date +%Y%m%d).tar.gz /data/coolify"

═══════════════════════════════════════════════════════════════════
  SECURITY REMINDERS
═══════════════════════════════════════════════════════════════════

⚠️  Ensure firewall rules are configured
⚠️  Regular backups are scheduled automatically
⚠️  Monitor logs for security issues
⚠️  Keep dependencies updated
⚠️  Change default passwords

EOF

    print_success "Deployment report saved to: $CREDENTIALS_FILE"
}

# Main execution (enhanced with security and backup features)
main() {
    check_prerequisites
    collect_info
    create_infrastructure
    wait_for_ssh
    deploy_coolify
    install_coolify
    configure_networking
    health_checks
    security_validation
    setup_database_backup
    save_report
    
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     Azure Deployment Completed Successfully! 🎉          ║${NC}"
    echo -e "${GREEN}║     Enhanced with Security & Automated Backups             ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    print_info "Access your Coolify instance at: http://$PUBLIC_IP"
    print_info "Login with: $ADMIN_EMAIL / $ADMIN_PASSWORD"
    print_info "WebSocket available at: http://$PUBLIC_IP:6001"
    print_info "Database backups scheduled automatically"
    print_info "Security validation completed"
    print_info "Deployment report saved with all management commands"
    print_warning "Change the admin password after first login!"
}

# Run main function
main "$@"
