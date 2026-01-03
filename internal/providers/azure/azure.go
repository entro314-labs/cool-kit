// Package azure implements the Azure provider.
package azure

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
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// AzureProvider handles Azure deployments
type AzureProvider struct {
	config     *config.Config
	gitManager *git.Manager
	sdkClient  *SDKClient
	useSDK     bool
}

// NewAzureProvider creates a new Azure provider
func NewAzureProvider(cfg *config.Config) (*AzureProvider, error) {
	// Ensure Settings map is initialized
	if cfg.Settings == nil {
		cfg.Settings = make(map[string]interface{})
	}

	provider := &AzureProvider{
		config:     cfg,
		gitManager: git.NewManager(cfg),
		useSDK:     true,
	}

	// Try to get subscription ID from Azure CLI if not set
	subID := cfg.Azure.SubscriptionID
	if subID == "" {
		output, err := provider.runAzCommand("account", "show", "--query", "id", "-o", "tsv")
		if err == nil && strings.TrimSpace(output) != "" {
			subID = strings.TrimSpace(output)
			cfg.Azure.SubscriptionID = subID
		}
	}

	// Initialize SDK client if we have subscription ID
	if subID != "" {
		sdkClient, err := NewSDKClient(subID, cfg.Azure.Location, cfg.Azure.ResourceGroup)
		if err == nil {
			provider.sdkClient = sdkClient
		} else {
			provider.useSDK = false
		}
	} else {
		provider.useSDK = false
	}

	return provider, nil
}

// GetDeploymentSteps returns the deployment steps for Azure
func (p *AzureProvider) GetDeploymentSteps() []ui.DeploymentStep {
	return []ui.DeploymentStep{
		{Name: "Validate Azure credentials", Description: "Checking Azure CLI and credentials"},
		{Name: "Clone Coolify repository", Description: "Fetching latest Coolify from GitHub"},
		{Name: "Create resource group", Description: "Setting up Azure resource group"},
		{Name: "Create network resources", Description: "Setting up NSG, VNet, and public IP"},
		{Name: "Create virtual machine", Description: "Launching Azure VM with cloud-init"},
		{Name: "Wait for VM ready", Description: "Waiting for VM to be ready"},
		{Name: "Deploy Coolify", Description: "Installing Coolify via cloud-init"},
		{Name: "Run health checks", Description: "Validating deployment"},
	}
}

// Deploy performs the Azure deployment with progress tracking
func (p *AzureProvider) Deploy(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	steps := []struct {
		name string
		fn   func(chan<- ui.StepProgressMsg, chan<- ui.LogMsg) error
	}{
		{"Validate Azure credentials", p.validateCredentials},
		{"Clone Coolify repository", p.cloneRepository},
		{"Create resource group", p.createResourceGroup},
		{"Create network resources", p.createNetworkResources},
		{"Create virtual machine", p.createVirtualMachine},
		{"Wait for VM ready", p.waitForVM},
		{"Deploy Coolify", p.deployCoolify},
		{"Run health checks", p.runHealthChecks},
	}

	for i, step := range steps {
		logChan <- ui.LogMsg{
			Level:   ui.LogInfo,
			Message: fmt.Sprintf("Starting: %s", step.name),
		}

		if err := step.fn(progressChan, logChan); err != nil {
			logChan <- ui.LogMsg{
				Level:   ui.LogError,
				Message: fmt.Sprintf("Failed: %s - %v", step.name, err),
			}
			return fmt.Errorf("step '%s' failed: %w", step.name, err)
		}

		progressChan <- ui.StepProgressMsg{
			StepIndex: i,
			Progress:  1.0,
			Message:   fmt.Sprintf("Completed: %s", step.name),
		}

		logChan <- ui.LogMsg{
			Level:   ui.LogSuccess,
			Message: fmt.Sprintf("✓ %s completed", step.name),
		}
	}

	return nil
}

