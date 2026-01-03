package azure

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/ui"
)

// Provisioner handles Azure VM provisioning
type Provisioner struct {
	ctx *DeploymentContext
}

// NewProvisioner creates a new Azure provisioner
func NewProvisioner(ctx *DeploymentContext) *Provisioner {
	return &Provisioner{ctx: ctx}
}

// Provision provisions the complete Azure infrastructure
func (p *Provisioner) Provision() error {
	ui.Section("Provisioning Azure Infrastructure")

	// Check Azure CLI
	if err := p.checkAzureCLI(); err != nil {
		return err
	}

	// Check login
	if err := p.checkAzureLogin(); err != nil {
		return err
	}

	// Create resource group
	if err := p.createResourceGroup(); err != nil {
		return err
	}

	// Create virtual network and subnet
	if err := p.createNetwork(); err != nil {
		return err
	}

	// Create network security group
	if err := p.createNSG(); err != nil {
		return err
	}

	// Create public IP
	if err := p.createPublicIP(); err != nil {
		return err
	}

	// Create network interface
	if err := p.createNIC(); err != nil {
		return err
	}

	// Create VM
	if err := p.createVM(); err != nil {
		return err
	}

	// Get public IP address
	if err := p.getPublicIP(); err != nil {
		return err
	}

	ui.Success("Infrastructure provisioned successfully")
	ui.Dim(fmt.Sprintf("Public IP: %s", p.ctx.PublicIP))

	return nil
}

// checkAzureCLI verifies Azure CLI is installed
func (p *Provisioner) checkAzureCLI() error {
	ui.Info("Checking Azure CLI")

	cmd := exec.Command("az", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Azure CLI is not installed. Install from: https://docs.microsoft.com/cli/azure/install-azure-cli")
	}

	ui.Dim("Azure CLI is available")
	return nil
}

// checkAzureLogin verifies user is logged into Azure
func (p *Provisioner) checkAzureLogin() error {
	ui.Info("Checking Azure login")

	cmd := exec.Command("az", "account", "show")
	if err := cmd.Run(); err != nil {
		ui.Warning("Not logged into Azure. Please run: az login")
		return fmt.Errorf("not authenticated with Azure")
	}

	ui.Dim("Authenticated with Azure")
	return nil
}

// createResourceGroup creates the Azure resource group
func (p *Provisioner) createResourceGroup() error {
	ui.Info(fmt.Sprintf("Creating resource group: %s", p.ctx.ResourceGroup))

	cmd := exec.Command("az", "group", "create",
		"--name", p.ctx.ResourceGroup,
		"--location", p.ctx.Location,
		"--output", "none")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create resource group: %w", err)
	}

	ui.Dim(fmt.Sprintf("Resource group created in %s", p.ctx.Location))
	return nil
}

// createNetwork creates virtual network and subnet
func (p *Provisioner) createNetwork() error {
	ui.Info("Creating virtual network and subnet")

	// Create VNet
	vnetCmd := exec.Command("az", "network", "vnet", "create",
		"--resource-group", p.ctx.ResourceGroup,
		"--name", p.ctx.VNetName,
		"--address-prefix", p.ctx.Config.Networking.VNetAddressPrefix,
		"--subnet-name", p.ctx.SubnetName,
		"--subnet-prefix", p.ctx.Config.Networking.SubnetAddressPrefix,
		"--output", "none")

	if err := vnetCmd.Run(); err != nil {
		return fmt.Errorf("failed to create virtual network: %w", err)
	}

	ui.Dim("Virtual network and subnet created")
	return nil
}

// createNSG creates and configures network security group
func (p *Provisioner) createNSG() error {
	ui.Info("Creating network security group")

	// Create NSG
	nsgCmd := exec.Command("az", "network", "nsg", "create",
		"--resource-group", p.ctx.ResourceGroup,
		"--name", p.ctx.NSGName,
		"--output", "none")

	if err := nsgCmd.Run(); err != nil {
		return fmt.Errorf("failed to create NSG: %w", err)
	}

	// Add rules
	rules := []struct {
		name     string
		port     string
		priority int
	}{
		{"allow-ssh", "22", 1000},
		{"allow-http", "80", 1001},
		{"allow-https", "443", 1002},
		{"allow-coolify", "8000", 1003},
	}

	for _, rule := range rules {
		ruleCmd := exec.Command("az", "network", "nsg", "rule", "create",
			"--resource-group", p.ctx.ResourceGroup,
			"--nsg-name", p.ctx.NSGName,
			"--name", rule.name,
			"--protocol", "tcp",
			"--direction", "inbound",
			"--source-address-prefix", "*",
			"--source-port-range", "*",
			"--destination-address-prefix", "*",
			"--destination-port-range", rule.port,
			"--access", "allow",
			"--priority", fmt.Sprintf("%d", rule.priority),
			"--output", "none")

		if err := ruleCmd.Run(); err != nil {
			return fmt.Errorf("failed to create NSG rule %s: %w", rule.name, err)
		}
	}

	ui.Dim("Network security group configured")
	return nil
}

