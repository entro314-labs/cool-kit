package appdeploy

import (
	"context"
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/api"
	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/smart"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// provisionServicesTask provisions detected services and updates environment variables
func provisionServicesTask(client *api.Client, projectCfg *config.ProjectConfig, deploymentConfig *smart.DeploymentConfig) ui.Task {
	return ui.Task{
		Name:         "provision-services",
		ActiveName:   fmt.Sprintf("Provisioning %d service(s)...", len(deploymentConfig.Services)),
		CompleteName: fmt.Sprintf("✓ Provisioned %d service(s)", len(deploymentConfig.Services)),
		Action: func() error {
			// Create provisioner
			provisioner := smart.NewServiceProvisioner(
				client,
				projectCfg.ProjectUUID,
				projectCfg.EnvironmentUUID,
				projectCfg.ServerUUID,
				projectCfg.Name,
			)

			// Provision all services
			result, err := provisioner.Provision(deploymentConfig)
			if err != nil {
				return fmt.Errorf("service provisioning failed: %w", err)
			}

			// Update application with generated environment variables
			if len(result.EnvironmentVars) > 0 {
				ctx := context.Background()
				var envVars []api.EnvironmentVariable
				for key, value := range result.EnvironmentVars {
					envVars = append(envVars, api.EnvironmentVariable{
						Key:   key,
						Value: value,
					})
				}

				_, err = client.UpdateApplicationEnvsBulk(ctx, projectCfg.AppUUID, envVars)
				if err != nil {
					return fmt.Errorf("failed to update environment variables: %w", err)
				}
			}

			// Report any errors from optional services
			if len(result.Errors) > 0 {
				for _, serviceErr := range result.Errors {
					ui.Dim(fmt.Sprintf("  Warning: %s", serviceErr.Error()))
				}
			}

			return nil
		},
	}
}
