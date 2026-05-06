package generator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// minJoomlaDirs is the minimum number of required directories that must be
// present (alongside configuration.php) to consider a project as Joomla.
const minJoomlaDirs = 3

// DetectJoomlaProject checks whether dir contains a Joomla project.
// It tries the Joomla CLI first (Joomla 5+), then falls back to checking
// for configuration.php and standard directory structure.
func DetectJoomlaProject(ctx context.Context, dir string) (bool, error) {
	// First, check that the directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("directory does not exist: %s", dir)
		}
		return false, fmt.Errorf("cannot access directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("not a directory: %s", dir)
	}

	// Try Joomla CLI first (authoritative for Joomla 5+)
	detected, err := detectViaCLI(ctx, dir)
	if err == nil && detected {
		return true, nil
	}

	// Fallback: check configuration.php + required directories
	return detectViaConfig(ctx, dir)
}

// detectViaCLI tries to detect Joomla via the Joomla CLI (cli/joomla.php).
func detectViaCLI(ctx context.Context, dir string) (bool, error) {
	cliPath := filepath.Join(dir, "cli", "joomla.php")
	if _, err := os.Stat(cliPath); err != nil {
		return false, err
	}

	cmd := exec.CommandContext(ctx, "php", cliPath, "list")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		// CLI exists but couldn't run (e.g. php not installed)
		return false, err
	}
	return true, nil
}

// detectViaConfig checks for configuration.php and standard Joomla directories.
// Requires at least minJoomlaDirs of the standard directories to be present.
func detectViaConfig(ctx context.Context, dir string) (bool, error) {
	requiredJoomlaDirs := []string{
		"administrator",
		"components",
		"modules",
		"plugins",
		"templates",
		"libraries",
		"includes",
	}

	configPath := filepath.Join(dir, "configuration.php")
	if _, err := os.Stat(configPath); err != nil {
		return false, nil
	}

	// Count how many required directories are present
	found := 0
	for _, d := range requiredJoomlaDirs {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		info, err := os.Stat(filepath.Join(dir, d))
		if err == nil && info.IsDir() {
			found++
		}
	}

	// Require at least minJoomlaDirs of the standard Joomla directories
	return found >= minJoomlaDirs, nil
}