// Destroy deletes the Azure resource group
func (p *AzureProvider) Destroy(logChan chan<- ui.LogMsg) error {
	rgName := p.config.Azure.ResourceGroup

	if p.useSDK && p.sdkClient != nil {
		logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Deleting resource group '%s' via SDK (this may take several minutes)...", rgName)}
		if err := p.sdkClient.DeleteResourceGroup(); err != nil {
			return err
		}
		logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Resource group deleted successfully"}
		return nil
	}

	// CLI fallback
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Deleting resource group '%s' via CLI...", rgName)}
	_, err := p.runAzCommand("group", "delete", "--name", rgName, "--yes", "--no-wait")
	if err != nil {
		return fmt.Errorf("failed to delete resource group: %w", err)
	}

	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Resource group deletion initiated (running in background)"}
	return nil
}

// runAzCommand runs an Azure CLI command and captures output
func (p *AzureProvider) runAzCommand(args ...string) (string, error) {
	cmd := exec.Command("az", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if err != nil {
		errOutput := strings.TrimSpace(stderr.String())
		if errOutput != "" {
			return output, fmt.Errorf("%s", errOutput)
		}
		return output, err
	}

	return output, nil
}

// validateCredentials validates Azure credentials
func (p *AzureProvider) validateCredentials(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Checking Azure CLI"}

	if p.useSDK && p.sdkClient != nil {
		progressChan <- ui.StepProgressMsg{Progress: 0.6, Message: "Validating SDK credentials"}
		if err := p.sdkClient.ValidateCredentials(); err != nil {
			return fmt.Errorf("Azure SDK validation failed: %w", err)
		}
		logChan <- ui.LogMsg{Level: ui.LogDebug, Message: "Azure SDK credentials validated"}
	} else {
		progressChan <- ui.StepProgressMsg{Progress: 0.6, Message: "Validating CLI credentials"}
		if _, err := p.runAzCommand("account", "show"); err != nil {
			return fmt.Errorf("Azure CLI not authenticated. Run 'az login': %w", err)
		}
		logChan <- ui.LogMsg{Level: ui.LogDebug, Message: "Azure CLI credentials validated"}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Credentials validated"}
	return nil
}

// cloneRepository clones the Coolify repository
func (p *AzureProvider) cloneRepository(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Fetching repository"}

	if err := p.gitManager.CloneOrPull(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.7, Message: "Getting commit info"}

	commitInfo, err := p.gitManager.GetLatestCommitInfo()
	if err != nil {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: "Could not get commit info"}
	} else {
		logChan <- ui.LogMsg{
			Level:   ui.LogInfo,
			Message: fmt.Sprintf("Using commit: %s", commitInfo.ShortHash),
		}
	}

	return nil
}

// createResourceGroup creates or uses existing Azure resource group
func (p *AzureProvider) createResourceGroup(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	rgName := p.config.Azure.ResourceGroup
	location := p.config.Azure.Location

	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: fmt.Sprintf("Checking resource group %s", rgName)}

	if p.useSDK && p.sdkClient != nil {
		if err := p.sdkClient.EnsureResourceGroup(); err != nil {
			return fmt.Errorf("failed to create resource group: %w", err)
		}
		// Update location if SDK detected existing RG with different location
		p.config.Azure.Location = p.sdkClient.GetLocation()
		logChan <- ui.LogMsg{Level: ui.LogDebug, Message: fmt.Sprintf("Using location: %s", p.config.Azure.Location)}
	} else {
		// Check if exists
		output, err := p.runAzCommand("group", "show", "--name", rgName, "--query", "location", "-o", "tsv")
		if err == nil && strings.TrimSpace(output) != "" {
			existingLocation := strings.TrimSpace(output)
			p.config.Azure.Location = existingLocation
			logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Using existing resource group in %s", existingLocation)}
		} else {
			progressChan <- ui.StepProgressMsg{Progress: 0.6, Message: "Creating resource group"}
			_, err = p.runAzCommand("group", "create", "--name", rgName, "--location", location)
			if err != nil {
				return fmt.Errorf("failed to create resource group: %w", err)
			}
		}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Resource group ready"}
	return nil
}

// createNetworkResources creates NSG, VNet, Public IP, and NIC
func (p *AzureProvider) createNetworkResources(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	vmName := p.config.Azure.VMName
	nsgName := vmName + "-nsg"
	vnetName := vmName + "-vnet"
	subnetName := "default"
	nicName := vmName + "-nic"
	publicIPName := vmName + "-ip"
	requiredPorts := []int{22, 80, 443, 8000, 6001}

	if p.useSDK && p.sdkClient != nil {
		// Create NSG
		progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Creating network security group"}
		nsgID, err := p.sdkClient.EnsureNSG(NSGCreateOpts{Name: nsgName, Ports: requiredPorts})
		if err != nil {
			return fmt.Errorf("failed to create NSG: %w", err)
		}
		p.config.Settings["nsg_id"] = nsgID

		// Create VNet
		progressChan <- ui.StepProgressMsg{Progress: 0.4, Message: "Creating virtual network"}
		subnetID, err := p.sdkClient.EnsureVNet(vnetName, subnetName)
		if err != nil {
			return fmt.Errorf("failed to create VNet: %w", err)
		}
		p.config.Settings["subnet_id"] = subnetID

		// Create Public IP
		progressChan <- ui.StepProgressMsg{Progress: 0.6, Message: "Creating public IP"}
		publicIPID, err := p.sdkClient.CreatePublicIP(publicIPName)
		if err != nil {
			return fmt.Errorf("failed to create public IP: %w", err)
		}
		p.config.Settings["public_ip_id"] = publicIPID

		// Create NIC
		progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "Creating network interface"}
		nicID, err := p.sdkClient.CreateNIC(nicName, subnetID, publicIPID, nsgID)
		if err != nil {
			return fmt.Errorf("failed to create NIC: %w", err)
		}
		p.config.Settings["nic_id"] = nicID
	} else {
		// CLI fallback - create NSG
		progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Creating NSG via CLI"}
		p.runAzCommand("network", "nsg", "create",
			"--resource-group", p.config.Azure.ResourceGroup,
			"--name", nsgName,
			"--location", p.config.Azure.Location)

		// Add rules
		for i, port := range requiredPorts {
			progress := 0.3 + (float64(i+1) / float64(len(requiredPorts)) * 0.5)
			progressChan <- ui.StepProgressMsg{Progress: progress, Message: fmt.Sprintf("Opening port %d", port)}
			p.runAzCommand("network", "nsg", "rule", "create",
				"--resource-group", p.config.Azure.ResourceGroup,
				"--nsg-name", nsgName,
				"--name", fmt.Sprintf("Allow%d", port),
				"--priority", fmt.Sprintf("%d", 100+i*10),
				"--destination-port-ranges", fmt.Sprintf("%d", port),
				"--access", "Allow", "--protocol", "Tcp", "--direction", "Inbound")
		}
	}

	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Network resources created"}
	return nil
}

