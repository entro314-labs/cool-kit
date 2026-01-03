package gcp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// InstanceStatus represents GCP Compute Engine instance status
type InstanceStatus struct {
	Name         string `json:"name"`
	Zone         string `json:"zone"`
	MachineType  string `json:"machineType"`
	Status       string `json:"status"`
	NetworkIP    string `json:"networkInterfaces[0].networkIP"`
	ExternalIP   string `json:"networkInterfaces[0].accessConfigs[0].natIP"`
	CreationTime string `json:"creationTimestamp"`
}

// CheckStatus checks the status of GCP Coolify deployment
func CheckStatus() error {
	ui.Section("GCP Coolify Instance Status")

	// Initialize config
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Check gcloud CLI
	if err := checkGCloudCLI(); err != nil {
		ui.Error(fmt.Sprintf("gcloud CLI check failed: %v", err))
		return err
	}
	ui.Success("gcloud CLI is configured")

	// Get current project
	project := cfg.GCP.Project
	if project == "" {
		// Try to get from gcloud config
		cmd := exec.Command("gcloud", "config", "get-value", "project")
		output, err := cmd.Output()
		if err == nil {
			project = strings.TrimSpace(string(output))
		}
	}
	if project == "" {
		ui.Warning("No GCP project configured. Use 'gcloud config set project PROJECT_ID' or set in cool-kit config")
		return nil
	}

	ui.Info(fmt.Sprintf("Project: %s", project))

	// Check for Coolify instance
	instance, err := getInstanceStatus(project, cfg.GCP.Zone)
	if err != nil {
		ui.Warning(fmt.Sprintf("Instance check: %v", err))
		ui.Info("Tip: Make sure you have deployed an instance with 'cool-kit gcp deploy'")
		return nil
	}

	displayInstanceStatus(instance)

	// If running, check SSH and services
	if instance.Status == "RUNNING" && instance.ExternalIP != "" {
		client := NewSSHClient(instance.ExternalIP, "coolify", cfg.GCP.SSHKeyPath)

		if err := client.TestConnection(); err != nil {
			ui.Warning(fmt.Sprintf("SSH connectivity: %v", err))
		} else {
			ui.Success("SSH connection available")
			checkServices(client)
			checkResources(client)
		}
	}

	return nil
}

// checkGCloudCLI verifies gcloud is configured
func checkGCloudCLI() error {
	cmd := exec.Command("gcloud", "auth", "list", "--format", "value(account)")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("gcloud CLI not configured: run 'gcloud auth login'")
	}
	if strings.TrimSpace(string(output)) == "" {
		return fmt.Errorf("no gcloud account configured")
	}
	return nil
}

