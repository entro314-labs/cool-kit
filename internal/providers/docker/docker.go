package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/git"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/entro314-labs/cool-kit/internal/utils"
)

// DockerProvider handles Docker/Docker Compose deployments
type DockerProvider struct {
	config     *config.Config
	gitManager *git.Manager
	logger     *utils.Logger
	health     *utils.HealthChecker
	profile    string
}

// NewDockerProvider creates a new Docker provider
func NewDockerProvider(cfg *config.Config, profile string) (*DockerProvider, error) {
	logger, err := utils.NewLogger("", false)
	if err != nil {
		return nil, err
	}

	return &DockerProvider{
		config:     cfg,
		gitManager: git.NewManager(cfg),
		logger:     logger,
		health:     utils.NewHealthChecker(logger),
		profile:    profile,
	}, nil
}

// GetDeploymentSteps returns deployment steps for Docker
func (p *DockerProvider) GetDeploymentSteps() []ui.DeploymentStep {
	return []ui.DeploymentStep{
		{Name: "Validate Docker installation", Description: "Checking Docker and Docker Compose"},
		{Name: "Clone Coolify repository", Description: "Fetching latest Coolify from GitHub"},
		{Name: "Generate credentials", Description: "Creating secure credentials"},
		{Name: "Configure environment", Description: "Setting up environment variables"},
		{Name: "Pull Docker images", Description: "Downloading required images"},
		{Name: "Start services", Description: "Starting Docker Compose services"},
		{Name: "Wait for services", Description: "Waiting for services to be ready"},
		{Name: "Run health checks", Description: "Validating deployment"},
	}
}

// Deploy performs Docker deployment
func (p *DockerProvider) Deploy(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	steps := []struct {
		name string
		fn   func(chan<- ui.StepProgressMsg, chan<- ui.LogMsg) error
	}{
		{"Validate Docker installation", p.validateDocker},
		{"Clone Coolify repository", p.cloneRepository},
		{"Generate credentials", p.generateCredentials},
		{"Configure environment", p.configureEnvironment},
		{"Pull Docker images", p.pullImages},
		{"Start services", p.startServices},
		{"Wait for services", p.waitForServices},
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

// validateDocker validates Docker installation
func (p *DockerProvider) validateDocker(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Checking Docker"}

	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker not installed or not running: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.6, Message: "Checking Docker Compose"}

	cmd = exec.Command("docker-compose", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker Compose not installed: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Docker validated"}
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Docker and Docker Compose available"}

	return nil
}

// cloneRepository clones Coolify repository
func (p *DockerProvider) cloneRepository(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Cloning repository"}

	if err := p.gitManager.CloneOrPull(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	commitInfo, _ := p.gitManager.GetLatestCommitInfo()
	if commitInfo != nil {
		logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Using commit: %s - %s", commitInfo.ShortHash, commitInfo.Message[:min(50, len(commitInfo.Message))])}
	}

	return nil
}

// generateCredentials generates secure credentials
func (p *DockerProvider) generateCredentials(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Generating credentials"}

	// Generate APP_KEY
	cmd := exec.Command("openssl", "rand", "-base64", "32")
	appKey, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to generate APP_KEY: %w", err)
	}

	p.config.Settings["app_key"] = fmt.Sprintf("base64:%s", string(appKey))

	progressChan <- ui.StepProgressMsg{Progress: 0.6, Message: "Generating database password"}

	// Generate DB password
	cmd = exec.Command("openssl", "rand", "-base64", "32")
	dbPassword, _ := cmd.Output()
	p.config.Settings["db_password"] = string(dbPassword)

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Credentials generated"}
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Secure credentials generated"}

	return nil
}

// configureEnvironment configures environment
func (p *DockerProvider) configureEnvironment(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Creating .env file"}

	workDir := p.config.Git.WorkDir
	envFile := filepath.Join(workDir, ".env")

	// Create .env file
	envContent := fmt.Sprintf(`APP_NAME=Coolify
APP_ENV=%s
APP_DEBUG=%s
APP_KEY=%s
DB_PASSWORD=%s
REDIS_PASSWORD=%s
`,
		p.getEnv(),
		p.getDebug(),
		p.config.Settings["app_key"],
		p.config.Settings["db_password"],
		p.config.Settings["redis_password"],
	)

	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "Environment configured"}
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Profile: %s", p.profile)}

	return nil
}

// pullImages pulls Docker images
func (p *DockerProvider) pullImages(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Pulling images"}

	workDir := p.config.Git.WorkDir
	composeFile := p.getComposeFile()

	cmd := exec.Command("docker-compose", "-f", composeFile, "pull")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Images pulled"}
	return nil
}

// startServices starts Docker Compose services
func (p *DockerProvider) startServices(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Starting services"}

	workDir := p.config.Git.WorkDir
	composeFile := p.getComposeFile()

	cmd := exec.Command("docker-compose", "-f", composeFile, "up", "-d")
	cmd.Dir = workDir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "Services started"}
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Docker Compose services started"}

	return nil
}

// waitForServices waits for services to be ready
func (p *DockerProvider) waitForServices(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Waiting for services"}

	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: "Waiting 30 seconds for services to initialize"}
	time.Sleep(30 * time.Second)

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Services ready"}
	return nil
}

// runHealthChecks runs health checks
func (p *DockerProvider) runHealthChecks(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Running health checks"}

	port := p.config.Local.AppPort
	if port == 0 {
		port = 8000
	}

	httpURL := fmt.Sprintf("http://localhost:%d", port)
	wsURL := fmt.Sprintf("http://localhost:%d", p.config.Local.WebSocketPort)

	if err := p.health.WaitForHTTP(httpURL, 10, 3*time.Second); err != nil {
		return err
	}

	p.health.CheckWebSocket(wsURL, 5*time.Second)

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Health checks passed"}
	return nil
}

// Helper methods
func (p *DockerProvider) getEnv() string {
	switch p.profile {
	case "production":
		return "production"
	case "staging":
		return "staging"
	default:
		return "local"
	}
}

func (p *DockerProvider) getDebug() string {
	if p.profile == "production" {
		return "false"
	}
	return "true"
}

func (p *DockerProvider) getComposeFile() string {
	switch p.profile {
	case "production":
		return "docker-compose.prod.yml"
	case "staging":
		return "docker-compose.staging.yml"
	default:
		return "docker-compose.local.yml"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
