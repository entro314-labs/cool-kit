package gcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
)

// ErrInstanceNotFound is returned when an instance cannot be found
var ErrInstanceNotFound = errors.New("instance not found")

// SDKClient wraps the GCP Compute SDK client
type SDKClient struct {
	instances *compute.InstancesClient
	project   string
	zone      string
	ctx       context.Context
}

// NewSDKClient creates a new GCP SDK client
func NewSDKClient(project, zone string) (*SDKClient, error) {
	ctx := context.Background()

	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create instances client: %w", err)
	}

	return &SDKClient{
		instances: instancesClient,
		project:   project,
		zone:      zone,
		ctx:       ctx,
	}, nil
}

// Close closes the client
func (c *SDKClient) Close() {
	if c.instances != nil {
		c.instances.Close()
	}
}

// InstanceCreateOpts defines options for creating a Compute Engine instance
type VMCreateOpts struct {
	Name          string
	MachineType   string
	ImageFamily   string
	ImageProject  string
	DiskSizeGB    int64
	Tags          []string
	Labels        map[string]string
	StartupScript string
}

// VMInfo contains information about a Compute Engine instance
type VMInfo struct {
	Name        string
	Zone        string
	Status      string
	MachineType string
	ExternalIP  string
	InternalIP  string
	Created     time.Time
}

// CreateInstance creates a new Compute Engine instance
func (c *SDKClient) CreateInstance(opts VMCreateOpts) (*VMInfo, error) {
	machineType := fmt.Sprintf("zones/%s/machineTypes/%s", c.zone, opts.MachineType)
	sourceImage := fmt.Sprintf("projects/%s/global/images/family/%s", opts.ImageProject, opts.ImageFamily)

	diskSizeGB := opts.DiskSizeGB
	if diskSizeGB == 0 {
		diskSizeGB = 30
	}

	// Build labels
	labels := opts.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["application"] = "coolify"
	labels["managed-by"] = "cool-kit"

	// Build network tags
	tags := opts.Tags
	if len(tags) == 0 {
		tags = []string{"http-server", "https-server"}
	}

	instance := &computepb.Instance{
		Name:        strPtr(opts.Name),
		MachineType: strPtr(machineType),
		Labels:      labels,
		Tags: &computepb.Tags{
			Items: tags,
		},
		Disks: []*computepb.AttachedDisk{
			{
				Boot:       boolPtr(true),
				AutoDelete: boolPtr(true),
				InitializeParams: &computepb.AttachedDiskInitializeParams{
					SourceImage: strPtr(sourceImage),
					DiskSizeGb:  int64Ptr(diskSizeGB),
					DiskType:    strPtr(fmt.Sprintf("zones/%s/diskTypes/pd-balanced", c.zone)),
				},
			},
		},
		NetworkInterfaces: []*computepb.NetworkInterface{
			{
				AccessConfigs: []*computepb.AccessConfig{
					{
						Name: strPtr("External NAT"),
						Type: strPtr("ONE_TO_ONE_NAT"),
					},
				},
			},
		},
	}

	// Add startup script if provided
	if opts.StartupScript != "" {
		instance.Metadata = &computepb.Metadata{
			Items: []*computepb.Items{
				{
					Key:   strPtr("startup-script"),
					Value: strPtr(opts.StartupScript),
				},
			},
		}
	}

	req := &computepb.InsertInstanceRequest{
		Project:          c.project,
		Zone:             c.zone,
		InstanceResource: instance,
	}

	op, err := c.instances.Insert(c.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Wait for operation to complete
	if err := op.Wait(c.ctx); err != nil {
		return nil, fmt.Errorf("failed waiting for instance creation: %w", err)
	}

	// Wait for instance to be running
	if err := c.WaitForInstance(opts.Name, "RUNNING"); err != nil {
		return nil, fmt.Errorf("instance created but not running: %w", err)
	}

	return c.GetInstance(opts.Name)
}

// GetInstance gets an instance by name
func (c *SDKClient) GetInstance(name string) (*VMInfo, error) {
	req := &computepb.GetInstanceRequest{
		Project:  c.project,
		Zone:     c.zone,
		Instance: name,
	}

	instance, err := c.instances.Get(c.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	return instanceToVMInfo(instance), nil
}

// GetCoolifyInstance finds a Coolify-labeled instance
func (c *SDKClient) GetCoolifyInstance() (*VMInfo, error) {
	req := &computepb.ListInstancesRequest{
		Project: c.project,
		Zone:    c.zone,
		Filter:  strPtr("labels.application=coolify"),
	}

	it := c.instances.List(c.ctx, req)

	for {
		instance, err := it.Next()
		if err != nil {
			break
		}
		if instance.GetStatus() != "TERMINATED" {
			return instanceToVMInfo(instance), nil
		}
	}

	return nil, ErrInstanceNotFound
}

// WaitForInstance waits for an instance to reach the specified status
func (c *SDKClient) WaitForInstance(name, targetStatus string) error {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for instance to reach status %s", targetStatus)
		case <-ticker.C:
			info, err := c.GetInstance(name)
			if err != nil {
				return err
			}
			if info.Status == targetStatus {
				return nil
			}
		}
	}
}

// DeleteInstance deletes an instance
func (c *SDKClient) DeleteInstance(name string) error {
	req := &computepb.DeleteInstanceRequest{
		Project:  c.project,
		Zone:     c.zone,
		Instance: name,
	}

	op, err := c.instances.Delete(c.ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	return op.Wait(c.ctx)
}

// ValidateCredentials validates GCP credentials by attempting to list instances
func (c *SDKClient) ValidateCredentials() error {
	req := &computepb.ListInstancesRequest{
		Project: c.project,
		Zone:    c.zone,
	}

	it := c.instances.List(c.ctx, req)
	_, err := it.Next()
	if err != nil && !strings.Contains(err.Error(), "no more items") {
		return fmt.Errorf("credential validation failed: %w", err)
	}
	return nil
}

// Helper functions
func instanceToVMInfo(instance *computepb.Instance) *VMInfo {
	info := &VMInfo{
		Name:   instance.GetName(),
		Zone:   instance.GetZone(),
		Status: instance.GetStatus(),
	}

	// Extract machine type name
	if mt := instance.GetMachineType(); mt != "" {
		parts := strings.Split(mt, "/")
		info.MachineType = parts[len(parts)-1]
	}

	// Get external IP
	for _, ni := range instance.GetNetworkInterfaces() {
		info.InternalIP = ni.GetNetworkIP()
		for _, ac := range ni.GetAccessConfigs() {
			if ip := ac.GetNatIP(); ip != "" {
				info.ExternalIP = ip
				break
			}
		}
	}

	return info
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}
