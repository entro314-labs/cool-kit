package azure

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/entro314-labs/cool-kit/internal/ui"
)

// StatusChecker handles status checks for Azure Coolify instances
type StatusChecker struct {
	ctx    *DeploymentContext
	client *SSHClient
}

// NewStatusChecker creates a new status checker
func NewStatusChecker(ctx *DeploymentContext, client *SSHClient) *StatusChecker {
	return &StatusChecker{
		ctx:    ctx,
		client: client,
	}
}

// CheckStatus performs comprehensive status check
func (s *StatusChecker) CheckStatus() error {
	ui.Section("Azure Coolify Instance Status")

	// Check VM status
	vmStatus, err := s.checkVMStatus()
	if err != nil {
		ui.Error(fmt.Sprintf("VM status check failed: %v", err))
	} else {
		s.displayVMStatus(vmStatus)
	}

	// Check SSH connectivity
	if err := s.checkSSH(); err != nil {
		ui.Error(fmt.Sprintf("SSH connectivity failed: %v", err))
		return err
	}

	// Check services
	serviceStatus, err := s.checkServices()
	if err != nil {
		ui.Error(fmt.Sprintf("Service status check failed: %v", err))
	} else {
		s.displayServiceStatus(serviceStatus)
	}

	// Check resource usage
	if err := s.checkResources(); err != nil {
		ui.Warning(fmt.Sprintf("Resource check failed: %v", err))
	}

	// Check Coolify version
	if err := s.checkCoolifyVersion(); err != nil {
		ui.Warning(fmt.Sprintf("Version check failed: %v", err))
	}

	return nil
}

// checkVMStatus checks Azure VM status using Azure CLI
func (s *StatusChecker) checkVMStatus() (*VMStatus, error) {
	ui.Info("Checking VM status")

	// Get VM details
	cmd := exec.Command("az", "vm", "show",
		"--resource-group", s.ctx.ResourceGroup,
		"--name", s.ctx.VMName,
		"--show-details",
		"--query", "{powerState:powerState,provisioningState:provisioningState,location:location,vmSize:hardwareProfile.vmSize,publicIP:publicIps,privateIP:privateIps}",
		"--output", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get VM status: %w", err)
	}

	var status VMStatus
	if err := json.Unmarshal(output, &status); err != nil {
		return nil, fmt.Errorf("failed to parse VM status: %w", err)
	}

	return &status, nil
}

// displayVMStatus displays VM status information
func (s *StatusChecker) displayVMStatus(status *VMStatus) {
	ui.Info("Virtual Machine")

	ui.Info(fmt.Sprintf("Power State: %s", status.PowerState))
	ui.Dim(fmt.Sprintf("Provisioning: %s", status.ProvisioningState))
	ui.Dim(fmt.Sprintf("VM Size: %s", status.VMSize))
	ui.Dim(fmt.Sprintf("Location: %s", status.Location))
	ui.Dim(fmt.Sprintf("Public IP: %s", status.PublicIP))
	ui.Dim(fmt.Sprintf("Private IP: %s", status.PrivateIP))
}

// checkSSH verifies SSH connectivity
func (s *StatusChecker) checkSSH() error {
	ui.Info("Checking SSH connectivity")

	_, err := s.client.Execute("echo 'SSH OK'")
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	ui.Dim("SSH connection is working")
	return nil
}

// checkServices checks Docker container status
func (s *StatusChecker) checkServices() (*ServiceStatus, error) {
	ui.Info("Checking service status")

	script := `#!/bin/bash
cd /data/coolify 2>/dev/null || exit 1

# Get container status
get_status() {
    local container=$1
    local running=$(docker inspect --format='{{.State.Running}}' $container 2>/dev/null || echo "false")
    local health=$(docker inspect --format='{{.State.Health.Status}}' $container 2>/dev/null || echo "none")
    local uptime=$(docker inspect --format='{{.State.StartedAt}}' $container 2>/dev/null || echo "unknown")

    echo "{\"running\":$running,\"health\":\"$health\",\"uptime\":\"$uptime\"}"
}

echo "{"
echo "  \"coolify\": $(get_status coolify),"
echo "  \"database\": $(get_status coolify-db),"
echo "  \"redis\": $(get_status coolify-redis),"
echo "  \"realtime\": $(get_status coolify-realtime)"
echo "}"
`

	output, err := s.client.Execute(script)
	if err != nil {
		return nil, fmt.Errorf("failed to check services: %w", err)
	}

	var status ServiceStatus
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		return nil, fmt.Errorf("failed to parse service status: %w", err)
	}

	return &status, nil
}

