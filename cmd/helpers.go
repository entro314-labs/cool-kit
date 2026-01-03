// Package cmd implements the command line interface for cool-kit.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/entro314-labs/cool-kit/internal/config"
)

// checkLogin ensures the user is authenticated
func checkLogin() error {
	if !config.IsLoggedIn() {
		return fmt.Errorf("not logged in: run '%s login' first", execName())
	}
	return nil
}

// IsVerbose checks if verbose mode is enabled
// Uses a package-level variable set during command execution to avoid init cycle
var verboseEnabled bool

func IsVerbose() bool {
	// Check environment variable as fallback
	if os.Getenv("COOLKIT_VERBOSE") == "true" || os.Getenv("COOLKIT_VERBOSE") == "1" {
		return true
	}
	return verboseEnabled
}

// SetVerbose sets the verbose flag (called by commands that have access to flags)
func SetVerbose(v bool) {
	verboseEnabled = v
}

// execName returns the executable name
func execName() string {
	name := filepath.Base(os.Args[0])
	if name == "" {
		return "coolify-deployer"
	}
	return name
}

// getWorkingDirName returns the name of the current working directory
func getWorkingDirName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "app"
	}
	return filepath.Base(cwd)
}

// Instance override set by --instance flag
var instanceOverride string

// SetInstanceOverride sets the instance override (called by PersistentPreRun)
func SetInstanceOverride(name string) {
	instanceOverride = name
}

// getCurrentInstance gets the current instance, respecting --instance flag override
func getCurrentInstance() (*config.Instance, error) {
	// Check for instance override
	if instanceOverride != "" {
		return config.GetInstance(instanceOverride)
	}

	// Use context-based instance
	return config.GetCurrentInstance()
}
