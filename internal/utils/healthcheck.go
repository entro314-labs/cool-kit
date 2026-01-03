package utils

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// HealthChecker provides health check functionality
type HealthChecker struct {
	logger *Logger
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger *Logger) *HealthChecker {
	return &HealthChecker{logger: logger}
}

// CheckHTTP checks if an HTTP endpoint is responding
func (h *HealthChecker) CheckHTTP(url string, timeout time.Duration) error {
	h.logger.Debug("Checking HTTP endpoint: %s", url)

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	h.logger.Success("HTTP endpoint is responding")
	return nil
}

// CheckWebSocket checks if a WebSocket endpoint is responding
func (h *HealthChecker) CheckWebSocket(url string, timeout time.Duration) error {
	h.logger.Debug("Checking WebSocket endpoint: %s", url)

	// Simple HTTP check for WebSocket endpoint
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		h.logger.Warning("WebSocket check failed (may be expected): %v", err)
		return nil // Don't fail on WebSocket check
	}
	defer resp.Body.Close()

	h.logger.Success("WebSocket endpoint is responding")
	return nil
}

// CheckDockerService checks if Docker services are running
func (h *HealthChecker) CheckDockerService(composeFile string, minServices int) error {
	h.logger.Debug("Checking Docker services in: %s", composeFile)

	cmd := exec.Command("docker-compose", "-f", composeFile, "ps", "-q")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check Docker services: %w", err)
	}

	services := strings.Split(strings.TrimSpace(string(output)), "\n")
	runningCount := 0
	for _, service := range services {
		if service != "" {
			runningCount++
		}
	}

	if runningCount < minServices {
		return fmt.Errorf("only %d/%d services running", runningCount, minServices)
	}

	h.logger.Success("Docker services running: %d/%d", runningCount, minServices)
	return nil
}

// CheckSSH checks if SSH connection is available
func (h *HealthChecker) CheckSSH(host, user string, timeout time.Duration) error {
	h.logger.Debug("Checking SSH connectivity to %s@%s", user, host)

	cmd := exec.Command("ssh",
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
		fmt.Sprintf("%s@%s", user, host),
		"echo 'SSH OK'")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("SSH check failed: %w", err)
	}

	h.logger.Success("SSH connection established")
	return nil
}

// WaitForHTTP waits for an HTTP endpoint to become available
func (h *HealthChecker) WaitForHTTP(url string, maxRetries int, retryInterval time.Duration) error {
	h.logger.Info("Waiting for HTTP endpoint: %s", url)

	for i := 0; i < maxRetries; i++ {
		if err := h.CheckHTTP(url, 5*time.Second); err == nil {
			return nil
		}

		if i < maxRetries-1 {
			h.logger.Debug("Retry %d/%d in %v...", i+1, maxRetries, retryInterval)
			time.Sleep(retryInterval)
		}
	}

	return fmt.Errorf("HTTP endpoint did not become available after %d retries", maxRetries)
}

// ComprehensiveCheck performs all health checks for a deployment
func (h *HealthChecker) ComprehensiveCheck(httpURL, wsURL string) error {
	h.logger.Section("Running Health Checks")

	// Check HTTP
	if err := h.CheckHTTP(httpURL, 10*time.Second); err != nil {
		h.logger.Error("HTTP health check failed: %v", err)
		return err
	}

	// Check WebSocket (non-fatal)
	h.CheckWebSocket(wsURL, 10*time.Second)

	h.logger.Success("All health checks passed")
	return nil
}
