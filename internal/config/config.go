// Package config handles configuration management.
package config

import (
	"encoding/json"
	"errors" // Added for errors.New()
	"fmt"
	"os"
	"path/filepath"
	"time" // Added for time.Now()

	"github.com/spf13/viper"
)

// Config holds all CLI configuration
type Config struct {
	// Multi-instance configuration
	Instances           []Instance `json:"instances"`
	CurrentContext      string     `json:"current_context"`
	LastUpdateCheckTime string     `json:"lastUpdateCheckTime"`

	// Legacy/Single-instance configuration (kept for backward compatibility)
	Provider    string                 `json:"provider"`
	Environment string                 `json:"environment"`
	Settings    map[string]interface{} `json:"settings"`
	Git         GitConfig              `json:"git"`
	Azure       AzureConfig            `json:"azure,omitempty"`
	AWS         AWSConfig              `json:"aws,omitempty"`
	GCP         GCPConfig              `json:"gcp,omitempty"`
	BareMetal   BareMetalConfig        `json:"baremetal,omitempty"`
	Local       LocalConfig            `json:"local,omitempty"`
	Production  ProductionConfig       `json:"production,omitempty"`

	path string // config file path (not serialized)
}

// New creates a new config with default values
func New() *Config {
	return &Config{
		Instances:           []Instance{},
		LastUpdateCheckTime: time.Now().Format(time.RFC3339),
		Settings:            make(map[string]interface{}),
		Environment:         "production",
		path:                Path(),
	}
}

// Instance represents a single Coolify instance
type Instance struct {
	Name    string `json:"name"`
	FQDN    string `json:"fqdn"`
	Token   string `json:"token"`
	Default bool   `json:"default"`
}

// Validate validates the instance configuration
func (i *Instance) Validate() error {
	if i.Name == "" {
		return errors.New("instance name is required")
	}
	if i.FQDN == "" {
		return errors.New("instance FQDN is required")
	}
	if i.Token == "" {
		return errors.New("instance token is required")
	}
	return nil
}

// GitConfig represents Git-related configuration
type GitConfig struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	WorkDir    string `json:"work_dir"`
}

// AzureConfig represents Azure-specific configuration
type AzureConfig struct {
	Location       string `json:"location" mapstructure:"location"`
	ResourceGroup  string `json:"resource_group" mapstructure:"resource_group"`
	VMName         string `json:"vm_name" mapstructure:"vm_name"`
	VMSize         string `json:"vm_size" mapstructure:"vm_size"`
	AdminUsername  string `json:"admin_username" mapstructure:"admin_username"`
	SSHKeyPath     string `json:"ssh_key_path" mapstructure:"ssh_key_path"`
	SubscriptionID string `json:"subscription_id" mapstructure:"subscription_id"`
}

// LocalConfig represents local development configuration
type LocalConfig struct {
	AppPort       int    `json:"app_port"`
	WebSocketPort int    `json:"websocket_port"`
	WorkDir       string `json:"work_dir"`
	Debug         bool   `json:"debug"`
}

// ProductionConfig represents production configuration
type ProductionConfig struct {
	Domain     string `json:"domain"`
	SSLEmail   string `json:"ssl_email"`
	Kubeconfig string `json:"kubeconfig"`
	Namespace  string `json:"namespace"`
}

// AWSConfig represents AWS-specific configuration
type AWSConfig struct {
	Region       string `json:"region"`
	InstanceType string `json:"instance_type"`
	AMI          string `json:"ami"`
	KeyName      string `json:"key_name"`
	VpcID        string `json:"vpc_id"`
	SubnetID     string `json:"subnet_id"`
	SSHKeyPath   string `json:"ssh_key_path"`
}

// GCPConfig represents GCP-specific configuration
type GCPConfig struct {
	Project     string `json:"project"`
	Zone        string `json:"zone"`
	MachineType string `json:"machine_type"`
	Network     string `json:"network"`
	SSHKeyPath  string `json:"ssh_key_path"`
}