// getInstanceStatus gets Compute Engine instance status
func getInstanceStatus(project, zone string) (*InstanceStatus, error) {
	// List instances with Coolify label
	cmd := exec.Command("gcloud", "compute", "instances", "list",
		"--project", project,
		"--filter", "labels.application=coolify",
		"--format", "json",
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	var instances []map[string]interface{}
	if err := json.Unmarshal(output, &instances); err != nil {
		return nil, fmt.Errorf("failed to parse instances: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no Coolify instance found")
	}

	// Get first instance
	inst := instances[0]
	status := &InstanceStatus{
		Name:   getString(inst, "name"),
		Zone:   getString(inst, "zone"),
		Status: getString(inst, "status"),
	}

	// Extract IPs from nested structure
	if networkInterfaces, ok := inst["networkInterfaces"].([]interface{}); ok && len(networkInterfaces) > 0 {
		if ni, ok := networkInterfaces[0].(map[string]interface{}); ok {
			status.NetworkIP = getString(ni, "networkIP")
			if accessConfigs, ok := ni["accessConfigs"].([]interface{}); ok && len(accessConfigs) > 0 {
				if ac, ok := accessConfigs[0].(map[string]interface{}); ok {
					status.ExternalIP = getString(ac, "natIP")
				}
			}
		}
	}

	if machineType, ok := inst["machineType"].(string); ok {
		parts := strings.Split(machineType, "/")
		status.MachineType = parts[len(parts)-1]
	}

	return status, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// displayInstanceStatus displays instance information
func displayInstanceStatus(status *InstanceStatus) {
	ui.Info("Compute Engine Instance")

	switch status.Status {
	case "RUNNING":
		ui.Success(fmt.Sprintf("  Status: %s", status.Status))
	case "TERMINATED", "STOPPED":
		ui.Error(fmt.Sprintf("  Status: %s", status.Status))
	default:
		ui.Warning(fmt.Sprintf("  Status: %s", status.Status))
	}

	ui.Dim(fmt.Sprintf("  Name: %s", status.Name))
	ui.Dim(fmt.Sprintf("  Machine Type: %s", status.MachineType))
	ui.Dim(fmt.Sprintf("  External IP: %s", status.ExternalIP))
	ui.Dim(fmt.Sprintf("  Internal IP: %s", status.NetworkIP))
}

// checkServices checks Docker container status via SSH
func checkServices(client *SSHClient) {
	ui.Info("Services")

	script := `for c in coolify coolify-db coolify-redis coolify-realtime; do
    status=$(docker inspect --format='{{.State.Running}}' $c 2>/dev/null || echo "false")
    echo "$c:$status"
done`

	output, err := client.Execute(script)
	if err != nil {
		ui.Warning(fmt.Sprintf("  Failed to check services: %v", err))
		return
	}

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		if parts[1] == "true" {
			ui.Success(fmt.Sprintf("  %s: running", parts[0]))
		} else {
			ui.Error(fmt.Sprintf("  %s: stopped", parts[0]))
		}
	}
}

// checkResources checks system resource usage
func checkResources(client *SSHClient) {
	ui.Info("Resources")

	script := `free -h | awk 'NR==2{printf "Memory: %s / %s\n", $3, $2}'
df -h / | awk 'NR==2{printf "Disk: %s / %s (%s)\n", $3, $2, $5}'`

	output, err := client.Execute(script)
	if err != nil {
		return
	}

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		ui.Dim(fmt.Sprintf("  %s", line))
	}
}

// SSHClient manages SSH connections to GCP instances
type SSHClient struct {
	Host     string
	Username string
	KeyPath  string
}

// NewSSHClient creates a new SSH client
func NewSSHClient(host, username, keyPath string) *SSHClient {
	if strings.HasPrefix(keyPath, "~/") {
		home, _ := os.UserHomeDir()
		keyPath = strings.Replace(keyPath, "~", home, 1)
	}
	keyPath = strings.TrimSuffix(keyPath, ".pub")

	return &SSHClient{
		Host:     host,
		Username: username,
		KeyPath:  keyPath,
	}
}

// Execute runs a command on the remote server
func (s *SSHClient) Execute(command string) (string, error) {
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-o", "ConnectTimeout=10",
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
func (s *SSHClient) TestConnection() error {
	_, err := s.Execute("echo 'SSH OK'")
	return err
}

// Interactive opens an interactive SSH session
func (s *SSHClient) Interactive() error {
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-i", s.KeyPath,
		fmt.Sprintf("%s@%s", s.Username, s.Host),
	}

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// WaitForSSH waits for SSH to become available
func (s *SSHClient) WaitForSSH(maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		if err := s.TestConnection(); err == nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("SSH did not become available within %v", maxWait)
}

// SSHIntoInstance opens an interactive SSH session to the GCP instance
func SSHIntoInstance() error {
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	project := cfg.GCP.Project
	if project == "" {
		cmd := exec.Command("gcloud", "config", "get-value", "project")
		output, err := cmd.Output()
		if err == nil {
			project = strings.TrimSpace(string(output))
		}
	}

	if project == "" {
		return fmt.Errorf("no GCP project configured")
	}

	instance, err := getInstanceStatus(project, cfg.GCP.Zone)
	if err != nil {
		return fmt.Errorf("failed to get instance status: %w", err)
	}

	if instance.ExternalIP == "" {
		return fmt.Errorf("instance has no external IP address")
	}

	if instance.Status != "RUNNING" {
		return fmt.Errorf("instance is not running (status: %s)", instance.Status)
	}

	client := NewSSHClient(instance.ExternalIP, "coolify", cfg.GCP.SSHKeyPath)
	return client.Interactive()
}