// createVirtualMachine creates the Azure VM
func (p *AzureProvider) createVirtualMachine(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	vmName := p.config.Azure.VMName
	cloudInit := p.getCloudInit()

	progressChan <- ui.StepProgressMsg{Progress: 0.1, Message: "Checking for existing VM"}

	// Check if VM already exists
	if p.vmExists(vmName) {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: fmt.Sprintf("VM '%s' already exists", vmName)}
		return ui.NewDeploymentError("azure", "Create VM", fmt.Errorf(
			"VM '%s' already exists in resource group '%s'. Delete it first or use a different name.\n\nTo delete: az vm delete --resource-group %s --name %s --yes",
			vmName, p.config.Azure.ResourceGroup, p.config.Azure.ResourceGroup, vmName,
		))
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Preparing VM configuration"}

	if p.useSDK && p.sdkClient != nil {
		// Get SSH public key
		sshKey, err := p.getSSHPublicKey()
		if err != nil {
			return ui.NewDeploymentError("azure", "Read SSH Key", fmt.Errorf("cannot read SSH public key: %w", err))
		}

		nicID, ok := p.config.Settings["nic_id"].(string)
		if !ok {
			return ui.NewDeploymentError("azure", "Create VM", fmt.Errorf("network interface not found - network setup may have failed"))
		}

		progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Creating VM via SDK"}

		_, err = p.sdkClient.CreateVM(VMCreateOpts{
			Name:          vmName,
			Size:          p.config.Azure.VMSize,
			AdminUsername: p.config.Azure.AdminUsername,
			SSHPublicKey:  sshKey,
			NicID:         nicID,
			CustomData:    cloudInit,
		})
		if err != nil {
			return ui.NewDeploymentError("azure", "Create VM", err)
		}
	} else {
		// CLI fallback
		progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Creating VM via CLI"}

		cloudInitFile := "/tmp/coolify-cloud-init.yaml"
		if err := os.WriteFile(cloudInitFile, []byte(cloudInit), 0600); err != nil {
			return ui.NewDeploymentError("azure", "Write cloud-init", err)
		}

		_, err := p.runAzCommand("vm", "create",
			"--resource-group", p.config.Azure.ResourceGroup,
			"--name", vmName,
			"--image", "Canonical:ubuntu-24_04-lts:server:latest",
			"--size", p.config.Azure.VMSize,
			"--admin-username", p.config.Azure.AdminUsername,
			"--generate-ssh-keys",
			"--location", p.config.Azure.Location,
			"--nsg", vmName+"-nsg",
			"--custom-data", cloudInitFile,
			"--public-ip-sku", "Standard")
		if err != nil {
			return ui.NewDeploymentError("azure", "Create VM", err)
		}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "VM created"}
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: fmt.Sprintf("VM %s created", vmName)}
	return nil
}

