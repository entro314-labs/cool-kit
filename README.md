# Cool-Kit: Coolify CLI Toolkit

🚀 **The complete command-line toolkit for Coolify deployment and management.**

Cool-Kit provides a comprehensive suite of commands for deploying, managing, and operating Coolify instances across local development and cloud platforms.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.25%2B-blue.svg)
![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-lightgrey.svg)

---

## ✨ Features

- 🏠 **Local Development** - Complete local Coolify setup with Docker Compose
- ☁️ **Cloud Deployment** - Full Azure, AWS, GCP, DigitalOcean, and Hetzner provisioning features
- 🔄 **Smart Updates** - Automatic updates with backup and rollback
- 💾 **Backup Management** - Create, list, and restore backups
- 📊 **Status Monitoring** - Comprehensive health checks and resource monitoring
- 🔐 **Secure by Default** - Cryptographically secure credential generation
- 🌍 **Cross-Platform** - Works on Linux, macOS, and Windows
- ⚡ **Type-Safe** - Built with Go for reliability and performance

---

## 📋 Requirements

### System Requirements

- **OS**: Linux, macOS, or Windows
- **RAM**: 2GB minimum (4GB recommended for local, 8GB for Azure)
- **Disk**: 20GB available space
- **Go**: 1.25+ (for building from source)

### Prerequisites

#### For Local Development

- Docker 20.10+
- Docker Compose v2+

#### For Azure Deployment

- Authenticated Azure session (via `az login` or Service Principal)
- Azure account with subscription
- SSH key pair (`~/.ssh/id_rsa.pub`)

---

## 🚀 Quick Start

### Installation

#### Download Pre-built Binary (Recommended)

```bash
# Download latest release (coming soon)
curl -fsSL https://github.com/entro314-labs/cool-kit/releases/latest/download/cool-kit-linux-amd64 -o cool-kit
chmod +x cool-kit
sudo mv cool-kit /usr/local/bin/
```

#### Build from Source

```bash
# Clone repository
git clone https://github.com/entro314-labs/cool-kit.git
cd cool-kit

# Build with Go
go build -o cool-kit

# Install (optional)
sudo mv cool-kit /usr/local/bin/
```

### Basic Usage

```bash
# Show all available commands
cool-kit --help

# Setup local Coolify development instance
cool-kit local setup

# Deploy Coolify to Azure
cool-kit azure deploy

# Check status of Azure instance
cool-kit azure status
```

---

## 🏠 Local Development Commands

Manage Coolify instances on your local machine using Docker Compose.

### `cool-kit local setup`

Interactive setup for local Coolify development:

- Collects configuration (email, password, ports)
- Generates secure credentials automatically
- Creates required directory structure
- Generates Docker Compose configuration
- Starts all services with health checks
- Runs database migrations

```bash
cool-kit local setup
```

### `cool-kit local start`

Start local Coolify services:

```bash
cool-kit local start
```

### `cool-kit local stop`

Stop services (preserves all data):

```bash
cool-kit local stop
```

### `cool-kit local logs`

View service logs:

```bash
# Follow logs in real-time
cool-kit local logs -f

# Show last 100 lines
cool-kit local logs --tail 100

# Logs for specific service
cool-kit local logs coolify
```

### `cool-kit local update`

Update to latest Coolify version:

- Pulls latest Docker images
- Recreates containers
- Runs database migrations
- Verifies health

```bash
cool-kit local update
```

### `cool-kit local reset`

Reset to clean state (⚠️ **deletes all data**):

```bash
# Interactive confirmation
cool-kit local reset

# Skip confirmation
cool-kit local reset --force
```

---

## ☁️ Azure Deployment Commands

Deploy and manage Coolify instances on Microsoft Azure.

### `cool-kit azure deploy`

Complete Azure deployment workflow:

1. **Load Configuration** - Interactive or from file
2. **Provision Infrastructure**:
   - Resource group
   - Virtual network and subnet
   - Network security group (ports 22, 80, 443, 8000)
   - Public IP address
   - Network interface
   - Virtual machine (Ubuntu 22.04)
3. **Install Dependencies**:
   - System updates
   - Docker and Docker Compose
