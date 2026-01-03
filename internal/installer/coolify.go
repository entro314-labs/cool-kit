package installer

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
)

const (
	CoolifyInstallScriptURL = "https://cdn.coollabs.io/coolify/install.sh"
	CoolifyGitHubRepo       = "https://github.com/coollabsio/coolify.git"
	DefaultInstallDir       = "/data/coolify"
)

// CoolifyDeployer handles official Coolify deployments
type CoolifyDeployer struct {
	config       *config.Config
	installDir   string
	progress     chan<- ProgressUpdate
	logs         chan<- string
	dashboardURL string
}

// ProgressUpdate represents deployment progress
type ProgressUpdate struct {
	Step        int
	TotalSteps  int
	StepName    string
	Percentage  float64
	Message     string
	Error       error
	IsCompleted bool
}

// NewCoolifyDeployer creates a new Coolify deployer
func NewCoolifyDeployer(cfg *config.Config, progressChan chan<- ProgressUpdate, logChan chan<- string) *CoolifyDeployer {
	installDir := DefaultInstallDir
	if cfg != nil && cfg.Git.WorkDir != "" {
		installDir = cfg.Git.WorkDir
	}

	return &CoolifyDeployer{
		config:     cfg,
		installDir: installDir,
		progress:   progressChan,
		logs:       logChan,
	}
}

// Deploy executes the official Coolify installation
func (d *CoolifyDeployer) Deploy() (string, error) {
	steps := []struct {
		name string
		fn   func() error
	}{
		{"Preparing environment", d.prepareEnvironment},
		{"Downloading Coolify installer", d.downloadInstaller},
		{"Executing official installation", d.executeInstaller},
		{"Retrieving dashboard URL", d.getDashboardURL},
		{"Running health checks", d.healthCheck},
	}

	totalSteps := len(steps)

	for i, step := range steps {
		d.sendProgress(i+1, totalSteps, step.name, 0, fmt.Sprintf("Starting: %s", step.name), nil)
		d.sendLog(fmt.Sprintf("[%d/%d] %s...", i+1, totalSteps, step.name))

		if err := step.fn(); err != nil {
			d.sendProgress(i+1, totalSteps, step.name, 0, "", err)
			d.sendLog(fmt.Sprintf("❌ Failed: %s - %v", step.name, err))
			return "", fmt.Errorf("step '%s' failed: %w", step.name, err)
		}

		percentage := float64(i+1) / float64(totalSteps) * 100
		d.sendProgress(i+1, totalSteps, step.name, percentage, fmt.Sprintf("✓ %s completed", step.name), nil)
		d.sendLog(fmt.Sprintf("✅ Completed: %s", step.name))
	}

	d.sendProgress(totalSteps, totalSteps, "Deployment Complete", 100, "🎉 Coolify deployed successfully!", nil)
	d.sendLog("🎉 Deployment completed successfully!")
	d.sendLog(fmt.Sprintf("🌐 Dashboard URL: %s", d.dashboardURL))

	return d.dashboardURL, nil
}

// prepareEnvironment prepares the deployment environment
func (d *CoolifyDeployer) prepareEnvironment() error {
	d.sendLog("Checking system requirements...")

	// Check if running as root or with sudo
	if os.Geteuid() != 0 {
		return fmt.Errorf("this installer must be run as root or with sudo")
	}

	// Check for required commands
	requiredCommands := []string{"curl", "wget", "git", "docker"}
	for _, cmd := range requiredCommands {
		if _, err := exec.LookPath(cmd); err != nil {
			d.sendLog(fmt.Sprintf("⚠️ %s not found, will be installed by Coolify installer", cmd))
		}
	}

	d.sendLog("✓ Environment preparation complete")
	return nil
}

// downloadInstaller downloads the official Coolify install script
func (d *CoolifyDeployer) downloadInstaller() error {
	d.sendLog(fmt.Sprintf("Downloading installer from %s...", CoolifyInstallScriptURL))

	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("coolify-cli-%d", time.Now().Unix()))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Download install script
	resp, err := http.Get(CoolifyInstallScriptURL)
	if err != nil {
		return fmt.Errorf("failed to download install script: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download install script: HTTP %d", resp.StatusCode)
	}

	// Save script
	scriptPath := filepath.Join(tmpDir, "install.sh")
	file, err := os.Create(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to create install script file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to save install script: %w", err)
	}

	// Make executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	d.installDir = tmpDir
	d.sendLog(fmt.Sprintf("✓ Installer downloaded to %s", scriptPath))
	return nil
}

