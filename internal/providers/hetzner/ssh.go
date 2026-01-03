package hetzner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
)

// SSHClient handles SSH connections to Hetzner servers
type SSHClient struct {
	Host     string
	Username string
	KeyPath  string
}

// NewSSHClient creates a new SSH client
func NewSSHClient(host, username string) *SSHClient {
	// Default to ~/.ssh/id_rsa
	keyPath := os.Getenv("HOME") + "/.ssh/id_rsa"

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
	_, err := s.Execute("echo OK")
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

// SSHIntoInstance opens an interactive SSH session to the Hetzner server
func SSHIntoInstance() error {
	// Get token
	token := os.Getenv("HCLOUD_TOKEN")
	if token == "" {
		if err := config.Initialize(); err == nil {
			cfg := config.Get()
			if cfg != nil {
				if t, ok := cfg.Settings["hetzner_token"].(string); ok {
					token = t
				}
			}
		}
	}

	if token == "" {
		return fmt.Errorf("no Hetzner Cloud token found. Set HCLOUD_TOKEN environment variable")
	}

	client, err := NewClient(token)
	if err != nil {
		return err
	}

	// Find Coolify server
	server, err := client.GetServerByLabel("application", "coolify")
	if err != nil {
		return fmt.Errorf("failed to find server: %w", err)
	}

	if server == nil {
		return fmt.Errorf("no Coolify server found. Deploy with: cool-kit hetzner deploy")
	}

	if server.Status != "running" {
		return fmt.Errorf("server is not running (status: %s)", server.Status)
	}

	if server.PublicIPv4 == "" {
		return fmt.Errorf("server has no public IPv4 address")
	}

	sshClient := NewSSHClient(server.PublicIPv4, "root")
	return sshClient.Interactive()
}

// DestroyInstance destroys the Hetzner Coolify server
func DestroyInstance() error {
	// Get token
	token := os.Getenv("HCLOUD_TOKEN")
	if token == "" {
		if err := config.Initialize(); err == nil {
			cfg := config.Get()
			if cfg != nil {
				if t, ok := cfg.Settings["hetzner_token"].(string); ok {
					token = t
				}
			}
		}
	}

	if token == "" {
		return fmt.Errorf("no Hetzner Cloud token found")
	}

	hclient, err := NewClient(token)
	if err != nil {
		return err
	}

	// Find Coolify server
	server, err := hclient.GetServerByLabel("application", "coolify")
	if err != nil {
		return fmt.Errorf("failed to find server: %w", err)
	}

	if server == nil {
		return fmt.Errorf("no Coolify server found")
	}

	// Confirm deletion
	fmt.Printf("This will delete server: %s (%s)\n", server.Name, server.PublicIPv4)
	fmt.Print("Type 'yes' to confirm: ")

	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "yes" {
		return fmt.Errorf("deletion cancelled")
	}

	if err := hclient.DeleteServer(server.ID); err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	fmt.Println("Server deleted successfully")
	return nil
}
