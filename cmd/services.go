package cmd

import (
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/api"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage services (databases, caches, etc.)",
	Long: `Manage services associated with your applications.

Services include databases (PostgreSQL, MySQL, MongoDB), caches (Redis),
search engines (Meilisearch, Elasticsearch), and other infrastructure.

Available Commands:
  services ls      - List all services
  services create  - Create a new service
  services rm      - Remove a service
  services info    - Show service details`,
}

var servicesListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all services",
	Long:  `List all databases and services in your Coolify instance.`,
	RunE:  runServicesList,
}

var servicesInfoCmd = &cobra.Command{
	Use:   "info [UUID]",
	Short: "Show service details",
	Long:  `Display detailed information about a specific service including connection details.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runServicesInfo,
}

var servicesRemoveCmd = &cobra.Command{
	Use:   "rm [UUID]",
	Short: "Remove a service",
	Long:  `Remove a service from Coolify. This will stop and delete the service.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runServicesRemove,
}

func init() {
	servicesCmd.AddCommand(servicesListCmd)
	servicesCmd.AddCommand(servicesInfoCmd)
	servicesCmd.AddCommand(servicesRemoveCmd)
}

func runServicesList(cmd *cobra.Command, args []string) error {
	if err := checkLogin(); err != nil {
		return err
	}

	instance, err := getCurrentInstance()
	if err != nil {
		return err
	}

	client := api.NewClient(instance.FQDN, instance.Token)

	ui.Section("Services")

	var databases []api.Database
	err = ui.RunTasks([]ui.Task{
		{
			Name:         "load-databases",
			ActiveName:   "Loading services...",
			CompleteName: "✓ Loaded services",
			Action: func() error {
				var err error
				databases, err = client.ListDatabases()
				return err
			},
		},
	})
	if err != nil {
		ui.Error("Failed to load services")
		return fmt.Errorf("failed to list databases: %w", err)
	}

	if len(databases) == 0 {
		ui.Dim("No services found")
		ui.Spacer()
		ui.NextSteps([]string{
			"Deploy an application with smart detection to auto-provision services",
			"Or create services manually in the Coolify dashboard",
		})
		return nil
	}

	ui.Spacer()
	for _, db := range databases {
		statusIcon := "●"
		if db.Status == "running" {
			statusIcon = ui.SuccessStyle.Render("●")
		} else if db.Status == "exited" || db.Status == "stopped" {
			statusIcon = ui.ErrorStyle.Render("●")
		}

		ui.KeyValue(
			fmt.Sprintf("%s %s", statusIcon, db.Name),
			fmt.Sprintf("%s (%s)", db.Type, db.UUID[:8]+"..."),
		)
	}

	return nil
}

func runServicesInfo(cmd *cobra.Command, args []string) error {
	uuid := args[0]

	if err := checkLogin(); err != nil {
		return err
	}

	instance, err := getCurrentInstance()
	if err != nil {
		return err
	}

	client := api.NewClient(instance.FQDN, instance.Token)

	ui.Section(fmt.Sprintf("Service: %s", uuid[:8]+"..."))

	var database *api.Database
	err = ui.RunTasks([]ui.Task{
		{
			Name:         "load-database",
			ActiveName:   "Loading service details...",
			CompleteName: "✓ Loaded service",
			Action: func() error {
				var err error
				database, err = client.GetDatabase(uuid)
				return err
			},
		},
	})
	if err != nil {
		ui.Error("Failed to load service")
		return fmt.Errorf("failed to get database: %w", err)
	}

	ui.Spacer()
	ui.KeyValue("Name", database.Name)
	ui.KeyValue("Type", database.Type)
	ui.KeyValue("UUID", database.UUID)
	ui.KeyValue("Image", database.Image)
	ui.KeyValue("Status", database.Status)
	ui.KeyValue("Public", fmt.Sprintf("%t", database.IsPublic))

	if database.Description != "" {
		ui.Spacer()
		ui.Dim(database.Description)
	}

	return nil
}

func runServicesRemove(cmd *cobra.Command, args []string) error {
	uuid := args[0]

	if err := checkLogin(); err != nil {
		return err
	}

	instance, err := getCurrentInstance()
	if err != nil {
		return err
	}

	client := api.NewClient(instance.FQDN, instance.Token)

	// Get service info first
	var database *api.Database
	err = ui.RunTasks([]ui.Task{
		{
			Name:         "load-database",
			ActiveName:   "Loading service...",
			CompleteName: "✓ Loaded service",
			Action: func() error {
				var err error
				database, err = client.GetDatabase(uuid)
				return err
			},
		},
	})
	if err != nil {
		ui.Error("Failed to load service")
		return fmt.Errorf("failed to get database: %w", err)
	}

	ui.Section(fmt.Sprintf("Remove Service: %s", database.Name))
	ui.KeyValue("Type", database.Type)
	ui.KeyValue("UUID", database.UUID)
	ui.Spacer()

	confirmed, err := ui.Confirm("Are you sure you want to remove this service?")
	if err != nil {
		return err
	}

	if !confirmed {
		ui.Dim("Cancelled")
		return nil
	}

	err = ui.RunTasks([]ui.Task{
		{
			Name:         "delete-database",
			ActiveName:   "Removing service...",
			CompleteName: "✓ Service removed",
			Action: func() error {
				return client.DeleteDatabase(uuid)
			},
		},
	})
	if err != nil {
		ui.Error("Failed to remove service")
		return fmt.Errorf("failed to delete database: %w", err)
	}

	ui.Success("Service removed successfully")

	return nil
}
