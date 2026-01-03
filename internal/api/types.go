package api

import "encoding/json"

// Server represents a Coolify server (enhanced from CAGC)
type Server struct {
	ID                            int             `json:"id,omitempty"`
	UUID                          string          `json:"uuid,omitempty"`
	Name                          string          `json:"name,omitempty"`
	Description                   string          `json:"description,omitempty"`
	IP                            string          `json:"ip,omitempty"`
	User                          string          `json:"user,omitempty"`
	Port                          int             `json:"port,omitempty"`
	PrivateKeyUUID                string          `json:"private_key_uuid,omitempty"`
	ProxyType                     string          `json:"proxy_type,omitempty"`
	Proxy                         json.RawMessage `json:"proxy,omitempty"`
	HighDiskUsageNotificationSent bool            `json:"high_disk_usage_notification_sent,omitempty"`
	UnreachableNotificationSent   bool            `json:"unreachable_notification_sent,omitempty"`
	UnreachableCount              int             `json:"unreachable_count,omitempty"`
	ValidationLogs                string          `json:"validation_logs,omitempty"`
	LogDrainNotificationSent      bool            `json:"log_drain_notification_sent,omitempty"`
	SwarmCluster                  string          `json:"swarm_cluster,omitempty"`
	Settings                      *ServerSetting  `json:"settings,omitempty"`
	IsBuildServer                 bool            `json:"is_build_server,omitempty"`
	InstantValidate               bool            `json:"instant_validate,omitempty"`
}

// ServerSetting represents settings for a Coolify server (enhanced from CAGC)
type ServerSetting struct {
	ID                                int    `json:"id,omitempty"`
	ConcurrentBuilds                  int    `json:"concurrent_builds,omitempty"`
	DynamicTimeout                    int    `json:"dynamic_timeout,omitempty"`
	ForceDisabled                     bool   `json:"force_disabled,omitempty"`
	ForceServerCleanup                bool   `json:"force_server_cleanup,omitempty"`
	IsBuildServer                     bool   `json:"is_build_server,omitempty"`
	IsCloudflareTunnel                bool   `json:"is_cloudflare_tunnel,omitempty"`
	IsJumpServer                      bool   `json:"is_jump_server,omitempty"`
	IsLogdrainAxiomEnabled            bool   `json:"is_logdrain_axiom_enabled,omitempty"`
	IsLogdrainCustomEnabled           bool   `json:"is_logdrain_custom_enabled,omitempty"`
	IsLogdrainHighlightEnabled        bool   `json:"is_logdrain_highlight_enabled,omitempty"`
	IsLogdrainNewrelicEnabled         bool   `json:"is_logdrain_newrelic_enabled,omitempty"`
	IsMetricsEnabled                  bool   `json:"is_metrics_enabled,omitempty"`
	IsReachable                       bool   `json:"is_reachable,omitempty"`
	IsSentinelEnabled                 bool   `json:"is_sentinel_enabled,omitempty"`
	IsSwarmManager                    bool   `json:"is_swarm_manager,omitempty"`
	IsSwarmWorker                     bool   `json:"is_swarm_worker,omitempty"`
	IsUsable                          bool   `json:"is_usable,omitempty"`
	LogdrainAxiomAPIKey               string `json:"logdrain_axiom_api_key,omitempty"`
	LogdrainAxiomDatasetName          string `json:"logdrain_axiom_dataset_name,omitempty"`
	LogdrainCustomConfig              string `json:"logdrain_custom_config,omitempty"`
	LogdrainCustomConfigParser        string `json:"logdrain_custom_config_parser,omitempty"`
	LogdrainHighlightProjectID        string `json:"logdrain_highlight_project_id,omitempty"`
	LogdrainNewrelicBaseURI           string `json:"logdrain_newrelic_base_uri,omitempty"`
	LogdrainNewrelicLicenseKey        string `json:"logdrain_newrelic_license_key,omitempty"`
	SentinelMetricsHistoryDays        int    `json:"sentinel_metrics_history_days,omitempty"`
	SentinelMetricsRefreshRateSeconds int    `json:"sentinel_metrics_refresh_rate_seconds,omitempty"`
	SentinelToken                     string `json:"sentinel_token,omitempty"`
	DockerCleanupFrequency            string `json:"docker_cleanup_frequency,omitempty"`
	DockerCleanupThreshold            int    `json:"docker_cleanup_threshold,omitempty"`
	ServerID                          int    `json:"server_id,omitempty"`
	WildcardDomain                    string `json:"wildcard_domain,omitempty"`
	CreatedAt                         string `json:"created_at,omitempty"`
	UpdatedAt                         string `json:"updated_at,omitempty"`
	DeleteUnusedVolumes               bool   `json:"delete_unused_volumes,omitempty"`
	DeleteUnusedNetworks              bool   `json:"delete_unused_networks,omitempty"`
}

