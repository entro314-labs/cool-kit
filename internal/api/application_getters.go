package api

// GetRepositoryProjectID returns the RepositoryProjectID or 0 if nil
func (a *Application) GetRepositoryProjectID() int {
	if a.RepositoryProjectID == nil {
		return 0
	}
	return *a.RepositoryProjectID
}

// GetFqdn returns the Fqdn or empty string if nil
func (a *Application) GetFqdn() string {
	if a.Fqdn == nil {
		return ""
	}
	return *a.Fqdn
}

// GetGitFullURL returns the GitFullURL or empty string if nil
func (a *Application) GetGitFullURL() string {
	if a.GitFullURL == nil {
		return ""
	}
	return *a.GitFullURL
}

// GetDockerRegistryImageName returns the DockerRegistryImageName or empty string if nil
func (a *Application) GetDockerRegistryImageName() string {
	if a.DockerRegistryImageName == nil {
		return ""
	}
	return *a.DockerRegistryImageName
}

// GetDockerRegistryImageTag returns the DockerRegistryImageTag or empty string if nil
func (a *Application) GetDockerRegistryImageTag() string {
	if a.DockerRegistryImageTag == nil {
		return ""
	}
	return *a.DockerRegistryImageTag
}

// GetPortsMappings returns the PortsMappings or empty string if nil
func (a *Application) GetPortsMappings() string {
	if a.PortsMappings == nil {
		return ""
	}
	return *a.PortsMappings
}

// GetHealthCheckPort returns the HealthCheckPort or empty string if nil
func (a *Application) GetHealthCheckPort() string {
	if a.HealthCheckPort == nil {
		return ""
	}
	return *a.HealthCheckPort
}

// GetHealthCheckHost returns the HealthCheckHost or empty string if nil
func (a *Application) GetHealthCheckHost() string {
	if a.HealthCheckHost == nil {
		return ""
	}
	return *a.HealthCheckHost
}

// GetHealthCheckResponseText returns the HealthCheckResponseText or empty string if nil
func (a *Application) GetHealthCheckResponseText() string {
	if a.HealthCheckResponseText == nil {
		return ""
	}
	return *a.HealthCheckResponseText
}

// GetLimitsCPUSet returns the LimitsCPUSet or empty string if nil
func (a *Application) GetLimitsCPUSet() string {
	if a.LimitsCPUSet == nil {
		return ""
	}
	return *a.LimitsCPUSet
}

// GetSourceID returns the SourceID or 0 if nil
func (a *Application) GetSourceID() int {
	if a.SourceID == nil {
		return 0
	}
	return *a.SourceID
}

// GetPrivateKeyID returns the PrivateKeyID or 0 if nil
func (a *Application) GetPrivateKeyID() int {
	if a.PrivateKeyID == nil {
		return 0
	}
	return *a.PrivateKeyID
}

// GetDockerfile returns the Dockerfile or empty string if nil
func (a *Application) GetDockerfile() string {
	if a.Dockerfile == nil {
		return ""
	}
	return *a.Dockerfile
}

// GetCustomLabels returns the CustomLabels or empty string if nil
func (a *Application) GetCustomLabels() string {
	if a.CustomLabels == nil {
		return ""
	}
	return *a.CustomLabels
}

// GetDockerfileTargetBuild returns the DockerfileTargetBuild or empty string if nil
func (a *Application) GetDockerfileTargetBuild() string {
	if a.DockerfileTargetBuild == nil {
		return ""
	}
	return *a.DockerfileTargetBuild
}

// GetManualWebhookSecretGithub returns the ManualWebhookSecretGithub or empty string if nil
func (a *Application) GetManualWebhookSecretGithub() string {
	if a.ManualWebhookSecretGithub == nil {
		return ""
	}
	return *a.ManualWebhookSecretGithub
}

// GetDockerCompose returns the DockerCompose or empty string if nil
func (a *Application) GetDockerCompose() string {
	if a.DockerCompose == nil {
		return ""
	}
	return *a.DockerCompose
}

// GetDockerComposeRaw returns the DockerComposeRaw or empty string if nil
func (a *Application) GetDockerComposeRaw() string {
	if a.DockerComposeRaw == nil {
		return ""
	}
	return *a.DockerComposeRaw
}

// GetDockerComposeDomains returns the DockerComposeDomains or empty string if nil
func (a *Application) GetDockerComposeDomains() string {
	if a.DockerComposeDomains == nil {
		return ""
	}
	return *a.DockerComposeDomains
}

// GetSwarmReplicas returns the SwarmReplicas or 0 if nil
func (a *Application) GetSwarmReplicas() int {
	if a.SwarmReplicas == nil {
		return 0
	}
	return *a.SwarmReplicas
}

// GetRedirect returns the Redirect or empty string if nil
func (a *Application) GetRedirect() string {
	if a.Redirect == nil {
		return ""
	}
	return *a.Redirect
}
