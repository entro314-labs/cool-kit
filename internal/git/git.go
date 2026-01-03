package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
)

// Manager handles Git operations
type Manager struct {
	config *config.Config
}

// NewManager creates a new Git manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{config: cfg}
}

// CloneOrPull clones the Coolify repository or uses existing
// Uses native git CLI for reliability
func (m *Manager) CloneOrPull() error {
	workDir := m.config.Git.WorkDir
	repoURL := m.config.Git.Repository
	branch := m.config.Git.Branch

	// Convert to absolute path
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if it's already a valid git repo
	if isGitRepo(absWorkDir) {
		// Just fetch/pull latest
		return m.pullLatest(absWorkDir, branch)
	}

	// Not a git repo - need to clone
	// Remove any existing directory first
	if _, err := os.Stat(absWorkDir); err == nil {
		if err := os.RemoveAll(absWorkDir); err != nil {
			return fmt.Errorf("failed to clean existing directory: %w", err)
		}
	}

	// Clone fresh
	return m.cloneFresh(absWorkDir, repoURL, branch)
}

// isGitRepo checks if directory is a valid git repository
func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// cloneFresh clones the repository
func (m *Manager) cloneFresh(workDir, repoURL, branch string) error {
	// Create parent directory
	parentDir := filepath.Dir(workDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Use git CLI - much more reliable than go-git
	cmd := exec.Command("git", "clone",
		"--depth", "1",
		"--branch", branch,
		"--single-branch",
		repoURL,
		workDir)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return fmt.Errorf("git clone failed: %s", errMsg)
		}
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

// pullLatest pulls the latest changes
func (m *Manager) pullLatest(workDir, branch string) error {
	// Try to pull
	cmd := exec.Command("git", "pull", "--ff-only", "origin", branch)
	cmd.Dir = workDir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Pull failed - might be already up to date or other issue
		// Just continue with what we have
		return nil
	}

	return nil
}

// GetLatestCommitInfo returns information about the latest commit
func (m *Manager) GetLatestCommitInfo() (*CommitInfo, error) {
	workDir := m.config.Git.WorkDir

	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// Get commit hash
	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	hashCmd.Dir = absWorkDir
	hashOutput, err := hashCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit hash: %w", err)
	}
	hash := strings.TrimSpace(string(hashOutput))

	// Get commit message
	msgCmd := exec.Command("git", "log", "-1", "--format=%s")
	msgCmd.Dir = absWorkDir
	msgOutput, _ := msgCmd.Output()
	message := strings.TrimSpace(string(msgOutput))

	// Get author
	authorCmd := exec.Command("git", "log", "-1", "--format=%an")
	authorCmd.Dir = absWorkDir
	authorOutput, _ := authorCmd.Output()
	author := strings.TrimSpace(string(authorOutput))

	shortHash := hash
	if len(hash) > 8 {
		shortHash = hash[:8]
	}

	return &CommitInfo{
		Hash:      hash,
		ShortHash: shortHash,
		Message:   message,
		Author:    author,
		Date:      time.Now(), // Simplified
	}, nil
}

// Clean removes the local repository clone
func (m *Manager) Clean() error {
	workDir := m.config.Git.WorkDir

	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absWorkDir); os.IsNotExist(err) {
		return nil
	}

	return os.RemoveAll(absWorkDir)
}

// CommitInfo represents information about a commit
type CommitInfo struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"short_hash"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Date      time.Time `json:"date"`
}

// PullRepo pulls the latest changes (public method for external use)
func (m *Manager) PullRepo(workDir, branch string) error {
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return err
	}
	return m.pullLatest(absWorkDir, branch)
}
