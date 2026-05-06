package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MCPServers represents the top-level structure of an mcp.json config file.
type MCPServers struct {
	Servers map[string]MCPEntry `json:"mcpServers"`
}

// MCPEntry represents a single MCP server entry in mcp.json.
type MCPEntry struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// readMCPConfig reads and parses an mcp.json file.
// Returns an empty config if the file does not exist.
func readMCPConfig(ctx context.Context, path string) (*MCPServers, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &MCPServers{Servers: make(map[string]MCPEntry)}, nil
		}
		return nil, fmt.Errorf("reading mcp config: %w", err)
	}

	var cfg MCPServers
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing mcp config: %w", err)
	}
	if cfg.Servers == nil {
		cfg.Servers = make(map[string]MCPEntry)
	}
	return &cfg, nil
}

// writeMCPConfig marshals and writes an MCPServers config to the given path.
func writeMCPConfig(ctx context.Context, path string, cfg *MCPServers) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling mcp config: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing mcp config: %w", err)
	}
	return nil
}

// backupFile creates a .bak copy of the given file if it exists.
func backupFile(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	orig, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading original for backup: %w", err)
	}
	if err := os.WriteFile(path+".bak", orig, 0644); err != nil {
		return fmt.Errorf("writing backup: %w", err)
	}
	return nil
}

// DeployMCP adds or replaces an MCP server entry in the named config file.
// It reads the existing config, merges the new entry, creates a backup
// (path.bak), then writes the merged config.
func DeployMCP(ctx context.Context, configPath, name string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var incoming MCPServers
	if err := json.Unmarshal(data, &incoming); err != nil {
		return fmt.Errorf("parsing MCP template: %w", err)
	}
	entry, ok := incoming.Servers[name]
	if !ok {
		return fmt.Errorf("template does not contain entry for %q", name)
	}

	cfg, err := readMCPConfig(ctx, configPath)
	if err != nil {
		return err
	}

	if err := backupFile(ctx, configPath); err != nil {
		return err
	}

	cfg.Servers[name] = entry
	return writeMCPConfig(ctx, configPath, cfg)
}

// RemoveMCP removes an MCP server entry from the named config file by key.
// It is a no-op (no error) if the key does not exist in the config.
func RemoveMCP(ctx context.Context, configPath, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	cfg, err := readMCPConfig(ctx, configPath)
	if err != nil {
		return err
	}

	if _, exists := cfg.Servers[name]; !exists {
		return nil
	}

	if err := backupFile(ctx, configPath); err != nil {
		return err
	}

	delete(cfg.Servers, name)
	return writeMCPConfig(ctx, configPath, cfg)
}
