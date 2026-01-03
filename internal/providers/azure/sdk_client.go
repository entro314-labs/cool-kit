package azure

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// ErrVMNotFound is returned when a VM cannot be found
var ErrVMNotFound = errors.New("VM not found")

// SDKClient wraps the Azure SDK clients
type SDKClient struct {
	ctx            context.Context
	subscriptionID string
	location       string
	resourceGroup  string
	cred           *azidentity.DefaultAzureCredential

	// Resource clients (lazy initialized)
	resourcesClient *armresources.ResourceGroupsClient
	computeClient   *armcompute.VirtualMachinesClient
	networkClient   *armnetwork.InterfacesClient
	nsgClient       *armnetwork.SecurityGroupsClient
	publicIPClient  *armnetwork.PublicIPAddressesClient
	vnetClient      *armnetwork.VirtualNetworksClient
	subnetClient    *armnetwork.SubnetsClient
}

// NewSDKClient creates a new Azure SDK client
func NewSDKClient(subscriptionID, location, resourceGroup string) (*SDKClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credentials: %w", err)
	}

	return &SDKClient{
		ctx:            context.Background(),
		subscriptionID: subscriptionID,
		location:       location,
		resourceGroup:  resourceGroup,
		cred:           cred,
	}, nil
}

// initClients initializes all Azure clients
func (c *SDKClient) initClients() error {
	var err error

	c.resourcesClient, err = armresources.NewResourceGroupsClient(c.subscriptionID, c.cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create resources client: %w", err)
	}

	c.computeClient, err = armcompute.NewVirtualMachinesClient(c.subscriptionID, c.cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create compute client: %w", err)
	}

	c.networkClient, err = armnetwork.NewInterfacesClient(c.subscriptionID, c.cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create network client: %w", err)
	}

	c.nsgClient, err = armnetwork.NewSecurityGroupsClient(c.subscriptionID, c.cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create NSG client: %w", err)
	}

	c.publicIPClient, err = armnetwork.NewPublicIPAddressesClient(c.subscriptionID, c.cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create public IP client: %w", err)
	}

	c.vnetClient, err = armnetwork.NewVirtualNetworksClient(c.subscriptionID, c.cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create vnet client: %w", err)
	}

	c.subnetClient, err = armnetwork.NewSubnetsClient(c.subscriptionID, c.cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create subnet client: %w", err)
	}

	return nil
}

// ValidateCredentials validates Azure credentials
func (c *SDKClient) ValidateCredentials() error {
	if c.resourcesClient == nil {
		if err := c.initClients(); err != nil {
			return err
		}
	}

	// Try to list resource groups to validate credentials
	pager := c.resourcesClient.NewListPager(nil)
	_, err := pager.NextPage(c.ctx)
	if err != nil {
		return fmt.Errorf("Azure credentials invalid: %w", err)
	}
	return nil
}

// EnsureResourceGroup creates the resource group if it doesn't exist
func (c *SDKClient) EnsureResourceGroup() error {
	if c.resourcesClient == nil {
		if err := c.initClients(); err != nil {
			return err
		}
	}

	// Check if resource group exists
	resp, err := c.resourcesClient.Get(c.ctx, c.resourceGroup, nil)
	if err == nil {
		// Resource group exists - use its location
		if resp.Location != nil && *resp.Location != c.location {
			c.location = *resp.Location
		}
		return nil
	}

	// Create resource group
	_, err = c.resourcesClient.CreateOrUpdate(c.ctx, c.resourceGroup, armresources.ResourceGroup{
		Location: to.Ptr(c.location),
		Tags: map[string]*string{
			"createdby":   to.Ptr("coolify-cli"),
			"environment": to.Ptr("production"),
		},
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to create resource group: %w", err)
	}

	return nil
}

// NSGCreateOpts options for creating network security group
type NSGCreateOpts struct {
	Name  string
	Ports []int
}

// EnsureNSG creates or updates a network security group with required rules
func (c *SDKClient) EnsureNSG(opts NSGCreateOpts) (string, error) {
	if c.nsgClient == nil {
		if err := c.initClients(); err != nil {
			return "", err
		}
	}

	// Build security rules
	rules := make([]*armnetwork.SecurityRule, len(opts.Ports))
	for i, port := range opts.Ports {
		rules[i] = &armnetwork.SecurityRule{
			Name: to.Ptr(fmt.Sprintf("Allow%d", port)),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Priority:                 to.Ptr(int32(100 + i*10)),
				Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
				Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
				Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
				SourceAddressPrefix:      to.Ptr("*"),
				SourcePortRange:          to.Ptr("*"),
				DestinationAddressPrefix: to.Ptr("*"),
				DestinationPortRange:     to.Ptr(fmt.Sprintf("%d", port)),
			},
		}
	}

	poller, err := c.nsgClient.BeginCreateOrUpdate(c.ctx, c.resourceGroup, opts.Name,
		armnetwork.SecurityGroup{
			Location: to.Ptr(c.location),
			Properties: &armnetwork.SecurityGroupPropertiesFormat{
				SecurityRules: rules,
			},
			Tags: map[string]*string{
				"createdby": to.Ptr("coolify-cli"),
			},
		}, nil)

	if err != nil {
		return "", fmt.Errorf("failed to create NSG: %w", err)
	}

	resp, err := poller.PollUntilDone(c.ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed waiting for NSG creation: %w", err)
	}

	return *resp.ID, nil
}