// displayServiceStatus displays service status information
func (s *StatusChecker) displayServiceStatus(status *ServiceStatus) {
	ui.Info("Services")

	services := []struct {
		name   string
		status ContainerStatus
	}{
		{"Coolify", status.Coolify},
		{"Database (PostgreSQL)", status.Database},
		{"Redis", status.Redis},
		{"Realtime (Soketi)", status.Realtime},
	}

	for _, svc := range services {
		statusStr := "Stopped"
		statusIcon := "✗"

		if svc.status.Running {
			statusStr = "Running"
			statusIcon = "✓"

			if svc.status.Health != "none" && svc.status.Health != "" {
				statusStr = fmt.Sprintf("Running (%s)", svc.status.Health)
			}
		}

		ui.Info(fmt.Sprintf("%s %s: %s", statusIcon, svc.name, statusStr))
		if svc.status.Uptime != "" && svc.status.Uptime != "unknown" {
			ui.Dim(fmt.Sprintf("  Started: %s", svc.status.Uptime))
		}
	}
}

// checkResources checks system resource usage
func (s *StatusChecker) checkResources() error {
	ui.Info("Checking resource usage")

	script := `#!/bin/bash
# CPU and Memory
echo "=== System Resources ==="
top -bn1 | grep "Cpu(s)" | awk '{print "CPU Usage: " $2}' || echo "CPU: N/A"
free -h | awk 'NR==2{printf "Memory: %s / %s (%.2f%%)\n", $3,$2,$3*100/$2}' || echo "Memory: N/A"

# Disk usage
echo ""
echo "=== Disk Usage ==="
df -h / | awk 'NR==2{printf "Root: %s / %s (%s)\n", $3,$2,$5}' || echo "Disk: N/A"
df -h /data/coolify 2>/dev/null | awk 'NR==2{printf "Coolify Data: %s / %s (%s)\n", $3,$2,$5}' || echo "Coolify Data: Using root partition"

# Docker disk usage
echo ""
echo "=== Docker Resources ==="
docker system df 2>/dev/null || echo "Docker stats: N/A"
`

	output, err := s.client.Execute(script)
	if err != nil {
		return fmt.Errorf("failed to check resources: %w", err)
	}

	ui.Info("Resources")
	fmt.Print(output)

	return nil
}

// checkCoolifyVersion checks installed Coolify version
func (s *StatusChecker) checkCoolifyVersion() error {
	ui.Info("Checking Coolify version")

	script := `#!/bin/bash
cd /data/coolify
VERSION=$(docker compose exec -T coolify php artisan --version 2>/dev/null | awk '{print $NF}' || echo "unknown")
echo "Installed Version: $VERSION"

# Check for updates (if possible)
LATEST=$(docker inspect ghcr.io/coollabsio/coolify:latest 2>/dev/null | jq -r '.[0].RepoDigests[0]' || echo "unknown")
echo "Latest Image: $LATEST"
`

	output, err := s.client.Execute(script)
	if err != nil {
		return fmt.Errorf("failed to check version: %w", err)
	}

	ui.Info("Version")
	fmt.Print(output)

	return nil
}

// GetQuickStatus returns a quick status summary
func (s *StatusChecker) GetQuickStatus() (string, error) {
	// Check if VM is running
	vmStatus, err := s.checkVMStatus()
	if err != nil {
		return "Unknown", err
	}

	if !strings.Contains(strings.ToLower(vmStatus.PowerState), "running") {
		return "VM Stopped", nil
	}

	// Check if SSH is available
	if err := s.checkSSH(); err != nil {
		return "VM Running (SSH unavailable)", nil
	}

	// Check if Coolify is running
	serviceStatus, err := s.checkServices()
	if err != nil {
		return "VM Running (Services unknown)", nil
	}

	if !serviceStatus.Coolify.Running {
		return "VM Running (Coolify stopped)", nil
	}

	return "Operational", nil
}
