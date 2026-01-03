package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/entro314-labs/cool-kit/internal/service"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:     "keys",
	Aliases: []string{"key", "private-key", "private-keys"},
	Short:   "Manage SSH private keys",
	Long:    "Manage SSH private keys in your Coolify instance.",
}

var keysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all private keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		svc := service.NewPrivateKeyService(client)
		keys, err := svc.List(context.Background())
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("format")
		return formatOutput(format, keys)
	},
}

var keysGetCmd = &cobra.Command{
	Use:   "get <uuid>",
	Short: "Get a private key by UUID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		svc := service.NewPrivateKeyService(client)
		key, err := svc.Get(context.Background(), args[0])
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("format")
		return formatOutput(format, key)
	},
}

var keysAddCmd = &cobra.Command{
	Use:   "add <name> <private-key>",
	Short: "Add a new private key",
	Long: `Add a new private key to Coolify.

Use @filename to read the key from a file:
  cool-kit keys add mykey @~/.ssh/id_rsa`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		name := args[0]
		keyContent := args[1]

		// Handle @filename syntax
		if strings.HasPrefix(keyContent, "@") {
			filename := strings.TrimPrefix(keyContent, "@")
			// Expand ~ to home directory
			if strings.HasPrefix(filename, "~") {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to get home directory: %w", err)
				}
				filename = strings.Replace(filename, "~", home, 1)
			}
			data, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("failed to read key file: %w", err)
			}
			keyContent = string(data)
		}

		description, _ := cmd.Flags().GetString("description")

		svc := service.NewPrivateKeyService(client)
		key, err := svc.Create(context.Background(), service.PrivateKeyCreateRequest{
			Name:        name,
			Description: description,
			PrivateKey:  keyContent,
		})
		if err != nil {
			return err
		}

		fmt.Printf("✅ Private key '%s' created with UUID: %s\n", key.Name, key.UUID)
		return nil
	},
}

var keysRemoveCmd = &cobra.Command{
	Use:     "remove <uuid>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a private key",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		svc := service.NewPrivateKeyService(client)
		if err := svc.Delete(context.Background(), args[0]); err != nil {
			return err
		}

		fmt.Printf("✅ Private key '%s' removed\n", args[0])
		return nil
	},
}

func init() {
	// Add format flags
	keysListCmd.Flags().String("format", "table", "Output format: table, json, pretty")
	keysGetCmd.Flags().String("format", "table", "Output format: table, json, pretty")
	keysAddCmd.Flags().String("description", "", "Key description")

	// Wire commands
	keysCmd.AddCommand(keysListCmd)
	keysCmd.AddCommand(keysGetCmd)
	keysCmd.AddCommand(keysAddCmd)
	keysCmd.AddCommand(keysRemoveCmd)
}