// Resource represents a resource on a server (from CAGC)
type Resource struct {
	ID        int    `json:"id,omitempty"`
	UUID      string `json:"uuid,omitempty"`
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	Status    string `json:"status,omitempty"`
}

// ServerDomain represents a domain configuration on a server (from CAGC)
type ServerDomain struct {
	IP      string   `json:"ip,omitempty"`
	Domains []string `json:"domains,omitempty"`
}

// Project represents a Coolify project
type Project struct {
	ID           int           `json:"id,omitempty"`
	UUID         string        `json:"uuid,omitempty"`
	Name         string        `json:"name,omitempty"`
	Description  string        `json:"description,omitempty"`
	Environments []Environment `json:"environments,omitempty"`
	CreatedAt    string        `json:"created_at,omitempty"`
	UpdatedAt    string        `json:"updated_at,omitempty"`
}

// Environment represents a Coolify environment within a project
type Environment struct {
	ID          int    `json:"id,omitempty"`
	UUID        string `json:"uuid,omitempty"`
	Name        string `json:"name,omitempty"`
	ProjectID   int    `json:"project_id,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// Application represents a Coolify application (enhanced from CAGC)
type Application struct {
	ID                             int     `json:"id,omitempty"`
	RepositoryProjectID            *int    `json:"repository_project_id,omitempty"`
	UUID                           string  `json:"uuid,omitempty"`
	Name                           string  `json:"name,omitempty"`
	Fqdn                           *string `json:"fqdn,omitempty"`
	ConfigHash                     string  `json:"config_hash,omitempty"`
	GitRepository                  string  `json:"git_repository,omitempty"`
	GitBranch                      string  `json:"git_branch,omitempty"`
	GitCommitSHA                   string  `json:"git_commit_sha,omitempty"`
	GitFullURL                     *string `json:"git_full_url,omitempty"`
	DockerRegistryImageName        *string `json:"docker_registry_image_name,omitempty"`
	DockerRegistryImageTag         *string `json:"docker_registry_image_tag,omitempty"`
	BuildPack                      string  `json:"build_pack,omitempty"`
	StaticImage                    string  `json:"static_image,omitempty"`
	InstallCommand                 string  `json:"install_command,omitempty"`
	BuildCommand                   string  `json:"build_command,omitempty"`
	StartCommand                   string  `json:"start_command,omitempty"`
	PortsExposes                   string  `json:"ports_exposes,omitempty"`
	PortsMappings                  *string `json:"ports_mappings,omitempty"`
	BaseDirectory                  string  `json:"base_directory,omitempty"`
	PublishDirectory               string  `json:"publish_directory,omitempty"`
	HealthCheckEnabled             bool    `json:"health_check_enabled,omitempty"`
	HealthCheckPath                string  `json:"health_check_path,omitempty"`
	HealthCheckPort                *string `json:"health_check_port,omitempty"`
	HealthCheckHost                *string `json:"health_check_host,omitempty"`
	HealthCheckMethod              string  `json:"health_check_method,omitempty"`
	HealthCheckReturnCode          int     `json:"health_check_return_code,omitempty"`
	HealthCheckScheme              string  `json:"health_check_scheme,omitempty"`
	HealthCheckResponseText        *string `json:"health_check_response_text,omitempty"`
	HealthCheckInterval            int     `json:"health_check_interval,omitempty"`
	HealthCheckTimeout             int     `json:"health_check_timeout,omitempty"`
	HealthCheckRetries             int     `json:"health_check_retries,omitempty"`
	HealthCheckStartPeriod         int     `json:"health_check_start_period,omitempty"`
	LimitsMemory                   string  `json:"limits_memory,omitempty"`
	LimitsMemorySwap               string  `json:"limits_memory_swap,omitempty"`
	LimitsMemorySwappiness         int     `json:"limits_memory_swappiness,omitempty"`
	LimitsMemoryReservation        string  `json:"limits_memory_reservation,omitempty"`
	LimitsCPUs                     string  `json:"limits_cpus,omitempty"`
	LimitsCPUSet                   *string `json:"limits_cpuset,omitempty"`
	LimitsCPUShares                int     `json:"limits_cpu_shares,omitempty"`
	Status                         string  `json:"status,omitempty"`
	PreviewURLTemplate             string  `json:"preview_url_template,omitempty"`
	DestinationType                string  `json:"destination_type,omitempty"`
	DestinationID                  int     `json:"destination_id,omitempty"`
	SourceID                       *int    `json:"source_id,omitempty"`
	PrivateKeyID                   *int    `json:"private_key_id,omitempty"`
	EnvironmentID                  int     `json:"environment_id,omitempty"`
	Dockerfile                     *string `json:"dockerfile,omitempty"`
	DockerfileLocation             string  `json:"dockerfile_location,omitempty"`
	CustomLabels                   *string `json:"custom_labels,omitempty"`
	DockerfileTargetBuild          *string `json:"dockerfile_target_build,omitempty"`
	ManualWebhookSecretGithub      *string `json:"manual_webhook_secret_github,omitempty"`
	ManualWebhookSecretGitlab      *string `json:"manual_webhook_secret_gitlab,omitempty"`
	ManualWebhookSecretBitbucket   *string `json:"manual_webhook_secret_bitbucket,omitempty"`
	ManualWebhookSecretGitea       *string `json:"manual_webhook_secret_gitea,omitempty"`
	DockerComposeLocation          string  `json:"docker_compose_location,omitempty"`
	DockerCompose                  *string `json:"docker_compose,omitempty"`
	DockerComposeRaw               *string `json:"docker_compose_raw,omitempty"`
	DockerComposeDomains           *string `json:"docker_compose_domains,omitempty"`
	DockerComposeCustomStartCmd    *string `json:"docker_compose_custom_start_command,omitempty"`
	DockerComposeCustomBuildCmd    *string `json:"docker_compose_custom_build_command,omitempty"`
	SwarmReplicas                  *int    `json:"swarm_replicas,omitempty"`
	SwarmPlacementConstraints      *string `json:"swarm_placement_constraints,omitempty"`
	CustomDockerRunOptions         *string `json:"custom_docker_run_options,omitempty"`
	PostDeploymentCommand          *string `json:"post_deployment_command,omitempty"`
	PostDeploymentCommandContainer *string `json:"post_deployment_command_container,omitempty"`
	PreDeploymentCommand           *string `json:"pre_deployment_command,omitempty"`
	PreDeploymentCommandContainer  *string `json:"pre_deployment_command_container,omitempty"`
	WatchPaths                     *string `json:"watch_paths,omitempty"`
	CustomHealthcheckFound         bool    `json:"custom_healthcheck_found,omitempty"`
	Redirect                       *string `json:"redirect,omitempty"`
	CreatedAt                      string  `json:"created_at,omitempty"`
	UpdatedAt                      string  `json:"updated_at,omitempty"`
	DeletedAt                      *string `json:"deleted_at,omitempty"`
	ComposeParsingVersion          string  `json:"compose_parsing_version,omitempty"`
	CustomNginxConfiguration       *string `json:"custom_nginx_configuration,omitempty"`
	Domains                        string  `json:"domains,omitempty"`
	IsHTTPBasicAuthEnabled         bool    `json:"is_http_basic_auth_enabled,omitempty"`
	HTTPBasicAuthUsername          *string `json:"http_basic_auth_username,omitempty"`
	HTTPBasicAuthPassword          *string `json:"http_basic_auth_password,omitempty"`
	ConnectToDockerNetwork         bool    `json:"connect_to_docker_network,omitempty"`
	ForceDomainOverride            bool    `json:"force_domain_override,omitempty"`
}

// CreatePublicAppRequest is the request body for creating a public app
type CreatePublicAppRequest struct {
	ProjectUUID      string `json:"project_uuid"`
	ServerUUID       string `json:"server_uuid"`
	EnvironmentName  string `json:"environment_name,omitempty"`
	EnvironmentUUID  string `json:"environment_uuid,omitempty"`
	GitRepository    string `json:"git_repository"`
	GitBranch        string `json:"git_branch"`
	BuildPack        string `json:"build_pack,omitempty"`
	Name             string `json:"name,omitempty"`
	Description      string `json:"description,omitempty"`
	Domains          string `json:"domains,omitempty"`
	InstantDeploy    bool   `json:"instant_deploy,omitempty"`
	InstallCommand   string `json:"install_command,omitempty"`
	BuildCommand     string `json:"build_command,omitempty"`
	StartCommand     string `json:"start_command,omitempty"`
	PortsExposes     string `json:"ports_exposes,omitempty"`
	PublishDirectory string `json:"publish_directory,omitempty"`
	BaseDirectory    string `json:"base_directory,omitempty"`
}

// CreateDockerImageAppRequest is the request body for creating a docker image app
type CreateDockerImageAppRequest struct {
	ProjectUUID             string `json:"project_uuid"`
	ServerUUID              string `json:"server_uuid"`
	EnvironmentName         string `json:"environment_name,omitempty"`
	EnvironmentUUID         string `json:"environment_uuid,omitempty"`
	Name                    string `json:"name,omitempty"`
	Description             string `json:"description,omitempty"`
	Domains                 string `json:"domains,omitempty"`
	InstantDeploy           bool   `json:"instant_deploy,omitempty"`
	DockerRegistryImageName string `json:"docker_registry_image_name"`
	DockerRegistryImageTag  string `json:"docker_registry_image_tag,omitempty"`
	PortsExposes            string `json:"ports_exposes,omitempty"`
}

// CreateAppResponse is the response from creating an app
type CreateAppResponse struct {
	UUID string `json:"uuid"`
}

// DeployResponse is the response from triggering a deployment
type DeployResponse struct {
	Deployments []DeploymentInfo `json:"deployments"`
}

// DeploymentInfo contains info about a triggered deployment
type DeploymentInfo struct {
	Message        string `json:"message"`
	ResourceUUID   string `json:"resource_uuid"`
	DeploymentUUID string `json:"deployment_uuid"`
}

// EnvVar represents an environment variable
type EnvVar struct {
	ID          int    `json:"id"`
	UUID        string `json:"uuid"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsBuildTime bool   `json:"is_build_time"`
	IsPreview   bool   `json:"is_preview"`
}

