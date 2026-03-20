package aws

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// ErrInstanceNotFound is returned when an instance cannot be found
var ErrInstanceNotFound = errors.New("instance not found")

// SDKClient wraps the AWS SDK v2 client
type SDKClient struct {
	ec2    *ec2.Client
	ctx    context.Context
	region string
}

// NewSDKClient creates a new AWS SDK client
func NewSDKClient(region string) (*SDKClient, error) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &SDKClient{
		ec2:    ec2.NewFromConfig(cfg),
		ctx:    ctx,
		region: region,
	}, nil
}

// InstanceCreateOpts defines options for creating an EC2 instance
type InstanceCreateOpts struct {
	Name           string
	InstanceType   string
	AMI            string
	KeyName        string
	SecurityGroups []string
	SubnetID       string
	DiskSizeGB     int32
	UserData       string
}

// InstanceInfo contains information about an EC2 instance
type InstanceInfo struct {
	InstanceID       string
	Name             string
	State            string
	InstanceType     string
	PublicIP         string
	PrivateIP        string
	AvailabilityZone string
	LaunchTime       time.Time
}

// CreateInstance creates a new EC2 instance
func (c *SDKClient) CreateInstance(opts InstanceCreateOpts) (*InstanceInfo, error) {
	// Build tag specifications
	tags := []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeInstance,
			Tags: []types.Tag{
				{Key: new("Name"), Value: new(opts.Name)},
				{Key: new("Application"), Value: new("Coolify")},
				{Key: new("ManagedBy"), Value: new("cool-kit")},
			},
		},
	}

	// Build input
	input := &ec2.RunInstancesInput{
		ImageId:           new(opts.AMI),
		InstanceType:      types.InstanceType(opts.InstanceType),
		MinCount:          new(int32(1)),
		MaxCount:          new(int32(1)),
		KeyName:           new(opts.KeyName),
		TagSpecifications: tags,
	}

	// Add security groups if provided
	if len(opts.SecurityGroups) > 0 {
		input.SecurityGroupIds = opts.SecurityGroups
	}

	// Add subnet if provided
	if opts.SubnetID != "" {
		input.SubnetId = new(opts.SubnetID)
	}

	// Add block device mapping for disk size
	if opts.DiskSizeGB > 0 {
		input.BlockDeviceMappings = []types.BlockDeviceMapping{
			{
				DeviceName: new("/dev/sda1"),
				Ebs: &types.EbsBlockDevice{
					VolumeSize: new(opts.DiskSizeGB),
					VolumeType: types.VolumeTypeGp3,
				},
			},
		}
	}

	// Add user data if provided
	if opts.UserData != "" {
		input.UserData = new(opts.UserData)
	}

	// Run instance
	result, err := c.ec2.RunInstances(c.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	if len(result.Instances) == 0 {
		return nil, fmt.Errorf("no instances created")
	}

	instanceID := *result.Instances[0].InstanceId

	// Wait for instance to be running
	if err := c.WaitForInstance(instanceID, "running"); err != nil {
		return nil, fmt.Errorf("instance created but failed to start: %w", err)
	}

	// Get instance info with public IP
	return c.GetInstance(instanceID)
}

// GetInstance gets an instance by ID
func (c *SDKClient) GetInstance(instanceID string) (*InstanceInfo, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	result, err := c.ec2.DescribeInstances(c.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, ErrInstanceNotFound
	}

	return instanceToInfo(&result.Reservations[0].Instances[0]), nil
}

// GetCoolifyInstance finds a Coolify-tagged instance
func (c *SDKClient) GetCoolifyInstance() (*InstanceInfo, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   new("tag:Application"),
				Values: []string{"Coolify"},
			},
			{
				Name:   new("instance-state-name"),
				Values: []string{"running", "pending", "stopping", "stopped"},
			},
		},
	}

	result, err := c.ec2.DescribeInstances(c.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, ErrInstanceNotFound
	}

	return instanceToInfo(&result.Reservations[0].Instances[0]), nil
}

// WaitForInstance waits for an instance to reach the specified state
func (c *SDKClient) WaitForInstance(instanceID, targetState string) error {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for instance to reach state %s", targetState)
		case <-ticker.C:
			info, err := c.GetInstance(instanceID)
			if err != nil {
				return err
			}
			if info.State == targetState {
				return nil
			}
		}
	}
}

// TerminateInstance terminates an instance
func (c *SDKClient) TerminateInstance(instanceID string) error {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}

	_, err := c.ec2.TerminateInstances(c.ctx, input)
	return err
}

// AllocateElasticIP allocates and associates an Elastic IP
func (c *SDKClient) AllocateElasticIP(instanceID string) (string, error) {
	// Allocate EIP
	allocResult, err := c.ec2.AllocateAddress(c.ctx, &ec2.AllocateAddressInput{
		Domain: types.DomainTypeVpc,
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeElasticIp,
				Tags: []types.Tag{
					{Key: new("Application"), Value: new("Coolify")},
					{Key: new("ManagedBy"), Value: new("cool-kit")},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to allocate Elastic IP: %w", err)
	}

	// Associate with instance
	_, err = c.ec2.AssociateAddress(c.ctx, &ec2.AssociateAddressInput{
		AllocationId: allocResult.AllocationId,
		InstanceId:   new(instanceID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to associate Elastic IP: %w", err)
	}

	return *allocResult.PublicIp, nil
}

// ValidateCredentials validates AWS credentials
func (c *SDKClient) ValidateCredentials() error {
	// Try to describe account attributes
	_, err := c.ec2.DescribeAccountAttributes(c.ctx, &ec2.DescribeAccountAttributesInput{})
	if err != nil {
		return fmt.Errorf("AWS credentials invalid or insufficient permissions: %w", err)
	}
	return nil
}

// Helper functions
func instanceToInfo(instance *types.Instance) *InstanceInfo {
	info := &InstanceInfo{
		InstanceID:   safeString(instance.InstanceId),
		InstanceType: string(instance.InstanceType),
		PrivateIP:    safeString(instance.PrivateIpAddress),
	}

	if instance.State != nil {
		info.State = string(instance.State.Name)
	}

	if instance.PublicIpAddress != nil {
		info.PublicIP = *instance.PublicIpAddress
	}

	if instance.Placement != nil && instance.Placement.AvailabilityZone != nil {
		info.AvailabilityZone = *instance.Placement.AvailabilityZone
	}

	if instance.LaunchTime != nil {
		info.LaunchTime = *instance.LaunchTime
	}

	// Get Name from tags
	for _, tag := range instance.Tags {
		if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
			info.Name = *tag.Value
			break
		}
	}

	return info
}

//go:fix inline
func strPtr(s string) *string {
	return new(s)
}

//go:fix inline
func int32Ptr(i int32) *int32 {
	return new(i)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
