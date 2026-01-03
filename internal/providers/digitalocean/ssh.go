package digitalocean

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// SSHClient handles SSH connections to DigitalOcean droplets
type SSHClient struct {
	Host     string
	Username string
	KeyPath  string
}

// NewSSHClient creates a new SSH client
func NewSSHClient(host, username string) *SSHClient {
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

// SSHIntoInstance opens an interactive SSH session to the DigitalOcean droplet
func SSHIntoInstance() error {
	token := getToken()
	if token == "" {
		return fmt.Errorf("no DigitalOcean token found. Set DIGITALOCEAN_TOKEN environment variable")
	}

	client, err := NewClient(token)
	if err != nil {
		return err
	}

	droplet, err := client.GetDropletByTag("coolify")
	if err != nil {
		return fmt.Errorf("failed to find droplet: %w", err)
	}

	if droplet == nil {
		return fmt.Errorf("no Coolify droplet found. Deploy with: cool-kit digitalocean deploy")
	}

	if droplet.Status != "active" {
		return fmt.Errorf("droplet is not active (status: %s)", droplet.Status)
	}

	if droplet.PublicIP == "" {
		return fmt.Errorf("droplet has no public IP address")
	}

	sshClient := NewSSHClient(droplet.PublicIP, "root")
	return sshClient.Interactive()
}

// DestroyInstance destroys the DigitalOcean Coolify droplet
func DestroyInstance() error {
	token := getToken()
	if token == "" {
		return fmt.Errorf("no DigitalOcean token found")
	}

	doclient, err := NewClient(token)
	if err != nil {
		return err
	}

	droplet, err := doclient.GetDropletByTag("coolify")
	if err != nil {
		return fmt.Errorf("failed to find droplet: %w", err)
	}

	if droplet == nil {
		return fmt.Errorf("no Coolify droplet found")
	}

	fmt.Printf("This will delete droplet: %s (%s)\n", droplet.Name, droplet.PublicIP)
	fmt.Print("Type 'yes' to confirm: ")

	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "yes" {
		return fmt.Errorf("deletion cancelled")
	}

	if err := doclient.DeleteDroplet(droplet.ID); err != nil {
		return fmt.Errorf("failed to delete droplet: %w", err)
	}

	fmt.Println("Droplet deleted successfully")
	return nil
}
