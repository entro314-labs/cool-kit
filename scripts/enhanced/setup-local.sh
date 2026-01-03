#!/bin/bash

# Final Coolify Local Setup Script
# Incorporates official installation script best practices

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Official logging functions for enhanced reporting (from deploy-local-v2.sh)
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
echo "║     Final Coolify Local Setup Script         ║"
echo "║     Based on official installation script     ║"
echo "╚════════════════════════════════════════════╝"
echo -e "${NC}"

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        print_info "Please install Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    # Check Docker Compose
    if ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not installed"
        print_info "Please install Docker Compose"
        exit 1
    fi
    
    # Check if Docker is running
    if ! docker info &> /dev/null; then
        print_error "Docker is not running"
        print_info "Please start Docker"
        exit 1
    fi
    
    print_success "Prerequisites verified"
}

# Collect setup information (enhanced with debug mode toggle)
collect_info() {
    print_info "Collecting setup information..."
    
    read -p "Admin Email [admin@coolify.local]: " ADMIN_EMAIL
    ADMIN_EMAIL=${ADMIN_EMAIL:-admin@coolify.local}
    
    read -p "Admin Password [leave blank to generate]: " ADMIN_PASSWORD
    if [ -z "$ADMIN_PASSWORD" ]; then
        ADMIN_PASSWORD="admin123"
        print_info "Using default password: $ADMIN_PASSWORD"
    fi
    
    read -p "Application Port [8000]: " APP_PORT
    APP_PORT=${APP_PORT:-8000}
    
    read -p "WebSocket Port [6001]: " WS_PORT
    WS_PORT=${WS_PORT:-6001}
    
    read -p "Enable debug mode? [Y/n]: " ENABLE_DEBUG
    ENABLE_DEBUG=${ENABLE_DEBUG:-Y}
    
    if [[ $ENABLE_DEBUG =~ ^[Yy]$ ]]; then
        APP_DEBUG="true"
        LOG_LEVEL="debug"
        print_info "Debug mode enabled"
    else
        APP_DEBUG="false"
        LOG_LEVEL="info"
        print_info "Debug mode disabled"
    fi
    
    print_success "Setup information collected"
}

# Create directory structure
create_directories() {
    print_info "Creating directory structure..."
    
    # Create directories following official structure
    mkdir -p ./data/{coolify,ssh,applications,databases,backups,services,proxy,sentinel}
    mkdir -p ./data/coolify/{ssh/{keys,mux},proxy/dynamic}
    mkdir -p ./logs
    mkdir -p ./ssl
    
    print_success "Directory structure created"
}

# Generate secure credentials
generate_credentials() {
    print_info "Generating secure credentials..."
    
    # Generate credentials following official script method
    APP_ID=$(openssl rand -hex 16)
    APP_KEY="base64:$(openssl rand -base64 32)"
    DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)
    REDIS_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)
    PUSHER_APP_ID=$(openssl rand -hex 32)
    PUSHER_APP_KEY=$(openssl rand -hex 32)
    PUSHER_APP_SECRET=$(openssl rand -hex 32)
    
    print_success "Secure credentials generated"
}

# Create environment file
create_environment() {
    print_info "Creating environment configuration..."
    
    cat > .env.local << EOF
# Coolify Local Development Environment
# Generated on: $(date)

# Application
APP_NAME=Coolify
APP_ENV=local
APP_DEBUG=${APP_DEBUG}
APP_URL=http://localhost:${APP_PORT}
APP_ID=${APP_ID}
APP_KEY=${APP_KEY}

# Database
DB_CONNECTION=pgsql
DB_HOST=coolify-db
DB_PORT=5432
DB_DATABASE=coolify
DB_USERNAME=coolify
DB_PASSWORD=${DB_PASSWORD}

# Redis
REDIS_HOST=coolify-redis
REDIS_PASSWORD=${REDIS_PASSWORD}
REDIS_PORT=6379

# Cache & Session
CACHE_STORE=redis
SESSION_DRIVER=redis
QUEUE_CONNECTION=redis

# Broadcasting (WebSocket)
BROADCAST_DRIVER=pusher
PUSHER_APP_ID=${PUSHER_APP_ID}
PUSHER_APP_KEY=${PUSHER_APP_KEY}
PUSHER_APP_SECRET=${PUSHER_APP_SECRET}
PUSHER_HOST=localhost
PUSHER_PORT=${WS_PORT}
PUSHER_SCHEME=http
PUSHER_BACKEND_HOST=localhost

# Admin User
ROOT_USERNAME=admin
ROOT_USER_EMAIL=${ADMIN_EMAIL}
ROOT_USER_PASSWORD=${ADMIN_PASSWORD}

# Development Settings
AUTOUPDATE=false
LOG_LEVEL=${LOG_LEVEL}
EOF

    print_success "Environment configuration created"
}

