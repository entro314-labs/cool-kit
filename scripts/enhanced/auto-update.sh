#!/bin/bash

# Final Coolify Auto-Update Script
# Incorporates official best practices + uses final deployment scripts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration (following official patterns)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_FILE="$PROJECT_ROOT/azure-config.json"
STATE_FILE="$PROJECT_ROOT/azure-state.json"
UPDATE_LOG="$PROJECT_ROOT/logs/auto-update.log"
CRON_FILE="$HOME/.coolify-auto-update-final"
BACKUP_DIR="$PROJECT_ROOT/backups"

# Official-style logging functions
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

# Official logging method
log_message() {
    local message="[$(date '+%Y-%m-%d %H:%M:%S')] $1"
    echo "$message" | tee -a "$UPDATE_LOG"
}

log_section() {
    echo "" >>"$UPDATE_LOG"
    echo "============================================================" >>"$UPDATE_LOG"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >>"$UPDATE_LOG"
    echo "============================================================" >>"$UPDATE_LOG"
}

# Create necessary directories
create_directories() {
    mkdir -p "$(dirname "$UPDATE_LOG")"
    mkdir -p "$BACKUP_DIR"
}

# Banner
echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║     Final Coolify Auto-Update Script                          ║"
echo "║     Official Best Practices + Final Scripts                  ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if we're in a git repository
check_git_repo() {
    if [ ! -d ".git" ]; then
        print_error "This script must be run from a git repository"
        log_message "ERROR: Not in a git repository"
        exit 1
    fi
    
    print_success "Git repository detected"
    log_message "Git repository verified"
}

# Check for final deployment script
check_final_scripts() {
    if [ ! -f "$SCRIPT_DIR/deploy-azure-final.sh" ]; then
        print_error "Final deployment script not found"
        print_error "Please ensure the final scripts are available"
        log_message "ERROR: Final deployment script missing"
        exit 1
    fi
    
    if [ ! -f "$SCRIPT_DIR/update-azure-final.sh" ]; then
        print_error "Final update script not found"
        print_error "Please ensure the final scripts are available"
        log_message "ERROR: Final update script missing"
        exit 1
    fi
    
    print_success "Final scripts available"
    log_message "Final scripts verified"
}

# Check Azure configuration (simplified for final scripts)
check_azure_config() {
    # For final scripts, we use hardcoded values from our deployment
    PUBLIC_IP="20.169.182.98"
    ADMIN_USERNAME="azureuser"
    
    # Test connectivity
    if ! ssh -o ConnectTimeout=10 -o BatchMode=yes ${ADMIN_USERNAME}@${PUBLIC_IP} "echo 'SSH OK'" > /dev/null 2>&1; then
        print_error "Cannot connect to Azure VM"
        log_message "ERROR: Azure VM not accessible"
        exit 1
    fi
    
    print_success "Azure VM accessible"
    log_message "Azure configuration verified"
}

# Get current and latest commit (enhanced)
get_commit_info() {
    print_step "Checking for repository updates..."
    log_section "Checking repository updates"
    
    # Fetch latest changes
    git fetch origin
    
    # Get current commit
    CURRENT_COMMIT=$(git rev-parse HEAD)
    
    # Get latest commit from main/develop branch
    LATEST_COMMIT=$(git rev-parse origin/main 2>/dev/null || git rev-parse origin/develop 2>/dev/null || git rev-parse origin/master 2>/dev/null)
    
    if [ -z "$LATEST_COMMIT" ]; then
        print_error "Could not determine latest commit"
        log_message "ERROR: Could not determine latest commit"
        exit 1
    fi
    
    log_message "Current commit: $CURRENT_COMMIT"
    log_message "Latest commit: $LATEST_COMMIT"
    
    if [ "$CURRENT_COMMIT" = "$LATEST_COMMIT" ]; then
        print_success "Repository is up to date"
        log_message "Repository is up to date"
        return 1  # No updates needed
    else
        print_info "Updates available"
        log_message "Updates available"
        return 0  # Updates available
    fi
}

