package cmd

import (
	"fmt"
	"strings"

	"github.com/entro314-labs/cool-kit/internal/api"
	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var instancesCmd = &cobra.Command{
	Use:     "instances",
	Aliases: []string{"instance"},
	Short:   "Manage Coolify instances",
	Long:    "Manage multiple Coolify instances with context switching.",
}

var instancesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all configured instances",
	RunE:    runInstancesList,
}

var instancesAddCmd = &cobra.Command{
	Use:   "add NAME",
	Short: "Add a new Coolify instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesAdd,
}

var instancesRemoveCmd = &cobra.Command{
	Use:     "remove NAME",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a Coolify instance",
	Args:    cobra.ExactArgs(1),
	RunE:    runInstancesRemove,
}

var instancesUseCmd = &cobra.Command{
	Use:   "use NAME",
	Short: "Switch to a different instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesUse,
}

var instancesCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the current instance",
	RunE:  runInstancesCurrent,
}

func init() {
	instancesCmd.AddCommand(instancesListCmd)
	instancesCmd.AddCommand(instancesAddCmd)
	instancesCmd.AddCommand(instancesRemoveCmd)
	instancesCmd.AddCommand(instancesUseCmd)
	instancesCmd.AddCommand(instancesCurrentCmd)

	// Add flags
	instancesAddCmd.Flags().String("url", "", "Coolify URL")
	instancesAddCmd.Flags().String("token", "", "API token")
	instancesAddCmd.Flags().Bool("default", false, "Set as default instance")
}

func runInstancesList(cmd *cobra.Command, args []string) error {
	instances, err := config.ListInstances()
	if err != nil {
		return fmt.Errorf("failed to load instances: %w", err)
	}

	if len(instances) == 0 {
		ui.Info("No instances configured")
		ui.Spacer()
		ui.NextSteps([]string{
			fmt.Sprintf("Run '%s login' to add your first instance", execName()),
			fmt.Sprintf("Run '%s instances add' to add another instance", execName()),
		})
		return nil
	}

	ui.Section("Configured Instances")
	ui.Spacer()

	currentInst, _ := config.GetCurrentInstance()
	currentName := ""
	if currentInst != nil {
		currentName = currentInst.Name
	}

	for _, inst := range instances {
		marker := "  "
		if inst.Name == currentName {
			marker = "→ "
		}

		defaultMarker := ""
		if inst.Default {
			defaultMarker = " " + ui.SuccessStyle.Render("(default)")
		}

		ui.Print(fmt.Sprintf("%s%s - %s%s",
			marker,
			ui.BoldStyle.Render(inst.Name),
			ui.DimStyle.Render(inst.FQDN),
			defaultMarker,
		))
	}

	ui.Spacer()
	ui.Spacer()
	ui.NextSteps([]string{
		fmt.Sprintf("Run '%s instances use NAME' to switch instance", execName()),
		fmt.Sprintf("Run '%s instances current' to show active instance", execName()),
	})

	return nil
}

func runInstancesAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Get flags
	url, _ := cmd.Flags().GetString("url")
	token, _ := cmd.Flags().GetString("token")
	setAsDefault, _ := cmd.Flags().GetBool("default")

	// Interactive prompts if flags not provided
	if url == "" {
		ui.Section("Add Coolify Instance")
		var err error
		url, err = ui.Input("Coolify URL", "https://coolify.example.com")
		if err != nil {
			return err
		}
	}

	url = strings.TrimSuffix(url, "/")
	if url == "" {
		return fmt.Errorf("Coolify URL is required")
	}

	if token == "" {
		ui.Spacer()
		ui.Dim("→ Get your API token from Settings → API Tokens in Coolify")
		var err error
		token, err = ui.Password("API Token")
		if err != nil {
			return err
		}
	}

	if token == "" {
		return fmt.Errorf("API token is required")
	}

	// Validate credentials
	ui.Spacer()
	ui.Info("Validating credentials...")
	client := api.NewClient(url, token)
	if err := client.HealthCheck(); err != nil {
		ui.Error("Connection failed")
		return fmt.Errorf("failed to connect: %w", err)
	}
	ui.Success("Connection verified")

	// Add instance
	if err := config.AddInstance(name, url, token, setAsDefault); err != nil {
		return fmt.Errorf("failed to add instance: %w", err)
	}

	ui.Spacer()
	ui.Success(fmt.Sprintf("Instance '%s' added successfully", name))

	if setAsDefault {
		ui.KeyValue("Status", "Set as default instance")
	}

	return nil
}

func runInstancesRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Confirm deletion
	confirmed, err := ui.Confirm(fmt.Sprintf("Remove instance '%s'?", name))
	if err != nil {
		return err
	}

	if !confirmed {
		ui.Dim("Cancelled")
		return nil
	}

	if err := config.RemoveInstance(name); err != nil {
		return fmt.Errorf("failed to remove instance: %w", err)
	}

	ui.Success(fmt.Sprintf("Instance '%s' removed", name))
	return nil
}

func runInstancesUse(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := config.UseInstance(name); err != nil {
		return fmt.Errorf("failed to switch instance: %w", err)
	}

	ui.Success(fmt.Sprintf("Switched to instance '%s'", name))
	return nil
}

func runInstancesCurrent(cmd *cobra.Command, args []string) error {
	inst, err := config.GetCurrentInstance()
	if err != nil {
		return err
	}

	ui.Section("Current Instance")
	ui.Spacer()
	ui.KeyValue("Name", inst.Name)
	ui.KeyValue("URL", inst.FQDN)

	if inst.Default {
		ui.KeyValue("Status", ui.SuccessStyle.Render("Default"))
	}

	return nil
}
