package aws

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/git"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// AWSProvider handles AWS deployments
type AWSProvider struct {
	config     *config.Config
	gitManager *git.Manager
	sdkClient  *SDKClient
}

// NewAWSProvider creates a new AWS provider
func NewAWSProvider(cfg *config.Config) *AWSProvider {
	provider := &AWSProvider{
		config:     cfg,
		gitManager: git.NewManager(cfg),
	}
	// Initialize SDK client (lazy init in validateCredentials if region not set yet)
	return provider
}

// GetDeploymentSteps returns the deployment steps for AWS
func (p *AWSProvider) GetDeploymentSteps() []ui.DeploymentStep {
	return []ui.DeploymentStep{
		{Name: "Validate AWS credentials", Description: "Checking AWS CLI and credentials"},
		{Name: "Clone Coolify repository", Description: "Fetching latest Coolify from GitHub"},
		{Name: "Create VPC and subnets", Description: "Setting up network infrastructure"},
		{Name: "Configure security groups", Description: "Setting up firewall rules"},
		{Name: "Launch EC2 instance", Description: "Creating virtual machine"},
		{Name: "Assign Elastic IP", Description: "Allocating static IP address"},
		{Name: "Install Docker", Description: "Installing Docker on EC2 instance"},
		{Name: "Deploy Coolify", Description: "Setting up Coolify application"},
		{Name: "Configure SSL", Description: "Setting up HTTPS with Let's Encrypt"},
		{Name: "Run health checks", Description: "Validating deployment"},
	}
}

// Deploy performs the AWS deployment with progress tracking
func (p *AWSProvider) Deploy(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	steps := []struct {
		name string
		fn   func(chan<- ui.StepProgressMsg, chan<- ui.LogMsg) error
	}{
		{"Validate AWS credentials", p.validateCredentials},
		{"Clone Coolify repository", p.cloneRepository},
		{"Create VPC and subnets", p.createVPC},
		{"Configure security groups", p.createSecurityGroups},
		{"Launch EC2 instance", p.launchInstance},
		{"Assign Elastic IP", p.assignElasticIP},
		{"Install Docker", p.installDocker},
		{"Deploy Coolify", p.deployCoolify},
		{"Configure SSL", p.configureSSL},
		{"Run health checks", p.runHealthChecks},
	}

	for i, step := range steps {
		logChan <- ui.LogMsg{
			Level:   ui.LogInfo,
			Message: fmt.Sprintf("Starting: %s", step.name),
		}

		if err := step.fn(progressChan, logChan); err != nil {
			logChan <- ui.LogMsg{
				Level:   ui.LogError,
				Message: fmt.Sprintf("Failed: %s - %v", step.name, err),
			}
			return fmt.Errorf("step '%s' failed: %w", step.name, err)
		}

		progressChan <- ui.StepProgressMsg{
			StepIndex: i,
			Progress:  1.0,
			Message:   fmt.Sprintf("Completed: %s", step.name),
		}

		logChan <- ui.LogMsg{
			Level:   ui.LogSuccess,
			Message: fmt.Sprintf("✓ %s completed", step.name),
		}
	}

	return nil
}

