# Coolify CLI - Quick Start Guide

## 🚀 **The Vision**

Install → Run → Choose → Deploy → Access Dashboard

**That's it!** The CLI handles everything else automatically.

---

## 📦 **Installation**

### **Option 1: Homebrew (macOS/Linux)**

```bash
brew install --cask entro314-labs/tap/cool-kit
```

### **Option 2: Winget (Windows)**

```powershell
winget install Entro314Labs.CoolKit
```

### **Option 3: AUR (Arch Linux)**

```bash
yay -S cool-kit-bin
```

### **Option 4: Download Binary**

```bash
# Download latest release
curl -fsSL https://github.com/entro314-labs/cool-kit/releases/latest/download/cool-kit_linux_x86_64.tar.gz -o cool-kit.tar.gz
tar -xzf cool-kit.tar.gz
chmod +x cool-kit
sudo mv cool-kit /usr/local/bin/

# Or download specific version
wget https://github.com/entro314-labs/cool-kit/releases/download/v1.0.0/cool-kit_linux_x86_64.tar.gz
tar -xzf cool-kit_linux_x86_64.tar.gz
chmod +x cool-kit
sudo mv cool-kit /usr/local/bin/cool-kit
```

### **Option 5: Build from Source**

```bash
# Clone repository
# Clone repository
git clone https://github.com/entro314-labs/cool-kit.git
cd cool-kit

# Install dependencies
go mod download

# Build
go build -o cool-kit main.go

# Install
sudo mv cool-kit /usr/local/bin/
```

---

## 🎯 **Usage - The Complete Flow**

### **Step 1: Run the CLI**

```bash
cool-kit
```

### **Step 2: Beautiful TUI Interface**

You'll see a sexy terminal interface:

```
╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║                    🚀 Coolify CLI Installer                    ║
║                                                                ║
║              Deploy Coolify Anywhere, Effortlessly             ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝

Choose your deployment target:

👉 ☁️  Azure
   Deploy to Microsoft Azure cloud

   🔶 AWS
   Deploy to Amazon Web Services

   🔵 Google Cloud
   Deploy to Google Cloud Platform

   🖥️  Bare Metal / VM
   Deploy to any Linux server via SSH

   🐳 Docker / Docker Compose
   Deploy locally with Docker

↑/↓ - Navigate  |  Enter - Select  |  Esc - Quit
```

### **Step 3: Automatic Deployment**

Once you select a provider, the CLI automatically:

1. **Downloads latest Coolify** from GitHub (main branch)
2. **Provisions infrastructure** (if cloud provider)
3. **Installs Docker** (if needed)
4. **Configures environment** with secure credentials
5. **Deploys Coolify** with all services
6. **Runs health checks** to ensure everything works
7. **Shows you the dashboard URL**

You'll see real-time progress:

```
🚀 Deploying Coolify to AWS

Overall Progress: 6/10 steps
████████████████████████░░░░░░░░ 60%

╭─────────────────────────────────────────╮
│ ✓ Validate AWS credentials              │
│ ✓ Clone Coolify repository               │
│ ✓ Create VPC and subnets                 │
│ ✓ Configure security groups              │
│ ✓ Launch EC2 instance                    │
│ ◐ Assign Elastic IP             [████░] │
│ ○ Install Docker                         │
│ ○ Deploy Coolify                         │
│ ○ Configure SSL                          │
│ ○ Run health checks                      │
╰─────────────────────────────────────────╯

📋 Recent Activity:
[16:23:45] ℹ Allocating Elastic IP
[16:23:47] • IP Address: 54.123.45.67
[16:23:50] ✓ IP associated with instance
```

### **Step 4: Access Your Dashboard**

When deployment completes:

```
╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║                   🎉 Deployment Successful! 🎉                 ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝

📊 Deployment Summary:
   Provider:     aws
   Duration:     5m 32s

🌐 Access Your Coolify Dashboard:
   → http://54.123.45.67:8000

📝 Next Steps:
   1. Open the dashboard URL in your browser
   2. Complete the initial setup wizard
   3. Start deploying your applications!

💡 Quick Commands:
   cool-kit status   - Check deployment status
   cool-kit update   - Update to latest Coolify version
   cool-kit --help   - View all available commands

✨ Happy deploying with Coolify! ✨
```

---

## 🎨 **Provider-Specific Examples**

### **Azure**

```bash
cool-kit

# Select "Azure" from menu
# CLI automatically:
# - Validates Azure credentials (az login required)
# - Creates resource group in Sweden Central
# - Provisions VM with Docker
# - Deploys Coolify
# - Shows: http://your-azure-ip:8000
```

### **AWS**

```bash
cool-kit

# Select "AWS" from menu
# CLI automatically:
# - Validates AWS credentials (aws configure required)
# - Creates VPC and security groups
# - Launches EC2 instance
# - Deploys Coolify
# - Shows: http://your-aws-ip:8000
```

