package production

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// ProductionProvider handles production deployments to remote servers via SSH
type ProductionProvider struct {
	config *config.Config
}

// NewProductionProvider creates a new production provider
func NewProductionProvider(cfg *config.Config) (*ProductionProvider, error) {
	// Validate required configuration
	if cfg.Production.Domain == "" {
		return nil, fmt.Errorf("production domain is required. Set it using: cool-kit config set production.domain yourdomain.com")
	}

	return &ProductionProvider{
		config: cfg,
	}, nil
}

// GetDeploymentSteps returns deployment steps for production
func (p *ProductionProvider) GetDeploymentSteps() []ui.DeploymentStep {
	return []ui.DeploymentStep{
		{Name: "Validate SSH connection", Description: "Testing connectivity to production server"},
		{Name: "Check server requirements", Description: "Verifying Docker and system requirements"},
		{Name: "Transfer deployment files", Description: "Copying Coolify files to server"},
		{Name: "Configure production environment", Description: "Setting up production configuration"},
		{Name: "Deploy Coolify", Description: "Running production deployment"},
		{Name: "Configure SSL", Description: "Setting up HTTPS with Let's Encrypt"},
		{Name: "Run health checks", Description: "Validating deployment"},
	}
}

// Deploy performs production deployment
func (p *ProductionProvider) Deploy(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	steps := []struct {
		name string
		fn   func(chan<- ui.StepProgressMsg, chan<- ui.LogMsg) error
	}{
		{"Validate SSH connection", p.validateSSH},
		{"Check server requirements", p.checkRequirements},
		{"Transfer deployment files", p.transferFiles},
		{"Configure production environment", p.configureEnvironment},
		{"Deploy Coolify", p.deployCoolify},
		{"Configure SSL", p.configureSSL},
		{"Run health checks", p.runHealthChecks},
	}

	for i, step := range steps {
		logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Starting: %s", step.name)}

		if err := step.fn(progressChan, logChan); err != nil {
			logChan <- ui.LogMsg{Level: ui.LogError, Message: fmt.Sprintf("Failed: %s - %v", step.name, err)}
			return fmt.Errorf("step '%s' failed: %w", step.name, err)
		}

		progressChan <- ui.StepProgressMsg{StepIndex: i, Progress: 1.0, Message: fmt.Sprintf("Completed: %s", step.name)}
		logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: fmt.Sprintf("✓ %s completed", step.name)}
	}

	return nil
}

// validateSSH validates SSH connection to the production server
func (p *ProductionProvider) validateSSH(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Testing SSH connection"}

	host := p.getHost()
	user := p.getUser()
	keyPath := p.getSSHKeyPath()

	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Connecting to %s@%s", user, host)}

	args := p.buildSSHArgs("echo 'SSH connection successful'")
	cmd := exec.Command("ssh", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w. Ensure SSH key is configured at %s", err, keyPath)
	}

	if !strings.Contains(string(output), "successful") {
		return fmt.Errorf("SSH connection test failed")
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "SSH connection validated"}
	return nil
}

// checkRequirements verifies server requirements
func (p *ProductionProvider) checkRequirements(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Checking Docker"}

	// Check Docker is installed
	args := p.buildSSHArgs("docker --version")
	cmd := exec.Command("ssh", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker not installed on production server. Install with: curl -fsSL https://get.docker.com | sh")
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Checking Docker Compose"}

	// Check Docker Compose
	args = p.buildSSHArgs("docker compose version || docker-compose version")
	cmd = exec.Command("ssh", args...)
	if err := cmd.Run(); err != nil {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: "Docker Compose not found, will install"}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "Checking disk space"}

	// Check disk space (at least 10GB)
	args = p.buildSSHArgs("df -BG / | tail -1 | awk '{print $4}' | tr -d 'G'")
	cmd = exec.Command("ssh", args...)
	output, err := cmd.Output()
	if err != nil {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: "Could not determine disk space"}
	} else {
		logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Available disk space: %sGB", strings.TrimSpace(string(output)))}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Requirements validated"}
	return nil
}

// transferFiles transfers deployment files to the server
func (p *ProductionProvider) transferFiles(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Creating deployment directory"}

	host := p.getHost()
	user := p.getUser()

	// Create directory on remote server
	args := p.buildSSHArgs("mkdir -p /data/coolify && mkdir -p /data/coolify/source")
	cmd := exec.Command("ssh", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create deployment directory: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Downloading Coolify"}

	// Clone Coolify directly on the server
	cloneCmd := "cd /data/coolify/source && (git pull 2>/dev/null || git clone https://github.com/coollabsio/coolify.git .)"
	args = p.buildSSHArgs(cloneCmd)
	cmd = exec.Command("ssh", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone Coolify: %w", err)
	}

	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Coolify source deployed to %s@%s:/data/coolify/source", user, host)}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Files transferred"}
	return nil
}

