package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// RegistryFilename is the name of the YAML registry file.
	RegistryFilename = "extensions.jkit.yaml"
)

// RegistryEntry represents a single extension in the registry.
type RegistryEntry struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	Version   string `yaml:"version,omitempty"`
	CreatedAt string `yaml:"created_at,omitempty"`
	UpdatedAt string `yaml:"updated_at,omitempty"`
}

// Registry holds the list of registered extensions.
type Registry struct {
	Extensions []RegistryEntry `yaml:"extensions"`
}

// Find looks up a registry entry by name. Returns nil if not found.
func (r *Registry) Find(name string) *RegistryEntry {
	for i := range r.Extensions {
		if r.Extensions[i].Name == name {
			return &r.Extensions[i]
		}
	}
	return nil
}

// ReadRegistry reads the extension registry from the given project directory.
// Returns an empty Registry if the file does not exist.
func ReadRegistry(ctx context.Context, projectDir string) (*Registry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path := filepath.Join(projectDir, RegistryFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{}, nil
		}
		return nil, fmt.Errorf("read registry: %w", err)
	}

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}
	if reg.Extensions == nil {
		reg.Extensions = []RegistryEntry{}
	}
	return &reg, nil
}

// WriteRegistry writes an extension entry to the registry.
// If an entry with the same name and type already exists, it updates it.
// Uses atomic write (write to temp file, rename) to prevent corruption.
func WriteRegistry(ctx context.Context, projectDir string, entry RegistryEntry) error {
	reg, err := ReadRegistry(ctx, projectDir)
	if err != nil {
		return err
	}

	if reg.Extensions == nil {
		reg.Extensions = []RegistryEntry{}
	}

	// Update existing entry or append
	found := false
	for i, e := range reg.Extensions {
		if e.Name == entry.Name && e.Type == entry.Type {
			reg.Extensions[i] = entry
			found = true
			break
		}
	}
	if !found {
		reg.Extensions = append(reg.Extensions, entry)
	}

	data, err := yaml.Marshal(&reg)
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	path := filepath.Join(projectDir, RegistryFilename)
	tmpPath := path + ".tmp"

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("write temp registry: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath) // best-effort cleanup
		return fmt.Errorf("rename registry: %w", err)
	}

	return nil
}