// executeInstaller runs the official Coolify installation script
func (d *CoolifyDeployer) executeInstaller() error {
	scriptPath := filepath.Join(d.installDir, "install.sh")
	d.sendLog(fmt.Sprintf("Executing installer: %s", scriptPath))
	d.sendLog("This may take several minutes depending on your server...")

	// Set environment variables if configured
	env := os.Environ()
	if d.config != nil {
		// Add custom environment variables from config
		if d.config.Settings != nil {
			if rootUser, ok := d.config.Settings["root_username"].(string); ok {
				env = append(env, fmt.Sprintf("ROOT_USERNAME=%s", rootUser))
			}
			if rootEmail, ok := d.config.Settings["root_email"].(string); ok {
				env = append(env, fmt.Sprintf("ROOT_USER_EMAIL=%s", rootEmail))
			}
			if rootPass, ok := d.config.Settings["root_password"].(string); ok {
				env = append(env, fmt.Sprintf("ROOT_USER_PASSWORD=%s", rootPass))
			}
		}
	}

	// Execute the install script
	cmd := exec.Command("bash", scriptPath)
	cmd.Env = env

	// Capture output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start installer: %w", err)
	}

	// Stream output
	go d.streamOutput(stdout, "STDOUT")
	go d.streamOutput(stderr, "STDERR")

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("installer failed: %w", err)
	}

	d.sendLog("✓ Official installation completed")
	return nil
}

// streamOutput streams command output to logs
func (d *CoolifyDeployer) streamOutput(reader io.Reader, prefix string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// Filter out empty lines
		if strings.TrimSpace(line) != "" {
			d.sendLog(fmt.Sprintf("[%s] %s", prefix, line))
		}
	}
}

// getDashboardURL retrieves the Coolify dashboard URL
func (d *CoolifyDeployer) getDashboardURL() error {
	d.sendLog("Retrieving dashboard URL...")

	// Get public IP
	publicIP, err := d.getPublicIP()
	if err != nil {
		d.sendLog("⚠️ Failed to get public IP, using localhost")
		publicIP = "localhost"
	}

	// Coolify default port is 8000
	d.dashboardURL = fmt.Sprintf("http://%s:8000", publicIP)
	d.sendLog(fmt.Sprintf("✓ Dashboard URL: %s", d.dashboardURL))

	return nil
}

// getPublicIP retrieves the server's public IP address
func (d *CoolifyDeployer) getPublicIP() (string, error) {
	resp, err := http.Get("https://ifconfig.io")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	ip := strings.TrimSpace(string(body))

	// Validate IP format
	ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if !ipRegex.MatchString(ip) {
		return "", fmt.Errorf("invalid IP format: %s", ip)
	}

	return ip, nil
}

// healthCheck performs health checks on the deployed Coolify instance
func (d *CoolifyDeployer) healthCheck() error {
	d.sendLog("Running health checks...")

	// Wait a bit for services to stabilize
	d.sendLog("Waiting for services to stabilize...")
	time.Sleep(10 * time.Second)

	// Check if Coolify container is running
	checkCmd := exec.Command("docker", "inspect", "--format={{.State.Health.Status}}", "coolify")
	output, err := checkCmd.Output()
	if err != nil {
		d.sendLog("⚠️ Could not check Coolify container health")
		return nil // Non-fatal
	}

	healthStatus := strings.TrimSpace(string(output))
	d.sendLog(fmt.Sprintf("Coolify container health: %s", healthStatus))

	// Check HTTP response
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(d.dashboardURL + "/up")
		if err == nil && resp.StatusCode == http.StatusOK {
			d.sendLog("✓ Coolify is responding to HTTP requests")
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}

		d.sendLog(fmt.Sprintf("Attempt %d/%d: Waiting for Coolify to be ready...", i+1, maxRetries))
		time.Sleep(5 * time.Second)
	}

	d.sendLog("⚠️ HTTP health check timed out (Coolify may still be starting)")
	return nil // Non-fatal
}

// sendProgress sends progress update
func (d *CoolifyDeployer) sendProgress(step, total int, name string, percentage float64, message string, err error) {
	if d.progress != nil {
		d.progress <- ProgressUpdate{
			Step:        step,
			TotalSteps:  total,
			StepName:    name,
			Percentage:  percentage,
			Message:     message,
			Error:       err,
			IsCompleted: step == total && err == nil,
		}
	}
}

// sendLog sends log message
func (d *CoolifyDeployer) sendLog(message string) {
	if d.logs != nil {
		d.logs <- message
	}
}
