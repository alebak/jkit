// Package init provides the interactive and parameterized initialization flow
// for JKit projects. It collects user inputs (TUI or flags) and orchestrates
// devcontainer setup, agent deployment, extension generation, and MCP configuration.
package initpkg

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alebak/jkit/internal/devcontainer"
	"gopkg.in/yaml.v3"
)

// ImageEntry represents a single Joomla Docker image entry from images.yaml.
type ImageEntry struct {
	Tag         string `yaml:"tag"`
	Description string `yaml:"description"`
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

// ParseImagesYAML reads and parses images.yaml from the given filesystem.
// It expects a top-level key "images" containing an array of ImageEntry.
// Returns an error if the file cannot be read, the YAML is malformed,
// the "images" key is missing, or the list is empty.
func ParseImagesYAML(fsys fs.FS) ([]ImageEntry, error) {
	data, err := fs.ReadFile(fsys, "images.yaml")
	if err != nil {
		return nil, fmt.Errorf("reading images.yaml: %w", err)
	}

	var f imagesFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing images.yaml: %w", err)
	}

	if len(f.Images) == 0 {
		return nil, fmt.Errorf("images.yaml: empty images list")
	}

	return f.Images, nil
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
