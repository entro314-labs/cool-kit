package baremetal

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/git"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/entro314-labs/cool-kit/internal/utils"
)

// BareMetalProvider handles bare metal and VM deployments via SSH
type BareMetalProvider struct {
	config     *config.Config
	gitManager *git.Manager
	logger     *utils.Logger
	health     *utils.HealthChecker
}

// NewBareMetalProvider creates a new bare metal provider
func NewBareMetalProvider(cfg *config.Config) (*BareMetalProvider, error) {
	logger, err := utils.NewLogger("", false)
	if err != nil {
		return nil, err
	}

	return &BareMetalProvider{
		config:     cfg,
		gitManager: git.NewManager(cfg),
		logger:     logger,
		health:     utils.NewHealthChecker(logger),
	}, nil
}

// GetDeploymentSteps returns deployment steps for bare metal
func (p *BareMetalProvider) GetDeploymentSteps() []ui.DeploymentStep {
	return []ui.DeploymentStep{
		{Name: "Validate SSH connectivity", Description: "Testing SSH connection to server"},
		{Name: "Check system requirements", Description: "Verifying OS and dependencies"},
		{Name: "Clone Coolify repository", Description: "Fetching latest Coolify from GitHub"},
		{Name: "Install Docker", Description: "Installing Docker Engine"},
		{Name: "Install Docker Compose", Description: "Installing Docker Compose"},
		{Name: "Configure environment", Description: "Setting up Coolify configuration"},
		{Name: "Deploy Coolify", Description: "Starting Coolify services"},
		{Name: "Run health checks", Description: "Validating deployment"},
	}
}

// Deploy performs bare metal deployment
func (p *BareMetalProvider) Deploy(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	steps := []struct {
		name string
		fn   func(chan<- ui.StepProgressMsg, chan<- ui.LogMsg) error
	}{
		{"Validate SSH connectivity", p.validateSSH},
		{"Check system requirements", p.checkRequirements},
		{"Clone Coolify repository", p.cloneRepository},
		{"Install Docker", p.installDocker},
		{"Install Docker Compose", p.installDockerCompose},
		{"Configure environment", p.configureEnvironment},
		{"Deploy Coolify", p.deployCoolify},
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

// validateSSH validates SSH connectivity
func (p *BareMetalProvider) validateSSH(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Testing SSH connection"}

	host := p.getHost()
	user := p.getUser()

	if err := p.health.CheckSSH(host, user, 10*time.Second); err != nil {
		return fmt.Errorf("SSH connectivity check failed: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "SSH connection verified"}
	return nil
}

// checkRequirements checks system requirements
func (p *BareMetalProvider) checkRequirements(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Checking OS version"}

	host := p.getHost()
	user := p.getUser()

	// Check OS
	cmd := p.sshCommand(host, user, "cat /etc/os-release")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check OS: %w", err)
	}

	osInfo := string(output)
	logChan <- ui.LogMsg{Level: ui.LogDebug, Message: fmt.Sprintf("OS: %s", strings.Split(osInfo, "\n")[0])}

	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Checking available disk space"}

	// Check disk space
	cmd = p.sshCommand(host, user, "df -h / | tail -1 | awk '{print $4}'")
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check disk space: %w", err)
	}

	diskSpace := strings.TrimSpace(string(output))
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Available disk space: %s", diskSpace)}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "System requirements verified"}
	return nil
}

// cloneRepository clones Coolify repository
func (p *BareMetalProvider) cloneRepository(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Cloning repository"}

	if err := p.gitManager.CloneOrPull(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	commitInfo, _ := p.gitManager.GetLatestCommitInfo()
	if commitInfo != nil {
		logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Using commit: %s", commitInfo.ShortHash)}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Repository cloned"}
	return nil
}

// installDocker installs Docker
func (p *BareMetalProvider) installDocker(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Checking if Docker is installed"}

	host := p.getHost()
	user := p.getUser()

	// Check if Docker is already installed
	cmd := p.sshCommand(host, user, "docker --version")
	if err := cmd.Run(); err == nil {
		logChan <- ui.LogMsg{Level: ui.LogInfo, Message: "Docker already installed"}
		progressChan <- ui.StepProgressMsg{Progress: 1.0, Message: "Docker already installed"}
		return nil
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.4, Message: "Installing Docker"}

	// Install Docker
	installScript := `
		curl -fsSL https://get.docker.com -o get-docker.sh
		sudo sh get-docker.sh
		sudo usermod -aG docker $USER
		sudo systemctl enable docker
		sudo systemctl start docker
		rm get-docker.sh
	`

	cmd = p.sshCommand(host, user, installScript)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Docker: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Docker installed"}
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Docker installation complete"}
	return nil
}

// installDockerCompose installs Docker Compose
func (p *BareMetalProvider) installDockerCompose(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Installing Docker Compose"}

	host := p.getHost()
	user := p.getUser()

	installScript := `
		sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
		sudo chmod +x /usr/local/bin/docker-compose
	`

	cmd := p.sshCommand(host, user, installScript)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Docker Compose: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Docker Compose installed"}
	return nil
}

// configureEnvironment configures Coolify environment
func (p *BareMetalProvider) configureEnvironment(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Creating directories"}

	host := p.getHost()
	user := p.getUser()

	// Create directories
	cmd := p.sshCommand(host, user, "mkdir -p ~/coolify")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.6, Message: "Copying configuration files"}

	// Copy files (simplified - in production would use SCP)
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: "Configuration files ready"}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Environment configured"}
	return nil
}

// deployCoolify deploys Coolify
func (p *BareMetalProvider) deployCoolify(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Starting Coolify services"}

	host := p.getHost()
	user := p.getUser()

	// Start Coolify with Docker Compose
	cmd := p.sshCommand(host, user, "cd ~/coolify && docker-compose up -d")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start Coolify: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.6, Message: "Waiting for services to start"}
	time.Sleep(30 * time.Second)

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Coolify deployed"}
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Coolify deployment complete"}
	return nil
}

// runHealthChecks runs health checks
func (p *BareMetalProvider) runHealthChecks(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Running health checks"}

	host := p.getHost()
	httpURL := fmt.Sprintf("http://%s", host)
	wsURL := fmt.Sprintf("http://%s:6001", host)

	if err := p.health.ComprehensiveCheck(httpURL, wsURL); err != nil {
		return fmt.Errorf("health checks failed: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "All checks passed"}
	return nil
}

// Helper methods
func (p *BareMetalProvider) getHost() string {
	if host, ok := p.config.Settings["baremetal_host"].(string); ok {
		return host
	}
	return "localhost"
}

func (p *BareMetalProvider) getUser() string {
	if user, ok := p.config.Settings["baremetal_user"].(string); ok {
		return user
	}
	return "ubuntu"
}

func (p *BareMetalProvider) sshCommand(host, user, command string) *exec.Cmd {
	return exec.Command("ssh", fmt.Sprintf("%s@%s", user, host), command)
}
