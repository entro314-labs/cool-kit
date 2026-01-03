package installer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/git"
	"github.com/entro314-labs/cool-kit/internal/providers/azure"
)

// AzureDeployer handles Azure deployments
type AzureDeployer struct {
	config     *config.Config
	gitManager *git.Manager
	sdkClient  *azure.SDKClient
	useSDK     bool
	logs       []string
}

// NewAzureDeployer creates a new Azure deployer
func NewAzureDeployer(cfg *config.Config) (*AzureDeployer, error) {
	// Ensure Settings map is initialized
	if cfg.Settings == nil {
		cfg.Settings = make(map[string]interface{})
	}

	deployer := &AzureDeployer{
		config:     cfg,
		gitManager: git.NewManager(cfg),
		logs:       []string{},
		useSDK:     true,
	}

	// Try to initialize SDK client
	subID := cfg.Azure.SubscriptionID
	if subID == "" {
		// Try to get subscription ID from Azure CLI
		output, err := deployer.runAzCommand("account", "show", "--query", "id", "-o", "tsv")
		if err == nil && strings.TrimSpace(output) != "" {
			subID = strings.TrimSpace(output)
			cfg.Azure.SubscriptionID = subID
		}
	}

	if subID != "" {
		sdkClient, err := azure.NewSDKClient(subID, cfg.Azure.Location, cfg.Azure.ResourceGroup)
		if err == nil {
			deployer.sdkClient = sdkClient
		} else {
			deployer.logs = append(deployer.logs, fmt.Sprintf("SDK init failed, falling back to CLI: %v", err))
			deployer.useSDK = false
		}
	} else {
		deployer.useSDK = false
	}

	return deployer, nil
}