// createPublicIP creates a public IP address
func (p *Provisioner) createPublicIP() error {
	ui.Info("Creating public IP address")

	cmd := exec.Command("az", "network", "public-ip", "create",
		"--resource-group", p.ctx.ResourceGroup,
		"--name", p.ctx.PublicIPName,
		"--sku", "Standard",
		"--allocation-method", "Static",
		"--output", "none")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create public IP: %w", err)
	}

	ui.Dim("Public IP address created")
	return nil
}

// createNIC creates network interface
func (p *Provisioner) createNIC() error {
	ui.Info("Creating network interface")

	cmd := exec.Command("az", "network", "nic", "create",
		"--resource-group", p.ctx.ResourceGroup,
		"--name", p.ctx.NICName,
		"--vnet-name", p.ctx.VNetName,
		"--subnet", p.ctx.SubnetName,
		"--public-ip-address", p.ctx.PublicIPName,
		"--network-security-group", p.ctx.NSGName,
		"--output", "none")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create network interface: %w", err)
	}

	ui.Dim("Network interface created")
	return nil
}

// createVM creates the Azure virtual machine
func (p *Provisioner) createVM() error {
	ui.Info("Creating virtual machine (this may take a few minutes)")

	// Read SSH public key
	sshKeyData, err := os.ReadFile(p.ctx.SSHKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read SSH key from %s: %w", p.ctx.SSHKeyPath, err)
	}

	cmd := exec.Command("az", "vm", "create",
		"--resource-group", p.ctx.ResourceGroup,
		"--name", p.ctx.VMName,
		"--location", p.ctx.Location,
		"--nics", p.ctx.NICName,
		"--size", p.ctx.VMSize,
		"--image", p.ctx.Config.Infrastructure.OSImage,
		"--admin-username", p.ctx.AdminUsername,
		"--ssh-key-values", string(sshKeyData),
		"--os-disk-size-gb", fmt.Sprintf("%d", p.ctx.Config.Infrastructure.OSDiskSizeGB),
		"--output", "none")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create VM: %w", err)
	}

	ui.Dim("Virtual machine created")
	return nil
}

// getPublicIP retrieves the public IP address
func (p *Provisioner) getPublicIP() error {
	ui.Info("Retrieving public IP address")

	cmd := exec.Command("az", "network", "public-ip", "show",
		"--resource-group", p.ctx.ResourceGroup,
		"--name", p.ctx.PublicIPName,
		"--query", "ipAddress",
		"--output", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get public IP: %w", err)
	}

	p.ctx.PublicIP = strings.TrimSpace(string(output))
	if p.ctx.PublicIP == "" {
		return fmt.Errorf("public IP not found")
	}

	return nil
}

// WaitForVM waits for VM to be fully ready
func (p *Provisioner) WaitForVM(maxWait time.Duration) error {
	ui.Info("Waiting for VM to be ready")

	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		cmd := exec.Command("az", "vm", "get-instance-view",
			"--resource-group", p.ctx.ResourceGroup,
			"--name", p.ctx.VMName,
			"--query", "instanceView.statuses[?starts_with(code, 'PowerState/')].displayStatus",
			"--output", "tsv")

		output, err := cmd.Output()
		if err == nil {
			status := strings.TrimSpace(string(output))
			if status == "VM running" {
				ui.Dim("VM is running")
				return nil
			}
		}

		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("VM did not become ready within %v", maxWait)
}

// Deprovision removes all Azure resources
func (p *Provisioner) Deprovision() error {
	ui.Section("Deprovisioning Azure Infrastructure")

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "az", "group", "delete",
		"--name", p.ctx.ResourceGroup,
		"--yes",
		"--no-wait")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete resource group: %w", err)
	}

	ui.Success(fmt.Sprintf("Resource group %s deletion initiated", p.ctx.ResourceGroup))
	return nil
}
