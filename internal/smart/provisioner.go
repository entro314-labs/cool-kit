package smart

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/entro314-labs/cool-kit/internal/api"
)

// ProvisioningResult represents the result of service provisioning
type ProvisioningResult struct {
	Services        []ProvisionedService
	EnvironmentVars map[string]string
	Errors          []error
}

// ProvisionedService represents a service that was created in Coolify
type ProvisionedService struct {
	Type          string
	UUID          string
	Name          string
	ConnectionURL string
	EnvVarName    string
}

// ServiceProvisioner handles automatic service creation in Coolify
type ServiceProvisioner struct {
	client          *api.Client
	projectUUID     string
	environmentUUID string
	serverUUID      string
	appName         string
}

// NewServiceProvisioner creates a new service provisioner
func NewServiceProvisioner(client *api.Client, projectUUID, environmentUUID, serverUUID, appName string) *ServiceProvisioner {
	return &ServiceProvisioner{
		client:          client,
		projectUUID:     projectUUID,
		environmentUUID: environmentUUID,
		serverUUID:      serverUUID,
		appName:         appName,
	}
}

// Provision creates all required services for a deployment
func (sp *ServiceProvisioner) Provision(config *DeploymentConfig) (*ProvisioningResult, error) {
	result := &ProvisioningResult{
		Services:        []ProvisionedService{},
		EnvironmentVars: make(map[string]string),
		Errors:          []error{},
	}

	// Provision each required service
	for _, service := range config.Services {
		provisioned, err := sp.provisionService(service)
		if err != nil {
			if service.Required {
				return nil, fmt.Errorf("failed to provision required service %s: %w", service.Type, err)
			}
			result.Errors = append(result.Errors, fmt.Errorf("failed to provision optional service %s: %w", service.Type, err))
			continue
		}

		result.Services = append(result.Services, *provisioned)
		result.EnvironmentVars[provisioned.EnvVarName] = provisioned.ConnectionURL
	}

	// Add user-defined environment variables from detection
	for _, envVar := range config.Environment {
		// Skip auto-generated vars - we've already set them from services
		if _, exists := result.EnvironmentVars[envVar.Key]; exists {
			continue
		}
		// Add user-defined vars with their default values
		result.EnvironmentVars[envVar.Key] = envVar.Value
	}

	return result, nil
}

// provisionService creates a single service in Coolify
func (sp *ServiceProvisioner) provisionService(service RequiredService) (*ProvisionedService, error) {
	switch service.Type {
	case "postgresql":
		return sp.provisionPostgreSQL(service)
	case "mysql":
		return sp.provisionMySQL(service)
	case "mongodb":
		return sp.provisionMongoDB(service)
	case "redis":
		return sp.provisionRedis(service)
	case "meilisearch":
		return sp.provisionMeilisearch(service)
	case "elasticsearch":
		return sp.provisionElasticsearch(service)
	default:
		return nil, fmt.Errorf("unsupported service type: %s", service.Type)
	}
}

// provisionPostgreSQL creates a PostgreSQL database
func (sp *ServiceProvisioner) provisionPostgreSQL(service RequiredService) (*ProvisionedService, error) {
	name := sp.generateServiceName("postgres")

	password, err := sp.generatePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	// Create PostgreSQL database via Coolify API
	dbConfig := map[string]interface{}{
		"name":              name,
		"description":       fmt.Sprintf("PostgreSQL database for %s (%s)", sp.appName, service.Reason),
		"project_uuid":      sp.projectUUID,
		"environment_name":  sp.environmentUUID,
		"server_uuid":       sp.serverUUID,
		"type":              "postgresql",
		"image":             fmt.Sprintf("postgres:%s", service.Version),
		"postgres_db":       sp.sanitizeDBName(sp.appName),
		"postgres_user":     sp.sanitizeDBName(sp.appName),
		"postgres_password": password,
		"is_public":         false,
	}

	response, err := sp.client.CreateDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL database: %w", err)
	}

	uuid, ok := response["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response from Coolify API: missing uuid")
	}

	// Generate connection URL
	connectionURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:5432/%s",
		dbConfig["postgres_user"],
		dbConfig["postgres_password"],
		name,
		dbConfig["postgres_db"],
	)

	return &ProvisionedService{
		Type:          "postgresql",
		UUID:          uuid,
		Name:          name,
		ConnectionURL: connectionURL,
		EnvVarName:    service.EnvVarName,
	}, nil
}

