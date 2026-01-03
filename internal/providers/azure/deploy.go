package azure

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/entro314-labs/cool-kit/internal/azureconfig"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// Deploy provisions Azure VM and deploys Coolify
func Deploy() error {
	ui.Section("Azure Coolify Deployment")
	ui.Dim("Deploying Coolify to Microsoft Azure")

	// Load or create configuration
	ctx, err := loadDeploymentContext()
	if err != nil {
		return err
	}

	// Display deployment plan
	displayDeploymentPlan(ctx)

	// Confirm deployment
	proceed, err := ui.Confirm("Proceed with deployment?")
	if err != nil {
		return err
	}
	if !proceed {
		ui.Info("Deployment cancelled")
		return nil
	}

	// Step 1: Provision Azure infrastructure
	provisioner := NewProvisioner(ctx)
	if err := provisioner.Provision(); err != nil {
		return fmt.Errorf("infrastructure provisioning failed: %w", err)
	}

	// Wait for VM to be fully ready
	if err := provisioner.WaitForVM(5 * time.Minute); err != nil {
		return fmt.Errorf("VM did not become ready: %w", err)
	}

	// Step 2: Create SSH client
	sshClient := NewSSHClient(ctx.PublicIP, ctx.AdminUsername, ctx.SSHKeyPath)

	// Step 3: Install Coolify
	installer := NewInstaller(ctx, sshClient)
	if err := installer.Install(); err != nil {
		return fmt.Errorf("Coolify installation failed: %w", err)
	}

	// Step 4: Display success information
	displaySuccessInfo(ctx, installer)

	return nil
}