// EnsureVNet creates a virtual network and subnet
func (c *SDKClient) EnsureVNet(vnetName, subnetName string) (string, error) {
	if c.vnetClient == nil {
		if err := c.initClients(); err != nil {
			return "", err
		}
	}

	// Create VNet with subnet
	poller, err := c.vnetClient.BeginCreateOrUpdate(c.ctx, c.resourceGroup, vnetName,
		armnetwork.VirtualNetwork{
			Location: to.Ptr(c.location),
			Properties: &armnetwork.VirtualNetworkPropertiesFormat{
				AddressSpace: &armnetwork.AddressSpace{
					AddressPrefixes: []*string{to.Ptr("10.0.0.0/16")},
				},
				Subnets: []*armnetwork.Subnet{
					{
						Name: to.Ptr(subnetName),
						Properties: &armnetwork.SubnetPropertiesFormat{
							AddressPrefix: to.Ptr("10.0.0.0/24"),
						},
					},
				},
			},
			Tags: map[string]*string{
				"createdby": to.Ptr("coolify-cli"),
			},
		}, nil)

	if err != nil {
		return "", fmt.Errorf("failed to create VNet: %w", err)
	}

	resp, err := poller.PollUntilDone(c.ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed waiting for VNet creation: %w", err)
	}

	// Return subnet ID
	if resp.Properties != nil && len(resp.Properties.Subnets) > 0 {
		return *resp.Properties.Subnets[0].ID, nil
	}

	return "", fmt.Errorf("subnet not found in VNet response")
}

// CreatePublicIP creates a public IP address
func (c *SDKClient) CreatePublicIP(name string) (string, error) {
	if c.publicIPClient == nil {
		if err := c.initClients(); err != nil {
			return "", err
		}
	}

	poller, err := c.publicIPClient.BeginCreateOrUpdate(c.ctx, c.resourceGroup, name,
		armnetwork.PublicIPAddress{
			Location: to.Ptr(c.location),
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic),
			},
			SKU: &armnetwork.PublicIPAddressSKU{
				Name: to.Ptr(armnetwork.PublicIPAddressSKUNameStandard),
			},
			Tags: map[string]*string{
				"createdby": to.Ptr("coolify-cli"),
			},
		}, nil)

	if err != nil {
		return "", fmt.Errorf("failed to create public IP: %w", err)
	}

	resp, err := poller.PollUntilDone(c.ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed waiting for public IP creation: %w", err)
	}

	return *resp.ID, nil
}

// CreateNIC creates a network interface
func (c *SDKClient) CreateNIC(name, subnetID, publicIPID, nsgID string) (string, error) {
	if c.networkClient == nil {
		if err := c.initClients(); err != nil {
			return "", err
		}
	}

	poller, err := c.networkClient.BeginCreateOrUpdate(c.ctx, c.resourceGroup, name,
		armnetwork.Interface{
			Location: to.Ptr(c.location),
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Name: to.Ptr("ipconfig1"),
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
							Subnet: &armnetwork.Subnet{
								ID: to.Ptr(subnetID),
							},
							PublicIPAddress: &armnetwork.PublicIPAddress{
								ID: to.Ptr(publicIPID),
							},
						},
					},
				},
				NetworkSecurityGroup: &armnetwork.SecurityGroup{
					ID: to.Ptr(nsgID),
				},
			},
			Tags: map[string]*string{
				"createdby": to.Ptr("coolify-cli"),
			},
		}, nil)

	if err != nil {
		return "", fmt.Errorf("failed to create NIC: %w", err)
	}

	resp, err := poller.PollUntilDone(c.ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed waiting for NIC creation: %w", err)
	}

	return *resp.ID, nil
}

// VMCreateOpts options for creating a virtual machine
type VMCreateOpts struct {
	Name          string
	Size          string
	AdminUsername string
	SSHPublicKey  string
	NicID         string
	CustomData    string // cloud-init script
}

// VMInfo contains information about a VM
type VMInfo struct {
	ID       string
	Name     string
	State    string
	PublicIP string
	Location string
}

