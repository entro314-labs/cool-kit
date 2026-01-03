package aws

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// StatusChecker handles status checks for AWS Coolify instances
type StatusChecker struct {
	cfg    *config.Config
	client *SSHClient
}

// EC2Status represents EC2 instance status
type EC2Status struct {
	InstanceID       string `json:"InstanceId"`
	InstanceType     string `json:"InstanceType"`
	State            string `json:"State"`
	PublicIP         string `json:"PublicIpAddress"`
	PrivateIP        string `json:"PrivateIpAddress"`
	AvailabilityZone string `json:"AvailabilityZone"`
	LaunchTime       string `json:"LaunchTime"`
}

// ContainerStatus represents a container's status
type ContainerStatus struct {
	Running bool   `json:"running"`
	Health  string `json:"health"`
	Uptime  string `json:"uptime"`
}

// ServiceStatus represents all Coolify service statuses
type ServiceStatus struct {
	Coolify  ContainerStatus `json:"coolify"`
	Database ContainerStatus `json:"database"`
	Redis    ContainerStatus `json:"redis"`
	Realtime ContainerStatus `json:"realtime"`
}

// NewStatusChecker creates a new status checker
func NewStatusChecker(cfg *config.Config, client *SSHClient) *StatusChecker {
	return &StatusChecker{
		cfg:    cfg,
		client: client,
	}
}

// CheckStatus performs comprehensive status check
func CheckStatus() error {
	ui.Section("AWS Coolify Instance Status")

	// Initialize config
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Check AWS CLI
	if err := checkAWSCLI(); err != nil {
		ui.Error(fmt.Sprintf("AWS CLI check failed: %v", err))
		return err
	}
	ui.Success("AWS CLI is configured")

	// Check EC2 instance status
	status, err := getEC2Status(cfg)
	if err != nil {
		ui.Warning(fmt.Sprintf("EC2 status check: %v", err))
		ui.Info("Tip: Make sure you have deployed an instance with 'cool-kit aws deploy'")
		return nil
	}

	displayEC2Status(status)

	// If instance is running, check SSH and services
	if status.State == "running" && status.PublicIP != "" {
		client := NewSSHClient(status.PublicIP, "ubuntu", cfg.AWS.SSHKeyPath)

		// Check SSH
		if err := client.TestConnection(); err != nil {
			ui.Warning(fmt.Sprintf("SSH connectivity: %v", err))
		} else {
			ui.Success("SSH connection available")

			// Check services
			if err := checkServices(client); err != nil {
				ui.Warning(fmt.Sprintf("Service check: %v", err))
			}

			// Check resources
			if err := checkResources(client); err != nil {
				ui.Warning(fmt.Sprintf("Resource check: %v", err))
			}
		}
	}

	return nil
}

// checkAWSCLI verifies AWS CLI is configured
func checkAWSCLI() error {
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("AWS CLI not configured: run 'aws configure'")
	}
	return nil
}

// getEC2Status gets EC2 instance status
func getEC2Status(cfg *config.Config) (*EC2Status, error) {
	// Query EC2 instances with Coolify tag
	cmd := exec.Command("aws", "ec2", "describe-instances",
		"--filters", "Name=tag:Application,Values=Coolify",
		"--query", "Reservations[0].Instances[0].{InstanceId:InstanceId,InstanceType:InstanceType,State:State.Name,PublicIpAddress:PublicIpAddress,PrivateIpAddress:PrivateIpAddress,AvailabilityZone:Placement.AvailabilityZone,LaunchTime:LaunchTime}",
		"--output", "json",
		"--region", cfg.AWS.Region,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get EC2 status: %w", err)
	}

	// Check for null response
	if strings.TrimSpace(string(output)) == "null" {
		return nil, fmt.Errorf("no Coolify instance found in region %s", cfg.AWS.Region)
	}

	var status EC2Status
	if err := json.Unmarshal(output, &status); err != nil {
		return nil, fmt.Errorf("failed to parse EC2 status: %w", err)
	}

	return &status, nil
}

// displayEC2Status displays EC2 instance information
func displayEC2Status(status *EC2Status) {
	ui.Info("EC2 Instance")

	stateDisplay := status.State
	switch status.State {
	case "running":
		ui.Success(fmt.Sprintf("  State: %s", stateDisplay))
	case "stopped", "terminated":
		ui.Error(fmt.Sprintf("  State: %s", stateDisplay))
	default:
		ui.Warning(fmt.Sprintf("  State: %s", stateDisplay))
	}

	ui.Dim(fmt.Sprintf("  Instance ID: %s", status.InstanceID))
	ui.Dim(fmt.Sprintf("  Instance Type: %s", status.InstanceType))
	ui.Dim(fmt.Sprintf("  Public IP: %s", status.PublicIP))
	ui.Dim(fmt.Sprintf("  Private IP: %s", status.PrivateIP))
	ui.Dim(fmt.Sprintf("  Availability Zone: %s", status.AvailabilityZone))
	ui.Dim(fmt.Sprintf("  Launch Time: %s", status.LaunchTime))
}

// checkServices checks Docker container status via SSH
func checkServices(client *SSHClient) error {
	ui.Info("Services")

	script := `#!/bin/bash
get_status() {
    local container=$1
    local running=$(docker inspect --format='{{.State.Running}}' $container 2>/dev/null || echo "false")
    local health=$(docker inspect --format='{{.State.Health.Status}}' $container 2>/dev/null || echo "none")
    echo "$running|$health"
}

for c in coolify coolify-db coolify-redis coolify-realtime; do
    status=$(get_status $c)
    echo "$c:$status"
done
`

	output, err := client.Execute(script)
	if err != nil {
		return fmt.Errorf("failed to check services: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		statusParts := strings.Split(parts[1], "|")
		running := statusParts[0] == "true"

		if running {
			ui.Success(fmt.Sprintf("  %s: running", name))
		} else {
			ui.Error(fmt.Sprintf("  %s: stopped", name))
		}
	}

	return nil
}

// checkResources checks system resource usage
func checkResources(client *SSHClient) error {
	ui.Info("Resources")

	script := `#!/bin/bash
# Memory
free -h | awk 'NR==2{printf "Memory: %s / %s\n", $3, $2}'
# Disk
df -h / | awk 'NR==2{printf "Disk: %s / %s (%s)\n", $3, $2, $5}'
`

	output, err := client.Execute(script)
	if err != nil {
		return fmt.Errorf("failed to check resources: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		ui.Dim(fmt.Sprintf("  %s", line))
	}

	return nil
}
