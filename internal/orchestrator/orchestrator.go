package orchestrator

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/service"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// Orchestrator manages the complete deployment flow
type Orchestrator struct {
	config            *config.Config
	provider          string
	deploymentService service.DeploymentService
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(cfg *config.Config, provider string) *Orchestrator {
	return &Orchestrator{
		config:            cfg,
		provider:          provider,
		deploymentService: service.NewDeploymentService(cfg),
	}
}

// Deploy performs the complete deployment with progress tracking
func (o *Orchestrator) Deploy() (*service.DeploymentResult, error) {
	// Get deployment steps based on provider
	steps := o.deploymentService.GetDeploymentSteps(o.provider)

	// Create progress model
	progressModel := ui.NewProgressModel(o.provider, steps)

	// Create channels for communication
	progressChan := make(chan ui.StepProgressMsg, 100)
	logChan := make(chan ui.LogMsg, 100)
	resultChan := make(chan *service.DeploymentResult, 1)
	errChan := make(chan error, 1)

	// Start deployment in background
	go func() {
		result, err := o.deploymentService.Deploy(context.Background(), o.provider, progressChan, logChan)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	// Start TUI
	program := tea.NewProgram(progressModel)

	// Forward messages from deployment to TUI
	go func() {
		for {
			select {
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				program.Send(progress)
			case log, ok := <-logChan:
				if !ok {
					return
				}
				program.Send(log)
			case result := <-resultChan:
				program.Send(ui.DeploymentCompleteMsg{
					Success: true,
					Message: fmt.Sprintf("Deployment complete! Access Coolify at: %s", result.DashboardURL),
				})
				// Send result back for return value
				resultChan <- result
				return
			case err := <-errChan:
				program.Send(ui.DeploymentCompleteMsg{
					Success: false,
					Message: fmt.Sprintf("Deployment failed: %v", err),
				})
				// Send error back for return value
				errChan <- err
				return
			}
		}
	}()

	// Wait for TUI to finish
	if _, err := program.Run(); err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	// Return the deployment result
	select {
	case res := <-resultChan:
		return res, nil
	case err := <-errChan:
		return nil, err
	default:
		// TUI exited without completion signal - possible user quit
		return nil, fmt.Errorf("deployment interrupted")
	}
}