// BareMetalConfig represents bare metal/VM SSH deployment configuration
type BareMetalConfig struct {
	Host       string `json:"host"`
	User       string `json:"user"`
	SSHKeyPath string `json:"ssh_key_path"`
	Port       int    `json:"port"`
}

var (
	globalConfig *Config
	configDir    string
)

// Default configuration values for CDP functionality
const (
	DefaultBranch   = "v4.x"
	DefaultPlatform = "linux/amd64" // Docker platform for server compatibility
)

// Initialize initializes the configuration system
func Initialize() error {
	// Set up config directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir = filepath.Join(home, ".cool-kit")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Initialize viper
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	// Set defaults
	setDefaults()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create default config
			return createDefaultConfig()
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Unmarshal config
	if err := viper.Unmarshal(&globalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("git.repository", "https://github.com/coollabsio/coolify.git")
	viper.SetDefault("git.branch", "v4.x")
	viper.SetDefault("git.work_dir", "./coolify-source")

	viper.SetDefault("azure.location", "swedencentral")
	viper.SetDefault("azure.resource_group", "coolify-rg")
	viper.SetDefault("azure.vm_name", "coolify-vm")
	viper.SetDefault("azure.vm_size", "Standard_B2s")
	viper.SetDefault("azure.admin_username", "azureuser")
	viper.SetDefault("azure.ssh_key_path", "~/.ssh/id_rsa.pub")

	viper.SetDefault("local.app_port", 8000)
	viper.SetDefault("local.websocket_port", 6001)
	viper.SetDefault("local.work_dir", "./coolify-local")
	viper.SetDefault("local.debug", true)

	viper.SetDefault("production.domain", "coolify.example.com")
	viper.SetDefault("production.ssl_email", "admin@example.com")
	viper.SetDefault("production.namespace", "coolify")

	// AWS defaults
	viper.SetDefault("aws.region", "us-east-1")
	viper.SetDefault("aws.instance_type", "t3.medium")
	viper.SetDefault("aws.ami", "ami-0c55b159cbfafe1f0")
	viper.SetDefault("aws.ssh_key_path", "~/.ssh/id_rsa.pub")

	// GCP defaults
	viper.SetDefault("gcp.zone", "us-central1-a")
	viper.SetDefault("gcp.machine_type", "e2-medium")
	viper.SetDefault("gcp.network", "default")
	viper.SetDefault("gcp.ssh_key_path", "~/.ssh/id_rsa.pub")

	// Bare metal defaults
	viper.SetDefault("baremetal.user", "root")
	viper.SetDefault("baremetal.port", 22)
	viper.SetDefault("baremetal.ssh_key_path", "~/.ssh/id_rsa")
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig() error {
	defaultConfig := &Config{
		Instances:           []Instance{},
		LastUpdateCheckTime: time.Now().Format(time.RFC3339),
		Provider:            "local",
		Environment:         "development",
		Settings:            make(map[string]interface{}),
		Git: GitConfig{
			Repository: "https://github.com/coollabsio/coolify.git",
			Branch:     "v4.x",
			WorkDir:    "./coolify-source",
		},
		Azure: AzureConfig{
			Location:      "swedencentral",
			ResourceGroup: "coolify-rg",
			VMName:        "coolify-vm",
			VMSize:        "Standard_B2s",
			AdminUsername: "azureuser",
			SSHKeyPath:    "~/.ssh/id_rsa.pub",
		},
		AWS: AWSConfig{
			Region:       "us-east-1",
			InstanceType: "t3.medium",
			AMI:          "ami-0c55b159cbfafe1f0",
			SSHKeyPath:   "~/.ssh/id_rsa.pub",
		},
		GCP: GCPConfig{
			Zone:        "us-central1-a",
			MachineType: "e2-medium",
			Network:     "default",
			SSHKeyPath:  "~/.ssh/id_rsa.pub",
		},
		BareMetal: BareMetalConfig{
			User:       "root",
			Port:       22,
			SSHKeyPath: "~/.ssh/id_rsa",
		},
		Local: LocalConfig{
			AppPort:       8000,
			WebSocketPort: 6001,
			WorkDir:       "./coolify-local",
			Debug:         true,
		},
		Production: ProductionConfig{
			Domain:    "coolify.example.com",
			SSLEmail:  "admin@example.com",
			Namespace: "coolify",
		},
	}

	// Save default config
	if err := Save(defaultConfig); err != nil {
		return fmt.Errorf("failed to save default config: %w", err)
	}

	globalConfig = defaultConfig
	return nil
}

// Get returns the global configuration
func Get() *Config {
	return globalConfig
}

// Save saves the configuration to file
func Save(cfg *Config) error {
	globalConfig = cfg

	// Update viper
	viper.Set("instances", cfg.Instances)
	viper.Set("current_context", cfg.CurrentContext)
	viper.Set("lastUpdateCheckTime", cfg.LastUpdateCheckTime)

	viper.Set("provider", cfg.Provider)
	viper.Set("environment", cfg.Environment)
	viper.Set("settings", cfg.Settings)
	viper.Set("git", cfg.Git)
	viper.Set("azure", cfg.Azure)
	viper.Set("local", cfg.Local)
	viper.Set("production", cfg.Production)
	viper.Set("aws", cfg.AWS)
	viper.Set("gcp", cfg.GCP)
	viper.Set("baremetal", cfg.BareMetal)

	// Write config file
	configFile := filepath.Join(configDir, "config.json")
	return viper.WriteConfigAs(configFile)
}

// UpdateProvider updates the provider and saves config
func UpdateProvider(provider string) error {
	if globalConfig == nil {
		return fmt.Errorf("configuration not initialized")
	}

	globalConfig.Provider = provider
	return Save(globalConfig)
}

// UpdateAzureConfig updates Azure-specific configuration
func UpdateAzureConfig(azureConfig AzureConfig) error {
	if globalConfig == nil {
		return fmt.Errorf("configuration not initialized")
	}

	globalConfig.Azure = azureConfig
	globalConfig.Provider = "azure"
	return Save(globalConfig)
}

// UpdateLocalConfig updates local development configuration
func UpdateLocalConfig(localConfig LocalConfig) error {
	if globalConfig == nil {
		return fmt.Errorf("configuration not initialized")
	}

	globalConfig.Local = localConfig
	globalConfig.Provider = "local"
	return Save(globalConfig)
}

// UpdateProductionConfig updates production configuration
func UpdateProductionConfig(prodConfig ProductionConfig) error {
	if globalConfig == nil {
		return fmt.Errorf("configuration not initialized")
	}

	globalConfig.Production = prodConfig
	globalConfig.Provider = "production"
	return Save(globalConfig)
}

// GetConfigDir returns the configuration directory
func GetConfigDir() string {
	return configDir
}

// Path returns the configuration file path
func Path() string {
	return filepath.Join(configDir, "config.json")
}

// ValidateConfig validates the current configuration
func ValidateConfig() error {
	if globalConfig == nil {
		return fmt.Errorf("configuration not initialized")
	}

	switch globalConfig.Provider {
	case "azure":
		if globalConfig.Azure.Location == "" {
			return fmt.Errorf("azure location is required")
		}
		if globalConfig.Azure.ResourceGroup == "" {
			return fmt.Errorf("azure resource group is required")
		}
	case "local":
		if globalConfig.Local.WorkDir == "" {
			return fmt.Errorf("local work directory is required")
		}
	case "production":
		if globalConfig.Production.Domain == "" {
			return fmt.Errorf("production domain is required")
		}
	case "aws":
		if globalConfig.AWS.Region == "" {
			return fmt.Errorf("AWS region is required")
		}
	case "gcp":
		if globalConfig.GCP.Project == "" {
			return fmt.Errorf("GCP project is required")
		}
	case "baremetal":
		if globalConfig.BareMetal.Host == "" {
			return fmt.Errorf("bare Metal host is required")
		}
	default:
		// Not returning error for unknown provider to allow for extensions
		// return fmt.Errorf("unknown provider: %s", globalConfig.Provider)
	}

	return nil
}

// ExportConfig exports configuration to JSON string
func ExportConfig() (string, error) {
	if globalConfig == nil {
		return "", fmt.Errorf("configuration not initialized")
	}

	data, err := json.MarshalIndent(globalConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	return string(data), nil
}

// ImportConfig imports configuration from JSON string
func ImportConfig(jsonData string) error {
	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return Save(&cfg)
}

// AddInstance adds a new instance to the configuration
func AddInstance(name, fqdn, token string, setAsDefault bool) error {
	cfg := Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Check if instance with this name already exists
	for _, inst := range cfg.Instances {
		if inst.Name == name {
			return fmt.Errorf("instance '%s' already exists", name)
		}
	}

	// If this is the first instance or setAsDefault is true, make it default
	if len(cfg.Instances) == 0 || setAsDefault {
		// Unset all other defaults
		for i := range cfg.Instances {
			cfg.Instances[i].Default = false
		}
		cfg.CurrentContext = name
	}

	// Add new instance
	cfg.Instances = append(cfg.Instances, Instance{
		Name:    name,
		FQDN:    fqdn,
		Token:   token,
		Default: len(cfg.Instances) == 0 || setAsDefault,
	})

	return Save(cfg)
}

// RemoveInstance removes an instance from the configuration
func RemoveInstance(name string) error {
	cfg := Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Find and remove instance
	found := false
	newInstances := []Instance{}
	wasDefault := false

	for _, inst := range cfg.Instances {
		if inst.Name == name {
			found = true
			wasDefault = inst.Default
			continue
		}
		newInstances = append(newInstances, inst)
	}

	if !found {
		return fmt.Errorf("instance '%s' not found", name)
	}

	cfg.Instances = newInstances

	// If we removed the default instance, set a new default
	if wasDefault && len(cfg.Instances) > 0 {
		cfg.Instances[0].Default = true
		cfg.CurrentContext = cfg.Instances[0].Name
	} else if len(cfg.Instances) == 0 {
		cfg.CurrentContext = ""
	}

	return Save(cfg)
}

// UseInstance sets an instance as the current context
func UseInstance(name string) error {
	cfg := Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Find instance
	found := false
	for i := range cfg.Instances {
		if cfg.Instances[i].Name == name {
			found = true
			cfg.Instances[i].Default = true
			cfg.CurrentContext = name
		} else {
			cfg.Instances[i].Default = false
		}
	}

	if !found {
		return fmt.Errorf("instance '%s' not found", name)
	}

	return Save(cfg)
}

// GetInstance returns a specific instance by name
func GetInstance(name string) (*Instance, error) {
	cfg := Get()
	if cfg == nil {
		return nil, fmt.Errorf("configuration not initialized")
	}

	for _, inst := range cfg.Instances {
		if inst.Name == name {
			return &inst, nil
		}
	}

	return nil, fmt.Errorf("instance '%s' not found", name)
}

// HasInstances checks if any instances are configured
func HasInstances() bool {
	cfg := Get()
	if cfg == nil {
		return false
	}

	return len(cfg.Instances) > 0
}

// GetCurrentInstance returns the current instance based on context
func GetCurrentInstance() (*Instance, error) {
	cfg := Get()
	if cfg == nil {
		return nil, fmt.Errorf("configuration not initialized")
	}

	if cfg.CurrentContext == "" && len(cfg.Instances) > 0 {
		// If no context set, use the first instance
		cfg.CurrentContext = cfg.Instances[0].Name
		cfg.Instances[0].Default = true
		if err := Save(cfg); err != nil {
			return nil, fmt.Errorf("failed to save default context: %w", err)
		}
		return &cfg.Instances[0], nil
	}

	// Find current instance
	for _, inst := range cfg.Instances {
		if inst.Name == cfg.CurrentContext {
			return &inst, nil
		}
	}

	if len(cfg.Instances) == 0 {
		return nil, fmt.Errorf("no instances configured: run 'cool-kit instances add' first")
	}

	return nil, fmt.Errorf("current instance '%s' not found", cfg.CurrentContext)
}

// ListInstances returns all configured instances
func ListInstances() ([]Instance, error) {
	cfg := Get()
	if cfg == nil {
		return nil, fmt.Errorf("configuration not initialized")
	}

	return cfg.Instances, nil
}