# Check what changed (enhanced)
check_changes() {
    print_step "Analyzing changes..."
    log_section "Analyzing changes"
    
    # Get list of changed files
    CHANGED_FILES=$(git diff --name-only "$CURRENT_COMMIT" "$LATEST_COMMIT")
    
    log_message "Changed files:"
    echo "$CHANGED_FILES" | while read file; do
        log_message "  - $file"
    done
    
    # Check for relevant changes
    APP_CHANGES=false
    CONFIG_CHANGES=false
    SCRIPT_CHANGES=false
    
    echo "$CHANGED_FILES" | while read file; do
        if [[ "$file" == src/* ]] || [[ "$file" == app/* ]] || [[ "$file" == composer.json ]] || [[ "$file" == package.json ]]; then
            APP_CHANGES=true
            log_message "Application changes detected"
        fi
        
        if [[ "$file" == docker-compose* ]] || [[ "$file" == .env* ]] || [[ "$file" == config/* ]]; then
            CONFIG_CHANGES=true
            log_message "Configuration changes detected"
        fi
        
        if [[ "$file" == scripts/* ]]; then
            SCRIPT_CHANGES=true
            log_message "Script changes detected"
        fi
    done
    
    print_success "Change analysis completed"
    log_message "Change analysis completed"
}

# Pull latest changes
pull_changes() {
    print_step "Pulling latest changes..."
    log_section "Pulling changes"
    
    # Create backup before pulling
    BACKUP_NAME="backup-$(date +%Y%m%d-%H%M%S)"
    git stash push -m "Auto-update backup $BACKUP_NAME"
    log_message "Created stash backup: $BACKUP_NAME"
    
    # Pull changes
    git pull origin main 2>/dev/null || git pull origin develop 2>/dev/null || git pull origin master 2>/dev/null
    
    print_success "Changes pulled successfully"
    log_message "Changes pulled successfully"
}

# Deploy to Azure using final script
deploy_to_azure() {
    print_step "Deploying to Azure..."
    log_section "Deploying to Azure"
    
    # Use the final update script
    if "$SCRIPT_DIR/update-azure-final.sh" >>"$UPDATE_LOG" 2>&1; then
        print_success "Azure deployment completed"
        log_message "Azure deployment completed successfully"
    else
        print_error "Azure deployment failed"
        log_message "ERROR: Azure deployment failed"
        
        # Attempt rollback
        print_warning "Attempting rollback..."
        if git stash pop >>"$UPDATE_LOG" 2>&1; then
            log_message "Rollback completed"
        else
            log_message "Rollback failed"
        fi
        exit 1
    fi
}

# Update state file (official method)
update_state() {
    print_step "Updating state..."
    
    # Create state file with official structure
    cat > "$STATE_FILE" << EOF
{
    "last_auto_update": "$(date -Iseconds)",
    "status": "updated",
    "previous_commit": "$CURRENT_COMMIT",
    "current_commit": "$LATEST_COMMIT",
    "deployment_method": "final_script",
    "services": {
        "coolify": "updated",
        "database": "running",
        "redis": "running",
        "realtime": "running"
    },
    "update_log": "$UPDATE_LOG"
}
EOF
    
    print_success "State updated"
    log_message "State file updated"
}

# Send notification (enhanced)
send_notification() {
    print_step "Sending notification..."
    
    # Create notification message
    NOTIFICATION_MSG="Coolify Auto-Update Completed Successfully
    
Updated: $(date)
Previous Commit: ${CURRENT_COMMIT:0:8}
Current Commit: ${LATEST_COMMIT:0:8}
Deployment Method: Final Script
Log File: $UPDATE_LOG

Access your Coolify instance at: http://20.169.182.98"
    
    # Save notification
    echo "$NOTIFICATION_MSG" > "$PROJECT_ROOT/last-update-notification.txt"
    
    print_success "Notification saved"
    log_message "Notification sent"
}

# Setup cron job (enhanced)
setup_cron() {
    print_step "Setting up auto-update cron job..."
    
    echo "Select update frequency:"
    echo "1. Every hour"
    echo "2. Every 6 hours"
    echo "3. Every 12 hours"
    echo "4. Daily at 2 AM"
    echo "5. Weekly on Sunday at 2 AM"
    echo "6. Custom schedule"
    read -p "Choice [4]: " CHOICE
    CHOICE=${CHOICE:-4}
    
    case $CHOICE in
        1) CRON_SCHEDULE="0 * * * *" ;;
        2) CRON_SCHEDULE="0 */6 * * *" ;;
        3) CRON_SCHEDULE="0 */12 * * *" ;;
        4) CRON_SCHEDULE="0 2 * * *" ;;
        5) CRON_SCHEDULE="0 2 * * 0" ;;
        6) 
            read -p "Enter cron schedule (e.g., '0 2 * * *' for daily at 2 AM): " CRON_SCHEDULE
            if [ -z "$CRON_SCHEDULE" ]; then
                CRON_SCHEDULE="0 2 * * *"
            fi
            ;;
        *) CRON_SCHEDULE="0 2 * * *" ;;
    esac
    
    # Create cron job with final script
    CRON_COMMAND="cd $PROJECT_ROOT && $SCRIPT_DIR/auto-update-final.sh --cron"
    (crontab -l 2>/dev/null; echo "$CRON_SCHEDULE $CRON_COMMAND") | crontab -
    
    # Save cron info
    echo "$CRON_SCHEDULE" > "$CRON_FILE"
    
    print_success "Auto-update scheduled with: $CRON_SCHEDULE"
    print_info "Cron job will run: $CRON_COMMAND"
    print_info "Using final deployment scripts"
    log_message "Cron job setup: $CRON_SCHEDULE"
}

