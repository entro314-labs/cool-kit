package azureconfig

import (
	"fmt"
	"os"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	var errors []error

	// Validate infrastructure
	if c.Infrastructure.Location == "" {
		errors = append(errors, &ValidationError{
			Field:   "infrastructure.location",
			Message: "Azure location is required",
		})
	}

	if c.Infrastructure.VMSize == "" {
		errors = append(errors, &ValidationError{
			Field:   "infrastructure.vm_size",
			Message: "VM size is required",
		})
	}

	if c.Infrastructure.AdminUsername == "" {
		errors = append(errors, &ValidationError{
			Field:   "infrastructure.admin_username",
			Message: "Admin username is required",
		})
	}

	if c.Infrastructure.SSHPublicKeyPath == "" {
		errors = append(errors, &ValidationError{
			Field:   "infrastructure.ssh_public_key_path",
			Message: "SSH public key path is required",
		})
	}

	// Validate SSH key file exists
	sshKeyPath := c.Infrastructure.SSHPublicKeyPath
	if strings.HasPrefix(sshKeyPath, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			sshKeyPath = strings.Replace(sshKeyPath, "~", home, 1)
		}
	}
	if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
		errors = append(errors, &ValidationError{
			Field:   "infrastructure.ssh_public_key_path",
			Message: fmt.Sprintf("SSH public key file does not exist: %s", sshKeyPath),
		})
	}

	// Validate networking
	if c.Networking.AppPort <= 0 || c.Networking.AppPort > 65535 {
		errors = append(errors, &ValidationError{
			Field:   "networking.app_port",
			Message: "App port must be between 1 and 65535",
		})
	}

	if c.Networking.SSHPort <= 0 || c.Networking.SSHPort > 65535 {
		errors = append(errors, &ValidationError{
			Field:   "networking.ssh_port",
			Message: "SSH port must be between 1 and 65535",
		})
	}

	if c.Networking.WebSocketPort <= 0 || c.Networking.WebSocketPort > 65535 {
		errors = append(errors, &ValidationError{
			Field:   "networking.websocket_port",
			Message: "WebSocket port must be between 1 and 65535",
		})
	}

	// Validate Coolify settings
	if c.Coolify.DefaultAdminEmail == "" {
		errors = append(errors, &ValidationError{
			Field:   "coolify.default_admin_email",
			Message: "Default admin email is required",
		})
	}

	if c.Coolify.DefaultAdminPassword == "" {
		errors = append(errors, &ValidationError{
			Field:   "coolify.default_admin_password",
			Message: "Default admin password is required",
		})
	}

	if c.Coolify.AppURLTemplate == "" {
		errors = append(errors, &ValidationError{
			Field:   "coolify.app_url_template",
			Message: "App URL template is required",
		})
	}

	// Validate paths
	if c.Paths.RemoteBase == "" {
		errors = append(errors, &ValidationError{
			Field:   "paths.remote_base",
			Message: "Remote base path is required",
		})
	}

	// Validate Docker settings
	if c.Docker.RegistryURL == "" {
		errors = append(errors, &ValidationError{
			Field:   "docker.registry_url",
			Message: "Docker registry URL is required",
		})
	}

	if c.Docker.AppImage == "" {
		errors = append(errors, &ValidationError{
			Field:   "docker.app_image",
			Message: "Docker app image is required",
		})
	}

	// Return combined errors
	if len(errors) > 0 {
		var messages []string
		for _, err := range errors {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("configuration validation failed:\n  %s", strings.Join(messages, "\n  "))
	}

	return nil
}

// ValidateRequired checks if required fields are present
func ValidateRequired(cfg *Config, requiredFields []string) error {
	var errors []error

	for _, field := range requiredFields {
		if err := validateField(cfg, field); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		var messages []string
		for _, err := range errors {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("required fields validation failed:\n  %s", strings.Join(messages, "\n  "))
	}

	return nil
}

func validateField(cfg *Config, field string) error {
	parts := strings.Split(field, ".")
	if len(parts) < 2 {
		return &ValidationError{
			Field:   field,
			Message: "Invalid field path",
		}
	}

	section := parts[0]
	fieldName := parts[1]

	switch section {
	case "infrastructure":
		return validateInfrastructureField(&cfg.Infrastructure, fieldName)
	case "networking":
		return validateNetworkingField(&cfg.Networking, fieldName)
	case "coolify":
		return validateCoolifyField(&cfg.Coolify, fieldName)
	case "paths":
		return validatePathsField(&cfg.Paths, fieldName)
	case "docker":
		return validateDockerField(&cfg.Docker, fieldName)
	default:
		return &ValidationError{
			Field:   field,
			Message: "Unknown configuration section",
		}
	}
}

func validateInfrastructureField(infra *InfrastructureConfig, field string) error {
	switch field {
	case "location":
		if infra.Location == "" {
			return &ValidationError{Field: "infrastructure.location", Message: "required"}
		}
	case "vm_size":
		if infra.VMSize == "" {
			return &ValidationError{Field: "infrastructure.vm_size", Message: "required"}
		}
	case "admin_username":
		if infra.AdminUsername == "" {
			return &ValidationError{Field: "infrastructure.admin_username", Message: "required"}
		}
	case "ssh_public_key_path":
		if infra.SSHPublicKeyPath == "" {
			return &ValidationError{Field: "infrastructure.ssh_public_key_path", Message: "required"}
		}
	}
	return nil
}

func validateNetworkingField(net *NetworkingConfig, field string) error {
	switch field {
	case "app_port":
		if net.AppPort <= 0 {
			return &ValidationError{Field: "networking.app_port", Message: "required"}
		}
	case "ssh_port":
		if net.SSHPort <= 0 {
			return &ValidationError{Field: "networking.ssh_port", Message: "required"}
		}
	case "websocket_port":
		if net.WebSocketPort <= 0 {
			return &ValidationError{Field: "networking.websocket_port", Message: "required"}
		}
	}
	return nil
}

func validateCoolifyField(coolify *CoolifyConfig, field string) error {
	switch field {
	case "default_admin_email":
		if coolify.DefaultAdminEmail == "" {
			return &ValidationError{Field: "coolify.default_admin_email", Message: "required"}
		}
	case "default_admin_password":
		if coolify.DefaultAdminPassword == "" {
			return &ValidationError{Field: "coolify.default_admin_password", Message: "required"}
		}
	}
	return nil
}

func validatePathsField(paths *PathsConfig, field string) error {
	switch field {
	case "remote_base":
		if paths.RemoteBase == "" {
			return &ValidationError{Field: "paths.remote_base", Message: "required"}
		}
	}
	return nil
}

func validateDockerField(docker *DockerConfig, field string) error {
	switch field {
	case "registry_url":
		if docker.RegistryURL == "" {
			return &ValidationError{Field: "docker.registry_url", Message: "required"}
		}
	case "app_image":
		if docker.AppImage == "" {
			return &ValidationError{Field: "docker.app_image", Message: "required"}
		}
	}
	return nil
}