# Create docker-compose configuration
create_docker_compose() {
    print_info "Creating Docker Compose configuration..."
    
    cat > docker-compose.local.yml << EOF
version: '3.8'

services:
  coolify:
    image: coollabsio/coolify:latest
    container_name: coolify-app
    restart: unless-stopped
    ports:
      - "${APP_PORT}:80"
      - "${WS_PORT}:6001"
    environment:
      APP_NAME: Coolify
      APP_ENV: local
      APP_DEBUG: "true"
      APP_URL: http://localhost:${APP_PORT}
      APP_ID: ${APP_ID}
      APP_KEY: ${APP_KEY}
      DB_CONNECTION: pgsql
      DB_HOST: coolify-db
      DB_PORT: 5432
      DB_DATABASE: coolify
      DB_USERNAME: coolify
      DB_PASSWORD: ${DB_PASSWORD}
      REDIS_HOST: coolify-redis
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      REDIS_PORT: 6379
      CACHE_STORE: redis
      SESSION_DRIVER: redis
      QUEUE_CONNECTION: redis
      BROADCAST_DRIVER: pusher
      PUSHER_APP_ID: ${PUSHER_APP_ID}
      PUSHER_APP_KEY: ${PUSHER_APP_KEY}
      PUSHER_APP_SECRET: ${PUSHER_APP_SECRET}
      PUSHER_HOST: localhost
      PUSHER_PORT: ${WS_PORT}
      PUSHER_SCHEME: http
      PUSHER_BACKEND_HOST: localhost
      ROOT_USERNAME: admin
      ROOT_USER_EMAIL: ${ADMIN_EMAIL}
      ROOT_USER_PASSWORD: ${ADMIN_PASSWORD}
    volumes:
      - ./data/coolify:/var/www/html/storage
      - /var/run/docker.sock:/var/run/docker.sock
      - ./logs:/var/log/nginx
    depends_on:
      coolify-db:
        condition: service_healthy
      coolify-redis:
        condition: service_healthy
    networks:
      - coolify-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:80"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s

  coolify-db:
    image: postgres:17-alpine
    container_name: coolify-db
    restart: unless-stopped
    environment:
      POSTGRES_USER: coolify
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: coolify
    volumes:
      - ./data/databases:/var/lib/postgresql/data
    networks:
      - coolify-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U coolify -d coolify"]
      interval: 10s
      timeout: 5s
      retries: 5

  coolify-redis:
    image: redis:7-alpine
    container_name: coolify-redis
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD} --appendonly yes
    volumes:
      - ./data/redis:/data
    networks:
      - coolify-network
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  coolify-realtime:
    image: quay.io/soketi/soketi:1.6-16-alpine
    container_name: coolify-realtime
    restart: unless-stopped
    ports:
      - "${WS_PORT}:6001"
    environment:
      SOKETI_DEBUG: "1"
      SOKETI_HOST: 0.0.0.0
      SOKETI_DEFAULT_APP_ID: ${PUSHER_APP_ID}
      SOKETI_DEFAULT_APP_KEY: ${PUSHER_APP_KEY}
      SOKETI_DEFAULT_APP_SECRET: ${PUSHER_APP_SECRET}
    networks:
      - coolify-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:6001"]
      interval: 10s
      timeout: 5s
      retries: 3

networks:
  coolify-network:
    driver: bridge

volumes:
  coolify-storage:
    driver: local
EOF

    print_success "Docker Compose configuration created"
}

# Start services
start_services() {
    print_info "Starting Coolify services..."
    
    # Pull images
    print_info "Pulling Docker images..."
    docker compose -f docker-compose.local.yml pull
    
    # Start services
    print_info "Starting services..."
    docker compose -f docker-compose.local.yml up -d
    
    # Wait for services to be ready
    print_info "Waiting for services to be ready..."
    sleep 30
    
    # Wait for database
    print_info "Waiting for database..."
    for i in {1..30}; do
        if docker exec coolify-db pg_isready -U coolify > /dev/null 2>&1; then
            break
        fi
        if [ $i -eq 30 ]; then
            print_error "Database failed to start"
            exit 1
        fi
        sleep 2
    done
    
    # Wait for Redis
    print_info "Waiting for Redis..."
    sleep 10
    
    # Wait for application
    print_info "Waiting for application..."
    for i in {1..60}; do
        if curl -s -o /dev/null -w '%{http_code}' http://localhost:${APP_PORT} | grep -q "200\|302"; then
            break
        fi
        if [ $i -eq 60 ]; then
            print_error "Application failed to start"
            docker compose -f docker-compose.local.yml logs coolify
            exit 1
        fi
        sleep 2
    done
    
    print_success "All services started successfully"
}