// HealthCheckResponse is the response from the health check endpoint
type HealthCheckResponse struct {
	Status string `json:"status"`
}

// Team represents a Coolify team
type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GitHubApp represents a GitHub App configured in Coolify
type GitHubApp struct {
	ID             int    `json:"id"`
	UUID           string `json:"uuid"`
	Name           string `json:"name"`
	Organization   string `json:"organization"`
	AppID          int    `json:"app_id"`
	InstallationID int    `json:"installation_id"`
	IsSystemWide   bool   `json:"is_system_wide"`
}

// CreatePrivateGitHubAppRequest is the request body for creating a private GitHub app
type CreatePrivateGitHubAppRequest struct {
	ProjectUUID        string `json:"project_uuid"`
	ServerUUID         string `json:"server_uuid"`
	EnvironmentName    string `json:"environment_name,omitempty"`
	EnvironmentUUID    string `json:"environment_uuid,omitempty"`
	GitHubAppUUID      string `json:"github_app_uuid"`
	GitRepository      string `json:"git_repository"`
	GitBranch          string `json:"git_branch"`
	BuildPack          string `json:"build_pack,omitempty"`
	IsStatic           bool   `json:"is_static,omitempty"`
	Name               string `json:"name,omitempty"`
	Description        string `json:"description,omitempty"`
	Domains            string `json:"domains,omitempty"`
	InstantDeploy      bool   `json:"instant_deploy,omitempty"`
	InstallCommand     string `json:"install_command,omitempty"`
	BuildCommand       string `json:"build_command,omitempty"`
	StartCommand       string `json:"start_command,omitempty"`
	PortsExposes       string `json:"ports_exposes,omitempty"`
	PublishDirectory   string `json:"publish_directory,omitempty"`
	BaseDirectory      string `json:"base_directory,omitempty"`
	HealthCheckEnabled bool   `json:"health_check_enabled,omitempty"`
	HealthCheckPath    string `json:"health_check_path,omitempty"`
}