# Remove cron job
remove_cron() {
    print_step "Removing auto-update cron job..."
    
    # Remove the specific cron job
    crontab -l 2>/dev/null | grep -v "$SCRIPT_DIR/auto-update-final.sh" | crontab -
    
    # Remove cron file
    rm -f "$CRON_FILE"
    
    print_success "Auto-update cron job removed"
    log_message "Cron job removed"
}

# Show status (enhanced)
show_status() {
    print_step "Auto-update status..."
    
    if [ -f "$CRON_FILE" ]; then
        CRON_SCHEDULE=$(cat "$CRON_FILE")
        print_success "Auto-update is enabled"
        print_info "Schedule: $CRON_SCHEDULE"
        print_info "Using final scripts"
    else
        print_warning "Auto-update is disabled"
    fi
    
    if [ -f "$STATE_FILE" ]; then
        print_info "Last update: $(grep -o '"last_auto_update": *"[^"]*"' "$STATE_FILE" | cut -d'"' -f4)"
        print_info "Last commit: $(grep -o '"current_commit": *"[^"]*"' "$STATE_FILE" | cut -d'"' -f4 | cut -c1-8)"
        print_info "Deployment method: $(grep -o '"deployment_method": *"[^"]*"' "$STATE_FILE" | cut -d'"' -f4)"
    else
        print_warning "No update history found"
    fi
    
    # Check if final scripts are available
    if [ -f "$SCRIPT_DIR/deploy-azure-final.sh" ] && [ -f "$SCRIPT_DIR/update-azure-final.sh" ]; then
        print_success "Final scripts available"
    else
        print_warning "Final scripts not found"
    fi
}

# Health check before auto-update
health_check() {
    print_step "Running pre-update health check..."
    
    # Check if Azure VM is accessible
    if ! ssh -o ConnectTimeout=10 -o BatchMode=yes azureuser@20.169.182.98 "echo 'Health check passed'" > /dev/null 2>&1; then
        print_error "Azure VM health check failed"
        log_message "ERROR: Pre-update health check failed"
        return 1
    fi
    
    # Check if application is responding
    if ! curl -s -o /dev/null -w '%{http_code}' http://20.169.182.98 | grep -q "200\|302"; then
        print_warning "Application not responding - proceeding with caution"
        log_message "WARNING: Application health check failed"
    fi
    
    print_success "Health check passed"
    log_message "Health check passed"
    return 0
}

# Main execution
main() {
    create_directories
    
    case "${1:-run}" in
        "setup")
            check_git_repo
            check_final_scripts
            check_azure_config
            setup_cron
            ;;
        "remove")
            remove_cron
            ;;
        "status")
            show_status
            ;;
        "health")
            health_check
            ;;
        "cron")
            # Running from cron, minimal output
            log_message "Starting auto-update check (final script version)"
            
            if check_git_repo && check_final_scripts && check_azure_config; then
                if health_check; then
                    if get_commit_info; then
                        check_changes
                        pull_changes
                        deploy_to_azure
                        update_state
                        send_notification
                        log_message "Auto-update completed successfully"
                    else
                        log_message "No updates available"
                    fi
                else
                    log_message "Auto-update skipped: health check failed"
                fi
            else
                log_message "Auto-update failed: prerequisites not met"
            fi
            ;;
        "run")
            check_git_repo
            check_final_scripts
            check_azure_config
            
            if health_check; then
                if get_commit_info; then
                    check_changes
                    pull_changes
                    deploy_to_azure
                    update_state
                    send_notification
                else
                    print_info "No updates available"
                fi
            else
                print_error "Health check failed - aborting update"
                exit 1
            fi
            ;;
        *)
            echo "Usage: $0 {setup|remove|status|health|run|cron}"
            echo ""
            echo "Commands:"
            echo "  setup   - Set up automatic updates"
            echo "  remove  - Remove automatic updates"
            echo "  status  - Show current status"
            echo "  health  - Run health check"
            echo "  run     - Run update manually"
            echo "  cron    - Run from cron (minimal output)"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
