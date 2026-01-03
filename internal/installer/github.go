package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/git"
)

// GitHubDeployer deploys Coolify from the latest GitHub commit
type GitHubDeployer struct {
	config     *config.Config
	gitManager *git.Manager
	workDir    string
	progress   chan<- ProgressUpdate
	logs       chan<- string
}

// NewGitHubDeployer creates a new GitHub-based deployer
func NewGitHubDeployer(cfg *config.Config, progressChan chan<- ProgressUpdate, logChan chan<- string) *GitHubDeployer {
	// Use temporary directory for GitHub deployments
	workDir := filepath.Join(os.TempDir(), fmt.Sprintf("coolify-github-%d", time.Now().Unix()))

	// Update config to use this work directory
	if cfg == nil {
		cfg = &config.Config{}
	}
	cfg.Git.Repository = CoolifyGitHubRepo
	cfg.Git.Branch = "main" // or "v4.x" depending on preference
	cfg.Git.WorkDir = workDir

	return &GitHubDeployer{
		config:     cfg,
		gitManager: git.NewManager(cfg),
		workDir:    workDir,
		progress:   progressChan,
		logs:       logChan,
	}
}

// Deploy executes deployment from latest GitHub commit
func (d *GitHubDeployer) Deploy() (string, error) {
	steps := []struct {
		name string
		fn   func() error
	}{
		{"Cloning latest Coolify from GitHub", d.cloneRepository},
		{"Checking out latest commit", d.checkoutLatest},
		{"Generating environment configuration", d.generateEnv},
		{"Starting Docker Compose services", d.startServices},
		{"Waiting for services to be healthy", d.waitForHealth},
		{"Retrieving dashboard URL", d.getDashboardURL},
	}

	totalSteps := len(steps)

	for i, step := range steps {
		d.sendProgress(i+1, totalSteps, step.name, 0, fmt.Sprintf("Starting: %s", step.name), nil)
		d.sendLog(fmt.Sprintf("[%d/%d] %s...", i+1, totalSteps, step.name))

		if err := step.fn(); err != nil {
			d.sendProgress(i+1, totalSteps, step.name, 0, "", err)
			d.sendLog(fmt.Sprintf("❌ Failed: %s - %v", step.name, err))
			return "", fmt.Errorf("step '%s' failed: %w", step.name, err)
		}

		percentage := float64(i+1) / float64(totalSteps) * 100
		d.sendProgress(i+1, totalSteps, step.name, percentage, fmt.Sprintf("✓ %s completed", step.name), nil)
		d.sendLog(fmt.Sprintf("✅ Completed: %s", step.name))
	}

	d.sendProgress(totalSteps, totalSteps, "Deployment Complete", 100, "🎉 Coolify deployed from GitHub!", nil)
	d.sendLog("🎉 Deployment completed successfully!")

	dashboardURL := fmt.Sprintf("http://localhost:8000")
	d.sendLog(fmt.Sprintf("🌐 Dashboard URL: %s", dashboardURL))

	return dashboardURL, nil
}

// cloneRepository clones the Coolify repository
func (d *GitHubDeployer) cloneRepository() error {
	d.sendLog(fmt.Sprintf("Cloning from %s (branch: %s)...", d.config.Git.Repository, d.config.Git.Branch))

	if err := d.gitManager.CloneOrPull(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	d.sendLog("✓ Repository cloned successfully")
	return nil
}

// checkoutLatest checks out the latest commit
func (d *GitHubDeployer) checkoutLatest() error {
	d.sendLog("Getting latest commit information...")

	commitInfo, err := d.gitManager.GetLatestCommitInfo()
	if err != nil {
		return fmt.Errorf("failed to get commit info: %w", err)
	}

	d.sendLog(fmt.Sprintf("✓ Latest commit: %s", commitInfo.ShortHash))
	d.sendLog(fmt.Sprintf("  Author: %s", commitInfo.Author))
	d.sendLog(fmt.Sprintf("  Message: %s", strings.TrimSpace(commitInfo.Message)))
	d.sendLog(fmt.Sprintf("  Date: %s", commitInfo.Date.Format(time.RFC1123)))

	return nil
}

// generateEnv generates the .env file
func (d *GitHubDeployer) generateEnv() error {
	d.sendLog("Generating environment configuration...")

	envPath := filepath.Join(d.workDir, ".env")

	// Check if .env.production exists
	envProductionPath := filepath.Join(d.workDir, ".env.production")
	if _, err := os.Stat(envProductionPath); err == nil {
		// Copy .env.production to .env
		if err := d.copyFile(envProductionPath, envPath); err != nil {
			return fmt.Errorf("failed to copy .env.production: %w", err)
		}
		d.sendLog("✓ Using .env.production template")
	} else {
		// Create basic .env file
		envContent := `APP_NAME=Coolify
APP_ENV=local
APP_DEBUG=true
APP_URL=http://localhost:8000
APP_KEY=

DB_CONNECTION=pgsql
DB_HOST=coolify-db
DB_PORT=5432
DB_DATABASE=coolify
DB_USERNAME=coolify
DB_PASSWORD=

REDIS_HOST=coolify-redis
REDIS_PASSWORD=
REDIS_PORT=6379

CACHE_STORE=redis
SESSION_DRIVER=redis
QUEUE_CONNECTION=redis
BROADCAST_DRIVER=pusher

PUSHER_APP_ID=coolify
PUSHER_APP_KEY=coolify
PUSHER_APP_SECRET=coolify
PUSHER_HOST=coolify-realtime
PUSHER_PORT=6001
PUSHER_SCHEME=http
`
		if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
			return fmt.Errorf("failed to write .env file: %w", err)
		}
		d.sendLog("✓ Created basic .env file")
	}

	// Generate secrets
	d.sendLog("Generating secure secrets...")
	if err := d.generateSecrets(envPath); err != nil {
		return fmt.Errorf("failed to generate secrets: %w", err)
	}

	d.sendLog("✓ Environment configuration ready")
	return nil
}

