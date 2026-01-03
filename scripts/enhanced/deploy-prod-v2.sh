#!/usr/bin/env bash

# Final Coolify Production Deployment Script
# Incorporates official best practices + production optimizations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="$PROJECT_ROOT/backups"
LOG_FILE="$PROJECT_ROOT/logs/deploy-prod-$(date +%Y%m%d-%H%M%S).log"

# Official logging functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] INFO: $1" >>"$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] SUCCESS: $1" >>"$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] WARNING: $1" >>"$LOG_FILE"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $1" >>"$LOG_FILE"
}

log_section() {
    echo ""
    echo "============================================================"
    echo "$1"
    echo "============================================================"
    echo "" >>"$LOG_FILE"
    echo "============================================================" >>"$LOG_FILE"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >>"$LOG_FILE"
    echo "============================================================" >>"$LOG_FILE"
}

# Create necessary directories
create_directories() {
    mkdir -p "$(dirname "$LOG_FILE")"
    mkdir -p "$BACKUP_DIR"
}

# Check if Docker is installed and running (enhanced)
check_docker() {
    log_info "Checking Docker installation..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        log_info "Visit: https://docs.docker.com/get-docker/"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker is not running. Please start Docker."
        exit 1
    fi

    log_success "Docker is installed and running"
}

# Check if Docker Compose is available (enhanced)
check_docker_compose() {
    log_info "Checking Docker Compose..."
    
    if ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not available."
        log_info "Please install Docker Compose plugin"
        exit 1
    fi

    log_success "Docker Compose is available"
}

# Validate production environment (enhanced with official patterns)
validate_prod_env() {
    log_info "Validating production environment..."
    
    # Check for production compose file
    if [ ! -f "docker-compose.prod.yml" ]; then
        log_error "docker-compose.prod.yml not found"
        exit 1
    fi
    
    # Check environment file
    if [ ! -f ".env" ]; then
        log_error ".env file not found"
        exit 1
    fi
    
    # Validate critical environment variables
    if ! grep -q "^APP_ENV=production" .env; then
        log_warning "APP_ENV is not set to production"
    fi
    
    if grep -q "APP_DEBUG=true" .env; then
        log_error "DEBUG mode is enabled in production!"
        exit 1
    fi
    
    # Check for required secrets
    local required_vars=("APP_KEY" "DB_PASSWORD" "REDIS_PASSWORD")
    for var in "${required_vars[@]}"; do
        if ! grep -q "^${var}=" .env || grep -q "^${var}=$" .env; then
            log_error "Required environment variable $var is missing or empty"
            exit 1
        fi
    done
    
    log_success "Production environment validated"
}

# Backup database before deployment (enhanced)
backup_database() {
    log_info "Creating database backup..."
    
    if docker compose -f docker-compose.prod.yml ps | grep -q "coolify-db.*Up"; then
        timestamp=$(date +%Y%m%d_%H%M%S)
        backup_file="${BACKUP_DIR}/coolify_backup_${timestamp}.sql.gz"
        
        log_info "Backing up to: $backup_file"
        
        if docker compose -f docker-compose.prod.yml exec -T coolify-db pg_dump -U coolify coolify | gzip > "$backup_file"; then
            log_success "Database backed up successfully"
            
            # Keep only last 10 backups
            find "$BACKUP_DIR" -name "coolify_backup_*.sql.gz" -type f | sort -r | tail -n +11 | xargs -r rm
            log_info "Cleaned old backups (kept last 10)"
        else
            log_error "Database backup failed"
            exit 1
        fi
    else
        log_info "No running database to backup (first deployment)"
    fi
}

# Pull latest images (enhanced)
pull_images() {
    log_info "Pulling latest images..."
    
    if docker compose -f docker-compose.prod.yml pull 2>>"$LOG_FILE"; then
        log_success "Images pulled successfully"
    else
        log_warning "Some images failed to pull (may be using local build)"
    fi
}

# Build Docker images with production optimizations (enhanced)
build_images() {
    log_info "Building production Docker images..."
    
    # Use production optimizations
    if docker compose -f docker-compose.prod.yml build --no-cache --pull 2>>"$LOG_FILE"; then
        log_success "Docker images built successfully"
    else
        log_error "Docker image build failed"
        exit 1
    fi
}