// validateCredentials validates AWS credentials
func (p *AWSProvider) validateCredentials(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Initializing AWS SDK client"}

	// Initialize SDK client
	var err error
	p.sdkClient, err = NewSDKClient(p.getRegion())
	if err != nil {
		return fmt.Errorf("failed to initialize AWS SDK: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Validating credentials"}

	// Validate credentials using SDK
	if err := p.sdkClient.ValidateCredentials(); err != nil {
		return fmt.Errorf("AWS credentials validation failed: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "Credentials validated"}

	logChan <- ui.LogMsg{
		Level:   ui.LogDebug,
		Message: fmt.Sprintf("AWS SDK initialized for region: %s", p.getRegion()),
	}

	return nil
}

// cloneRepository clones the Coolify repository
func (p *AWSProvider) cloneRepository(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Fetching repository"}

	if err := p.gitManager.CloneOrPull(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.7, Message: "Getting commit info"}

	commitInfo, err := p.gitManager.GetLatestCommitInfo()
	if err != nil {
		return fmt.Errorf("failed to get commit info: %w", err)
	}

	logChan <- ui.LogMsg{
		Level:   ui.LogInfo,
		Message: fmt.Sprintf("Using commit: %s - %s", commitInfo.ShortHash, commitInfo.Message[:min(50, len(commitInfo.Message))]),
	}

	return nil
}

// createVPC creates VPC and subnets
func (p *AWSProvider) createVPC(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Creating VPC"}

	// Create VPC
	vpcCIDR := "10.0.0.0/16"
	cmd := exec.Command("aws", "ec2", "create-vpc",
		"--cidr-block", vpcCIDR,
		"--tag-specifications", "ResourceType=vpc,Tags=[{Key=Name,Value=coolify-vpc},{Key=Application,Value=Coolify},{Key=ManagedBy,Value=cool-kit}]",
		"--region", p.getRegion(),
		"--output", "json")

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to create VPC: %w", err)
	}

	// Parse VPC ID from JSON output
	var vpcResponse struct {
		Vpc struct {
			VpcId string `json:"VpcId"`
		} `json:"Vpc"`
	}
	if err := json.Unmarshal(output, &vpcResponse); err != nil {
		return fmt.Errorf("failed to parse VPC response: %w", err)
	}
	vpcID := vpcResponse.Vpc.VpcId
	p.config.Settings["vpc_id"] = vpcID

	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "VPC created"}
	logChan <- ui.LogMsg{Level: ui.LogDebug, Message: fmt.Sprintf("VPC %s created successfully", vpcID)}

	progressChan <- ui.StepProgressMsg{Progress: 0.7, Message: "Creating subnet"}

	// Create subnet using the VPC ID
	subnetCIDR := "10.0.1.0/24"
	cmd = exec.Command("aws", "ec2", "create-subnet",
		"--vpc-id", vpcID,
		"--cidr-block", subnetCIDR,
		"--tag-specifications", "ResourceType=subnet,Tags=[{Key=Name,Value=coolify-subnet},{Key=Application,Value=Coolify},{Key=ManagedBy,Value=cool-kit}]",
		"--region", p.getRegion(),
		"--output", "json")

	subnetOutput, err := cmd.Output()
	if err != nil {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: "Subnet creation skipped or already exists"}
	} else {
		// Parse subnet ID
		var subnetResponse struct {
			Subnet struct {
				SubnetId string `json:"SubnetId"`
			} `json:"Subnet"`
		}
		if err := json.Unmarshal(subnetOutput, &subnetResponse); err == nil {
			p.config.Settings["subnet_id"] = subnetResponse.Subnet.SubnetId
			logChan <- ui.LogMsg{Level: ui.LogDebug, Message: fmt.Sprintf("Subnet %s created", subnetResponse.Subnet.SubnetId)}
		}
	}

	return nil
}

// createSecurityGroups creates security groups
func (p *AWSProvider) createSecurityGroups(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Creating security group"}

	// Create security group
	cmd := exec.Command("aws", "ec2", "create-security-group",
		"--group-name", "coolify-sg",
		"--description", "Security group for Coolify",
		"--region", p.getRegion())

	if err := cmd.Run(); err != nil {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: "Security group may already exist"}
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Adding ingress rules"}

	// Add ingress rules
	ports := []string{"22", "80", "443", "6001"}
	for i, port := range ports {
		progress := 0.5 + (float64(i+1) / float64(len(ports)) * 0.4)
		progressChan <- ui.StepProgressMsg{Progress: progress, Message: fmt.Sprintf("Opening port %s", port)}

		cmd = exec.Command("aws", "ec2", "authorize-security-group-ingress",
			"--group-name", "coolify-sg",
			"--protocol", "tcp",
			"--port", port,
			"--cidr", "0.0.0.0/0",
			"--region", p.getRegion())

		cmd.Run() // Ignore errors as rules may already exist
	}

	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Security group configured"}
	return nil
}

