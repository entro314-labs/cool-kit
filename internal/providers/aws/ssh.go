package aws

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
)

// SSHClient manages SSH connections to AWS EC2 instances
type SSHClient struct {
	Host     string
	Username string
	KeyPath  string
}

// NewSSHClient creates a new SSH client
func NewSSHClient(host, username, keyPath string) *SSHClient {
	// Expand ~ in key path
	if strings.HasPrefix(keyPath, "~/") {
		home, _ := os.UserHomeDir()
		keyPath = strings.Replace(keyPath, "~", home, 1)
	}
	// Use private key (remove .pub if present)
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
		return string(output), fmt.Errorf("SSH command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// TestConnection tests SSH connectivity
func (s *SSHClient) TestConnection() error {
	_, err := s.Execute("echo 'SSH OK'")
	return err
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

// SSHIntoInstance opens an interactive SSH session to the AWS instance
func SSHIntoInstance() error {
	// Initialize config
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Get instance IP
	status, err := getEC2Status(cfg)
	if err != nil {
		return fmt.Errorf("failed to get instance status: %w", err)
	}

	if status.PublicIP == "" {
		return fmt.Errorf("instance has no public IP address")
	}

	if status.State != "running" {
		return fmt.Errorf("instance is not running (state: %s)", status.State)
	}

	client := NewSSHClient(status.PublicIP, "ubuntu", cfg.AWS.SSHKeyPath)
	return client.Interactive()
}