// loadDeploymentContext loads or creates deployment configuration
func loadDeploymentContext() (*DeploymentContext, error) {
	ui.Info("Loading configuration")

	// Check for existing config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".coolify", "azure-config.json")
	cfg, err := azureconfig.LoadOrDefault(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		ui.Warning("Configuration validation failed")
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Collect deployment information
	resourceGroup, err := ui.InputWithDefault("Resource Group name", "coolify-rg")
	if err != nil {
		return nil, err
	}

	vmName, err := ui.InputWithDefault("VM name", "coolify-vm")
	if err != nil {
		return nil, err
	}

	adminEmail, err := ui.Input("Admin email", "")
	if err != nil {
		return nil, err
	}

	adminPassword, err := ui.Password("Admin password")
	if err != nil {
		return nil, err
	}

	// Expand SSH key path
	sshKeyPath := cfg.Infrastructure.SSHPublicKeyPath
	if sshKeyPath[:2] == "~/" {
		sshKeyPath = filepath.Join(homeDir, sshKeyPath[2:])
	}

	// Create deployment context
	ctx := &DeploymentContext{
		Config:        cfg,
		ResourceGroup: resourceGroup,
		VMName:        vmName,
		Location:      cfg.Infrastructure.Location,
		VMSize:        cfg.Infrastructure.VMSize,
		VNetName:      fmt.Sprintf("%s-vnet", resourceGroup),
		SubnetName:    fmt.Sprintf("%s-subnet", resourceGroup),
		NSGName:       fmt.Sprintf("%s-nsg", resourceGroup),
		PublicIPName:  fmt.Sprintf("%s-ip", resourceGroup),
		NICName:       fmt.Sprintf("%s-nic", resourceGroup),
		AdminUsername: cfg.Infrastructure.AdminUsername,
		SSHKeyPath:    sshKeyPath,
		AdminEmail:    adminEmail,
		AdminPassword: adminPassword,
	}

	ui.Dim("Configuration loaded successfully")
	return ctx, nil
}

// displayDeploymentPlan shows the deployment plan
func displayDeploymentPlan(ctx *DeploymentContext) {
	ui.Info("Deployment Plan")
	ui.Dim(fmt.Sprintf("Resource Group: %s", ctx.ResourceGroup))
	ui.Dim(fmt.Sprintf("Location: %s", ctx.Location))
	ui.Dim(fmt.Sprintf("VM Name: %s", ctx.VMName))
	ui.Dim(fmt.Sprintf("VM Size: %s", ctx.VMSize))
	ui.Dim(fmt.Sprintf("Admin User: %s", ctx.AdminUsername))
	ui.Dim("")
	ui.Dim("Resources to create:")
	ui.Dim("  • Virtual Network")
	ui.Dim("  • Subnet")
	ui.Dim("  • Network Security Group (SSH, HTTP, HTTPS, Coolify ports)")
	ui.Dim("  • Public IP Address")
	ui.Dim("  • Network Interface")
	ui.Dim("  • Virtual Machine")
	ui.Dim("  • Coolify containers (app, database, redis, realtime)")
	ui.Dim("")
}

// displaySuccessInfo shows success information after deployment
func displaySuccessInfo(ctx *DeploymentContext, installer *Installer) {
	ui.Success("Deployment completed successfully!")
	ui.Dim("")
	ui.Info("Access Information")
	ui.Info(fmt.Sprintf("Coolify URL: %s", installer.GetInstallationURL()))
	ui.Info(fmt.Sprintf("Admin Email: %s", ctx.AdminEmail))
	ui.Dim(fmt.Sprintf("Public IP: %s", ctx.PublicIP))
	ui.Dim(fmt.Sprintf("SSH: ssh %s@%s", ctx.AdminUsername, ctx.PublicIP))
	ui.Dim("")
	ui.Info("Next Steps")
	ui.Dim("1. Visit the Coolify URL to complete setup")
	ui.Dim("2. Log in with your admin credentials")
	ui.Dim("3. Configure your first server")
	ui.Dim("4. Deploy your applications")
	ui.Dim("")
	ui.Info("Management Commands")
	ui.Dim("  cool-kit azure status    - Check instance status")
	ui.Dim("  cool-kit azure update    - Update Coolify")
	ui.Dim("  cool-kit azure backup    - Create backup")
	ui.Dim("  cool-kit azure ssh       - SSH into instance")
	ui.Dim("")
}

// Update updates an existing Azure Coolify instance
func Update() error {
	// Load deployment context
	ctx, err := loadExistingDeploymentContext()
	if err != nil {
		return err
	}

	// Create SSH client
	sshClient := NewSSHClient(ctx.PublicIP, ctx.AdminUsername, ctx.SSHKeyPath)

	// Ask about auto-rollback
	autoRollback, err := ui.Confirm("Enable automatic rollback on failure?")
	if err != nil {
		return err
	}

	// Create updater and perform update
	updater := NewUpdater(ctx, sshClient)
	if err := updater.Update(autoRollback); err != nil {
		return err
	}

	ui.Dim("")
	ui.Info("Update Complete")
	ui.Dim("Your Coolify instance has been updated to the latest version")
	ui.Dim(fmt.Sprintf("Access: http://%s:8000", ctx.PublicIP))

	return nil
}

// Status checks the status of Azure Coolify instance
func Status() error {
	// Load deployment context
	ctx, err := loadExistingDeploymentContext()
	if err != nil {
		return err
	}

	// Create SSH client
	sshClient := NewSSHClient(ctx.PublicIP, ctx.AdminUsername, ctx.SSHKeyPath)

	// Create status checker and perform check
	checker := NewStatusChecker(ctx, sshClient)
	return checker.CheckStatus()
}

// Backup creates a backup of the Azure instance
func Backup() error {
	// Load deployment context
	ctx, err := loadExistingDeploymentContext()
	if err != nil {
		return err
	}

	// Create SSH client
	sshClient := NewSSHClient(ctx.PublicIP, ctx.AdminUsername, ctx.SSHKeyPath)

	// Create backup manager
	manager := NewBackupManager(ctx, sshClient)

	// Ask for backup type
	backupType, err := ui.InputWithDefault("Backup type", "manual")
	if err != nil {
		return err
	}

	// Create backup
	backup, err := manager.CreateBackup(backupType)
	if err != nil {
		return err
	}

	ui.Dim("")
	ui.Info("Backup Information")
	ui.Dim(fmt.Sprintf("Backup ID: %s", backup.ID))
	ui.Dim(fmt.Sprintf("Created: %s", backup.Timestamp))
	if backup.Version != "" {
		ui.Dim(fmt.Sprintf("Coolify Version: %s", backup.Version))
	}
	ui.Dim("")
	ui.Dim("Use 'cool-kit azure rollback' to restore from this backup")

	return nil
}

// Rollback rolls back to a previous version
func Rollback() error {
	// Load deployment context
	ctx, err := loadExistingDeploymentContext()
	if err != nil {
		return err
	}

	// Create SSH client
	sshClient := NewSSHClient(ctx.PublicIP, ctx.AdminUsername, ctx.SSHKeyPath)

	// Create backup manager
	manager := NewBackupManager(ctx, sshClient)

	// List available backups
	backups, err := manager.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backups) == 0 {
		ui.Warning("No backups available")
		return nil
	}

	ui.Info("Available Backups")
	for i, backup := range backups {
		ui.Dim(fmt.Sprintf("%d. %s (Type: %s, Created: %s)", i+1, backup.ID, backup.Type, backup.Timestamp))
		if backup.Version != "" {
			ui.Dim(fmt.Sprintf("   Version: %s", backup.Version))
		}
	}
	ui.Dim("")

	// Get backup selection
	backupNum, err := ui.InputWithDefault("Select backup number to restore", "1")
	if err != nil {
		return err
	}

	var backupID string

	// Parse selection
	var selectedIndex int
	fmt.Sscanf(backupNum, "%d", &selectedIndex)
	if selectedIndex < 1 || selectedIndex > len(backups) {
		return fmt.Errorf("invalid backup selection")
	}
	backupID = backups[selectedIndex-1].ID

	// Confirm rollback
	confirmed, err := ui.Confirm(fmt.Sprintf("Rollback to backup %s? This will stop Coolify and restore all data", backupID))
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Info("Rollback cancelled")
		return nil
	}

	// Perform rollback
	if err := manager.RestoreBackup(backupID); err != nil {
		return err
	}

	ui.Dim("")
	ui.Dim(fmt.Sprintf("Access: http://%s:8000", ctx.PublicIP))

	return nil
}

