// Package config handles configuration management.
package config

import (
	"encoding/json"
	"errors" // Added for errors.New()
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time" // Added for time.Now()

	"github.com/spf13/viper"
)

// Config holds all CLI configuration
type Config struct {
	// Multi-instance configuration
	Instances           []Instance `json:"instances" mapstructure:"instances"`
	CurrentContext      string     `json:"current_context" mapstructure:"current_context"`
	LastUpdateCheckTime string     `json:"lastUpdateCheckTime" mapstructure:"lastUpdateCheckTime"`

	// Deployment configuration
	Provider    string           `json:"provider" mapstructure:"provider"`
	Environment string           `json:"environment" mapstructure:"environment"`
	Settings    map[string]any   `json:"settings" mapstructure:"settings"`
	Git         GitConfig        `json:"git" mapstructure:"git"`
	Azure       AzureConfig      `json:"azure,omitzero" mapstructure:"azure"`
	AWS         AWSConfig        `json:"aws,omitzero" mapstructure:"aws"`
	GCP         GCPConfig        `json:"gcp,omitzero" mapstructure:"gcp"`
	BareMetal   BareMetalConfig  `json:"baremetal,omitzero" mapstructure:"baremetal"`
	Local       LocalConfig      `json:"local,omitzero" mapstructure:"local"`
	Production  ProductionConfig `json:"production,omitzero" mapstructure:"production"`

	path string // config file path (not serialized)
}

// New creates a new config with default values
func New() *Config {
	return &Config{
		Instances:           []Instance{},
		LastUpdateCheckTime: time.Now().Format(time.RFC3339),
		Settings:            make(map[string]any),
		Environment:         "production",
		path:                Path(),
	}
}

// Instance represents a single Coolify instance
type Instance struct {
	Name    string `json:"name" mapstructure:"name"`
	FQDN    string `json:"fqdn" mapstructure:"fqdn"`
	Token   string `json:"token" mapstructure:"token"`
	Default bool   `json:"default,omitzero" mapstructure:"default"`
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
	Repository string `json:"repository" mapstructure:"repository"`
	Branch     string `json:"branch" mapstructure:"branch"`
	WorkDir    string `json:"work_dir" mapstructure:"work_dir"`
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
	AppPort       int    `json:"app_port" mapstructure:"app_port"`
	WebSocketPort int    `json:"websocket_port" mapstructure:"websocket_port"`
	WorkDir       string `json:"work_dir" mapstructure:"work_dir"`
	Debug         bool   `json:"debug,omitzero" mapstructure:"debug"`
}

// ProductionConfig represents production configuration
type ProductionConfig struct {
	Domain     string `json:"domain" mapstructure:"domain"`
	SSLEmail   string `json:"ssl_email" mapstructure:"ssl_email"`
	Kubeconfig string `json:"kubeconfig" mapstructure:"kubeconfig"`
	Namespace  string `json:"namespace" mapstructure:"namespace"`
}

// AWSConfig represents AWS-specific configuration
type AWSConfig struct {
	Region       string `json:"region" mapstructure:"region"`
	InstanceType string `json:"instance_type" mapstructure:"instance_type"`
	AMI          string `json:"ami" mapstructure:"ami"`
	KeyName      string `json:"key_name" mapstructure:"key_name"`
	VpcID        string `json:"vpc_id" mapstructure:"vpc_id"`
	SubnetID     string `json:"subnet_id" mapstructure:"subnet_id"`
	SSHKeyPath   string `json:"ssh_key_path" mapstructure:"ssh_key_path"`
}

// GCPConfig represents GCP-specific configuration
type GCPConfig struct {
	Project     string `json:"project" mapstructure:"project"`
	Zone        string `json:"zone" mapstructure:"zone"`
	MachineType string `json:"machine_type" mapstructure:"machine_type"`
	Network     string `json:"network" mapstructure:"network"`
	SSHKeyPath  string `json:"ssh_key_path" mapstructure:"ssh_key_path"`
}

// BareMetalConfig represents bare metal/VM SSH deployment configuration
type BareMetalConfig struct {
	Host       string `json:"host" mapstructure:"host"`
	User       string `json:"user" mapstructure:"user"`
	SSHKeyPath string `json:"ssh_key_path" mapstructure:"ssh_key_path"`
	Port       int    `json:"port" mapstructure:"port"`
}

var (
	globalConfig *Config
	configDir    string
)

