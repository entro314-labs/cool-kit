package baremetal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// CheckStatus checks the status of a bare metal Coolify deployment
func CheckStatus() error {
	ui.Section("Bare Metal Coolify Instance Status")

	// Initialize config
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Check host is configured
	if cfg.BareMetal.Host == "" {
		ui.Warning("No host configured. Use 'cool-kit config set baremetal.host YOUR_HOST'")
		ui.Info("Or provide host with '--host' flag when running commands")
		return nil
	}

	ui.Info(fmt.Sprintf("Host: %s", cfg.BareMetal.Host))
	ui.Dim(fmt.Sprintf("User: %s", cfg.BareMetal.User))

	// Create SSH client
	client := NewSSHClientForStatus(cfg.BareMetal.Host, cfg.BareMetal.User, cfg.BareMetal.SSHKeyPath, cfg.BareMetal.Port)

	// Test SSH connectivity
	if err := client.TestConnection(); err != nil {
		ui.Error(fmt.Sprintf("SSH connectivity failed: %v", err))
		return nil
	}
	ui.Success("SSH connection available")

	// Check system info
	checkSystemInfo(client)

	// Check Docker
	checkDocker(client)

	// Check Coolify services
	checkCoolifyServices(client)

	// Check resources
	checkSystemResources(client)

	return nil
}

// checkSystemInfo checks basic system information
func checkSystemInfo(client *StatusSSHClient) {
	ui.Info("System Information")

	script := `hostnamectl 2>/dev/null | grep -E "Operating System|Kernel" || (uname -a | head -1)`

	output, err := client.Execute(script)
	if err != nil {
		ui.Warning(fmt.Sprintf("  Failed to get system info: %v", err))
		return
	}

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		ui.Dim(fmt.Sprintf("  %s", strings.TrimSpace(line)))
	}
}

// checkDocker checks Docker installation
func checkDocker(client *StatusSSHClient) {
	ui.Info("Docker Status")

	// Check Docker version
	output, err := client.Execute("docker --version 2>/dev/null")
	if err != nil {
		ui.Error("  Docker: not installed or not accessible")
		return
	}
	ui.Success(fmt.Sprintf("  %s", strings.TrimSpace(output)))

	// Check Docker daemon
	_, err = client.Execute("docker ps >/dev/null 2>&1")
	if err != nil {
		ui.Warning("  Docker daemon: not running or no permissions")
	} else {
		ui.Success("  Docker daemon: running")
	}
}

// checkCoolifyServices checks Coolify container status
func checkCoolifyServices(client *StatusSSHClient) {
	ui.Info("Coolify Services")

	script := `for c in coolify coolify-db coolify-redis coolify-realtime coolify-soketi; do
    status=$(docker inspect --format='{{.State.Running}}' $c 2>/dev/null || echo "notfound")
    health=$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' $c 2>/dev/null || echo "none")
    echo "$c:$status:$health"
done`

	output, err := client.Execute(script)
	if err != nil {
		ui.Warning(fmt.Sprintf("  Failed to check services: %v", err))
		return
	}

	foundAny := false
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		running := parts[1]

		if running == "notfound" {
			continue
		}

		foundAny = true
		if running == "true" {
			healthInfo := ""
			if len(parts) >= 3 && parts[2] != "none" {
				healthInfo = fmt.Sprintf(" (%s)", parts[2])
			}
			ui.Success(fmt.Sprintf("  %s: running%s", name, healthInfo))
		} else {
			ui.Error(fmt.Sprintf("  %s: stopped", name))
		}
	}

	if !foundAny {
		ui.Warning("  No Coolify containers found")
	}
}

// checkSystemResources checks CPU, memory, and disk usage
func checkSystemResources(client *StatusSSHClient) {
	ui.Info("System Resources")

	script := `# Memory
free -h 2>/dev/null | awk 'NR==2{printf "Memory: %s used / %s total\n", $3, $2}'
# Disk
df -h / 2>/dev/null | awk 'NR==2{printf "Disk (/): %s used / %s total (%s)\n", $3, $2, $5}'
# Load average
uptime 2>/dev/null | awk -F'load average:' '{print "Load: " $2}'`

	output, err := client.Execute(script)
	if err != nil {
		ui.Warning(fmt.Sprintf("  Failed to check resources: %v", err))
		return
	}

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		ui.Dim(fmt.Sprintf("  %s", strings.TrimSpace(line)))
	}
}

// StatusSSHClient is a simpler SSH client for status checks
type StatusSSHClient struct {
	Host     string
	Username string
	KeyPath  string
	Port     int
}

// NewSSHClientForStatus creates a new SSH client for status checks
func NewSSHClientForStatus(host, username, keyPath string, port int) *StatusSSHClient {
	if strings.HasPrefix(keyPath, "~/") {
		home, _ := getUserHomeDir()
		keyPath = strings.Replace(keyPath, "~", home, 1)
	}
	keyPath = strings.TrimSuffix(keyPath, ".pub")

	if port == 0 {
		port = 22
	}

	return &StatusSSHClient{
		Host:     host,
		Username: username,
		KeyPath:  keyPath,
		Port:     port,
	}
}

func getUserHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	return home, err
}

// Execute runs a command via SSH
func (s *StatusSSHClient) Execute(command string) (string, error) {
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-o", "ConnectTimeout=10",
		"-p", fmt.Sprintf("%d", s.Port),
		"-i", s.KeyPath,
		fmt.Sprintf("%s@%s", s.Username, s.Host),
		command,
	}

	cmd := exec.Command("ssh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("SSH command failed: %w", err)
	}

	return string(output), nil
}

// TestConnection tests SSH connectivity
func (s *StatusSSHClient) TestConnection() error {
	_, err := s.Execute("echo 'SSH OK'")
	return err
}

// SSHIntoInstance opens an interactive SSH session
func SSHIntoInstance() error {
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	if cfg.BareMetal.Host == "" {
		return fmt.Errorf("no host configured. Use --host flag or 'cool-kit config set baremetal.host YOUR_HOST'")
	}

	return runInteractiveSSH(cfg.BareMetal.Host, cfg.BareMetal.User, cfg.BareMetal.SSHKeyPath, cfg.BareMetal.Port)
}

// runInteractiveSSH opens an interactive SSH session
func runInteractiveSSH(host, user, keyPath string, port int) error {
	if strings.HasPrefix(keyPath, "~/") {
		home, _ := os.UserHomeDir()
		keyPath = strings.Replace(keyPath, "~", home, 1)
	}
	keyPath = strings.TrimSuffix(keyPath, ".pub")
	if port == 0 {
		port = 22
	}

	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-p", fmt.Sprintf("%d", port),
		"-i", keyPath,
		fmt.Sprintf("%s@%s", user, host),
	}

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
