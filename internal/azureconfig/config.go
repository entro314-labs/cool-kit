package azureconfig

// Config represents the complete Azure deployment configuration
type Config struct {
	Infrastructure InfrastructureConfig `json:"infrastructure"`
	Networking     NetworkingConfig     `json:"networking"`
	Coolify        CoolifyConfig        `json:"coolify"`
	Paths          PathsConfig          `json:"paths"`
	Docker         DockerConfig         `json:"docker"`
}

// InfrastructureConfig contains Azure infrastructure settings
type InfrastructureConfig struct {
	Location         string `json:"location"`
	VMSize           string `json:"vm_size"`
	AdminUsername    string `json:"admin_username"`
	SSHPublicKeyPath string `json:"ssh_public_key_path"`
	ImageOffer       string `json:"image_offer"`
	ImagePublisher   string `json:"image_publisher"`
	ImageSKU         string `json:"image_sku"`
	ImageVersion     string `json:"image_version"`
	OSImage          string `json:"os_image"`
	OSDiskSizeGB     int    `json:"os_disk_size_gb"`
}

// NetworkingConfig contains networking settings
type NetworkingConfig struct {
	AppPort            int    `json:"app_port"`
	SSHPort            int    `json:"ssh_port"`
	WebSocketPort      int    `json:"websocket_port"`
	VNetAddressPrefix  string `json:"vnet_address_prefix"`
	SubnetAddressPrefix string `json:"subnet_address_prefix"`
}

// CoolifyConfig contains Coolify-specific settings
type CoolifyConfig struct {
	DefaultAdminEmail    string `json:"default_admin_email"`
	DefaultAdminPassword string `json:"default_admin_password"`
	AppURLTemplate       string `json:"app_url_template"`
	PusherHostTemplate   string `json:"pusher_host_template"`
	PusherPort           int    `json:"pusher_port"`
	AppID                string `json:"app_id"`
	AppKey               string `json:"app_key"`
	DBPassword           string `json:"db_password"`
	RedisPassword        string `json:"redis_password"`
	PusherAppID          string `json:"pusher_app_id"`
	PusherAppKey         string `json:"pusher_app_key"`
	PusherAppSecret      string `json:"pusher_app_secret"`
}

// PathsConfig contains remote path configurations
type PathsConfig struct {
	RemoteBase    string `json:"remote_base"`
	RemoteEnv     string `json:"remote_env"`
	RemoteStatus  string `json:"remote_status"`
	RemoteLogs    string `json:"remote_logs"`
	RemoteBackups string `json:"remote_backups"`
}

// DockerConfig contains Docker image configurations
type DockerConfig struct {
	RegistryURL    string `json:"registry_url"`
	HelperImage    string `json:"helper_image"`
	AppImage       string `json:"app_image"`
	RealtimeImage  string `json:"realtime_image"`
}

// DefaultConfig returns the default Azure configuration
func DefaultConfig() *Config {
	return &Config{
		Infrastructure: InfrastructureConfig{
			Location:         "swedencentral",
			VMSize:           "Standard_B2s",
			AdminUsername:    "azureuser",
			SSHPublicKeyPath: "~/.ssh/id_rsa.pub",
			ImageOffer:       "0001-com-ubuntu-containers",
			ImagePublisher:   "Canonical",
			ImageSKU:         "20_04-lts-gen2",
			ImageVersion:     "latest",
			OSImage:          "Canonical:0001-com-ubuntu-server-jammy:22_04-lts-gen2:latest",
			OSDiskSizeGB:     30,
		},
		Networking: NetworkingConfig{
			AppPort:            80,
			SSHPort:            22,
			WebSocketPort:      6001,
			VNetAddressPrefix:  "10.0.0.0/16",
			SubnetAddressPrefix: "10.0.1.0/24",
		},
		Coolify: CoolifyConfig{
			DefaultAdminEmail:    "admin@coolify.local",
			DefaultAdminPassword: "admin123",
			AppURLTemplate:       "http://{public_ip}",
			PusherHostTemplate:   "{public_ip}",
			PusherPort:           6001,
		},
		Paths: PathsConfig{
			RemoteBase:    "/home/azureuser/coolify",
			RemoteEnv:     "/home/azureuser/coolify/.env",
			RemoteStatus:  "/home/azureuser/coolify/.upgrade-status",
			RemoteLogs:    "/home/azureuser/coolify/logs",
			RemoteBackups: "/home/azureuser/coolify/backups",
		},
		Docker: DockerConfig{
			RegistryURL:   "ghcr.io",
			HelperImage:   "coollabsio/coolify-helper",
			AppImage:      "coollabsio/coolify",
			RealtimeImage: "coollabsio/coolify-realtime",
		},
	}
}
