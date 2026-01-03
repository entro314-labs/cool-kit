package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			currentVal := viper.Get(key)
			currentStr := fmt.Sprintf("%v", currentVal)
			if currentVal == nil {
				currentStr = ""
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

		// Get config path and remove it
		configPath := viper.ConfigFileUsed()
		if configPath == "" {
			ui.Warning("No configuration file found")
			return nil
		}

		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove config file: %w", err)
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
		// Get all settings
		allSettings := viper.AllSettings()

		// Pretty print JSON
		data, err := json.MarshalIndent(allSettings, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		if len(args) == 0 {
			// Output to stdout
			fmt.Println(string(data))
		} else {
			// Write to file
			filePath := args[0]
			if err := os.WriteFile(filePath, data, 0644); err != nil {
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

		// Parse JSON
		var imported map[string]interface{}
		if err := json.Unmarshal(data, &imported); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		// Confirm import
		confirm, err := ui.Confirm(fmt.Sprintf("Import %d settings from %s?", len(imported), filePath))
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}

		// Merge into viper
		for key, value := range imported {
			viper.Set(key, value)
		}

		// Save
		if err := saveConfig(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		ui.Success(fmt.Sprintf("Imported configuration from %s", filePath))
		return nil
	},
}

// setConfigValue sets a config value with proper type conversion
func setConfigValue(key, value string) error {
	// Try to parse as int
	if i, err := strconv.Atoi(value); err == nil {
		viper.Set(key, i)
	} else if b, err := strconv.ParseBool(value); err == nil {
		// Try to parse as bool
		viper.Set(key, b)
	} else {
		// Keep as string
		viper.Set(key, value)
	}

	return saveConfig()
}

// saveConfig writes viper config to file
func saveConfig() error {
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		// Create default config path
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir := home + "/.coolify-cli"
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
		configPath = configDir + "/config.yaml"
	}

	// Ensure directory exists
	dir := configPath[:strings.LastIndex(configPath, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return viper.WriteConfigAs(configPath)
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configExportCmd)
	configCmd.AddCommand(configImportCmd)

	rootCmd.AddCommand(configCmd)
}
