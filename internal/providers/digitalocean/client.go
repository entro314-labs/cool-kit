package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

// ErrDropletNotFound is returned when a droplet cannot be found
var ErrDropletNotFound = errors.New("droplet not found")

// Client wraps the DigitalOcean SDK client
type Client struct {
	godo *godo.Client
	ctx  context.Context
}

// tokenSource implements oauth2.TokenSource for DigitalOcean
type tokenSource struct {
	AccessToken string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: t.AccessToken,
	}, nil
}

// NewClient creates a new DigitalOcean client
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("DigitalOcean token is required. Set DIGITALOCEAN_TOKEN env var or in config")
	}

	ts := &tokenSource{AccessToken: token}
	oauthClient := oauth2.NewClient(context.Background(), ts)
	client := godo.NewClient(oauthClient)

	return &Client{
		godo: client,
		ctx:  context.Background(),
	}, nil
}

// DropletCreateOpts defines options for creating a droplet
type DropletCreateOpts struct {
	Name            string
	Region          string
	Size            string
	Image           string
	SSHFingerprints []string
	Tags            []string
	UserData        string
}

// DropletInfo contains information about a DigitalOcean droplet
type DropletInfo struct {
	ID       int
	Name     string
	Status   string
	PublicIP string
	Region   string
	Size     string
	Image    string
	Created  time.Time
}

// CreateDroplet creates a new DigitalOcean droplet
func (c *Client) CreateDroplet(opts DropletCreateOpts) (*DropletInfo, error) {
	// Build SSH keys from fingerprints
	var sshKeys []godo.DropletCreateSSHKey
	for _, fp := range opts.SSHFingerprints {
		sshKeys = append(sshKeys, godo.DropletCreateSSHKey{Fingerprint: fp})
	}

	// Add default tags
	tags := append(opts.Tags, "coolify", "managed-by-cool-kit")

	createRequest := &godo.DropletCreateRequest{
		Name:   opts.Name,
		Region: opts.Region,
		Size:   opts.Size,
		Image: godo.DropletCreateImage{
			Slug: opts.Image,
		},
		SSHKeys:  sshKeys,
		Tags:     tags,
		UserData: opts.UserData,
	}

	droplet, _, err := c.godo.Droplets.Create(c.ctx, createRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create droplet: %w", err)
	}

	// Wait for droplet to be active
	if err := c.WaitForDroplet(droplet.ID, "active"); err != nil {
		return nil, fmt.Errorf("droplet created but not active: %w", err)
	}

	// Get updated droplet info with IP
	droplet, _, err = c.godo.Droplets.Get(c.ctx, droplet.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet info: %w", err)
	}

	return dropletToInfo(droplet), nil
}

// GetDroplet gets a droplet by ID
func (c *Client) GetDroplet(id int) (*DropletInfo, error) {
	droplet, _, err := c.godo.Droplets.Get(c.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet: %w", err)
	}
	return dropletToInfo(droplet), nil
}

// GetDropletByTag finds a droplet by tag
func (c *Client) GetDropletByTag(tag string) (*DropletInfo, error) {
	droplets, _, err := c.godo.Droplets.ListByTag(c.ctx, tag, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list droplets: %w", err)
	}
	if len(droplets) == 0 {
		return nil, ErrDropletNotFound
	}
	return dropletToInfo(&droplets[0]), nil
}

// WaitForDroplet waits for a droplet to reach the specified status
func (c *Client) WaitForDroplet(dropletID int, targetStatus string) error {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for droplet to reach status %s", targetStatus)
		case <-ticker.C:
			droplet, _, err := c.godo.Droplets.Get(c.ctx, dropletID)
			if err != nil {
				return err
			}
			if droplet.Status == targetStatus {
				return nil
			}
		}
	}
}

// DeleteDroplet deletes a droplet
func (c *Client) DeleteDroplet(dropletID int) error {
	_, err := c.godo.Droplets.Delete(c.ctx, dropletID)
	return err
}

// ListSSHKeys lists all SSH keys in the account
func (c *Client) ListSSHKeys() ([]godo.Key, error) {
	keys, _, err := c.godo.Keys.List(c.ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSH keys: %w", err)
	}
	return keys, nil
}

// CreateSSHKey creates a new SSH key
func (c *Client) CreateSSHKey(name, publicKey string) (*godo.Key, error) {
	key, _, err := c.godo.Keys.Create(c.ctx, &godo.KeyCreateRequest{
		Name:      name,
		PublicKey: publicKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH key: %w", err)
	}
	return key, nil
}

// AssignFloatingIP creates and assigns a floating IP
func (c *Client) AssignFloatingIP(dropletID int, region string) (string, error) {
	fip, _, err := c.godo.FloatingIPs.Create(c.ctx, &godo.FloatingIPCreateRequest{
		Region:    region,
		DropletID: dropletID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create floating IP: %w", err)
	}

	return fip.IP, nil
}

// GetAccount gets account information (for validation)
func (c *Client) GetAccount() (*godo.Account, error) {
	account, _, err := c.godo.Account.Get(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return account, nil
}

// dropletToInfo converts godo.Droplet to DropletInfo
func dropletToInfo(droplet *godo.Droplet) *DropletInfo {
	info := &DropletInfo{
		ID:     droplet.ID,
		Name:   droplet.Name,
		Status: droplet.Status,
		Region: droplet.Region.Slug,
		Size:   droplet.Size.Slug,
	}

	// Parse created time from string
	if droplet.Created != "" {
		if t, err := time.Parse(time.RFC3339, droplet.Created); err == nil {
			info.Created = t
		}
	}

	if droplet.Image != nil {
		info.Image = droplet.Image.Slug
	}

	// Get public IPv4
	for _, network := range droplet.Networks.V4 {
		if network.Type == "public" {
			info.PublicIP = network.IPAddress
			break
		}
	}

	return info
}