// vmExists checks if a VM already exists
func (p *AzureProvider) vmExists(vmName string) bool {
	if p.useSDK && p.sdkClient != nil {
		return p.sdkClient.VMExists(vmName)
	}
	// CLI fallback
	_, err := p.runAzCommand("vm", "show",
		"--resource-group", p.config.Azure.ResourceGroup,
		"--name", vmName,
		"--query", "id", "-o", "tsv")
	return err == nil
}

// waitForVM waits for VM to be ready and gets public IP
func (p *AzureProvider) waitForVM(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	vmName := p.config.Azure.VMName
	publicIPName := vmName + "-ip"

	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Waiting for VM provisioning"}

	if p.useSDK && p.sdkClient != nil {
		if err := p.sdkClient.WaitForVM(vmName, 5*time.Minute); err != nil {
			return fmt.Errorf("timeout waiting for VM: %w", err)
		}

		progressChan <- ui.StepProgressMsg{Progress: 0.7, Message: "Getting public IP"}
		ip, err := p.sdkClient.GetPublicIPAddress(publicIPName)
		if err != nil {
			return fmt.Errorf("failed to get public IP: %w", err)
		}
		p.config.Settings["public_ip"] = ip
	} else {
		// CLI - get IP
		time.Sleep(30 * time.Second) // Wait for provisioning

		progressChan <- ui.StepProgressMsg{Progress: 0.7, Message: "Getting public IP"}
		output, err := p.runAzCommand("vm", "list-ip-addresses",
			"--resource-group", p.config.Azure.ResourceGroup,
			"--name", vmName,
			"--query", "[0].virtualMachine.network.publicIpAddresses[0].ipAddress",
			"--output", "tsv")
		if err != nil {
			return fmt.Errorf("failed to get VM IP: %w", err)
		}
		p.config.Settings["public_ip"] = strings.TrimSpace(output)
	}

	publicIP := p.config.Settings["public_ip"].(string)
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("VM public IP: %s", publicIP)}
	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "VM ready"}
	return nil
}