### **Google Cloud**

```bash
cool-kit

# Select "Google Cloud" from menu
# CLI automatically:
# - Validates GCP credentials (gcloud auth login required)
# - Creates VPC network
# - Launches Compute Engine instance
# - Deploys Coolify
# - Shows: http://your-gcp-ip:8000
```

### **Bare Metal / VM**

```bash
cool-kit

# Select "Bare Metal / VM" from menu
# Enter your server details:
# - Host: 192.168.1.100
# - User: ubuntu
# - SSH Key: ~/.ssh/id_rsa

# CLI automatically:
# - Connects via SSH
# - Installs Docker
# - Deploys Coolify
# - Shows: http://192.168.1.100:8000
```

### **Docker (Local)**

```bash
cool-kit

# Select "Docker / Docker Compose" from menu
# CLI automatically:
# - Validates Docker installation
# - Clones Coolify repository
# - Generates secure credentials
# - Starts all services
# - Shows: http://localhost:8000
```

---

## 🔧 **Prerequisites**

### **All Deployments**

- Internet connection
- Terminal access

### **Cloud Providers**

- **Azure**: Authenticated session (run `az login` or Set `AZURE_AUTH_LOCATION`)
- **AWS**: AWS CLI installed + `aws configure`
- **GCP**: gcloud CLI installed + `gcloud auth login`

### **Bare Metal / VM**

- SSH access to target server
- Ubuntu/Debian/CentOS/RHEL server

### **Docker (Local)**

- Docker installed and running
- Docker Compose installed

---

## 💡 **Key Features**

### **✅ Hands-Off Deployment**

- No manual configuration needed
- Automatic credential generation
- Smart defaults for everything
- One command to deploy

### **✅ Latest Coolify Always**

- Pulls from official GitHub repository
- Uses main/master branch
- Always up-to-date
- No stale versions

### **✅ Beautiful Interface**

- Real-time progress tracking
- Color-coded logs
- Step-by-step visualization
- Professional appearance

### **✅ Multi-Cloud Support**

- Azure, AWS, GCP
- Bare metal servers
- Local Docker
- Consistent experience everywhere

### **✅ Production Ready**

- Secure credential generation
- Health checks and validation
- Proper error handling
- Rollback capabilities

---

## 📊 **What Gets Deployed**

The CLI deploys the complete Coolify stack:

### **Services**

- **Coolify App** - Main application (port 8000)
- **PostgreSQL** - Database (port 5432)
- **Redis** - Cache and queues (port 6379)
- **Soketi** - WebSocket server (port 6001)
- **Queue Worker** - Background jobs

### **Configuration**

- Secure APP_KEY generated
- Random database passwords
- Redis authentication
- WebSocket configuration
- All environment variables

### **Networking**

- Proper port exposure
- Health checks configured
- Service dependencies
- Internal network

---

## 🎯 **After Deployment**

### **1. Access Dashboard**

Open the provided URL in your browser:

```
http://your-server-ip:8000
```

### **2. Initial Setup**

Complete the Coolify setup wizard:

- Create admin account
- Configure settings
- Add SSH keys
- Connect servers

### **3. Start Deploying**

Use Coolify to deploy your applications:

- Connect Git repositories
- Deploy Docker containers
- Manage databases
- Configure domains

---

## 🔄 **Updating Coolify**

```bash
# Update to latest version
cool-kit update

# Check current status
cool-kit status

# View version info
cool-kit version
```

---

## 🆘 **Troubleshooting**

### **Deployment Failed**

```bash
# Check logs
cool-kit logs

# Retry deployment
cool-kit deploy --retry

# Clean and redeploy
cool-kit clean
cool-kit deploy
```

### **Can't Access Dashboard**

```bash
# Check service status
cool-kit status

# Restart services
cool-kit restart

# View health checks
cool-kit health
```

### **Need Help**

```bash
# View all commands
cool-kit --help

# Get provider-specific help
cool-kit azure --help
cool-kit aws --help
```

---

## 🎉 **That's It!**

The Coolify CLI makes deploying Coolify as simple as:

1. **Install** the CLI
2. **Run** `cool-kit`
3. **Choose** your provider
4. **Wait** for automatic deployment
5. **Access** your dashboard

**No configuration files. No manual steps. Just works.** ✨

---

## 📚 **Additional Resources**

- **Official Docs**: https://coolify.io/docs
- **GitHub**: https://github.com/coollabsio/coolify
- **Discord**: https://discord.gg/coolify
- **CLI Issues**: https://github.com/entro314-labs/cool-kit/issues

---

## 🙏 **Credits**

Built with ❤️ for the Coolify community

- **Coolify**: https://coolify.io
- **CLI Framework**: Cobra + BubbleTea
- **Cloud SDKs**: Azure, AWS, GCP official CLIs