// runAzCommand runs an Azure CLI command and captures output
func (d *AzureDeployer) runAzCommand(args ...string) (string, error) {
	cmd := exec.Command("az", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if err != nil {
		errOutput := stderr.String()
		if errOutput != "" {
			d.logs = append(d.logs, fmt.Sprintf("stderr: %s", errOutput))
		}
		return output, fmt.Errorf("%w: %s", err, errOutput)
	}

	return output, nil
}

// Deploy performs the complete Azure deployment
func (d *AzureDeployer) Deploy() error {
	steps := []struct {
		name string
		fn   func() error
	}{
		{"Validating Azure credentials", d.validateCredentials},
		{"Cloning Coolify repository", d.cloneRepository},
		{"Creating resource group", d.createResourceGroup},
		{"Creating network resources", d.createNetworkResources},
		{"Creating virtual machine", d.createVirtualMachine},
		{"Configuring Coolify", d.configureCoolify},
		{"Running health checks", d.runHealthChecks},
	}

	for _, step := range steps {
		if err := step.fn(); err != nil {
			return fmt.Errorf("step '%s' failed: %w", step.name, err)
		}
	}

	return nil
}

// validateCredentials validates Azure credentials
func (d *AzureDeployer) validateCredentials() error {
	if d.useSDK && d.sdkClient != nil {
		return d.sdkClient.ValidateCredentials()
	}

	// CLI fallback
	_, err := d.runAzCommand("account", "show")
	if err != nil {
		return fmt.Errorf("Azure credentials validation failed. Please run 'az login': %w", err)
	}
	return nil
}

// cloneRepository clones the Coolify repository
func (d *AzureDeployer) cloneRepository() error {
	return d.gitManager.CloneOrPull()
}

// createResourceGroup creates or uses existing Azure resource group
func (d *AzureDeployer) createResourceGroup() error {
	if d.useSDK && d.sdkClient != nil {
		err := d.sdkClient.EnsureResourceGroup()
		if err == nil {
			// Update location if SDK detected existing RG with different location
			d.config.Azure.Location = d.sdkClient.GetLocation()
		}
		return err
	}

	// CLI fallback
	rgName := d.config.Azure.ResourceGroup
	location := d.config.Azure.Location

	// Check if resource group exists
	output, err := d.runAzCommand("group", "show", "--name", rgName, "--query", "location", "-o", "tsv")
	if err == nil && strings.TrimSpace(output) != "" {
		existingLocation := strings.TrimSpace(output)
		if existingLocation != location {
			d.config.Azure.Location = existingLocation
		}
		return nil
	}

	// Create new resource group
	_, err = d.runAzCommand("group", "create",
		"--name", rgName,
		"--location", location,
		"--tags", "createdby=coolify-cli", "environment=production")

	return err
}

// createNetworkResources creates network security group with proper rules
func (d *AzureDeployer) createNetworkResources() error {
	vmName := d.config.Azure.VMName
	nsgName := vmName + "-nsg"
	vnetName := vmName + "-vnet"
	subnetName := "default"
	nicName := vmName + "-nic"
	publicIPName := vmName + "-ip"

	requiredPorts := []int{22, 80, 443, 8000, 6001}

	if d.useSDK && d.sdkClient != nil {
		// Create NSG
		nsgID, err := d.sdkClient.EnsureNSG(azure.NSGCreateOpts{
			Name:  nsgName,
			Ports: requiredPorts,
		})
		if err != nil {
			return err
		}
		d.config.Settings["nsg_id"] = nsgID

		// Create VNet and Subnet
		subnetID, err := d.sdkClient.EnsureVNet(vnetName, subnetName)
		if err != nil {
			return err
		}
		d.config.Settings["subnet_id"] = subnetID

		// Create Public IP
		publicIPID, err := d.sdkClient.CreatePublicIP(publicIPName)
		if err != nil {
			return err
		}
		d.config.Settings["public_ip_id"] = publicIPID

		// Create NIC
		nicID, err := d.sdkClient.CreateNIC(nicName, subnetID, publicIPID, nsgID)
		if err != nil {
			return err
		}
		d.config.Settings["nic_id"] = nicID

		return nil
	}

	// CLI fallback - create NSG and rules
	_, err := d.runAzCommand("network", "nsg", "create",
		"--resource-group", d.config.Azure.ResourceGroup,
		"--name", nsgName,
		"--location", d.config.Azure.Location)
	if err != nil {
		return err
	}

	// Add security rules
	for i, port := range requiredPorts {
		d.runAzCommand("network", "nsg", "rule", "create",
			"--resource-group", d.config.Azure.ResourceGroup,
			"--nsg-name", nsgName,
			"--name", fmt.Sprintf("Allow%d", port),
			"--priority", fmt.Sprintf("%d", 100+i*10),
			"--destination-port-ranges", fmt.Sprintf("%d", port),
			"--access", "Allow",
			"--protocol", "Tcp",
			"--direction", "Inbound")
	}

	return nil
}

// createVirtualMachine creates the Azure virtual machine
func (d *AzureDeployer) createVirtualMachine() error {
	vmName := d.config.Azure.VMName
	cloudInit := d.getCloudInit()

	if d.useSDK && d.sdkClient != nil {
		// Get SSH public key
		sshKey, err := d.getSSHPublicKey()
		if err != nil {
			return fmt.Errorf("failed to read SSH key: %w", err)
		}

		nicID, ok := d.config.Settings["nic_id"].(string)
		if !ok {
			return fmt.Errorf("NIC ID not found - network setup failed")
		}

		_, err = d.sdkClient.CreateVM(azure.VMCreateOpts{
			Name:          vmName,
			Size:          d.config.Azure.VMSize,
			AdminUsername: d.config.Azure.AdminUsername,
			SSHPublicKey:  sshKey,
			NicID:         nicID,
			CustomData:    cloudInit,
		})
		if err != nil {
			return err
		}

		// Wait for VM to be ready
		if err := d.sdkClient.WaitForVM(vmName, 5*time.Minute); err != nil {
			return err
		}

		// Get public IP
		ip, err := d.sdkClient.GetPublicIPAddress(vmName + "-ip")
		if err != nil {
			return err
		}
		d.config.Settings["public_ip"] = ip

		return nil
	}

	// CLI fallback
	cloudInitFile := "/tmp/coolify-cloud-init.yaml"
	if err := os.WriteFile(cloudInitFile, []byte(cloudInit), 0644); err != nil {
		return fmt.Errorf("failed to write cloud-init: %w", err)
	}

	_, err := d.runAzCommand("vm", "create",
		"--resource-group", d.config.Azure.ResourceGroup,
		"--name", vmName,
		"--image", "Ubuntu2204",
		"--size", d.config.Azure.VMSize,
		"--admin-username", d.config.Azure.AdminUsername,
		"--generate-ssh-keys",
		"--location", d.config.Azure.Location,
		"--nsg", vmName+"-nsg",
		"--custom-data", cloudInitFile,
		"--public-ip-sku", "Standard")

	if err != nil {
		return fmt.Errorf("failed to create VM: %w", err)
	}

	return d.getPublicIP()
}

// getSSHPublicKey reads the SSH public key for the user
func (d *AzureDeployer) getSSHPublicKey() (string, error) {
	keyPath := d.config.Azure.SSHKeyPath
	if keyPath == "" {
		keyPath = "~/.ssh/id_rsa.pub"
	}

	// Expand home directory
	if strings.HasPrefix(keyPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		keyPath = filepath.Join(home, keyPath[2:])
	}

	data, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read SSH key from %s: %w", keyPath, err)
	}

	return strings.TrimSpace(string(data)), nil
}