4. **Deploy Coolify**:
   - Container deployment
   - Database initialization
   - Admin user creation
5. **Verify Health** - All services running

```bash
cool-kit azure deploy
```

**What you'll be asked:**

- Resource Group name (default: `coolify-rg`)
- VM name (default: `coolify-vm`)
- Admin email
- Admin password

**What you get:**

- Fully provisioned Azure VM
- Coolify running on `http://<public-ip>:8000`
- SSH access via `ssh azureuser@<public-ip>`

### `cool-kit azure update`

Update existing Azure instance with automatic rollback:

- Creates pre-update backup
- Pulls latest images
- Recreates containers
- Runs migrations
- Verifies health
- Automatic rollback on failure (optional)

```bash
cool-kit azure update
```

### `cool-kit azure status`

Comprehensive health check:

- ✅ VM status (power state, provisioning)
- ✅ SSH connectivity
- ✅ Service health (Coolify, DB, Redis, Realtime)
- ✅ Resource usage (CPU, memory, disk)
- ✅ Coolify version

```bash
cool-kit azure status
```

### `cool-kit azure backup`

Create full backup:

- PostgreSQL database dump
- Redis data
- Docker volumes (applications, services, etc.)
- Configuration files
- Metadata with timestamp

```bash
cool-kit azure backup
```

### `cool-kit azure rollback`

Restore from backup:

- Lists available backups
- Interactive selection
- Complete restoration
- Health verification

```bash
cool-kit azure rollback
```

### `cool-kit azure config`

Manage Azure configuration:

1. View current configuration
2. Edit configuration file
3. Validate configuration
4. Reset to defaults
5. Show config file path

```bash
cool-kit azure config
```

**Configuration file**: `~/.coolify/azure-config.json`

### `cool-kit azure ssh`

Interactive SSH access to Azure VM:

```bash
cool-kit azure ssh
```

---

## ⚙️ Configuration

### Azure Configuration

Default configuration is created at `~/.coolify/azure-config.json`:

```json
{
  "infrastructure": {
    "location": "swedencentral",
    "vm_size": "Standard_B2s",
    "admin_username": "azureuser",
    "ssh_public_key_path": "~/.ssh/id_rsa.pub",
    "os_image": "Canonical:0001-com-ubuntu-server-jammy:22_04-lts-gen2:latest",
    "os_disk_size_gb": 30
  },
  "networking": {
    "app_port": 80,
    "ssh_port": 22,
    "websocket_port": 6001,
    "vnet_address_prefix": "10.0.0.0/16",
    "subnet_address_prefix": "10.0.1.0/24"
  },
  "coolify": {
    "default_admin_email": "admin@coolify.local",
    "app_url_template": "http://{public_ip}",
    "pusher_host_template": "{public_ip}",
    "pusher_port": 6001
  }
}
```

Customize before deployment:

```bash
# Edit configuration
cool-kit azure config
# Choose option 2 to edit, then option 3 to validate
```

### Local Configuration

Local setup is interactive - you'll be prompted for:

- Admin email
- Admin password (or auto-generated)
- App port (default: 8000)
- Debug mode (default: false)

Configuration is saved to `~/.coolify/local-config.json`

---

## 🔐 Security

### Credentials

- All passwords generated with `crypto/rand` (cryptographically secure)
- Minimum 32 characters for sensitive credentials
- SSH key-based authentication for Azure
- No credentials stored in plaintext (except config files you create)

### Azure Security

- Network Security Group with minimal required ports
- SSH key authentication (no password login)
- Automatic system updates on first boot
- Secure credential generation for all services

### Best Practices

- ✅ Always use strong passwords
- ✅ Keep SSH keys secure
- ✅ Regularly update Coolify (`cool-kit azure update`)
- ✅ Regular backups (`cool-kit azure backup`)
- ✅ Monitor with `cool-kit azure status`

---

## 📊 Architecture

### Local Development

```
Docker Compose Stack:
├── coolify (app)
├── coolify-db (PostgreSQL 15)
├── coolify-redis (Redis 7)
└── coolify-soketi (WebSocket)
```

### Azure Deployment

