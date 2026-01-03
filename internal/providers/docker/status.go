package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// CheckStatus checks the status of Docker Coolify deployment
func CheckStatus() error {
	ui.Section("Docker Coolify Status")

	// Check Docker daemon
	if err := checkDockerDaemon(); err != nil {
		ui.Error(fmt.Sprintf("Docker daemon: %v", err))
		return err
	}
	ui.Success("Docker daemon is running")

	// Check Docker Compose
	if err := checkDockerCompose(); err != nil {
		ui.Warning(fmt.Sprintf("Docker Compose: %v", err))
	} else {
		ui.Success("Docker Compose is available")
	}

	// Check containers
	if err := checkContainers(); err != nil {
		ui.Warning(fmt.Sprintf("Container check: %v", err))
	}

	// Check resource usage
	if err := checkResources(); err != nil {
		ui.Warning(fmt.Sprintf("Resource check: %v", err))
	}

	return nil
}

// checkDockerDaemon verifies Docker daemon is running
func checkDockerDaemon() error {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker daemon not running")
	}
	return nil
}

// checkDockerCompose verifies Docker Compose is available
func checkDockerCompose() error {
	// Try new docker compose
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Try legacy docker-compose
	cmd = exec.Command("docker-compose", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker Compose not found")
	}
	return nil
}

// checkContainers checks Coolify container status
func checkContainers() error {
	ui.Info("Container Status")

	containers := []string{"coolify", "coolify-db", "coolify-redis", "coolify-realtime"}

	for _, container := range containers {
		cmd := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", container)
		output, err := cmd.Output()
		if err != nil {
			ui.Dim(fmt.Sprintf("  %s: not found", container))
			continue
		}

		status := strings.TrimSpace(string(output))
		switch status {
		case "running":
			ui.Success(fmt.Sprintf("  %s: %s", container, status))
		case "exited", "dead":
			ui.Error(fmt.Sprintf("  %s: %s", container, status))
		default:
			ui.Warning(fmt.Sprintf("  %s: %s", container, status))
		}
	}

	return nil
}

// checkResources displays Docker resource usage
func checkResources() error {
	ui.Info("Resource Usage")

	cmd := exec.Command("docker", "system", "df", "--format", "{{.Type}}\t{{.Size}}\t{{.Reclaimable}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get Docker stats: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		ui.Dim(fmt.Sprintf("  %s", line))
	}

	return nil
}

// ViewLogs displays logs from Docker services
func ViewLogs(service string, follow bool, tail int) error {
	cfg := config.Get()
	workDir := ""
	if cfg != nil && cfg.Local.WorkDir != "" {
		workDir = cfg.Local.WorkDir
	} else {
		workDir = "./coolify-source"
	}

	args := []string{"compose", "logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}
	if service != "" {
		args = append(args, service)
	}

	cmd := exec.Command("docker", args...)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// StopServices stops Docker Coolify services
func StopServices() error {
	ui.Info("Stopping Docker services...")

	cfg := config.Get()
	workDir := ""
	if cfg != nil && cfg.Local.WorkDir != "" {
		workDir = cfg.Local.WorkDir
	} else {
		workDir = "./coolify-source"
	}

	cmd := exec.Command("docker", "compose", "stop")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	ui.Success("Services stopped")
	return nil
}

// StartServices starts Docker Coolify services
func StartServices() error {
	ui.Info("Starting Docker services...")

	cfg := config.Get()
	workDir := ""
	if cfg != nil && cfg.Local.WorkDir != "" {
		workDir = cfg.Local.WorkDir
	} else {
		workDir = "./coolify-source"
	}

	cmd := exec.Command("docker", "compose", "start")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	ui.Success("Services started")
	return nil
}