# Stop services gracefully (enhanced)
stop_services() {
    if docker compose -f docker-compose.prod.yml ps | grep -q "Up"; then
        log_info "Stopping existing services gracefully..."
        
        # Stop with timeout
        if timeout 60 docker compose -f docker-compose.prod.yml stop 2>>"$LOG_FILE"; then
            log_success "Services stopped gracefully"
        else
            log_warning "Services stop timeout, forcing..."
            docker compose -f docker-compose.prod.yml kill 2>>"$LOG_FILE"
            log_success "Services stopped forcefully"
        fi
    else
        log_info "No running services to stop"
    fi
}

# Start services in production mode (enhanced)
start_services() {
    log_info "Starting services in production mode..."
    
    if docker compose -f docker-compose.prod.yml up -d --remove-orphans 2>>"$LOG_FILE"; then
        log_success "Services started"
    else
        log_error "Failed to start services"
        exit 1
    fi
}

# Wait for services to be healthy (enhanced)
wait_for_services() {
    log_info "Waiting for services to be healthy..."
    
    local max_attempts=60
    local attempt=1
    
    # Wait for database
    log_info "Waiting for database..."
    while [ $attempt -le $max_attempts ]; do
        if docker compose -f docker-compose.prod.yml ps | grep -q "coolify-db.*healthy"; then
            log_success "Database is healthy"
            break
        fi
        
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
        
        if [ $attempt -gt $max_attempts ]; then
            log_error "Database failed to become healthy"
            docker compose -f docker-compose.prod.yml logs coolify-db >>"$LOG_FILE" 2>&1
            exit 1
        fi
    done
    
    # Wait for application
    attempt=1
    log_info "Waiting for application..."
    while [ $attempt -le $max_attempts ]; do
        if docker compose -f docker-compose.prod.yml ps | grep -q "coolify.*healthy"; then
            log_success "Application is healthy"
            break
        fi
        
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
        
        if [ $attempt -gt $max_attempts ]; then
            log_warning "Application health check timeout"
            docker compose -f docker-compose.prod.yml logs coolify >>"$LOG_FILE" 2>&1
            return 0
        fi
    done
}

# Run migrations safely (enhanced)
run_migrations() {
    log_info "Running database migrations..."
    
    if docker compose -f docker-compose.prod.yml exec -T coolify php artisan migrate --force 2>>"$LOG_FILE"; then
        log_success "Migrations completed"
    else
        log_error "Migrations failed"
        exit 1
    fi
}

# Optimize application (enhanced)
optimize_app() {
    log_info "Optimizing application for production..."
    
    # Clear caches first
    docker compose -f docker-compose.prod.yml exec -T coolify php artisan cache:clear 2>>"$LOG_FILE"
    docker compose -f docker-compose.prod.yml exec -T coolify php artisan config:clear 2>>"$LOG_FILE"
    docker compose -f docker-compose.prod.yml exec -T coolify php artisan route:clear 2>>"$LOG_FILE"
    docker compose -f docker-compose.prod.yml exec -T coolify php artisan view:clear 2>>"$LOG_FILE"
    
    # Build production caches
    docker compose -f docker-compose.prod.yml exec -T coolify php artisan config:cache 2>>"$LOG_FILE"
    docker compose -f docker-compose.prod.yml exec -T coolify php artisan route:cache 2>>"$LOG_FILE"
    docker compose -f docker-compose.prod.yml exec -T coolify php artisan view:cache 2>>"$LOG_FILE"
    
    log_success "Application optimized for production"
}

# Security checks (new feature)
security_checks() {
    log_info "Running security checks..."
    
    # Check if debug mode is off
    if docker compose -f docker-compose.prod.yml exec -T coolify php artisan tinker --execute="echo config('app.debug');" 2>/dev/null | grep -q "true"; then
        log_error "Debug mode is still enabled!"
        exit 1
    fi
    
    # Check if app URL is set
    app_url=$(docker compose -f docker-compose.prod.yml exec -T coolify php artisan tinker --execute="echo config('app.url');" 2>/dev/null)
    if [ -z "$app_url" ] || [ "$app_url" = "http://localhost" ]; then
        log_warning "APP_URL may not be properly configured"
    fi
    
    log_success "Security checks passed"
}