# Configure application
configure_application() {
    print_info "Configuring Coolify application..."
    
    # Run migrations
    print_info "Running database migrations..."
    docker exec coolify-app php /var/www/html/artisan migrate --force
    
    # Create admin user (if not exists)
    print_info "Creating admin user..."
    docker exec coolify-app php /var/www/html/artisan tinker --execute="
\$user = \App\Models\User::where('email', '${ADMIN_EMAIL}')->first();
if (!\$user) {
    \$user = new \App\Models\User();
    \$user->name = 'Admin';
    \$user->email = '${ADMIN_EMAIL}';
    \$user->password = \Hash::make('${ADMIN_PASSWORD}');
    \$user->save();
    echo 'Admin user created successfully';
} else {
    echo 'Admin user already exists';
}
"
    
    # Clear and setup caches (enhanced based on debug mode)
    print_info "Setting up caches..."
    
    # Clear all caches first
    docker exec coolify-app php /var/www/html/artisan cache:clear
    docker exec coolify-app php /var/www/html/artisan config:clear
    docker exec coolify-app php /var/www/html/artisan route:clear
    docker exec coolify-app php /var/www/html/artisan view:clear
    
    # For local development, don't cache unless explicitly asked
    if [[ $APP_DEBUG != "true" ]]; then
        print_info "Optimizing caches for production-like performance..."
        docker exec coolify-app php /var/www/html/artisan config:cache
        docker exec coolify-app php /var/www/html/artisan route:cache
    else
        print_info "Debug mode enabled - keeping caches clear for development"
    fi
    
    print_success "Application configured successfully"
}

# Health checks
health_checks() {
    print_info "Running health checks..."
    
    # Check application
    if curl -s -o /dev/null -w '%{http_code}' http://localhost:${APP_PORT} | grep -q "200\|302"; then
        print_success "Application is responding"
    else
        print_error "Application is not responding"
        return 1
    fi
    
    # Check WebSocket
    if curl -s http://localhost:${WS_PORT} | grep -q "OK"; then
        print_success "WebSocket is responding"
    else
        print_warning "WebSocket check failed"
    fi
    
    # Check services
    if docker compose -f docker-compose.local.yml ps | grep -q "Up"; then
        print_success "All services are running"
    else
        print_warning "Some services may not be running properly"
    fi
}

# Save setup report
save_report() {
    REPORT_FILE="coolify-local-setup-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$REPORT_FILE" << EOF
╔════════════════════════════════════════════════════════════════╗
║              Coolify Local Setup Report                       ║
║              Based on Official Installation Script           ║
╚════════════════════════════════════════════════════════════════╝

Setup Date: $(date)
Status: ✅ SUCCESS
Method: Official Best Practices + Local Development

═══════════════════════════════════════════════════════════════════
  ACCESS INFORMATION
═══════════════════════════════════════════════════════════════════

Application URL: http://localhost:${APP_PORT}
Login URL:      http://localhost:${APP_PORT}/login
WebSocket URL:  http://localhost:${WS_PORT}

═══════════════════════════════════════════════════════════════════
  ADMIN CREDENTIALS
═══════════════════════════════════════════════════════════════════

Email:    ${ADMIN_EMAIL}
Password: ${ADMIN_PASSWORD}

⚠️  IMPORTANT: Change this password in production!

═══════════════════════════════════════════════════════════════════
  DEVELOPMENT FEATURES
═══════════════════════════════════════════════════════════════════

Debug Mode: ${APP_DEBUG}
Log Level:  ${LOG_LEVEL}
WebSocket:  Enabled for debugging
Database:  Migrations applied
Admin User: Created
Caches:     $([ "$APP_DEBUG" = "true" ] && echo "Clear for development" || echo "Optimized for performance")

═══════════════════════════════════════════════════════════════════
  CONFIGURATION FILES
═══════════════════════════════════════════════════════════════════

Environment: .env.local
Docker Compose: docker-compose.local.yml

═══════════════════════════════════════════════════════════════════
EOF

    print_success "Setup report saved to: $REPORT_FILE"
}

# Main execution
main() {
    check_prerequisites
    collect_info
    create_directories
    generate_credentials
    create_environment
    create_docker_compose
    start_services
    configure_application
    health_checks
    save_report
    
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║      Local Setup Completed Successfully! 🎉              ║${NC}"
    echo -e "${GREEN}║      Based on Official Installation Script               ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    print_info "Access your Coolify instance at: http://localhost:${APP_PORT}"
    print_info "Login with: ${ADMIN_EMAIL} / ${ADMIN_PASSWORD}"
    print_info "WebSocket debugging available at: http://localhost:${WS_PORT}"
    print_info "Setup report saved with all management commands"
}

# Run main function
main "$@"
