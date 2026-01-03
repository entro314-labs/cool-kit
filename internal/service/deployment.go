package service

import (
	"context"
	"fmt"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/providers/aws"
	"github.com/entro314-labs/cool-kit/internal/providers/azure"
	"github.com/entro314-labs/cool-kit/internal/providers/baremetal"
	"github.com/entro314-labs/cool-kit/internal/providers/digitalocean"
	"github.com/entro314-labs/cool-kit/internal/providers/docker"
	"github.com/entro314-labs/cool-kit/internal/providers/gcp"
	"github.com/entro314-labs/cool-kit/internal/providers/hetzner"
	"github.com/entro314-labs/cool-kit/internal/providers/production"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// DeploymentResult contains deployment results
type DeploymentResult struct {
	Provider     string
	DashboardURL string
	Success      bool
	Duration     time.Duration
	Resources    map[string]string
}

// DeploymentService handles the deployment logic for different providers
type DeploymentService interface {
	Deploy(ctx context.Context, provider string, progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) (*DeploymentResult, error)
	GetDeploymentSteps(provider string) []ui.DeploymentStep
}

type deploymentService struct {
	config *config.Config
}

// NewDeploymentService creates a new deployment service
func NewDeploymentService(cfg *config.Config) DeploymentService {
	return &deploymentService{
		config: cfg,
	}
}

// Deploy performs the deployment
func (s *deploymentService) Deploy(ctx context.Context, provider string, progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) (*DeploymentResult, error) {
	startTime := time.Now()
	var err error
	var dashboardURL string

	switch provider {
	case "azure":
		dashboardURL, err = s.deployAzure(progressChan, logChan)
	case "aws":
		dashboardURL, err = s.deployAWS(progressChan, logChan)
	case "gcp":
		dashboardURL, err = s.deployGCP(progressChan, logChan)
	case "baremetal":
		dashboardURL, err = s.deployBareMetal(progressChan, logChan)
	case "hetzner":
		dashboardURL, err = s.deployHetzner(progressChan, logChan)
	case "digitalocean":
		dashboardURL, err = s.deployDigitalOcean(progressChan, logChan)
	case "docker", "local":
		dashboardURL, err = s.deployDocker(progressChan, logChan)
	case "production":
		dashboardURL, err = s.deployProduction(progressChan, logChan)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	if err != nil {
		return nil, err
	}

	return &DeploymentResult{
		Provider:     provider,
		DashboardURL: dashboardURL,
		Success:      true,
		Duration:     time.Since(startTime),
	}, nil
}

// GetDeploymentSteps returns steps for the provider
func (s *deploymentService) GetDeploymentSteps(provider string) []ui.DeploymentStep {
	switch provider {
	case "azure":
		p, _ := azure.NewAzureProvider(s.config)
		if p != nil {
			return p.GetDeploymentSteps()
		}
		return []ui.DeploymentStep{{Name: "Configure Azure credentials first"}}
	case "aws":
		provider := aws.NewAWSProvider(s.config)
		return provider.GetDeploymentSteps()
	case "gcp":
		provider, _ := gcp.NewGCPProvider(s.config)
		return provider.GetDeploymentSteps()
	case "baremetal":
		p, _ := baremetal.NewBareMetalProvider(s.config)
		return p.GetDeploymentSteps()
	case "hetzner":
		p, _ := hetzner.NewHetznerProvider(s.config)
		if p != nil {
			return p.GetDeploymentSteps()
		}
		return []ui.DeploymentStep{{Name: "Configure Hetzner credentials first"}}
	case "digitalocean":
		p, _ := digitalocean.NewDigitalOceanProvider(s.config)
		if p != nil {
			return p.GetDeploymentSteps()
		}
		return []ui.DeploymentStep{{Name: "Configure DigitalOcean credentials first"}}
	case "docker", "local":
		p, _ := docker.NewDockerProvider(s.config, "development")
		return p.GetDeploymentSteps()
	case "production":
		p, _ := production.NewProductionProvider(s.config)
		if p != nil {
			return p.GetDeploymentSteps()
		}
		return []ui.DeploymentStep{
			{Name: "Configure production settings first"},
		}
	default:
		return []ui.DeploymentStep{}
	}
}

// deployAzure deploys to Azure
func (s *deploymentService) deployAzure(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) (string, error) {
	provider, err := azure.NewAzureProvider(s.config)
	if err != nil {
		return "", err
	}

	if err := provider.Deploy(progressChan, logChan); err != nil {
		return "", err
	}

	// Get public IP from config
	publicIP, ok := s.config.Settings["public_ip"].(string)
	if !ok {
		publicIP = "your-azure-vm-ip"
	}

	return fmt.Sprintf("http://%s:8000", publicIP), nil
}

// deployAWS deploys to AWS
func (s *deploymentService) deployAWS(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) (string, error) {
	provider := aws.NewAWSProvider(s.config)

	if err := provider.Deploy(progressChan, logChan); err != nil {
		return "", err
	}

	publicIP, ok := s.config.Settings["public_ip"].(string)
	if !ok {
		publicIP = "your-aws-instance-ip"
	}

	return fmt.Sprintf("http://%s:8000", publicIP), nil
}

// deployGCP deploys to GCP
func (s *deploymentService) deployGCP(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) (string, error) {
	provider, err := gcp.NewGCPProvider(s.config)
	if err != nil {
		return "", err
	}

	if err := provider.Deploy(progressChan, logChan); err != nil {
		return "", err
	}

	publicIP, ok := s.config.Settings["public_ip"].(string)
	if !ok {
		publicIP = "your-gcp-instance-ip"
	}

	return fmt.Sprintf("http://%s:8000", publicIP), nil
}

// deployBareMetal deploys to bare metal/VM
func (s *deploymentService) deployBareMetal(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) (string, error) {
	provider, err := baremetal.NewBareMetalProvider(s.config)
	if err != nil {
		return "", err
	}

	if err := provider.Deploy(progressChan, logChan); err != nil {
		return "", err
	}

	host, ok := s.config.Settings["baremetal_host"].(string)
	if !ok {
		host = "your-server-ip"
	}

	return fmt.Sprintf("http://%s:8000", host), nil
}

// deployDocker deploys with Docker
func (s *deploymentService) deployDocker(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) (string, error) {
	provider, err := docker.NewDockerProvider(s.config, "development")
	if err != nil {
		return "", err
	}

	if err := provider.Deploy(progressChan, logChan); err != nil {
		return "", err
	}

	port := s.config.Local.AppPort
	if port == 0 {
		port = 8000
	}

	return fmt.Sprintf("http://localhost:%d", port), nil
}

// deployProduction deploys to production via SSH
func (s *deploymentService) deployProduction(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) (string, error) {
	provider, err := production.NewProductionProvider(s.config)
	if err != nil {
		return "", fmt.Errorf("production configuration error: %w", err)
	}

	if err := provider.Deploy(progressChan, logChan); err != nil {
		return "", err
	}

	domain := s.config.Production.Domain
	return fmt.Sprintf("https://%s", domain), nil
}

// deployHetzner deploys to Hetzner Cloud
func (s *deploymentService) deployHetzner(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) (string, error) {
	provider, err := hetzner.NewHetznerProvider(s.config)
	if err != nil {
		return "", err
	}

	if err := provider.Deploy(progressChan, logChan); err != nil {
		return "", err
	}

	publicIP, ok := s.config.Settings["hetzner_server_ip"].(string)
	if !ok {
		publicIP = "your-hetzner-server-ip"
	}

	return fmt.Sprintf("http://%s:8000", publicIP), nil
}

// deployDigitalOcean deploys to DigitalOcean
func (s *deploymentService) deployDigitalOcean(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) (string, error) {
	provider, err := digitalocean.NewDigitalOceanProvider(s.config)
	if err != nil {
		return "", err
	}

	if err := provider.Deploy(progressChan, logChan); err != nil {
		return "", err
	}

	publicIP, ok := s.config.Settings["do_droplet_ip"].(string)
	if !ok {
		publicIP = "your-digitalocean-droplet-ip"
	}

	return fmt.Sprintf("http://%s:8000", publicIP), nil
}