// getPublicIP retrieves and stores the VM's public IP (CLI method)
func (d *AzureDeployer) getPublicIP() error {
	output, err := d.runAzCommand("vm", "list-ip-addresses",
		"--resource-group", d.config.Azure.ResourceGroup,
		"--name", d.config.Azure.VMName,
		"--query", "[0].virtualMachine.network.publicIpAddresses[0].ipAddress",
		"--output", "tsv")

	if err != nil {
		return fmt.Errorf("failed to get VM IP address: %w", err)
	}

	publicIP := strings.TrimSpace(output)
	if publicIP == "" {
		return fmt.Errorf("no public IP found for VM")
	}

	d.config.Settings["public_ip"] = publicIP
	return nil
}

// configureCoolify waits for cloud-init to complete
func (d *AzureDeployer) configureCoolify() error {
	publicIP, ok := d.config.Settings["public_ip"].(string)
	if !ok || publicIP == "" {
		return fmt.Errorf("public IP not found in config")
	}

	d.logs = append(d.logs, "Waiting for cloud-init to install Docker and Coolify...")

	maxWait := 5 * time.Minute
	start := time.Now()

	for time.Since(start) < maxWait {
		cmd := exec.Command("ssh",
			"-o", "StrictHostKeyChecking=no",
			"-o", "ConnectTimeout=10",
			"-o", "BatchMode=yes",
			fmt.Sprintf("%s@%s", d.config.Azure.AdminUsername, publicIP),
			"docker ps")

		if err := cmd.Run(); err == nil {
			return nil
		}

		time.Sleep(30 * time.Second)
	}

	return fmt.Errorf("timeout waiting for Docker to be ready")
}

// runHealthChecks performs health checks on the deployed instance
func (d *AzureDeployer) runHealthChecks() error {
	publicIP, ok := d.config.Settings["public_ip"].(string)
	if !ok || publicIP == "" {
		return fmt.Errorf("public IP not found in config")
	}

	time.Sleep(30 * time.Second)

	curlCmd := exec.Command("curl", "-f", "-s", "-o", "/dev/null", "--max-time", "10",
		fmt.Sprintf("http://%s:8000", publicIP))

	if err := curlCmd.Run(); err != nil {
		d.logs = append(d.logs, fmt.Sprintf("HTTP check failed (may need more time): %v", err))
	}

	return nil
}

// getCloudInit returns the cloud-init script for installing Docker and Coolify
func (d *AzureDeployer) getCloudInit() string {
	return `#cloud-config
package_update: true
package_upgrade: true

packages:
  - curl
  - git

runcmd:
  - curl -fsSL https://get.docker.com | sh
  - systemctl enable docker
  - systemctl start docker
  - mkdir -p /data/coolify
  - curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash
`
}

// GetLogs returns deployment logs
func (d *AzureDeployer) GetLogs() []string {
	return d.logs
}

// IsUsingSDK returns whether the deployer is using the SDK
func (d *AzureDeployer) IsUsingSDK() bool {
	return d.useSDK
}
