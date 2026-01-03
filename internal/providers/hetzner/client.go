package hetzner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// ErrServerNotFound is returned when a server cannot be found
var ErrServerNotFound = errors.New("server not found")

// Client wraps the Hetzner Cloud SDK client
type Client struct {
	hcloud *hcloud.Client
	ctx    context.Context
}

// NewClient creates a new Hetzner Cloud client
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("Hetzner Cloud token is required. Set HCLOUD_TOKEN env var or in config")
	}

	client := hcloud.NewClient(hcloud.WithToken(token))

	return &Client{
		hcloud: client,
		ctx:    context.Background(),
	}, nil
}

// ServerCreateOpts defines options for creating a server
type ServerCreateOpts struct {
	Name       string
	ServerType string
	Image      string
	Location   string
	SSHKeyIDs  []int64
	Labels     map[string]string
	UserData   string
}

// ServerInfo contains information about a Hetzner server
type ServerInfo struct {
	ID         int64
	Name       string
	Status     string
	PublicIPv4 string
	PublicIPv6 string
	ServerType string
	Location   string
	Created    time.Time
}

// CreateServer creates a new Hetzner Cloud server
func (c *Client) CreateServer(opts ServerCreateOpts) (*ServerInfo, error) {
	// Get server type
	serverType, _, err := c.hcloud.ServerType.GetByName(c.ctx, opts.ServerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get server type '%s': %w", opts.ServerType, err)
	}
	if serverType == nil {
		return nil, fmt.Errorf("server type '%s' not found", opts.ServerType)
	}

	// Get image
	image, _, err := c.hcloud.Image.GetByNameAndArchitecture(c.ctx, opts.Image, hcloud.ArchitectureX86)
	if err != nil {
		return nil, fmt.Errorf("failed to get image '%s': %w", opts.Image, err)
	}
	if image == nil {
		return nil, fmt.Errorf("image '%s' not found", opts.Image)
	}

	// Get location
	location, _, err := c.hcloud.Location.GetByName(c.ctx, opts.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to get location '%s': %w", opts.Location, err)
	}
	if location == nil {
		return nil, fmt.Errorf("location '%s' not found", opts.Location)
	}

	// Get SSH keys
	var sshKeys []*hcloud.SSHKey
	for _, keyID := range opts.SSHKeyIDs {
		key, _, err := c.hcloud.SSHKey.GetByID(c.ctx, keyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get SSH key %d: %w", keyID, err)
		}
		if key != nil {
			sshKeys = append(sshKeys, key)
		}
	}

	// Set labels
	labels := opts.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["application"] = "coolify"
	labels["managed-by"] = "cool-kit"

	// Create server
	result, _, err := c.hcloud.Server.Create(c.ctx, hcloud.ServerCreateOpts{
		Name:       opts.Name,
		ServerType: serverType,
		Image:      image,
		Location:   location,
		SSHKeys:    sshKeys,
		Labels:     labels,
		UserData:   opts.UserData,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	// Wait for server to be running
	if err := c.WaitForServer(result.Server.ID, hcloud.ServerStatusRunning); err != nil {
		return nil, fmt.Errorf("server created but failed to start: %w", err)
	}

	// Get updated server info
	server, _, err := c.hcloud.Server.GetByID(c.ctx, result.Server.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	return serverToInfo(server), nil
}

// GetServer gets a server by name
func (c *Client) GetServer(name string) (*ServerInfo, error) {
	server, _, err := c.hcloud.Server.GetByName(c.ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}
	if server == nil {
		return nil, ErrServerNotFound
	}
	return serverToInfo(server), nil
}

// GetServerByLabel finds a server by label
func (c *Client) GetServerByLabel(key, value string) (*ServerInfo, error) {
	servers, err := c.hcloud.Server.AllWithOpts(c.ctx, hcloud.ServerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: fmt.Sprintf("%s=%s", key, value),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}
	if len(servers) == 0 {
		return nil, ErrServerNotFound
	}
	return serverToInfo(servers[0]), nil
}

// WaitForServer waits for a server to reach the specified status
func (c *Client) WaitForServer(serverID int64, targetStatus hcloud.ServerStatus) error {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for server to reach status %s", targetStatus)
		case <-ticker.C:
			server, _, err := c.hcloud.Server.GetByID(c.ctx, serverID)
			if err != nil {
				return err
			}
			if server.Status == targetStatus {
				return nil
			}
		}
	}
}

// DeleteServer deletes a server
func (c *Client) DeleteServer(serverID int64) error {
	_, _, err := c.hcloud.Server.DeleteWithResult(c.ctx, &hcloud.Server{ID: serverID})
	return err
}

// ListSSHKeys lists all SSH keys in the account
func (c *Client) ListSSHKeys() ([]*hcloud.SSHKey, error) {
	keys, err := c.hcloud.SSHKey.All(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSH keys: %w", err)
	}
	return keys, nil
}

// CreateSSHKey creates a new SSH key
func (c *Client) CreateSSHKey(name, publicKey string) (*hcloud.SSHKey, error) {
	key, _, err := c.hcloud.SSHKey.Create(c.ctx, hcloud.SSHKeyCreateOpts{
		Name:      name,
		PublicKey: publicKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH key: %w", err)
	}
	return key, nil
}

// AssignFloatingIP assigns a floating IP to a server (for static IP)
func (c *Client) AssignFloatingIP(serverID int64) (string, error) {
	// Create a new floating IP
	result, _, err := c.hcloud.FloatingIP.Create(c.ctx, hcloud.FloatingIPCreateOpts{
		Type:        hcloud.FloatingIPTypeIPv4,
		Server:      &hcloud.Server{ID: serverID},
		Description: hcloud.Ptr("Coolify static IP"),
		Labels: map[string]string{
			"application": "coolify",
			"managed-by":  "cool-kit",
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create floating IP: %w", err)
	}

	return result.FloatingIP.IP.String(), nil
}

// serverToInfo converts hcloud.Server to ServerInfo
func serverToInfo(server *hcloud.Server) *ServerInfo {
	info := &ServerInfo{
		ID:         server.ID,
		Name:       server.Name,
		Status:     string(server.Status),
		ServerType: server.ServerType.Name,
		Location:   server.Datacenter.Location.Name,
		Created:    server.Created,
	}

	if server.PublicNet.IPv4.IP != nil {
		info.PublicIPv4 = server.PublicNet.IPv4.IP.String()
	}
	if server.PublicNet.IPv6.IP != nil {
		info.PublicIPv6 = server.PublicNet.IPv6.IP.String()
	}

	return info
}