// ManageConfig manages Azure deployment configuration
func ManageConfig() error {
	ui.Section("Azure Configuration Management")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".coolify", "azure-config.json")

	// Show menu
	ui.Info("Configuration Options")
	ui.Dim("1. View current configuration")
	ui.Dim("2. Edit configuration file")
	ui.Dim("3. Validate configuration")
	ui.Dim("4. Reset to defaults")
	ui.Dim("5. Show config file path")
	ui.Dim("")

	choice, err := ui.InputWithDefault("Select option", "1")
	if err != nil {
		return err
	}

	switch choice {
	case "1":
		// View configuration
		cfg, err := azureconfig.LoadOrDefault(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		ui.Info("Current Configuration")
		ui.Dim(fmt.Sprintf("Location: %s", cfg.Infrastructure.Location))
		ui.Dim(fmt.Sprintf("VM Size: %s", cfg.Infrastructure.VMSize))
		ui.Dim(fmt.Sprintf("Admin Username: %s", cfg.Infrastructure.AdminUsername))
		ui.Dim(fmt.Sprintf("VNet Prefix: %s", cfg.Networking.VNetAddressPrefix))
		ui.Dim(fmt.Sprintf("Subnet Prefix: %s", cfg.Networking.SubnetAddressPrefix))

	case "2":
		// Edit configuration
		ui.Info(fmt.Sprintf("Edit configuration at: %s", configPath))
		ui.Dim("After editing, run 'cool-kit azure config' and choose option 3 to validate")

	case "3":
		// Validate configuration
		cfg, err := azureconfig.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		if err := cfg.Validate(); err != nil {
			ui.Error("Configuration validation failed")
			return err
		}

		ui.Success("Configuration is valid")

	case "4":
		// Reset to defaults
		confirmed, err := ui.Confirm("Reset configuration to defaults?")
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Info("Reset cancelled")
			return nil
		}

		cfg := azureconfig.DefaultConfig()
		if err := azureconfig.Save(cfg, configPath); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		ui.Success("Configuration reset to defaults")

	case "5":
		// Show config path
		ui.Info(fmt.Sprintf("Configuration file: %s", configPath))

	default:
		ui.Warning("Invalid option")
	}

	return nil
}

// SSH opens an SSH connection to the Azure instance
func SSH() error {
	ui.Section("Azure SSH Connection")

	// Load deployment context
	ctx, err := loadExistingDeploymentContext()
	if err != nil {
		return err
	}

	ui.Dim(fmt.Sprintf("Connecting to %s@%s", ctx.AdminUsername, ctx.PublicIP))
	ui.Dim("")

	// Create SSH client
	sshClient := NewSSHClient(ctx.PublicIP, ctx.AdminUsername, ctx.SSHKeyPath)

	// Open interactive session
	return sshClient.Interactive()
}

// loadExistingDeploymentContext loads deployment context for existing deployment
func loadExistingDeploymentContext() (*DeploymentContext, error) {
	ui.Info("Loading deployment configuration")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".coolify", "azure-config.json")
	cfg, err := azureconfig.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w\nRun 'cool-kit azure deploy' first", err)
	}

	// Get resource group and VM name from user or config
	resourceGroup, err := ui.InputWithDefault("Resource Group name", "coolify-rg")
	if err != nil {
		return nil, err
	}

	vmName, err := ui.InputWithDefault("VM name", "coolify-vm")
	if err != nil {
		return nil, err
	}

	// Get public IP from Azure
	ctx := &DeploymentContext{
		Config:        cfg,
		ResourceGroup: resourceGroup,
		VMName:        vmName,
		Location:      cfg.Infrastructure.Location,
		VMSize:        cfg.Infrastructure.VMSize,
		AdminUsername: cfg.Infrastructure.AdminUsername,
	}

	// Expand SSH key path
	sshKeyPath := cfg.Infrastructure.SSHPublicKeyPath
	if len(sshKeyPath) >= 2 && sshKeyPath[:2] == "~/" {
		sshKeyPath = filepath.Join(homeDir, sshKeyPath[2:])
	}
	ctx.SSHKeyPath = sshKeyPath

	// Get public IP from Azure
	provisioner := NewProvisioner(ctx)
	if err := provisioner.getPublicIP(); err != nil {
		return nil, fmt.Errorf("failed to get public IP: %w", err)
	}

	ui.Dim("Configuration loaded")
	return ctx, nil
}