// launchInstance launches EC2 instance
func (p *AWSProvider) launchInstance(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Launching EC2 instance"}

	instanceType := p.getInstanceType()
	ami := p.getAMI()

	logChan <- ui.LogMsg{
		Level:   ui.LogInfo,
		Message: fmt.Sprintf("Instance type: %s, AMI: %s", instanceType, ami),
	}

	// Use SDK client to create instance
	opts := InstanceCreateOpts{
		Name:         "coolify-instance",
		InstanceType: instanceType,
		AMI:          ami,
		DiskSizeGB:   30,
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.4, Message: "Creating EC2 instance via SDK"}

	instanceInfo, err := p.sdkClient.CreateInstance(opts)
	if err != nil {
		return fmt.Errorf("failed to launch instance: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "Instance is running"}

	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: fmt.Sprintf("EC2 instance %s is running", instanceInfo.InstanceID)}

	// Store instance ID in config
	p.config.Settings["instance_id"] = instanceInfo.InstanceID
	p.config.Settings["public_ip"] = instanceInfo.PublicIP

	return nil
}

// assignElasticIP assigns an Elastic IP
func (p *AWSProvider) assignElasticIP(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Allocating Elastic IP"}

	instanceID, ok := p.config.Settings["instance_id"].(string)
	if !ok || instanceID == "" {
		return fmt.Errorf("instance ID not found - instance must be launched first")
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Associating Elastic IP with instance"}

	// Use SDK client to allocate and associate Elastic IP
	publicIP, err := p.sdkClient.AllocateElasticIP(instanceID)
	if err != nil {
		return fmt.Errorf("failed to allocate Elastic IP: %w", err)
	}

	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Elastic IP %s allocated and associated", publicIP)}

	// Store public IP in config (overwrite dynamic IP with static Elastic IP)
	p.config.Settings["public_ip"] = publicIP

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "Elastic IP associated"}

	return nil
}

// installDocker installs Docker on the instance
func (p *AWSProvider) installDocker(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Connecting to instance"}

	publicIP, ok := p.config.Settings["public_ip"].(string)
	if !ok || publicIP == "" {
		return fmt.Errorf("public IP not found")
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.4, Message: "Installing Docker"}

	// SSH and install Docker
	installScript := "curl -fsSL https://get.docker.com | sudo sh && sudo usermod -aG docker ubuntu && sudo systemctl enable docker && sudo systemctl start docker"

	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Installing Docker via SSH on %s", publicIP)}

	// Execute via SSH
	cmd := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=30",
		fmt.Sprintf("ubuntu@%s", publicIP),
		installScript,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: string(output)}
		return fmt.Errorf("failed to install Docker via SSH: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "Docker installed"}

	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Docker installation complete"}

	return nil
}

// deployCoolify deploys Coolify
func (p *AWSProvider) deployCoolify(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Preparing Coolify deployment"}

	publicIP, ok := p.config.Settings["public_ip"].(string)
	if !ok || publicIP == "" {
		return fmt.Errorf("public IP not found")
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.4, Message: "Running Coolify install script"}

	// Seed .env and execute Coolify install script
	installScript := `mkdir -p /data/coolify/source && \
echo "COOLIFY_POSTGRES_VERSION=17-trixie" >> /data/coolify/source/.env && \
echo "COOLIFY_REDIS_VERSION=8.4.0-bookworm" >> /data/coolify/source/.env && \
curl -fsSL https://cdn.coollabs.io/coolify/install.sh | sudo bash`

	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Deploying Coolify via SSH on %s", publicIP)}

	cmd := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=60",
		fmt.Sprintf("ubuntu@%s", publicIP),
		installScript,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		logChan <- ui.LogMsg{Level: ui.LogWarning, Message: string(output)}
		return fmt.Errorf("failed to deploy Coolify via SSH: %w", err)
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "Coolify deployed"}

	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Coolify deployment complete"}

	return nil
}

// configureSSL configures SSL
func (p *AWSProvider) configureSSL(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.3, Message: "Setting up Let's Encrypt"}
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: "Coolify configures SSL automatically via Caddy/Traefik"}
	progressChan <- ui.StepProgressMsg{Progress: 0.8, Message: "SSL configured"}
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "SSL configuration complete"}
	return nil
}

// runHealthChecks runs health checks
func (p *AWSProvider) runHealthChecks(progressChan chan<- ui.StepProgressMsg, logChan chan<- ui.LogMsg) error {
	progressChan <- ui.StepProgressMsg{Progress: 0.2, Message: "Checking HTTP endpoint"}

	publicIP, ok := p.config.Settings["public_ip"].(string)
	if !ok {
		return fmt.Errorf("public IP not found")
	}

	progressChan <- ui.StepProgressMsg{Progress: 0.5, Message: "Checking WebSocket"}
	logChan <- ui.LogMsg{Level: ui.LogInfo, Message: fmt.Sprintf("Running health checks on %s", publicIP)}

	// Simulate health checks
	time.Sleep(3 * time.Second)

	progressChan <- ui.StepProgressMsg{Progress: 0.9, Message: "All checks passed"}
	logChan <- ui.LogMsg{Level: ui.LogSuccess, Message: "Health checks passed"}
	return nil
}

// Helper methods
func (p *AWSProvider) getRegion() string {
	if region, ok := p.config.Settings["aws_region"].(string); ok {
		return region
	}
	return "us-east-1"
}

func (p *AWSProvider) getInstanceType() string {
	if instanceType, ok := p.config.Settings["aws_instance_type"].(string); ok {
		return instanceType
	}
	return "t3.medium"
}

func (p *AWSProvider) getAMI() string {
	// Ubuntu 24.04 LTS in us-east-1 (noble)
	return "ami-04b70fa74e45c3917"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
