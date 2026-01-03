package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/entro314-labs/cool-kit/internal/templates"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project with Coolify templates",
	RunE:  runInit,
}

func init() {
	// Add flags if needed in the future
}

func runInit(cmd *cobra.Command, args []string) error {
	ui.Section("Initialize Project")

	// List available templates
	files, err := templates.ListTemplates()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	selectedFiles, err := ui.MultiSelect("Select templates to generate:", files)
	if err != nil {
		return err
	}

	if len(selectedFiles) == 0 {
		ui.Dim("No templates selected")
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	for _, file := range selectedFiles {
		targetPath := filepath.Join(cwd, file)

		// Check if file exists
		if _, err := os.Stat(targetPath); err == nil {
			overwrite, err := ui.Confirm(fmt.Sprintf("File %s already exists. Overwrite?", file))
			if err != nil {
				return err
			}
			if !overwrite {
				ui.Dim(fmt.Sprintf("Skipping %s", file))
				continue
			}
		}

		if err := templates.WriteTemplate(file, targetPath); err != nil {
			return fmt.Errorf("failed to write %s: %w", file, err)
		}
		ui.Success(fmt.Sprintf("Created %s", file))
	}

	ui.Spacer()
	ui.Success("Project initialized successfully!")
	return nil
}
