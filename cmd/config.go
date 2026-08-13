package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "View and manage Coolify CLI configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show and edit configuration interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		for {
			key, err := ui.RunConfigMenu()
			if err != nil {
				return err
			}
			if key == "" {
				return nil // User exited
			}

			// Get current value
			currentStr, err := getConfigValue(key)
			if err != nil {
				ui.Error(fmt.Sprintf("Failed to load %s: %v", key, err))
				continue
			}

			// Ask for new value
			newVal, err := ui.Input(fmt.Sprintf("Enter new value for %s", key), currentStr)
			if err != nil {
				return err
			}

			// Set and save
			if err := setConfigValue(key, newVal); err != nil {
				ui.Error(fmt.Sprintf("Failed to set %s: %v", key, err))
			} else {
				ui.Success(fmt.Sprintf("Set %s = %s", key, newVal))
			}
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set KEY VALUE",
	Short: "Set configuration value",
	Long: `Set a configuration value. Supports nested keys with dot notation.

Examples:
  cool-kit config set provider azure
  cool-kit config set azure.location westeurope
  cool-kit config set azure.vm_size Standard_D2s_v3
  cool-kit config set local.app_port 8000`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if err := setConfigValue(key, value); err != nil {
			return err
		}

		ui.Success(fmt.Sprintf("Set %s = %s", key, value))
		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Long:  "Reset all configuration values to their defaults. This cannot be undone.",
	RunE: func(cmd *cobra.Command, args []string) error {
		confirm, err := ui.Confirm("Reset all configuration to defaults? This cannot be undone.")
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}

		// Reset in-memory config to defaults and persist.
		defaultCfg := config.New()
		if err := config.Save(defaultCfg); err != nil {
			return fmt.Errorf("failed to reset configuration: %w", err)
		}

		ui.Success("Configuration reset to defaults")
		return nil
	},
}

var configExportCmd = &cobra.Command{
	Use:   "export [FILE]",
	Short: "Export configuration to JSON file",
	Long: `Export current configuration to a JSON file.
If no file is specified, outputs to stdout.

Examples:
  cool-kit config export > my-config.json
  cool-kit config export my-backup.json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := config.ExportConfig()
		if err != nil {
			return fmt.Errorf("failed to export configuration: %w", err)
		}

		if len(args) == 0 {
			// Output to stdout
			fmt.Println(data)
		} else {
			// Write to file
			filePath := args[0]
			if err := os.WriteFile(filePath, []byte(data), 0644); err != nil {
				return fmt.Errorf("failed to write config to %s: %w", filePath, err)
			}
			ui.Success(fmt.Sprintf("Configuration exported to %s", filePath))
		}

		return nil
	},
}

var configImportCmd = &cobra.Command{
	Use:   "import FILE",
	Short: "Import configuration from JSON file",
	Long: `Import configuration from a JSON file.
This will merge the imported values with existing configuration.

Examples:
  cool-kit config import my-backup.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Read file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		// Confirm import
		confirm, err := ui.Confirm(fmt.Sprintf("Import configuration from %s?", filePath))
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}

		if err := config.ImportConfig(string(data)); err != nil {
			return fmt.Errorf("failed to import configuration: %w", err)
		}

		ui.Success(fmt.Sprintf("Imported configuration from %s", filePath))
		return nil
	},
}

// setConfigValue sets a config value with proper type conversion
func setConfigValue(key, value string) error {
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	switch key {
	case "provider":
		cfg.Provider = value
	case "environment":
		cfg.Environment = value
	case "local.app_port":
		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer for %s: %q", key, value)
		}
		cfg.Local.AppPort = i
	case "local.websocket_port":
		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer for %s: %q", key, value)
		}
		cfg.Local.WebSocketPort = i
	case "azure.location":
		cfg.Azure.Location = value
	case "azure.vm_size":
		cfg.Azure.VMSize = value
	default:
		return fmt.Errorf("unsupported config key: %s", key)
	}

	return config.Save(cfg)
}

func getConfigValue(key string) (string, error) {
	cfg := config.Get()
	if cfg == nil {
		return "", fmt.Errorf("configuration not initialized")
	}

	switch key {
	case "provider":
		return cfg.Provider, nil
	case "environment":
		return cfg.Environment, nil
	case "local.app_port":
		return strconv.Itoa(cfg.Local.AppPort), nil
	case "local.websocket_port":
		return strconv.Itoa(cfg.Local.WebSocketPort), nil
	case "azure.location":
		return cfg.Azure.Location, nil
	case "azure.vm_size":
		return cfg.Azure.VMSize, nil
	default:
		return "", fmt.Errorf("unsupported config key: %s", key)
	}
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configExportCmd)
	configCmd.AddCommand(configImportCmd)

}
