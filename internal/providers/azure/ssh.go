package azure

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// SSHClient manages SSH connections to Azure VMs
type SSHClient struct {
	Host     string
	Username string
	KeyPath  string
}

// NewSSHClient creates a new SSH client
func NewSSHClient(host, username, keyPath string) *SSHClient {
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
		"-i", s.KeyPath,
		fmt.Sprintf("%s@%s", s.Username, s.Host),
		command,
	}

	cmd := exec.Command("ssh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("SSH command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// ExecuteWithRetry executes a command with retries
func (s *SSHClient) ExecuteWithRetry(command string, maxRetries int, retryDelay time.Duration) (string, error) {
	var lastErr error
	var output string

	for i := 0; i < maxRetries; i++ {
		output, lastErr = s.Execute(command)
		if lastErr == nil {
			return output, nil
		}

		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	return output, fmt.Errorf("command failed after %d retries: %w", maxRetries, lastErr)
}

// CopyFile copies a local file to the remote server
func (s *SSHClient) CopyFile(localPath, remotePath string) error {
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-i", s.KeyPath,
		localPath,
		fmt.Sprintf("%s@%s:%s", s.Username, s.Host, remotePath),
	}

	cmd := exec.Command("scp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("SCP failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// CopyFileContent copies file content (as string) to remote server
func (s *SSHClient) CopyFileContent(content, remotePath string) error {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "coolify-remote-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write content to temp file
	if _, err := tmpFile.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Copy temp file to remote
	return s.CopyFile(tmpFile.Name(), remotePath)
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
		_, err := s.Execute("echo 'SSH Ready'")
		if err == nil {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("SSH did not become available within %v", maxWait)
}

// CheckDocker checks if Docker is installed and running
func (s *SSHClient) CheckDocker() (bool, error) {
	output, err := s.Execute("docker --version && docker ps")
	if err != nil {
		return false, nil
	}

	return strings.Contains(output, "Docker version"), nil
}

// InstallDocker installs Docker on the remote server
func (s *SSHClient) InstallDocker() error {
	script := `#!/bin/bash
set -e

# Install Docker
if ! command -v docker &> /dev/null; then
    echo "Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh

    # Start Docker
    systemctl start docker
    systemctl enable docker

    # Add current user to docker group
    usermod -aG docker $USER

    echo "Docker installed successfully"
else
    echo "Docker is already installed"
fi

# Verify Docker is running
docker --version
docker ps
`

	_, err := s.Execute(script)
	return err
}

// GetDockerContainerStatus gets the status of a Docker container
func (s *SSHClient) GetDockerContainerStatus(containerName string) (string, error) {
	cmd := fmt.Sprintf("docker inspect --format='{{.State.Status}}' %s 2>/dev/null || echo 'not found'", containerName)
	output, err := s.Execute(cmd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}
