package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Provider interface that deployment providers must implement
type Provider interface {
	// GetDeploymentSteps returns the steps this provider will execute
	GetDeploymentSteps() []DeploymentStep
	// Deploy executes the deployment, sending progress through channels
	Deploy(progressChan chan<- StepProgressMsg, logChan chan<- LogMsg) error
}

// DeploymentRunner wraps a provider with the BubbleTea TUI
type DeploymentRunner struct {
	provider     Provider
	providerName string
}

// NewDeploymentRunner creates a new deployment runner
func NewDeploymentRunner(providerName string, provider Provider) *DeploymentRunner {
	return &DeploymentRunner{
		provider:     provider,
		providerName: providerName,
	}
}

// RunWithTUI executes the deployment with the interactive TUI
func (r *DeploymentRunner) RunWithTUI() error {
	steps := r.provider.GetDeploymentSteps()
	model := NewProgressModel(r.providerName, steps)

	// Create channels for communicating with the provider
	progressChan := make(chan StepProgressMsg, 100)
	logChan := make(chan LogMsg, 100)
	errChan := make(chan error, 1)

	// Create a tea program with the model
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Start the deployment in a goroutine
	go func() {
		err := r.provider.Deploy(progressChan, logChan)
		errChan <- err
		close(progressChan)
		close(logChan)
	}()

	// Forward progress updates to the TUI
	go func() {
		for {
			select {
			case progress, ok := <-progressChan:
				if !ok {
					progressChan = nil
					continue
				}
				p.Send(progress)

			case log, ok := <-logChan:
				if !ok {
					logChan = nil
					continue
				}
				p.Send(log)

			case err := <-errChan:
				if err != nil {
					p.Send(DeploymentCompleteMsg{
						Success: false,
						Message: err.Error(),
					})
				} else {
					p.Send(DeploymentCompleteMsg{
						Success: true,
						Message: fmt.Sprintf("Successfully deployed to %s!", r.providerName),
					})
				}
				return
			}

			if progressChan == nil && logChan == nil {
				return
			}
		}
	}()

	// Run the TUI
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check if deployment had errors
	if m, ok := finalModel.(ProgressModel); ok && m.err != nil {
		return m.err
	}

	return nil
}

// RunSimple executes the deployment without TUI (for non-interactive mode)
func (r *DeploymentRunner) RunSimple() error {
	steps := r.provider.GetDeploymentSteps()

	Section(fmt.Sprintf("Deploying to %s", r.providerName))
	Spacer()
	Bold("Deployment Steps:")
	for i, step := range steps {
		Dim(fmt.Sprintf("  %d. %s - %s", i+1, step.Name, step.Description))
	}
	Spacer()

	progressChan := make(chan StepProgressMsg, 100)
	logChan := make(chan LogMsg, 100)

	// Run deployment in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- r.provider.Deploy(progressChan, logChan)
		close(progressChan)
		close(logChan)
	}()

	// Process progress updates
	for {
		select {
		case progress, ok := <-progressChan:
			if !ok {
				progressChan = nil
			} else {
				Dim(fmt.Sprintf("[Step %d] %.0f%% - %s", progress.StepIndex+1, progress.Progress*100, progress.Message))
			}
		case log, ok := <-logChan:
			if !ok {
				logChan = nil
			} else {
				switch log.Level {
				case LogSuccess:
					Success(log.Message)
				case LogError:
					Error(log.Message)
				case LogWarning:
					Warning(log.Message)
				default:
					Dim(log.Message)
				}
			}
		case err := <-errChan:
			if err != nil {
				Error(fmt.Sprintf("Deployment failed: %v", err))
				return err
			}
			Spacer()
			Success(fmt.Sprintf("%s deployment completed successfully!", r.providerName))
			return nil
		}

		if progressChan == nil && logChan == nil {
			break
		}
	}

	return nil
}