// provisionMySQL creates a MySQL database
func (sp *ServiceProvisioner) provisionMySQL(service RequiredService) (*ProvisionedService, error) {
	name := sp.generateServiceName("mysql")

	password, err := sp.generatePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}
	rootPassword, err := sp.generatePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate root password: %w", err)
	}

	dbConfig := map[string]interface{}{
		"name":                name,
		"description":         fmt.Sprintf("MySQL database for %s (%s)", sp.appName, service.Reason),
		"project_uuid":        sp.projectUUID,
		"environment_name":    sp.environmentUUID,
		"server_uuid":         sp.serverUUID,
		"type":                "mysql",
		"image":               fmt.Sprintf("mysql:%s", service.Version),
		"mysql_database":      sp.sanitizeDBName(sp.appName),
		"mysql_user":          sp.sanitizeDBName(sp.appName),
		"mysql_password":      password,
		"mysql_root_password": rootPassword,
		"is_public":           false,
	}

	response, err := sp.client.CreateDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create MySQL database: %w", err)
	}

	uuid, ok := response["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response from Coolify API: missing uuid")
	}

	connectionURL := fmt.Sprintf(
		"mysql://%s:%s@%s:3306/%s",
		dbConfig["mysql_user"],
		dbConfig["mysql_password"],
		name,
		dbConfig["mysql_database"],
	)

	return &ProvisionedService{
		Type:          "mysql",
		UUID:          uuid,
		Name:          name,
		ConnectionURL: connectionURL,
		EnvVarName:    service.EnvVarName,
	}, nil
}

// provisionMongoDB creates a MongoDB database
func (sp *ServiceProvisioner) provisionMongoDB(service RequiredService) (*ProvisionedService, error) {
	name := sp.generateServiceName("mongodb")

	password, err := sp.generatePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	dbConfig := map[string]interface{}{
		"name":                       name,
		"description":                fmt.Sprintf("MongoDB database for %s (%s)", sp.appName, service.Reason),
		"project_uuid":               sp.projectUUID,
		"environment_name":           sp.environmentUUID,
		"server_uuid":                sp.serverUUID,
		"type":                       "mongodb",
		"image":                      fmt.Sprintf("mongo:%s", service.Version),
		"mongo_initdb_root_username": sp.sanitizeDBName(sp.appName),
		"mongo_initdb_root_password": password,
		"is_public":                  false,
	}

	response, err := sp.client.CreateDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB database: %w", err)
	}

	uuid, ok := response["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response from Coolify API: missing uuid")
	}

	connectionURL := fmt.Sprintf(
		"mongodb://%s:%s@%s:27017",
		dbConfig["mongo_initdb_root_username"],
		dbConfig["mongo_initdb_root_password"],
		name,
	)

	return &ProvisionedService{
		Type:          "mongodb",
		UUID:          uuid,
		Name:          name,
		ConnectionURL: connectionURL,
		EnvVarName:    service.EnvVarName,
	}, nil
}

// provisionRedis creates a Redis instance
func (sp *ServiceProvisioner) provisionRedis(service RequiredService) (*ProvisionedService, error) {
	name := sp.generateServiceName("redis")

	password, err := sp.generatePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	dbConfig := map[string]interface{}{
		"name":             name,
		"description":      fmt.Sprintf("Redis for %s (%s)", sp.appName, service.Reason),
		"project_uuid":     sp.projectUUID,
		"environment_name": sp.environmentUUID,
		"server_uuid":      sp.serverUUID,
		"type":             "redis",
		"image":            fmt.Sprintf("redis:%s", service.Version),
		"redis_password":   password,
		"is_public":        false,
	}

	response, err := sp.client.CreateDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis instance: %w", err)
	}

	uuid, ok := response["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response from Coolify API: missing uuid")
	}

	connectionURL := fmt.Sprintf(
		"redis://:%s@%s:6379",
		dbConfig["redis_password"],
		name,
	)

	return &ProvisionedService{
		Type:          "redis",
		UUID:          uuid,
		Name:          name,
		ConnectionURL: connectionURL,
		EnvVarName:    service.EnvVarName,
	}, nil
}