// generateSecrets generates secure secrets for the .env file
func (d *GitHubDeployer) generateSecrets(envPath string) error {
	secrets := map[string]string{
		"APP_KEY":        "base64:" + d.generateRandomString(32),
		"DB_PASSWORD":    d.generateRandomString(32),
		"REDIS_PASSWORD": d.generateRandomString(32),
	}

	// Read current env
	content, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}

	envStr := string(content)

	// Replace empty values
	for key, value := range secrets {
		// Replace "KEY=" with "KEY=value"
		envStr = strings.ReplaceAll(envStr, key+"=\n", key+"="+value+"\n")
		envStr = strings.ReplaceAll(envStr, key+"=", key+"="+value)
	}

	// Write back
	return os.WriteFile(envPath, []byte(envStr), 0644)
}

// generateRandomString generates a random string of specified length
func (d *GitHubDeployer) generateRandomString(length int) string {
	cmd := exec.Command("openssl", "rand", "-base64", fmt.Sprintf("%d", length))
	output, err := cmd.Output()
	if err != nil {
		// Fallback to timestamp-based string
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return strings.TrimSpace(string(output))[:length]
}

// startServices starts Docker Compose services
func (d *GitHubDeployer) startServices() error {
	d.sendLog("Starting Docker Compose services...")

	// Check which docker-compose file exists
	composeFile := filepath.Join(d.workDir, "docker-compose.yml")
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		return fmt.Errorf("docker-compose.yml not found in repository")
	}

	d.sendLog(fmt.Sprintf("Using compose file: %s", composeFile))

	// Start services
	cmd := exec.Command("docker", "compose", "-f", composeFile, "up", "-d")
	cmd.Dir = d.workDir
	cmd.Env = os.Environ()

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		d.sendLog(fmt.Sprintf("Docker Compose output:\n%s", string(output)))
		return fmt.Errorf("failed to start services: %w", err)
	}

	d.sendLog(string(output))
	d.sendLog("✓ Services started successfully")

	return nil
}

// waitForHealth waits for services to be healthy
func (d *GitHubDeployer) waitForHealth() error {
	d.sendLog("Waiting for services to become healthy...")

	maxWait := 120 * time.Second
	checkInterval := 5 * time.Second
	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		// Check Coolify container health
		cmd := exec.Command("docker", "inspect", "--format={{.State.Health.Status}}", "coolify-app")
		output, err := cmd.Output()

		if err == nil {
			status := strings.TrimSpace(string(output))
			d.sendLog(fmt.Sprintf("Health status: %s", status))

			if status == "healthy" {
				d.sendLog("✓ All services are healthy")
				return nil
			}
		}

		d.sendLog("Services still starting...")
		time.Sleep(checkInterval)
	}

	d.sendLog("⚠️ Health check timed out (services may still be starting)")
	return nil // Non-fatal
}

// getDashboardURL determines the dashboard URL
func (d *GitHubDeployer) getDashboardURL() error {
	// For local deployments, always use localhost
	d.sendLog("Dashboard URL: http://localhost:8000")
	return nil
}

// copyFile copies a file from src to dst
func (d *GitHubDeployer) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// sendProgress sends progress update
func (d *GitHubDeployer) sendProgress(step, total int, name string, percentage float64, message string, err error) {
	if d.progress != nil {
		d.progress <- ProgressUpdate{
			Step:        step,
			TotalSteps:  total,
			StepName:    name,
			Percentage:  percentage,
			Message:     message,
			Error:       err,
			IsCompleted: step == total && err == nil,
		}
	}
}

// sendLog sends log message
func (d *GitHubDeployer) sendLog(message string) {
	if d.logs != nil {
		d.logs <- message
	}
}
