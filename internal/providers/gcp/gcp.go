package gcp

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/git"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/entro314-labs/cool-kit/internal/utils"
)

// GCPProvider handles Google Cloud Platform deployments
type GCPProvider struct {
	config     *config.Config
	gitManager *git.Manager
	logger     *utils.Logger
	health     *utils.HealthChecker
}

// NewGCPProvider creates a new GCP provider
func NewGCPProvider(cfg *config.Config) (*GCPProvider, error) {
	logger, err := utils.NewLogger("", false)
	if err != nil {
		return nil, err
	}

	return &GCPProvider{
		config:     cfg,
		gitManager: git.NewManager(cfg),
		logger:     logger,
		health:     utils.NewHealthChecker(logger),
	}, nil
}

// GetDeploymentSteps returns deployment steps for GCP
func (p *GCPProvider) GetDeploymentSteps() []ui.DeploymentStep {
	return []ui.DeploymentStep{
		{Name: "Validate GCP credentials", Description: "Checking gcloud CLI and credentials"},
		{Name: "Clone Coolify repository", Description: "Fetching latest Coolify from GitHub"},
		{Name: "Create VPC network", Description: "Setting up network infrastructure"},
		{Name: "Configure firewall rules", Description: "Setting up security rules"},
		{Name: "Launch Compute Engine instance", Description: "Creating virtual machine"},
		{Name: "Assign static IP", Description: "Allocating external IP address"},
		{Name: "Install Docker", Description: "Installing Docker on instance"},
		{Name: "Deploy Coolify", Description: "Setting up Coolify application"},
		{Name: "Configure Cloud DNS", Description: "Setting up DNS records"},
		{Name: "Run health checks", Description: "Validating deployment"},
	}
}

