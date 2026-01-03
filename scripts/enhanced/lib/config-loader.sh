#!/bin/bash

# Configuration Loader for Coolify Scripts
# Loads dynamic configuration from JSON files

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Default configuration paths
AZURE_CONFIG_FILE="$PROJECT_ROOT/scripts/config/azure-config.json"
LOCAL_CONFIG_FILE="$PROJECT_ROOT/scripts/config/local-config.json"
PROD_CONFIG_FILE="$PROJECT_ROOT/scripts/config/prod-config.json"

# Load Azure configuration
load_azure_config() {
    local config_file="${1:-$AZURE_CONFIG_FILE}"
    
    if [ ! -f "$config_file" ]; then
        echo "ERROR: Azure configuration file not found: $config_file" >&2
        return 1
    fi
    
    # Parse JSON and export variables
    eval "$(jq -r '
        .infrastructure | to_entries[] | 
        "AZURE_\(.key | ascii_upcase)=\(.value | @sh)"
    ' "$config_file")"
    
    eval "$(jq -r '
        .networking | to_entries[] | 
        "NETWORK_\(.key | ascii_upcase)=\(.value | @sh)"
    ' "$config_file")"
    
    eval "$(jq -r '
        .coolify | to_entries[] | 
        "COOLIFY_\(.key | ascii_upcase)=\(.value | @sh)"
    ' "$config_file")"
    
    eval "$(jq -r '
        .paths | to_entries[] | 
        "REMOTE_\(.key | ascii_upcase)=\(.value | @sh)"
    ' "$config_file")"
    
    eval "$(jq -r '
        .docker | to_entries[] | 
        "DOCKER_\(.key | ascii_upcase)=\(.value | @sh)"
    ' "$config_file")"
    
    export $(jq -r 'paths | to_entries[] | "REMOTE_\(.key | ascii_upcase)"' "$config_file")
    export $(jq -r 'infrastructure | to_entries[] | "AZURE_\(.key | ascii_upcase)"' "$config_file")
    export $(jq -r 'networking | to_entries[] | "NETWORK_\(.key | ascii_upcase)"' "$config_file")
    export $(jq -r 'coolify | to_entries[] | "COOLIFY_\(.key | ascii_upcase)"' "$config_file")
    export $(jq -r 'docker | to_entries[] | "DOCKER_\(.key | ascii_upcase)"' "$config_file")
}

# Load local configuration
load_local_config() {
    local config_file="${1:-$LOCAL_CONFIG_FILE}"
    
    if [ ! -f "$config_file" ]; then
        echo "WARNING: Local configuration file not found: $config_file" >&2
        return 0
    fi
    
    eval "$(jq -r '
        .local | to_entries[] | 
        "LOCAL_\(.key | ascii_upcase)=\(.value | @sh)"
    ' "$config_file")"
    
    eval "$(jq -r '
        .docker | to_entries[] | 
        "DOCKER_\(.key | ascii_upcase)=\(.value | @sh)"
    ' "$config_file")"
    
    export $(jq -r 'local | to_entries[] | "LOCAL_\(.key | ascii_upcase)"' "$config_file")
    export $(jq -r 'docker | to_entries[] | "DOCKER_\(.key | ascii_upcase)"' "$config_file")
}

# Load production configuration
load_prod_config() {
    local config_file="${1:-$PROD_CONFIG_FILE}"
    
    if [ ! -f "$config_file" ]; then
        echo "WARNING: Production configuration file not found: $config_file" >&2
        return 0
    fi
    
    eval "$(jq -r '
        .production | to_entries[] | 
        "PROD_\(.key | ascii_upcase)=\(.value | @sh)"
    ' "$config_file")"
    
    eval "$(jq -r '
        .docker | to_entries[] | 
        "DOCKER_\(.key | ascii_upcase)=\(.value | @sh)"
    ' "$config_file")"
    
    export $(jq -r 'production | to_entries[] | "PROD_\(.key | ascii_upcase)"' "$config_file")
    export $(jq -r 'docker | to_entries[] | "DOCKER_\(.key | ascii_upcase)"' "$config_file")
}

# Get dynamic value with fallback
get_config_value() {
    local key="$1"
    local default="$2"
    local config_file="$3"
    
    if [ -f "$config_file" ]; then
        local value=$(jq -r ".$key // \"$default\"" "$config_file")
        echo "$value"
    else
        echo "$default"
    fi
}

# Update configuration value
update_config_value() {
    local key="$1"
    local value="$2"
    local config_file="$3"
    
    if [ -f "$config_file" ]; then
        jq ".$key = \"$value\"" "$config_file" > "${config_file}.tmp" && \
        mv "${config_file}.tmp" "$config_file"
    fi
}

# Validate required configuration
validate_config() {
    local config_file="$1"
    local required_keys="$2"
    
    if [ ! -f "$config_file" ]; then
        echo "ERROR: Configuration file not found: $config_file" >&2
        return 1
    fi
    
    for key in $required_keys; do
        if ! jq -e ".$key" "$config_file" > /dev/null 2>&1; then
            echo "ERROR: Required configuration key missing: $key" >&2
            return 1
        fi
    done
    
    return 0
}

# Export configuration for other scripts
export_config() {
    local config_type="$1"
    
    case "$config_type" in
        "azure")
            load_azure_config
            ;;
        "local")
            load_local_config
            ;;
        "prod")
            load_prod_config
            ;;
        *)
            echo "ERROR: Unknown configuration type: $config_type" >&2
            return 1
            ;;
    esac
}

# Show current configuration
show_config() {
    local config_file="$1"
    
    if [ -f "$config_file" ]; then
        echo "Current configuration from: $config_file"
        jq '.' "$config_file"
    else
        echo "Configuration file not found: $config_file"
    fi
}