// NOTE: Database, Service, and Deployment types are in their respective files
// (databases.go, deployments.go) to avoid duplication

// EnvironmentVariable represents a Coolify environment variable (from CAGC)
type EnvironmentVariable struct {
	UUID        string `json:"uuid,omitempty"`
	Key         string `json:"key,omitempty"`
	Value       string `json:"value,omitempty"`
	IsPreview   bool   `json:"is_preview,omitempty"`
	IsBuildTime bool   `json:"is_build_time,omitempty"`
	IsLiteral   bool   `json:"is_literal,omitempty"`
	IsMultiline bool   `json:"is_multiline,omitempty"`
	IsShownOnce bool   `json:"is_shown_once,omitempty"`
}

// CreateResponse is a generic response for create operations (from CAGC)
type CreateResponse struct {
	UUID    string `json:"uuid,omitempty"`
	Message string `json:"message,omitempty"`
}

// DeploymentResponse represents a response for deployment operations (from CAGC)
type DeploymentResponse struct {
	Message        string `json:"message,omitempty"`
	DeploymentUUID string `json:"deployment_uuid,omitempty"`
}

// CommandResponse represents a command execution response (from CAGC)
type CommandResponse struct {
	Message  string `json:"message,omitempty"`
	Response string `json:"response,omitempty"`
}

// MessageResponse represents a simple message response (from CAGC)
type MessageResponse struct {
	Message string `json:"message,omitempty"`
}

// LogsResponse represents a logs response (from CAGC)
type LogsResponse struct {
	Logs string `json:"logs,omitempty"`
}

// PrivateKey represents a Coolify private key for SSH access (from CAGC)
type PrivateKey struct {
	ID           int    `json:"id,omitempty"`
	UUID         string `json:"uuid,omitempty"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	PrivateKey   string `json:"private_key,omitempty"`
	IsGitRelated bool   `json:"is_git_related,omitempty"`
	TeamID       int    `json:"team_id,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
	PublicKey    string `json:"public_key,omitempty"`
	Fingerprint  string `json:"fingerprint,omitempty"`
}

// Destination represents a Coolify destination (from CAGC)
type Destination struct {
	UUID          string `json:"uuid,omitempty"`
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	ServerUUID    string `json:"server_uuid,omitempty"`
	EngineType    string `json:"engine_type,omitempty"`
	NetworkUUID   string `json:"network_uuid,omitempty"`
	NetworkName   string `json:"network_name,omitempty"`
	Engine        string `json:"engine,omitempty"`
	ResourceCount int    `json:"resource_count,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}
