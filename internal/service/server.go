package service

import (
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/config"
)

// ServerService handles Coolify server instance management
type ServerService interface {
	ListInstances() []config.Instance
	AddInstance(instance config.Instance) error
	RemoveInstance(fqdn string) error
	UseInstance(fqdn string) error
	GetCurrentInstance() (*config.Instance, error)
}

type serverService struct {
	config *config.Config
}

// NewServerService creates a new server service
func NewServerService(cfg *config.Config) ServerService {
	return &serverService{
		config: cfg,
	}
}

// ListInstances lists all configured instances
func (s *serverService) ListInstances() []config.Instance {
	return s.config.Instances
}

// AddInstance adds a new instance
func (s *serverService) AddInstance(instance config.Instance) error {
	// Validate instance
	if err := instance.Validate(); err != nil {
		return err
	}

	// Check for duplicates
	for _, inst := range s.config.Instances {
		if inst.FQDN == instance.FQDN {
			return fmt.Errorf("instance with FQDN %s already exists", instance.FQDN)
		}
	}

	// Add instance
	s.config.Instances = append(s.config.Instances, instance)

	// If this is the first instance, set it as current
	if len(s.config.Instances) == 1 {
		s.config.CurrentContext = instance.FQDN
	}

	return config.Save(s.config)
}

// RemoveInstance removes an instance
func (s *serverService) RemoveInstance(fqdn string) error {
	found := false
	var newInstances []config.Instance

	for _, inst := range s.config.Instances {
		if inst.FQDN == fqdn {
			found = true
			continue
		}
		newInstances = append(newInstances, inst)
	}

	if !found {
		return fmt.Errorf("instance not found: %s", fqdn)
	}

	s.config.Instances = newInstances

	// If we removed the current context, reset it
	if s.config.CurrentContext == fqdn {
		if len(s.config.Instances) > 0 {
			s.config.CurrentContext = s.config.Instances[0].FQDN
		} else {
			s.config.CurrentContext = ""
		}
	}

	return config.Save(s.config)
}

// UseInstance switches the current context
func (s *serverService) UseInstance(fqdn string) error {
	found := false
	for _, inst := range s.config.Instances {
		if inst.FQDN == fqdn {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("instance not found: %s", fqdn)
	}

	s.config.CurrentContext = fqdn
	return config.Save(s.config)
}

// GetCurrentInstance returns the current active instance
func (s *serverService) GetCurrentInstance() (*config.Instance, error) {
	if s.config.CurrentContext == "" {
		return nil, fmt.Errorf("no instance selected")
	}

	for _, inst := range s.config.Instances {
		if inst.FQDN == s.config.CurrentContext {
			return &inst, nil
		}
	}

	return nil, fmt.Errorf("current instance %s not found in configuration", s.config.CurrentContext)
}
