package local

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// CheckPrerequisites verifies Docker and Docker Compose are installed and running
func CheckPrerequisites() error {
	// Check if Docker command exists
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker is not installed. Please install Docker: https://docs.docker.com/get-docker/")
	}

	// Check if Docker is running by running a simple command
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not running. Please start Docker")
	}

	// Check Docker Compose
	if err := checkDockerCompose(); err != nil {
		return err
	}

	return nil
}

// checkDockerCompose verifies Docker Compose is available
func checkDockerCompose() error {
	// Try 'docker compose' (new syntax)
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err != nil {
		// Try 'docker-compose' (old syntax)
		cmd = exec.Command("docker-compose", "version")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("docker compose is not installed. Please install Docker Compose")
		}
	}
	return nil
}

// ComposeUp starts services using docker compose
func ComposeUp(workDir string, detach bool) error {
	args := []string{"compose", "up"}
	if detach {
		args = append(args, "-d")
	}

	cmd := exec.Command("docker", args...)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	return nil
}

// ComposeDown stops services using docker compose
func ComposeDown(workDir string, removeVolumes bool) error {
	args := []string{"compose", "down"}
	if removeVolumes {
		args = append(args, "-v")
	}

	cmd := exec.Command("docker", args...)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	return nil
}

// ComposeLogs shows logs from services
func ComposeLogs(workDir string, service string, follow bool, tail int) error {
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

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to show logs: %w", err)
	}

	return nil
}

// ComposePull pulls latest images
func ComposePull(workDir string) error {
	cmd := exec.Command("docker", "compose", "pull")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}

	return nil
}

// ComposeExecOutput executes a command and captures output
func ComposeExecOutput(ctx context.Context, workDir string, service string, command []string) (string, error) {
	args := []string{"compose", "exec", "-T", service}
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %w", err)
	}

	return string(output), nil
}

// CheckServiceHealth checks if a service is healthy
func CheckServiceHealth(workDir string, service string) (bool, error) {
	cmd := exec.Command("docker", "compose", "ps", service, "--format", "json")
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to check service health: %w", err)
	}

	// Simple check - if there's output and no error, service exists
	// More sophisticated health checking could parse the JSON
	return len(output) > 0, nil
}