```
Azure Resources:
├── Resource Group
├── Virtual Network (10.0.0.0/16)
├── Subnet (10.0.1.0/24)
├── Network Security Group
│   ├── SSH (22)
│   ├── HTTP (80)
│   ├── HTTPS (443)
│   └── Coolify (8000)
├── Public IP Address
├── Network Interface
└── Virtual Machine (Ubuntu 22.04)
    └── Docker Stack:
        ├── coolify
        ├── coolify-db
        ├── coolify-redis
        └── coolify-realtime
```

---

## 🐛 Troubleshooting

### Local Development

**Docker not running:**

```bash
# Start Docker Desktop (macOS/Windows)
# Or start Docker daemon (Linux):
sudo systemctl start docker
```

**Port 8000 already in use:**

```bash
# Find what's using it
lsof -i :8000

# Or choose different port during setup
```

**Services not healthy:**

```bash
# Check logs
cool-kit local logs -f

# Reset and try again
cool-kit local reset --force
cool-kit local setup
```

### Azure Deployment

**Not logged into Azure:**

```bash
`az login` works best for local development. For CI/CD, use Service Principals.
```

**SSH connection fails:**

```bash
# Verify SSH key exists
ls -la ~/.ssh/id_rsa.pub

# Generate if missing
ssh-keygen -t rsa -b 4096
```

**Deployment fails:**

- Check Azure CLI authentication: `az account show`
- Verify subscription has quota
- Check network connectivity
- Review error messages

---

## 🛣️ Roadmap

- [x] AWS deployment support
- [x] Google Cloud deployment support
- [x] DigitalOcean deployment support
- [x] Hetzner deployment support
- [ ] Kubernetes deployment support
- [ ] Monitoring stack integration
- [ ] Service template deployment
- [ ] Multi-region deployment
- [ ] Automated SSL certificate management

---

## 📚 Documentation

- [Migration Complete Summary](AZURE_IMPLEMENTATION_COMPLETE.md) - Full implementation details
- [Script Migration Analysis](SCRIPT_MIGRATION_ANALYSIS.md) - Original bash script analysis
- [Migration Progress](MIGRATION_PROGRESS.md) - Detailed progress tracking

---

## 🤝 Contributing

We welcome contributions!

### Development Setup

```bash
git clone https://github.com/entro314-labs/cool-kit.git
cd cool-kit

# Install dependencies
go mod tidy

# Build
go build

# Run tests
go test ./...

# Format code
go fmt ./...
```

### Code Structure

```
cool-kit/
├── cmd/                    # Command definitions
│   ├── azure.go           # Azure commands
│   ├── local.go           # Local commands
│   └── root.go            # Root command
├── internal/
│   ├── api/               # Coolify API client
│   ├── azure/             # Azure deployment logic
│   ├── azureconfig/       # Azure configuration
│   ├── local/             # Local deployment logic
│   ├── smart/             # Smart service detection
│   └── ui/                # UI components
└── main.go                # Entry point
```

---

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 🔗 Links

- **Coolify Main Repository**: https://github.com/coollabsio/coolify
- **Documentation**: https://coolify.io/docs
- **Community**: https://coolify.io/discord
- **Issues**: https://github.com/coollabsio/coolify/issues

---

## ❤️ Support Coolify

Coolify is an open-source project. If you use **Hetzner Cloud** for your deployments, please consider using our referral link. It helps support the project at no extra cost to you!

- **Sign up for Hetzner**: [https://coolify.io/hetzner](https://coolify.io/hetzner)
- **Direct Link**: [Hetzner Console](https://console.hetzner.com/refer?pk_campaign=referral-invite&pk_medium=referral-program&pk_source=reflink&pk_content=VBVO47VycYLt)

---

## 🎉 Acknowledgments

Cool-Kit is the result of migrating **3,397 lines of bash scripts** to **~3,800 lines of type-safe Go code**, providing:

- ✅ Cross-platform compatibility
- ✅ Better error handling
- ✅ Interactive user experience
- ✅ Complete Azure integration
- ✅ Production-ready deployment workflows

**Migration completed**: 2025-12-27 (100% complete)

---

<p align="center">
  Made with ❤️ by the Coolify Community
</p>
