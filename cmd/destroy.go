package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/providers/azure"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy [provider]",
	Short: "Destroy cloud resources",
	Long: `Destroy cloud infrastructure created by cool-kit.

This command will delete all resources created during deployment
including VMs, networks, and storage. This action cannot be undone.

Examples:
  cool-kit destroy azure
  cool-kit destroy aws
  cool-kit destroy gcp
  cool-kit destroy hetzner
  cool-kit destroy digitalocean
  cool-kit destroy --all  # Destroy based on last deployment`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDestroy,
}

var (
	destroyForce         bool
	destroyAll           bool
	destroyResourceGroup string
)

func init() {
	destroyCmd.Flags().BoolVarP(&destroyForce, "force", "f", false, "Skip confirmation prompts")
	destroyCmd.Flags().BoolVar(&destroyAll, "all", false, "Destroy based on last deployment provider")
	destroyCmd.Flags().StringVar(&destroyResourceGroup, "resource-group", "", "Azure resource group to delete")

	rootCmd.AddCommand(destroyCmd)
}

func runDestroy(cmd *cobra.Command, args []string) error {
	ui.Section("Destroy Resources")

	var provider string

	if len(args) > 0 {
		provider = strings.ToLower(args[0])
	} else if destroyAll {
		cfg := config.Get()
		provider = cfg.Provider
		if provider == "" {
			return fmt.Errorf("no provider specified and no last deployment found")
		}
	} else {
		return fmt.Errorf("please specify a provider: cool-kit destroy [azure|aws|gcp|hetzner|digitalocean]")
	}

	ui.Warning(fmt.Sprintf("This will PERMANENTLY DELETE all %s resources!", strings.ToUpper(provider)))
	fmt.Println()

	switch provider {
	case "azure":
		return destroyAzure()
	case "aws":
		return destroyAWS()
	case "gcp":
		return destroyGCP()
	case "hetzner":
		return destroyHetzner()
	case "digitalocean", "do":
		return destroyDigitalOcean()
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}
}

func destroyAzure() error {
	cfg := config.Get()
	rgName := destroyResourceGroup
	if rgName == "" {
		rgName = cfg.Azure.ResourceGroup
	}
	if rgName == "" {
		rgName = "coolify-rg"
		cfg.Azure.ResourceGroup = rgName
	}

	ui.Info(fmt.Sprintf("Resource Group: %s", rgName))

	// Initialize provider
	provider, err := azure.NewAzureProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize Azure provider: %w", err)
	}

	// Confirm
	if !destroyForce {
		confirm, err := ui.Confirm(fmt.Sprintf("Delete resource group '%s' and ALL its resources?", rgName))
		if err != nil {
			return err
		}
		if !confirm {
			ui.Info("Cancelled")
			return nil
		}
	}

	// Create a channel for logs
	logChan := make(chan ui.LogMsg)

	// Start a goroutine to print logs
	go func() {
		for msg := range logChan {
			switch msg.Level {
			case ui.LogInfo:
				ui.Info(msg.Message)
			case ui.LogSuccess:
				ui.Success(msg.Message)
			case ui.LogWarning:
				ui.Warning(msg.Message)
			case ui.LogError:
				ui.Error(msg.Message)
			case ui.LogDebug:
				// Skip debug logs in destroy command unless verbose
				if IsVerbose() {
					ui.Dim(msg.Message)
				}
			}
		}
	}()

	// Execute Destroy
	if err := provider.Destroy(logChan); err != nil {
		close(logChan)
		return err
	}

	close(logChan)
	return nil
}