// CreateVM creates a virtual machine
func (c *SDKClient) CreateVM(opts VMCreateOpts) (*VMInfo, error) {
	if c.computeClient == nil {
		if err := c.initClients(); err != nil {
			return nil, err
		}
	}

	// Prepare SSH configuration
	sshConfig := &armcompute.SSHConfiguration{
		PublicKeys: []*armcompute.SSHPublicKey{
			{
				Path:    to.Ptr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", opts.AdminUsername)),
				KeyData: to.Ptr(opts.SSHPublicKey),
			},
		},
	}

	// Encode custom data (cloud-init)
	customData := ""
	if opts.CustomData != "" {
		customData = base64.StdEncoding.EncodeToString([]byte(opts.CustomData))
	}

	poller, err := c.computeClient.BeginCreateOrUpdate(c.ctx, c.resourceGroup, opts.Name,
		armcompute.VirtualMachine{
			Location: to.Ptr(c.location),
			Properties: &armcompute.VirtualMachineProperties{
				HardwareProfile: &armcompute.HardwareProfile{
					VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes(opts.Size)),
				},
				StorageProfile: &armcompute.StorageProfile{
					ImageReference: &armcompute.ImageReference{
						Publisher: to.Ptr("Canonical"),
						Offer:     to.Ptr("0001-com-ubuntu-server-jammy"),
						SKU:       to.Ptr("22_04-lts-gen2"),
						Version:   to.Ptr("latest"),
					},
					OSDisk: &armcompute.OSDisk{
						CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
						ManagedDisk: &armcompute.ManagedDiskParameters{
							StorageAccountType: to.Ptr(armcompute.StorageAccountTypesStandardLRS),
						},
						DiskSizeGB: to.Ptr(int32(30)),
					},
				},
				OSProfile: &armcompute.OSProfile{
					ComputerName:  to.Ptr(opts.Name),
					AdminUsername: to.Ptr(opts.AdminUsername),
					CustomData:    to.Ptr(customData),
					LinuxConfiguration: &armcompute.LinuxConfiguration{
						DisablePasswordAuthentication: to.Ptr(true),
						SSH:                           sshConfig,
					},
				},
				NetworkProfile: &armcompute.NetworkProfile{
					NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
						{
							ID: to.Ptr(opts.NicID),
							Properties: &armcompute.NetworkInterfaceReferenceProperties{
								Primary: to.Ptr(true),
							},
						},
					},
				},
			},
			Tags: map[string]*string{
				"createdby":   to.Ptr("coolify-cli"),
				"application": to.Ptr("coolify"),
			},
		}, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	resp, err := poller.PollUntilDone(c.ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed waiting for VM creation: %w", err)
	}

	return &VMInfo{
		ID:       *resp.ID,
		Name:     *resp.Name,
		State:    *resp.Properties.ProvisioningState,
		Location: c.location,
	}, nil
}

// GetVM retrieves VM information
func (c *SDKClient) GetVM(name string) (*VMInfo, error) {
	if c.computeClient == nil {
		if err := c.initClients(); err != nil {
			return nil, err
		}
	}

	resp, err := c.computeClient.Get(c.ctx, c.resourceGroup, name, nil)
	if err != nil {
		if strings.Contains(err.Error(), "ResourceNotFound") {
			return nil, ErrVMNotFound
		}
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}

	info := &VMInfo{
		ID:       *resp.ID,
		Name:     *resp.Name,
		Location: c.location,
	}

	if resp.Properties != nil && resp.Properties.ProvisioningState != nil {
		info.State = *resp.Properties.ProvisioningState
	}

	return info, nil
}

// VMExists checks if a VM exists
func (c *SDKClient) VMExists(name string) bool {
	_, err := c.GetVM(name)
	return err == nil
}

// GetPublicIPAddress retrieves the public IP address value
func (c *SDKClient) GetPublicIPAddress(name string) (string, error) {
	if c.publicIPClient == nil {
		if err := c.initClients(); err != nil {
			return "", err
		}
	}

	resp, err := c.publicIPClient.Get(c.ctx, c.resourceGroup, name, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get public IP: %w", err)
	}

	if resp.Properties != nil && resp.Properties.IPAddress != nil {
		return *resp.Properties.IPAddress, nil
	}

	return "", fmt.Errorf("public IP address not allocated yet")
}

// WaitForVM waits for VM to be in running state
func (c *SDKClient) WaitForVM(name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for VM")
			}

			vm, err := c.GetVM(name)
			if err != nil {
				continue
			}

			if vm.State == "Succeeded" {
				return nil
			}
		}
	}
}

// DeleteResourceGroup deletes the entire resource group
func (c *SDKClient) DeleteResourceGroup() error {
	if c.resourcesClient == nil {
		if err := c.initClients(); err != nil {
			return err
		}
	}

	poller, err := c.resourcesClient.BeginDelete(c.ctx, c.resourceGroup, nil)
	if err != nil {
		return fmt.Errorf("failed to delete resource group: %w", err)
	}

	_, err = poller.PollUntilDone(c.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed waiting for resource group deletion: %w", err)
	}

	return nil
}

// GetLocation returns the current location
func (c *SDKClient) GetLocation() string {
	return c.location
}
