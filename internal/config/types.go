package config

// GlobalConfig represents user-level configuration for CDP functionality
type GlobalConfig struct {
	CoolifyURL     string          `json:"coolify_url"`
	CoolifyToken   string          `json:"coolify_token"`
	GitHubToken    string          `json:"github_token,omitempty"`
	DockerRegistry *DockerRegistry `json:"docker_registry,omitempty"`
}

// DockerRegistry represents Docker registry credentials
type DockerRegistry struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ProjectConfig represents per-project deployment configuration
type ProjectConfig struct {
	Name            string `json:"name"`
	DeployMethod    string `json:"deploy_method"` // "git" or "docker"
	ProjectUUID     string `json:"project_uuid"`
	ServerUUID      string `json:"server_uuid"`
	EnvironmentUUID string `json:"environment_uuid"`
	AppUUID         string `json:"app_uuid"`
	Framework       string `json:"framework"`
	BuildPack       string `json:"build_pack"`
	InstallCommand  string `json:"install_command"`
	BuildCommand    string `json:"build_command"`
	StartCommand    string `json:"start_command"`
	PublishDir      string `json:"publish_dir"`
	Port            string `json:"port"`
	Platform        string `json:"platform"` // Docker platform
	Branch          string `json:"branch"`
	Domain          string `json:"domain,omitempty"`
	DockerImage     string `json:"docker_image,omitempty"`
	GitHubRepo      string `json:"github_repo,omitempty"`
	GitHubPrivate   bool   `json:"github_private,omitempty"`
	GitHubAppUUID   string `json:"github_app_uuid,omitempty"`
}

// Deployment methods
const (
	DeployMethodGit    = "git"
	DeployMethodDocker = "docker"
	DefaultPort        = "3000"
)