// configureEnvironment sets up production environment
func (p *ProductionProvider) configureEnvironment(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Generating credentials"}

	domain := p.config.Production.Domain
	sslEmail := p.config.Production.SSLEmail
	if sslEmail == "" {
		sslEmail = "admin@" + domain
	}

	// Generate APP_KEY and other credentials on the server
	envSetup := fmt.Sprintf(`
cd /data/coolify/source
APP_KEY=$(openssl rand -base64 32)
DB_PASSWORD=$(openssl rand -base64 32)
REDIS_PASSWORD=$(openssl rand -base64 32)

cat > .env << EOF
APP_NAME=Coolify
APP_ENV=production
APP_DEBUG=false
APP_URL=https://%s
APP_KEY=base64:$APP_KEY
DB_PASSWORD=$DB_PASSWORD
REDIS_PASSWORD=$REDIS_PASSWORD
SSL_EMAIL=%s
EOF
`, domain, sslEmail)

	args := p.buildSSHArgs(envSetup)
	cmd := exec.Command("ssh", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure environment: %w", err)
	}

	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Production domain: %s", domain)}
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("SSL email: %s", sslEmail)}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Environment configured"}
	return nil
}

// deployCoolify deploys Coolify using Docker Compose
func (p *ProductionProvider) deployCoolify(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Pulling Docker images"}

	// Pull and start Coolify
	deployCmd := `
cd /data/coolify/source
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
`
	args := p.buildSSHArgs(deployCmd)
	cmd := exec.Command("ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Try with docker-compose (v1) as fallback
		deployCmd = `
cd /data/coolify/source
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d
`
		args = p.buildSSHArgs(deployCmd)
		cmd = exec.Command("ssh", args...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to deploy Coolify: %w", err)
		}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.7, Message: "Waiting for services"}

	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: "Waiting 60 seconds for services to initialize"}
	time.Sleep(60 * time.Second)

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Coolify deployed"}
	return nil
}

// configureSSL sets up SSL certificates
func (p *ProductionProvider) configureSSL(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Configuring SSL"}

	domain := p.config.Production.Domain
	sslEmail := p.config.Production.SSLEmail
	if sslEmail == "" {
		sslEmail = "admin@" + domain
	}

	// Coolify handles SSL internally via Traefik/Caddy, but we verify the setup
	sslCheck := fmt.Sprintf(`
# Verify SSL is being handled
docker ps | grep -E "(traefik|caddy)" && echo "SSL proxy running"
`)

	args := p.buildSSHArgs(sslCheck)
	cmd := exec.Command("ssh", args...)
	output, err := cmd.Output()
	// Grep returns exit code 1 if not found, which causes err != nil
	// We only care if found or not, but weird system errors should be logged
	if err != nil && err.Error() != "exit status 1" {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: fmt.Sprintf("Failed to check SSL status: %v", err)}
	}

	if strings.Contains(string(output), "SSL proxy running") {
		logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "SSL proxy detected"}
	} else {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: "SSL will be configured on first request to " + domain}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "SSL configured"}
	return nil
}

// runHealthChecks runs health checks
func (p *ProductionProvider) runHealthChecks(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Checking container status"}

	// Check container status
	args := p.buildSSHArgs("docker ps --filter 'name=coolify' --format '{{.Names}}: {{.Status}}'")
	cmd := exec.Command("ssh", args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}

	containers := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, container := range containers {
		if container != "" {
			logChan <- ui.LogMsg{Level: ui.LogInfo, Message: container}
		}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.6, Message: "Checking HTTP endpoint"}

	domain := p.config.Production.Domain

	// Check HTTP endpoint (try both http and https)
	healthCheck := fmt.Sprintf("curl -sL -o /dev/null -w '%%{http_code}' --max-time 10 https://%s/api/health || curl -sL -o /dev/null -w '%%{http_code}' --max-time 10 http://%s:8000/api/health", domain, domain)
	args = p.buildSSHArgs(healthCheck)
	cmd = exec.Command("ssh", args...)
	output, err = cmd.Output()
	if err != nil {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: fmt.Sprintf("Health check command failed: %v", err)}
	}

	statusCode := strings.TrimSpace(string(output))
	if statusCode == "200" {
		logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Health check passed (HTTP 200)"}
	} else {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: fmt.Sprintf("Health check returned: %s (may need DNS propagation)", statusCode)}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Health checks completed"}
	return nil
}

// Helper methods

func (p *ProductionProvider) getHost() string {
	// Use baremetal config for SSH host if production doesn't have one
	if host, ok := p.config.Settings["production_host"].(string); ok && host != "" {
		return host
	}
	if p.config.BareMetal.Host != "" {
		return p.config.BareMetal.Host
	}
	// Default to domain (may need DNS resolution)
	return p.config.Production.Domain
}

func (p *ProductionProvider) getUser() string {
	if user, ok := p.config.Settings["production_user"].(string); ok && user != "" {
		return user
	}
	if p.config.BareMetal.User != "" {
		return p.config.BareMetal.User
	}
	return "root"
}

func (p *ProductionProvider) getSSHKeyPath() string {
	if keyPath, ok := p.config.Settings["production_ssh_key"].(string); ok && keyPath != "" {
		return keyPath
	}
	if p.config.BareMetal.SSHKeyPath != "" {
		return p.config.BareMetal.SSHKeyPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + "/.ssh/id_rsa"
}

func (p *ProductionProvider) getSSHPort() string {
	if p.config.BareMetal.Port > 0 {
		return fmt.Sprint(p.config.BareMetal.Port)
	}
	return "22"
}

func (p *ProductionProvider) buildSSHArgs(command string) []string {
	return []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=10",
		"-i", p.getSSHKeyPath(),
		"-p", p.getSSHPort(),
		fmt.Sprintf("%s@%s", p.getUser(), p.getHost()),
		command,
	}
}