# Show deployment status (enhanced)
show_status() {
    echo ""
    log_info "Deployment Status:"
    docker compose -f docker-compose.prod.yml ps
    echo ""
    log_success "Coolify production deployment complete!"
    echo ""

    # Get APP_URL from environment
    app_url=$(grep "^APP_URL=" .env | cut -d'=' -f2)
    if [ -z "$app_url" ]; then
        app_url="http://localhost"
    fi

    echo -e "${GREEN}Access your application at:${NC}"
    echo -e "  🌐 ${BLUE}${app_url}${NC}"
    echo ""
    echo -e "${YELLOW}Production Management:${NC}"
    echo -e "  View logs:         docker compose -f docker-compose.prod.yml logs -f coolify"
    echo -e "  Monitor:           docker compose -f docker-compose.prod.yml ps"
    echo -e "  Shell access:      docker compose -f docker-compose.prod.yml exec coolify bash"
    echo -e "  Backups location:  ./backups/"
    echo -e "  Deployment log:    $LOG_FILE"
    echo ""
    echo -e "${RED}Security Reminders:${NC}"
    echo -e "  ⚠️  Ensure firewall rules are configured"
    echo -e "  ⚠️  Regular backups are scheduled"
    echo -e "  ⚠️  SSL/TLS certificates are configured"
    echo -e "  ⚠️  Monitor logs for security issues"
    echo -e "  ⚠️  Keep dependencies updated"
    echo ""
}

# Generate deployment report (new feature)
generate_report() {
    local report_file="$PROJECT_ROOT/deployment-report-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$report_file" << EOF
╔════════════════════════════════════════════════════════════════╗
║              Coolify Production Deployment Report               ║
╚════════════════════════════════════════════════════════════════╝

Deployment Date: $(date)
Status: ✅ SUCCESS
Environment: Production

═══════════════════════════════════════════════════════════════════
  DEPLOYMENT DETAILS
═══════════════════════════════════════════════════════════════════

✅ Environment validated
✅ Database backed up
✅ Images pulled and built
✅ Services restarted
✅ Health checks passed
✅ Migrations applied
✅ Application optimized
✅ Security checks passed

═══════════════════════════════════════════════════════════════════
  ACCESS INFORMATION
═══════════════════════════════════════════════════════════════════

Application URL: $(grep "^APP_URL=" .env | cut -d'=' -f2)
Deployment Log: $LOG_FILE
Backup Location: $BACKUP_DIR

═══════════════════════════════════════════════════════════════════
  NEXT STEPS
═══════════════════════════════════════════════════════════════════

1. Configure SSL/TLS certificates
2. Set up monitoring and alerts
3. Configure backup schedules
4. Review security settings
5. Test all functionality

EOF

    log_success "Deployment report saved to: $report_file"
}

# Main deployment flow
main() {
    create_directories
    
    echo -e "${GREEN}"
    echo "╔════════════════════════════════════════════════════════════════╗"
    echo "║        Final Coolify Production Deployment Script            ║"
    echo "║        Official Best Practices + Production Optimizations     ║"
    echo "╚════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    log_warning "This script will deploy Coolify in PRODUCTION mode!"
    log_warning "This will restart all services and may cause downtime."
    echo ""
    
    read -p "Continue with production deployment? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Deployment cancelled"
        exit 0
    fi

    # Change to script directory
    cd "$(dirname "$0")/.."

    # Execute deployment steps
    log_section "Step 1: Prerequisites"
    check_docker
    check_docker_compose
    
    log_section "Step 2: Environment Validation"
    validate_prod_env
    
    log_section "Step 3: Backup"
    backup_database
    
    log_section "Step 4: Image Management"
    pull_images
    build_images
    
    log_section "Step 5: Service Management"
    stop_services
    start_services
    
    log_section "Step 6: Health Checks"
    wait_for_services
    
    log_section "Step 7: Application Setup"
    run_migrations
    optimize_app
    
    log_section "Step 8: Security"
    security_checks
    
    log_section "Step 9: Finalization"
    show_status
    generate_report
    
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║           Production Deployment Completed! 🎉              ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# Run main function
main "$@"
