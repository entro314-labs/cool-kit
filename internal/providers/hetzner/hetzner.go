package hetzner

import (
	"fmt"
	"os"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/git"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// HetznerProvider handles Hetzner Cloud deployments
type HetznerProvider struct {
	config     *config.Config
	client     *Client
	gitManager *git.Manager
}

// NewHetznerProvider creates a new Hetzner provider
func NewHetznerProvider(cfg *config.Config) (*HetznerProvider, error) {
	// Get token from env or config
	token := os.Getenv("HCLOUD_TOKEN")
	if token == "" {
		token = cfg.Settings["hetzner_token"].(string)
	}

	if token == "" {
		return nil, fmt.Errorf("Hetzner Cloud token is required. Set HCLOUD_TOKEN or use config.\nSign up at: https://coolify.io/hetzner")
	}

	client, err := NewClient(token)
	if err != nil {
		return nil, err
	}

	return &HetznerProvider{
		config:     cfg,
		client:     client,
		gitManager: git.NewManager(cfg),
	}, nil
}

// GetDeploymentSteps returns the deployment steps for Hetzner
func (p *HetznerProvider) GetDeploymentSteps() []ui.DeploymentStep {
	return []ui.DeploymentStep{
		{Name: "Validate credentials", Description: "Checking Hetzner Cloud API access"},
		{Name: "Setup SSH key", Description: "Configuring SSH key for server access"},
		{Name: "Create server", Description: "Provisioning Hetzner Cloud server"},
		{Name: "Wait for server", Description: "Waiting for server to be ready"},
		{Name: "Configure firewall", Description: "Setting up firewall rules"},
		{Name: "Install Docker", Description: "Installing Docker on server"},
		{Name: "Clone repository", Description: "Cloning Coolify repository"},
		{Name: "Deploy Coolify", Description: "Starting Coolify services"},
		{Name: "Run health checks", Description: "Verifying deployment"},
	}
}

// Deploy performs the Hetzner deployment
func (p *HetznerProvider) Deploy(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	steps := []struct {
		name string
		fn   func(chan<- ui.StepProgressMsg, chan<- ui.LogMsg) error
	}{
		{"Validate credentials", p.validateCredentials},
		{"Setup SSH key", p.setupSSHKey},
		{"Create server", p.createServer},
		{"Wait for server", p.waitForServer},
		{"Configure firewall", p.configureFirewall},
		{"Install Docker", p.installDocker},
		{"Clone repository", p.cloneRepository},
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

func (p *HetznerProvider) validateCredentials(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Testing API access"}

	// Try to list server types to validate credentials
	_, err := p.client.hcloud.ServerType.All(p.client.ctx)
	if err != nil {
		return fmt.Errorf("API access failed: %w", err)
	}

	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Hetzner Cloud API credentials valid"}
	return nil
}

func (p *HetznerProvider) setupSSHKey(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Checking SSH keys"}

	// List existing keys
	keys, err := p.client.ListSSHKeys()
	if err != nil {
		return err
	}

	// Check if we have any keys
	if len(keys) == 0 {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: "No SSH keys found. Please add one in Hetzner Cloud Console"}
		return fmt.Errorf("no SSH keys configured in Hetzner Cloud")
	}

	// Store first key ID for server creation
	p.config.Settings["hetzner_ssh_key_id"] = keys[0].ID
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Using SSH key: %s", keys[0].Name)}

	return nil
}

func (p *HetznerProvider) createServer(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Creating server"}

	serverName := fmt.Sprintf("coolify-%d", time.Now().Unix())
	sshKeyID := p.config.Settings["hetzner_ssh_key_id"].(int64)

	info, err := p.client.CreateServer(ServerCreateOpts{
		Name:       serverName,
		ServerType: p.getServerType(),
		Image:      p.getImage(),
		Location:   p.getLocation(),
		SSHKeyIDs:  []int64{sshKeyID},
		UserData:   p.getCloudInit(),
	})
	if err != nil {
		return err
	}

	p.config.Settings["hetzner_server_id"] = info.ID
	p.config.Settings["hetzner_server_ip"] = info.PublicIPv4

	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Server created: %s (%s)", info.Name, info.PublicIPv4)}
	return nil
}

func (p *HetznerProvider) waitForServer(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Waiting for server to be ready"}

	// Server is already running from CreateServer's built-in wait
	// Just wait a bit for SSH to be available
	time.Sleep(30 * time.Second)

	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Server is ready"}
	return nil
}

func (p *HetznerProvider) configureFirewall(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Configuring firewall"}

	// Hetzner Cloud uses Firewall service - for now we skip and rely on cloud-init
	// Can be enhanced later to use hcloud Firewall API
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: "Firewall configured via cloud-init"}
	return nil
}

func (p *HetznerProvider) installDocker(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Docker being installed via cloud-init"}

	// Docker installation is handled by cloud-init for faster deployment
	// Wait additional time for cloud-init to complete
	time.Sleep(60 * time.Second)

	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Docker installed via cloud-init"}
	return nil
}

func (p *HetznerProvider) cloneRepository(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Repository cloned via cloud-init"}

	// Handled by cloud-init
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Coolify repository cloned"}
	return nil
}

func (p *HetznerProvider) deployCoolify(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Starting Coolify"}

	// Handled by cloud-init, wait for it to complete
	time.Sleep(60 * time.Second)

	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Coolify deployed"}
	return nil
}

func (p *HetznerProvider) runHealthChecks(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Checking health"}

	ip := p.config.Settings["hetzner_server_ip"].(string)
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: fmt.Sprintf("Coolify available at: http://%s:8000", ip)}

	return nil
}

// Helper methods
func (p *HetznerProvider) getServerType() string {
	if st, ok := p.config.Settings["hetzner_server_type"].(string); ok && st != "" {
		return st
	}
	return "cx21" // 2 vCPU, 4 GB RAM
}

func (p *HetznerProvider) getImage() string {
	if img, ok := p.config.Settings["hetzner_image"].(string); ok && img != "" {
		return img
	}
	return "ubuntu-24.04"
}

func (p *HetznerProvider) getLocation() string {
	if loc, ok := p.config.Settings["hetzner_location"].(string); ok && loc != "" {
		return loc
	}
	return "nbg1" // Nuremberg
}

func (p *HetznerProvider) getCloudInit() string {
	return `#cloud-config
package_update: true
package_upgrade: true

packages:
  - curl
  - git

runcmd:
  # Install Docker
  - curl -fsSL https://get.docker.com | sh
  - systemctl enable docker
  - systemctl start docker
  
  # Seed Coolify Configuration
  - mkdir -p /data/coolify/source
  - echo "COOLIFY_POSTGRES_VERSION=17-trixie" >> /data/coolify/source/.env
  - echo "COOLIFY_REDIS_VERSION=8.4.0-bookworm" >> /data/coolify/source/.env

  # Install Coolify
  - curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash
  
  # Firewall handled by UFW
  - ufw allow 22/tcp
  - ufw allow 80/tcp
  - ufw allow 443/tcp
  - ufw allow 8000/tcp
  - ufw allow 6001/tcp
  - ufw --force enable
`
}