// provisionMeilisearch creates a Meilisearch instance
func (sp *ServiceProvisioner) provisionMeilisearch(service RequiredService) (*ProvisionedService, error) {
	name := sp.generateServiceName("meilisearch")
	masterKey, err := sp.generatePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}

	// Meilisearch is typically deployed as a service, not a database
	serviceConfig := map[string]interface{}{
		"name":             name,
		"description":      fmt.Sprintf("Meilisearch for %s (%s)", sp.appName, service.Reason),
		"project_uuid":     sp.projectUUID,
		"environment_name": sp.environmentUUID,
		"server_uuid":      sp.serverUUID,
		"type":             "meilisearch",
		"image":            "getmeili/meilisearch:latest",
		"environment_variables": map[string]string{
			"MEILI_MASTER_KEY": masterKey,
			"MEILI_ENV":        "production",
		},
		"is_public": false,
	}

	response, err := sp.client.CreateService(serviceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Meilisearch service: %w", err)
	}

	uuid, ok := response["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response from Coolify API: missing uuid")
	}

	connectionURL := fmt.Sprintf("http://%s:7700", name)

	return &ProvisionedService{
		Type:          "meilisearch",
		UUID:          uuid,
		Name:          name,
		ConnectionURL: connectionURL,
		EnvVarName:    service.EnvVarName,
	}, nil
}

// provisionElasticsearch creates an Elasticsearch instance
func (sp *ServiceProvisioner) provisionElasticsearch(service RequiredService) (*ProvisionedService, error) {
	name := sp.generateServiceName("elasticsearch")
	password, err := sp.generatePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	serviceConfig := map[string]interface{}{
		"name":             name,
		"description":      fmt.Sprintf("Elasticsearch for %s (%s)", sp.appName, service.Reason),
		"project_uuid":     sp.projectUUID,
		"environment_name": sp.environmentUUID,
		"server_uuid":      sp.serverUUID,
		"type":             "elasticsearch",
		"image":            fmt.Sprintf("elasticsearch:%s", service.Version),
		"environment_variables": map[string]string{
			"discovery.type":         "single-node",
			"ELASTIC_PASSWORD":       password,
			"xpack.security.enabled": "true",
		},
		"is_public": false,
	}

	response, err := sp.client.CreateService(serviceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch service: %w", err)
	}

	uuid, ok := response["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response from Coolify API: missing uuid")
	}

	connectionURL := fmt.Sprintf("https://elastic:%s@%s:9200", password, name)

	return &ProvisionedService{
		Type:          "elasticsearch",
		UUID:          uuid,
		Name:          name,
		ConnectionURL: connectionURL,
		EnvVarName:    service.EnvVarName,
	}, nil
}

// Helper functions

// generateServiceName creates a unique service name
func (sp *ServiceProvisioner) generateServiceName(serviceType string) string {
	// Sanitize app name for use in service name
	sanitized := strings.ToLower(sp.appName)
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ReplaceAll(sanitized, "_", "-")

	return fmt.Sprintf("%s-%s", sanitized, serviceType)
}

// sanitizeDBName creates a valid database/username
func (sp *ServiceProvisioner) sanitizeDBName(name string) string {
	// Convert to lowercase and replace invalid characters
	sanitized := strings.ToLower(name)
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	sanitized = strings.ReplaceAll(sanitized, " ", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")

	// Remove any non-alphanumeric characters except underscore
	result := ""
	for _, char := range sanitized {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			result += string(char)
		}
	}

	// Ensure it starts with a letter
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = "db_" + result
	}

	return result
}

// generatePassword creates a secure random password using crypto/rand
func (sp *ServiceProvisioner) generatePassword() (string, error) {
	// Generate 32 random bytes (256 bits of entropy)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("crypto/rand failed: %w", err)
	}

	// Encode to base64 for a password-safe string
	// Use URL encoding to avoid special characters that might cause issues
	password := base64.URLEncoding.EncodeToString(bytes)

	// Trim padding and limit to 43 characters (standard base64 length for 32 bytes)
	return strings.TrimRight(password, "="), nil
}