func destroyAWS() error {
	cfg := config.Get()

	// Get instance ID from settings
	var instanceID string
	if cfg.Settings != nil {
		if id, ok := cfg.Settings["instance_id"].(string); ok {
			instanceID = id
		}
	}

	if instanceID == "" {
		ui.Warning("No AWS instance ID found in configuration")
		ui.Info("To manually terminate instances, use: aws ec2 terminate-instances --instance-ids <ID>")
		return nil
	}

	region := "us-east-1"
	if r, ok := cfg.Settings["aws_region"].(string); ok {
		region = r
	}

	ui.Info(fmt.Sprintf("Instance: %s (Region: %s)", instanceID, region))

	if !destroyForce {
		confirm, err := ui.Confirm("Terminate this EC2 instance?")
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}
	}

	// Terminate instance
	ui.Info("Terminating EC2 instance...")
	terminateCmd := exec.Command("aws", "ec2", "terminate-instances",
		"--instance-ids", instanceID,
		"--region", region)

	if output, err := terminateCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to terminate instance: %s", string(output))
	}

	ui.Success("EC2 instance terminated")

	// Note about VPC
	ui.Dim("Note: VPC and security groups may need manual cleanup if not using default VPC")
	ui.NextSteps([]string{
		fmt.Sprintf("Check AWS Console: https://%s.console.aws.amazon.com/ec2/", region),
	})

	return nil
}

func destroyGCP() error {
	cfg := config.Get()

	var instanceName, zone, project string
	if cfg.Settings != nil {
		if n, ok := cfg.Settings["gcp_instance"].(string); ok {
			instanceName = n
		}
		if z, ok := cfg.Settings["gcp_zone"].(string); ok {
			zone = z
		}
		if p, ok := cfg.Settings["gcp_project"].(string); ok {
			project = p
		}
	}

	if instanceName == "" {
		instanceName = "coolify-vm"
	}
	if zone == "" {
		zone = "us-central1-a"
	}

	ui.Info(fmt.Sprintf("Instance: %s (Zone: %s)", instanceName, zone))

	if !destroyForce {
		confirm, err := ui.Confirm("Delete this Compute Engine instance?")
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}
	}

	// Delete instance
	ui.Info("Deleting Compute Engine instance...")
	args := []string{"compute", "instances", "delete", instanceName, "--zone", zone, "--quiet"}
	if project != "" {
		args = append(args, "--project", project)
	}

	deleteCmd := exec.Command("gcloud", args...)
	if output, err := deleteCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete instance: %s", string(output))
	}

	ui.Success("Compute Engine instance deleted")
	return nil
}

func destroyHetzner() error {
	cfg := config.Get()

	var serverID string
	if cfg.Settings != nil {
		if id, ok := cfg.Settings["hetzner_server_id"].(int64); ok {
			serverID = fmt.Sprintf("%d", id)
		} else if idStr, ok := cfg.Settings["hetzner_server_id"].(string); ok {
			serverID = idStr
		}
	}

	if serverID == "" {
		ui.Warning("No Hetzner server ID found in configuration")
		ui.Info("Use: hcloud server delete <name>")
		return nil
	}

	ui.Info(fmt.Sprintf("Server ID: %s", serverID))

	if !destroyForce {
		confirm, err := ui.Confirm("Delete this Hetzner server?")
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}
	}

	// Delete server
	ui.Info("Deleting Hetzner server...")
	deleteCmd := exec.Command("hcloud", "server", "delete", serverID)
	if output, err := deleteCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete server: %s", string(output))
	}

	ui.Success("Hetzner server deleted")
	return nil
}

func destroyDigitalOcean() error {
	cfg := config.Get()

	var dropletID string
	if cfg.Settings != nil {
		if id, ok := cfg.Settings["do_droplet_id"].(int); ok {
			dropletID = fmt.Sprintf("%d", id)
		} else if idStr, ok := cfg.Settings["do_droplet_id"].(string); ok {
			dropletID = idStr
		}
	}

	if dropletID == "" {
		ui.Warning("No DigitalOcean droplet ID found in configuration")
		ui.Info("Use: doctl compute droplet delete <name>")
		return nil
	}

	ui.Info(fmt.Sprintf("Droplet ID: %s", dropletID))

	if !destroyForce {
		confirm, err := ui.Confirm("Delete this DigitalOcean droplet?")
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}
	}

	// Delete droplet
	ui.Info("Deleting DigitalOcean droplet...")
	deleteCmd := exec.Command("doctl", "compute", "droplet", "delete", dropletID, "--force")
	if output, err := deleteCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete droplet: %s", string(output))
	}

	ui.Success("DigitalOcean droplet deleted")
	return nil
}
