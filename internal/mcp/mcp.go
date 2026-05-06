// Package mcp provides Model Context Protocol integration for AI tools.
package mcp

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MCPConfigPathFor returns the mcp.json config path for the given agent name.
// Environment variables override the default paths:
//
//	CLAUDE_MCP_CONFIG    — override for Claude Code
//	OPENCODE_MCP_CONFIG  — override for OpenCode
//	GEMINI_MCP_CONFIG    — override for Gemini CLI
func MCPConfigPathFor(ctx context.Context, agentName string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	switch agentName {
	case "claude":
		if p := os.Getenv("CLAUDE_MCP_CONFIG"); p != "" {
			return p, nil
		}
		return filepath.Join(home, ".claude", "mcp.json"), nil
	case "opencode":
		if p := os.Getenv("OPENCODE_MCP_CONFIG"); p != "" {
			return p, nil
		}
		return filepath.Join(home, ".config", "opencode", "mcp.json"), nil
	case "gemini":
		if p := os.Getenv("GEMINI_MCP_CONFIG"); p != "" {
			return p, nil
		}
		return filepath.Join(home, ".gemini", "mcp.json"), nil
	default:
		return "", fmt.Errorf("unknown agent: %s", agentName)
	}
}

// ListAvailable returns MCP server names discovered from .json files in
// the templates/mcp/ directory of the given filesystem. Names are
// the filenames without the .json extension, sorted alphabetically.
func ListAvailable(ctx context.Context, fsys fs.FS) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := fs.ReadDir(fsys, "templates/mcp")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading mcp directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), ".json"))
	}
	sort.Strings(names)
	return names, nil
}