// Deploy performs GCP deployment
func (p *GCPProvider) Deploy(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	steps := []struct {
		name string
		fn   func(chan<- ui.StepProgressMsg, chan<- ui.LogMsg) error
	}{
		{"Validate GCP credentials", p.validateCredentials},
		{"Clone Coolify repository", p.cloneRepository},
		{"Create VPC network", p.createVPC},
		{"Configure firewall rules", p.createFirewallRules},
		{"Launch Compute Engine instance", p.launchInstance},
		{"Assign static IP", p.assignStaticIP},
		{"Install Docker", p.installDocker},
		{"Deploy Coolify", p.deployCoolify},
		{"Configure Cloud DNS", p.configureDNS},
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

// validateCredentials validates GCP credentials
func (p *GCPProvider) validateCredentials(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Checking gcloud CLI"}

	cmd := exec.Command("gcloud", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gcloud CLI not installed. Please install it first: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Validating credentials"}

	cmd = exec.Command("gcloud", "auth", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("GCP credentials validation failed. Please run 'gcloud auth login': %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "Credentials validated"}
	logChan <- ui.LogMsg{Level: ui.LogDebug, Message: fmt.Sprintf("GCP Auth: %s", strings.TrimSpace(string(output)))}

	return nil
}

// cloneRepository clones Coolify repository
func (p *GCPProvider) cloneRepository(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Fetching repository"}

	if err := p.gitManager.CloneOrPull(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	commitInfo, _ := p.gitManager.GetLatestCommitInfo()
	if commitInfo != nil {
		logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Using commit: %s", commitInfo.ShortHash)}
	}

	return nil
}

// createVPC creates VPC network
func (p *GCPProvider) createVPC(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Creating VPC network"}

	project := p.getProject()
	networkName := p.getNetwork()

	cmd := exec.Command("gcloud", "compute", "networks", "create", networkName,
		"--project", project,
		"--subnet-mode=auto",
		"--bgp-routing-mode=regional")

	if err := cmd.Run(); err != nil {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: "VPC may already exist"}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "VPC network ready"}
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "VPC network configured"}

	return nil
}

// createFirewallRules creates firewall rules
func (p *GCPProvider) createFirewallRules(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Creating firewall rules"}

	project := p.getProject()
	networkName := p.getNetwork()

	rules := []struct {
		name  string
		ports string
	}{
		{"coolify-ssh", "22"},
		{"coolify-http", "80"},
		{"coolify-https", "443"},
		{"coolify-websocket", "6001"},
	}

	for i, rule := range rules {
		progress := 0.2 + (float64(i+1) / float64(len(rules)) * 0.7)
		progressChan <- ui.StepProgressMsg{Progress: progress, Message: fmt.Sprintf("Creating rule: %s", rule.name)}

		cmd := exec.Command("gcloud", "compute", "firewall-rules", "create", rule.name,
			"--project", project,
			"--network", networkName,
			"--allow", fmt.Sprintf("tcp:%s", rule.ports),
			"--source-ranges", "0.0.0.0/0")

		cmd.Run() // Ignore errors as rules may exist
	}

	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Firewall rules configured"}
	return nil
}

// launchInstance launches Compute Engine instance
func (p *GCPProvider) launchInstance(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Launching instance"}

	project := p.getProject()
	zone := p.getZone()
	machineType := p.getMachineType()
	instanceName := p.getInstanceName()
	networkName := p.getNetwork()

	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Machine type: %s, Zone: %s", machineType, zone)}

	cmd := exec.Command("gcloud", "compute", "instances", "create", instanceName,
		"--project", project,
		"--zone", zone,
		"--machine-type", machineType,
		"--image-family", "ubuntu-2004-lts",
		"--image-project", "ubuntu-os-cloud",
		"--boot-disk-size", "30GB",
		"--network", networkName,
		"--tags", "coolify,http-server,https-server",
		"--labels", "application=coolify,managed-by=cool-kit")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch instance: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.6, Message: "Instance launched, waiting for running state"}
	if err := p.waitForInstanceStatus(instanceName, "RUNNING", 60, 5*time.Second); err != nil {
		return fmt.Errorf("instance did not reach RUNNING state: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Instance is running"}
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Compute Engine instance running"}

	return nil
}

// assignStaticIP assigns static external IP
func (p *GCPProvider) assignStaticIP(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Reserving static IP"}

	project := p.getProject()
	region := p.getRegion()
	zone := p.getZone()
	instanceName := p.getInstanceName()

	cmd := exec.Command("gcloud", "compute", "addresses", "create", "coolify-ip",
		"--project", project,
		"--region", region)

	if err := cmd.Run(); err != nil {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: "IP may already exist"}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.7, Message: "Getting IP address"}

	cmd = exec.Command("gcloud", "compute", "addresses", "describe", "coolify-ip",
		"--project", project,
		"--region", region,
		"--format", "get(address)")

	output, _ := cmd.Output()
	publicIP := strings.TrimSpace(string(output))

	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "Attaching static IP to instance"}

	// Replace any existing external IP access config with the reserved static IP.
	accessConfigName := "external-nat"
	accessConfigCmd := exec.Command("gcloud", "compute", "instances", "describe", instanceName,
		"--project", project,
		"--zone", zone,
		"--format", "value(networkInterfaces[0].accessConfigs[0].name)")
	if nameOutput, err := accessConfigCmd.Output(); err == nil {
		if existing := strings.TrimSpace(string(nameOutput)); existing != "" {
			accessConfigName = existing
		}
	}

	detachCmd := exec.Command("gcloud", "compute", "instances", "delete-access-config", instanceName,
		"--project", project,
		"--zone", zone,
		"--access-config-name", accessConfigName)
	_ = detachCmd.Run() // Ignore errors when no access config exists.

	attachCmd := exec.Command("gcloud", "compute", "instances", "add-access-config", instanceName,
		"--project", project,
		"--zone", zone,
		"--access-config-name", accessConfigName,
		"--address", publicIP)
	if output, err := attachCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to attach static IP %s to instance %s: %w (%s)", publicIP, instanceName, err, strings.TrimSpace(string(output)))
	}

	p.config.Settings["public_ip"] = publicIP
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Static IP: %s", publicIP)}

	return nil
}

// installDocker installs Docker
func (p *GCPProvider) installDocker(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Waiting for SSH access"}
	if err := p.waitForSSH(24, 5*time.Second); err != nil {
		return fmt.Errorf("instance SSH not ready: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.4, Message: "Installing Docker"}
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: "Installing Docker via gcloud SSH"}

	installScript := "set -euo pipefail; if ! command -v docker >/dev/null 2>&1; then curl -fsSL https://get.docker.com | sudo sh; fi; sudo usermod -aG docker $USER || true; sudo systemctl enable docker; sudo systemctl start docker; sudo docker --version >/dev/null"
	output, err := p.runSSHCommand(installScript)
	if err != nil {
		return fmt.Errorf("failed to install Docker via SSH: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "Docker installed"}
	return nil
}

// deployCoolify deploys Coolify
func (p *GCPProvider) deployCoolify(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Deploying Coolify"}

	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: "Deploying Coolify via gcloud SSH"}
	installScript := "set -euo pipefail; sudo mkdir -p /data/coolify/source; sudo chown -R $USER:$USER /data/coolify; touch /data/coolify/source/.env; grep -q '^COOLIFY_POSTGRES_VERSION=' /data/coolify/source/.env || echo 'COOLIFY_POSTGRES_VERSION=17-trixie' >> /data/coolify/source/.env; grep -q '^COOLIFY_REDIS_VERSION=' /data/coolify/source/.env || echo 'COOLIFY_REDIS_VERSION=8.4.0-bookworm' >> /data/coolify/source/.env; curl -fsSL https://cdn.coollabs.io/coolify/install.sh | sudo bash"
	output, err := p.runSSHCommand(installScript)
	if err != nil {
		return fmt.Errorf("failed to deploy Coolify via SSH: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Coolify deployed"}
	return nil
}

// configureDNS configures Cloud DNS
func (p *GCPProvider) configureDNS(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Configuring DNS"}

	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: "DNS configuration skipped (optional)"}
	return nil
}

// runHealthChecks runs health checks
func (p *GCPProvider) runHealthChecks(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Running health checks"}

	publicIP, ok := p.config.Settings["public_ip"].(string)
	if !ok || publicIP == "" {
		return fmt.Errorf("public IP not found")
	}

	httpURL := fmt.Sprintf("http://%s:8000", publicIP)
	wsURL := fmt.Sprintf("http://%s:6001", publicIP)
	httpClient := &http.Client{Timeout: 10 * time.Second}

	var lastErr error
	for attempt := 1; attempt <= 20; attempt++ {
		progress := 0.2 + (float64(attempt) / 20.0 * 0.55)
		progressChan <- ui.StepProgressMsg{
			Progress: progress,
			Message:  fmt.Sprintf("Waiting for Coolify endpoint (attempt %d/20)", attempt),
		}

		resp, err := httpClient.Get(httpURL)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 500 {
				lastErr = nil
				break
			}
			lastErr = fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		time.Sleep(15 * time.Second)
	}
	if lastErr != nil {
		return fmt.Errorf("Coolify endpoint did not become reachable at %s: %w", httpURL, lastErr)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.85, Message: "Checking WebSocket endpoint"}
	if resp, err := httpClient.Get(wsURL); err == nil {
		_ = resp.Body.Close()
		logChan <- ui.LogMsg{Level: ui.LogInfo, Message: "WebSocket endpoint responded"}
	} else {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: fmt.Sprintf("WebSocket endpoint check failed: %v", err)}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.95, Message: "All checks passed"}
	return nil
}

// Helper methods
func (p *GCPProvider) getProject() string {
	if project := strings.TrimSpace(p.config.GCP.Project); project != "" {
		return project
	}
	return "my-project"
}

func (p *GCPProvider) getZone() string {
	if zone := strings.TrimSpace(p.config.GCP.Zone); zone != "" {
		return zone
	}
	return "us-central1-a"
}

func (p *GCPProvider) getRegion() string {
	zone := p.getZone()
	parts := strings.Split(zone, "-")
	if len(parts) >= 2 {
		return strings.Join(parts[:2], "-")
	}
	return "us-central1"
}

func (p *GCPProvider) getMachineType() string {
	if machineType := strings.TrimSpace(p.config.GCP.MachineType); machineType != "" {
		return machineType
	}
	return "e2-medium"
}

func (p *GCPProvider) getNetwork() string {
	if p.config.GCP.Network != "" {
		return p.config.GCP.Network
	}
	return "default"
}

func (p *GCPProvider) getInstanceName() string {
	if name, ok := p.config.Settings["gcp_instance_name"].(string); ok && name != "" {
		return name
	}
	return "coolify-instance"
}

func (p *GCPProvider) runSSHCommand(remoteCommand string) ([]byte, error) {
	cmd := exec.Command(
		"gcloud", "compute", "ssh", p.getInstanceName(),
		"--project", p.getProject(),
		"--zone", p.getZone(),
		"--quiet",
		"--command", remoteCommand,
	)
	return cmd.CombinedOutput()
}

func (p *GCPProvider) waitForSSH(maxAttempts int, retryDelay time.Duration) error {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		output, err := p.runSSHCommand("echo SSH ready")
		if err == nil && strings.Contains(string(output), "SSH ready") {
			return nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("unexpected SSH readiness output: %s", strings.TrimSpace(string(output)))
		}
		time.Sleep(retryDelay)
	}
	return fmt.Errorf("SSH did not become available after %d attempts: %w", maxAttempts, lastErr)
}

func (p *GCPProvider) waitForInstanceStatus(name, target string, maxAttempts int, retryDelay time.Duration) error {
	var lastStatus string
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		cmd := exec.Command("gcloud", "compute", "instances", "describe", name,
			"--project", p.getProject(),
			"--zone", p.getZone(),
			"--format", "value(status)")
		output, err := cmd.Output()
		if err == nil {
			lastStatus = strings.TrimSpace(string(output))
			if strings.EqualFold(lastStatus, target) {
				return nil
			}
		}
		time.Sleep(retryDelay)
	}
	if lastStatus == "" {
		return fmt.Errorf("unable to determine instance status after %d attempts", maxAttempts)
	}
	return fmt.Errorf("instance status stayed at %q, expected %q", lastStatus, target)
}