// Default configuration values for cool-kit functionality
const (
	DefaultBranch   = "v4.x"
	DefaultPlatform = "linux/amd64"
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

	// Load tokens from keyring
	for i, inst := range globalConfig.Instances {
		if inst.Token == "" || inst.Token == "keyring" {
			key := fmt.Sprintf("instance-%s-token", inst.Name)
			token, err := getToken(key)
			if err == nil && token != "" {
				globalConfig.Instances[i].Token = token
			}
		}
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
	viper.SetDefault("aws.ami", "ami-04b70fa74e45c3917")
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
		Settings:            make(map[string]any),
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
			AMI:          "ami-04b70fa74e45c3917",
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
	// Update viper
	// Prepare safe instances (tokens in keyring)
	safeInstances := make([]Instance, len(cfg.Instances))
	for i, inst := range cfg.Instances {
		safeInstances[i] = inst
		// Save token to keyring
		if inst.Token != "" {
			key := fmt.Sprintf("instance-%s-token", inst.Name)
			if err := setToken(key, inst.Token); err == nil {
				safeInstances[i].Token = "keyring"
			}
		}
	}
	viper.Set("instances", safeInstances)

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

// Merge applies overrides to the configuration
func (c *Config) Merge(overrides map[string]string) error {
	for k, v := range overrides {
		switch k {
		// Azure overrides
		case "location":
			c.Azure.Location = v
		case "vm_size":
			c.Azure.VMSize = v
		case "disk_size":
			// Assuming disk size might be int in struct but string in override
			// c.Azure.DiskSize = v (AzureConfig doesn't seem to have DiskSize based on file view, checking...)
			// Actually AzureConfig only has: Location, ResourceGroup, VMName, VMSize, AdminUsername, SSHKeyPath, SubscriptionID
			// The Model collected "disk_size" but AzureConfig doesn't have it.
			// skipping for now or adding to Settings if needed.
			c.Settings["disk_size"] = v
		case "ssh_key":
			// Map "ssh_key" to actual SSHKeyPath fields for all providers
			c.Azure.SSHKeyPath = v
			c.AWS.SSHKeyPath = v
			c.GCP.SSHKeyPath = v
			c.BareMetal.SSHKeyPath = v

		// Local overrides
		case "app_port":
			if p, err := strconv.Atoi(v); err == nil {
				c.Local.AppPort = p
			}
		case "ws_port":
			if p, err := strconv.Atoi(v); err == nil {
				c.Local.WebSocketPort = p
			}
		case "data_dir":
			// LocalConfig works with WorkDir
			c.Local.WorkDir = v

		// Production overrides
		case "domain":
			c.Production.Domain = v
		case "ssh_host":
			c.BareMetal.Host = v
		case "ssh_user":
			c.BareMetal.User = v

		// Generic/Other
		default:
			c.Settings[k] = v
		}
	}
	return nil
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("configuration not initialized")
	}

	switch c.Provider {
	case "azure":
		if c.Azure.Location == "" {
			return fmt.Errorf("azure location is required")
		}
		if c.Azure.ResourceGroup == "" {
			return fmt.Errorf("azure resource group is required")
		}
	case "local":
		if c.Local.WorkDir == "" {
			return fmt.Errorf("local work directory is required")
		}
	case "production":
		if c.Production.Domain == "" {
			return fmt.Errorf("production domain is required")
		}
	case "aws":
		if c.AWS.Region == "" {
			return fmt.Errorf("AWS region is required")
		}
	case "gcp":
		if c.GCP.Project == "" {
			return fmt.Errorf("GCP project is required")
		}
	case "baremetal":
		if c.BareMetal.Host == "" {
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

// UpsertInstance creates or updates an instance in the configuration.
// If setAsDefault is true, the upserted instance becomes the current default context.
func UpsertInstance(name, fqdn, token string, setAsDefault bool) error {
	cfg := Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	updated := false
	for i := range cfg.Instances {
		if cfg.Instances[i].Name != name {
			continue
		}

		cfg.Instances[i].FQDN = fqdn
		cfg.Instances[i].Token = token
		updated = true
		break
	}

	if !updated {
		cfg.Instances = append(cfg.Instances, Instance{
			Name:  name,
			FQDN:  fqdn,
			Token: token,
		})
	}

	// Ensure at least one default exists.
	if len(cfg.Instances) == 1 || setAsDefault {
		for i := range cfg.Instances {
			cfg.Instances[i].Default = cfg.Instances[i].Name == name
		}
		cfg.CurrentContext = name
	} else {
		// Preserve existing default, but make the upserted instance default if none exists.
		hasDefault := false
		for _, inst := range cfg.Instances {
			if inst.Default {
				hasDefault = true
				break
			}
		}
		if !hasDefault {
			for i := range cfg.Instances {
				cfg.Instances[i].Default = cfg.Instances[i].Name == name
			}
			cfg.CurrentContext = name
		}
	}

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

// ClearInstances removes all configured instances and their stored keyring tokens.
func ClearInstances() error {
	cfg := Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	for _, inst := range cfg.Instances {
		_ = deleteToken(fmt.Sprintf("instance-%s-token", inst.Name))
	}

	cfg.Instances = []Instance{}
	cfg.CurrentContext = ""

	return Save(cfg)
}