// deployCoolify waits for cloud-init to complete
func (p *AzureProvider) deployCoolify(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	publicIP, ok := p.config.Settings["public_ip"].(string)
	if !ok || publicIP == "" {
		return fmt.Errorf("public IP not found")
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.1, Message: "Waiting for cloud-init"}
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: "Cloud-init is installing Docker and Coolify (this takes 5-10 min)..."}

	// Wait for Docker to be ready (cloud-init completion indicator)
	// Cloud-init typically takes 8-10 minutes for: package updates + docker install + Coolify install
	maxWait := 10 * time.Minute
	start := time.Now()
	attempt := 0

	for time.Since(start) < maxWait {
		attempt++
		elapsed := time.Since(start).Round(time.Second)
		progress := 0.1 + (float64(elapsed.Seconds()) / maxWait.Seconds() * 0.8)
		progressChan <- ui.StepProgressMsg{Progress: progress, Message: fmt.Sprintf("Checking Docker... (%s elapsed)", elapsed)}

		cmd := exec.Command("ssh",
			"-o", "StrictHostKeyChecking=no",
			"-o", "ConnectTimeout=10",
			"-o", "BatchMode=yes",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "LogLevel=ERROR",
			fmt.Sprintf("%s@%s", p.config.Azure.AdminUsername, publicIP),
			"sudo docker ps 2>/dev/null")

		if err := cmd.Run(); err == nil {
			logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: fmt.Sprintf("Docker is ready (attempt %d, %s)", attempt, elapsed)}
			progressChan <- ui.StepProgressMsg{Progress: 0.95, Message: "Coolify deployed"}
			return nil
		}

		// Log status every 5 attempts (every 2.5 min)
		if attempt%5 == 0 {
			logChan <- ui.LogMsg{Level: ui.LogDebug, Message: fmt.Sprintf("Still waiting... attempt %d (%s elapsed)", attempt, elapsed)}
		}

		time.Sleep(30 * time.Second)
	}

	return fmt.Errorf("timeout waiting for Docker/Coolify after 10 minutes (SSH to %s@%s may not be ready)", p.config.Azure.AdminUsername, publicIP)
}

// runHealthChecks performs health checks
func (p *AzureProvider) runHealthChecks(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	publicIP, ok := p.config.Settings["public_ip"].(string)
	if !ok {
		return fmt.Errorf("public IP not found")
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Checking HTTP endpoint"}

	time.Sleep(10 * time.Second)

	curlCmd := exec.Command("curl", "-f", "-s", "-o", "/dev/null", "--max-time", "10",
		fmt.Sprintf("http://%s:8000", publicIP))

	if err := curlCmd.Run(); err != nil {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: "HTTP check failed (Coolify may need more time to start)"}
	} else {
		logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Coolify is responding on port 8000"}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Health checks complete"}
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: fmt.Sprintf("🚀 Coolify available at: http://%s:8000", publicIP)}

	return nil
}

// getCloudInit returns the cloud-init script
func (p *AzureProvider) getCloudInit() string {
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
  - usermod -aG docker ubuntu
  - mkdir -p /data/coolify/source
  - echo "COOLIFY_POSTGRES_VERSION=17-trixie" >> /data/coolify/source/.env
  - echo "COOLIFY_REDIS_VERSION=8.4.0-bookworm" >> /data/coolify/source/.env
  - curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash
`
}

// getSSHPublicKey reads the SSH public key
func (p *AzureProvider) getSSHPublicKey() (string, error) {
	keyPath := p.config.Azure.SSHKeyPath
	if keyPath == "" {
		keyPath = "~/.ssh/id_rsa.pub"
	}

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
