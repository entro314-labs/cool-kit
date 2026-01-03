package azure

import (
	"github.com/entro314-labs/cool-kit/internal/azureconfig"
)

// DeploymentContext holds all information needed for Azure deployment
type DeploymentContext struct {
	Config *azureconfig.Config

	// Resource identifiers
	ResourceGroup string
	VMName        string
	Location      string
	VMSize        string

	// Network information
	VNetName     string
	SubnetName   string
	NSGName      string
	PublicIPName string
	NICName      string

	// Generated values
	PublicIP      string
	AdminUsername string
	SSHKeyPath    string

	// Coolify settings
	AdminEmail    string
	AdminPassword string
}

// BackupInfo contains backup metadata
type BackupInfo struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Version   string `json:"version"`
	Path      string `json:"path,omitempty"`
	Size      int64  `json:"size"`
}

// VMStatus represents the status of an Azure VM
type VMStatus struct {
	PowerState        string
	ProvisioningState string
	VMSize            string
	Location          string
	PublicIP          string
	PrivateIP         string
}

// ServiceStatus represents the status of Coolify services
type ServiceStatus struct {
	Coolify  ContainerStatus
	Database ContainerStatus
	Redis    ContainerStatus
	Realtime ContainerStatus
}

// ContainerStatus represents the status of a Docker container
type ContainerStatus struct {
	Running bool
	Health  string
	Uptime  string
}
