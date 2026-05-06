// Package init provides the interactive and parameterized initialization flow
// for JKit projects. It collects user inputs (TUI or flags) and orchestrates
// devcontainer setup, agent deployment, extension generation, and MCP configuration.
package initpkg

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alebak/jkit/internal/devcontainer"
	"gopkg.in/yaml.v3"
)

const (
	// Default images.yaml remote URL (DD-03: remote-first, no local file dependency).
	imagesRemoteURL = "https://raw.githubusercontent.com/alebak/jkit/main/images.yaml"

	// Local cache path in user's home for offline use (R-DEVC-13).
	imagesCacheFile = ".cache/jkit/images.yaml"

	// Max cache age before refreshing from remote.
	imagesCacheMaxAge = 24 * time.Hour
)

// ImageEntry represents a single Joomla Docker image entry from images.yaml.
type ImageEntry struct {
	Tag         string `yaml:"tag"`
	Description string `yaml:"description"`
}

// Built-in fallback images when no source is available.
var defaultImages = []ImageEntry{
	{Tag: "joomla:6.1-php8.4-apache", Description: "Joomla 6.1 with PHP 8.4 (recommended)"},
	{Tag: "joomla:5.3-php8.3-apache", Description: "Joomla 5.3 LTS with PHP 8.3"},
	{Tag: "joomla:5.3-php8.4-apache", Description: "Joomla 5.3 LTS with PHP 8.4"},
	{Tag: "joomla:6.1-php8.5-apache", Description: "Joomla 6.1 with PHP 8.5 (experimental)"},
}

// imagesFile holds the top-level YAML structure for images.yaml.
type imagesFile struct {
	Images []ImageEntry `yaml:"images"`
}

// InitConfig holds all user-provided configuration for initializing a project.
type InitConfig struct {
	ProjectName string
	JoomlaImage string
	Agents      []string
	Quickstart  string // path to .zip, empty if none
	Timezone    string
	Force       bool
}

// DefaultInitConfig returns InitConfig with sensible defaults.
// ProjectName is empty (must be provided), JoomlaImage and Timezone
// have defaults, Agents is nil (meaning none selected), Quickstart
// is empty, Force is false.
func DefaultInitConfig() InitConfig {
	return InitConfig{
		ProjectName: "",
		JoomlaImage: "joomla:6.1-php8.4-apache",
		Timezone:    "UTC",
		Agents:      nil,
		Quickstart:  "",
		Force:       false,
	}
}

// ToDevcontainerData maps InitConfig fields to DevcontainerData.
// Fields not set in InitConfig receive defaults from DefaultDevcontainerData().
func (c InitConfig) ToDevcontainerData() devcontainer.DevcontainerData {
	d := devcontainer.DefaultDevcontainerData()

	d.ProjectName = c.ProjectName
	if c.JoomlaImage != "" {
		d.JoomlaImage = c.JoomlaImage
	}
	if c.Timezone != "" {
		d.Timezone = c.Timezone
	}
	if c.Agents != nil {
		d.SelectedAgents = c.Agents
	}

	return d
}

// LoadImages loads Joomla images from the following cascade (DD-03):
//
//  1. Remote fetch from imagesRemoteURL, cached to ~/.cache/jkit/images.yaml
//  2. Local cache (from previous fetch)
//  3. Built-in defaultImages as last resort
//
// The binary ships standalone — no images.yaml alongside it.
func LoadImages() ([]ImageEntry, error) {
	// 1. Try remote fetch + cache
	if images, ok := fetchRemoteImages(); ok {
		return images, nil
	}

	// 2. Try local cache from ~/.cache/jkit/
	cachePath := filepath.Join(userHomeDir(), imagesCacheFile)
	if images, ok := loadFromPath(cachePath); ok {
		return images, nil
	}

	// 3. Built-in defaults
	return defaultImages, nil
}

// loadFromPath reads and parses images from the given path.
func loadFromPath(path string) ([]ImageEntry, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	images, err := parseImagesYAML(data)
	if err != nil {
		return nil, false
	}
	return images, true
}

// fetchRemoteImages fetches images.yaml from the remote URL and caches it
// to ~/.cache/jkit/images.yaml.
func fetchRemoteImages() ([]ImageEntry, bool) {
	resp, err := httpGet(imagesRemoteURL)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false
	}

	images, err := parseImagesYAML(data)
	if err != nil {
		return nil, false
	}

	// Cache to ~/.cache/jkit/
	cachePath := filepath.Join(userHomeDir(), imagesCacheFile)
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err == nil {
		_ = os.WriteFile(cachePath, data, 0644)
	}

	return images, true
}


// parseImagesYAML parses images.yaml bytes into ImageEntry slice.
func parseImagesYAML(data []byte) ([]ImageEntry, error) {
	var f imagesFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing images.yaml: %w", err)
	}
	if len(f.Images) == 0 {
		return nil, fmt.Errorf("images.yaml: empty images list")
	}
	return f.Images, nil
}

// userHomeDir returns the user's home directory, or "." if unavailable.
func userHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}

var httpGet = func(url string) (*http.Response, error) {
	return http.Get(url)
}

// detectQuickstart looks for .zip files in the given directory.
// Returns the path to the single .zip file found, or an error if
// none or multiple are found.
func detectQuickstart(dir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.zip"))
	if err != nil {
		return "", fmt.Errorf("searching for .zip files: %w", err)
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no quickstart .zip found in current directory")
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("multiple .zip files found (%d); specify one with --quickstart", len(matches))
	}
}

// extractQuickstart extracts a .zip archive into the target directory.
// It is a no-op if quickstartPath is empty.
func extractQuickstart(quickstartPath, targetDir string) error {
	if quickstartPath == "" {
		return nil
	}
	// Read and extract the zip
	data, err := os.ReadFile(quickstartPath)
	if err != nil {
		return fmt.Errorf("reading quickstart %s: %w", quickstartPath, err)
	}
	return extractZipBytes(data, targetDir)
}
