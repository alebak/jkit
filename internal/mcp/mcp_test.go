package mcp

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/alebak/jkit"
)

func TestListAvailable(t *testing.T) {
	tests := []struct {
		name    string
		fsys    fs.FS
		want    []string
		wantErr bool
	}{
		{
			name: "all MCPs from .json files",
			fsys: fstest.MapFS{
				"templates/mcp/playwright.json": &fstest.MapFile{Data: []byte(`{"mcpServers":{}}`)},
				"templates/mcp/mariadb.json":    &fstest.MapFile{Data: []byte(`{"mcpServers":{}}`)},
			},
			want: []string{"mariadb", "playwright"},
		},
		{
			name: "non-json files ignored",
			fsys: fstest.MapFS{
				"templates/mcp/playwright.json": &fstest.MapFile{Data: []byte(`{"mcpServers":{}}`)},
				"templates/mcp/readme.txt":      &fstest.MapFile{Data: []byte("readme")},
			},
			want: []string{"playwright"},
		},
		{
			name: "subdirectories ignored",
			fsys: fstest.MapFS{
				"templates/mcp/playwright.json": &fstest.MapFile{Data: []byte(`{"mcpServers":{}}`)},
				"templates/mcp/backup/old.json": &fstest.MapFile{Data: []byte(`{"mcpServers":{}}`)},
			},
			want: []string{"playwright"},
		},
		{
			name: "empty filesystem",
			fsys: fstest.MapFS{},
			want: nil,
		},
		{
			name: "only txt files",
			fsys: fstest.MapFS{
				"templates/mcp/list.txt": &fstest.MapFile{Data: []byte("list")},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListAvailable(context.Background(), tt.fsys)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAvailable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// restoreEnv saves and restores environment variables for test isolation.
func restoreEnv(keys ...string) func() {
	orig := make(map[string]string, len(keys))
	for _, k := range keys {
		orig[k] = os.Getenv(k)
	}
	return func() {
		for k, v := range orig {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}
}

func TestMCPConfigPathFor(t *testing.T) {
	t.Run("default paths", func(t *testing.T) {
		defer restoreEnv("CLAUDE_MCP_CONFIG", "OPENCODE_MCP_CONFIG", "GEMINI_MCP_CONFIG")()
		os.Unsetenv("CLAUDE_MCP_CONFIG")
		os.Unsetenv("OPENCODE_MCP_CONFIG")
		os.Unsetenv("GEMINI_MCP_CONFIG")

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			agent string
			want  string
		}{
			{"claude", filepath.Join(home, ".claude", "mcp.json")},
			{"opencode", filepath.Join(home, ".config", "opencode", "mcp.json")},
			{"gemini", filepath.Join(home, ".gemini", "mcp.json")},
		}

		for _, tt := range tests {
			t.Run(tt.agent, func(t *testing.T) {
				got, err := MCPConfigPathFor(context.Background(), tt.agent)
				if err != nil {
					t.Fatalf("MCPConfigPathFor() error = %v", err)
				}
				if got != tt.want {
					t.Errorf("MCPConfigPathFor() = %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("env var overrides", func(t *testing.T) {
		defer restoreEnv("CLAUDE_MCP_CONFIG", "OPENCODE_MCP_CONFIG", "GEMINI_MCP_CONFIG")()
		os.Setenv("CLAUDE_MCP_CONFIG", "/custom/claude/mcp.json")
		os.Setenv("OPENCODE_MCP_CONFIG", "/custom/opencode/mcp.json")
		os.Setenv("GEMINI_MCP_CONFIG", "/custom/gemini/mcp.json")

		tests := []struct {
			agent string
			want  string
		}{
			{"claude", "/custom/claude/mcp.json"},
			{"opencode", "/custom/opencode/mcp.json"},
			{"gemini", "/custom/gemini/mcp.json"},
		}

		for _, tt := range tests {
			t.Run(tt.agent, func(t *testing.T) {
				got, err := MCPConfigPathFor(context.Background(), tt.agent)
				if err != nil {
					t.Fatalf("MCPConfigPathFor() error = %v", err)
				}
				if got != tt.want {
					t.Errorf("MCPConfigPathFor() = %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("unknown agent", func(t *testing.T) {
		_, err := MCPConfigPathFor(context.Background(), "nonexistent")
		if err == nil {
			t.Fatal("expected error for unknown agent")
		}
		if !strings.Contains(err.Error(), "unknown agent") {
			t.Errorf("expected 'unknown agent', got: %v", err)
		}
	})
}

func TestDeployMCP(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "mcp.json")

	playwrightData := []byte(`{"mcpServers":{"playwright":{"command":"npx","args":["@playwright/mcp"]}}}`)

	err := DeployMCP(context.Background(), configPath, "playwright", playwrightData)
	if err != nil {
		t.Fatalf("DeployMCP() error = %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var cfg MCPServers
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}

	entry, ok := cfg.Servers["playwright"]
	if !ok {
		t.Fatal("expected playwright entry in config")
	}
	if entry.Command != "npx" {
		t.Errorf("command = %q, want %q", entry.Command, "npx")
	}
	if len(entry.Args) != 1 || entry.Args[0] != "@playwright/mcp" {
		t.Errorf("args = %v, want [@playwright/mcp]", entry.Args)
	}
}

func TestDeployMCP_CreatesBackup(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "mcp.json")

	initial := []byte(`{"mcpServers":{"existing":{"command":"echo","args":["hello"]}}}`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	playwrightData := []byte(`{"mcpServers":{"playwright":{"command":"npx","args":["@playwright/mcp"]}}}`)

	err := DeployMCP(context.Background(), configPath, "playwright", playwrightData)
	if err != nil {
		t.Fatalf("DeployMCP() error = %v", err)
	}

	// Verify backup
	backupPath := configPath + ".bak"
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("backup not found: %v", err)
	}
	if string(backupData) != string(initial) {
		t.Errorf("backup content = %q, want %q", string(backupData), string(initial))
	}

	// Verify merged content
	var cfg MCPServers
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Servers["existing"]; !ok {
		t.Error("expected existing entry to be preserved")
	}
	if _, ok := cfg.Servers["playwright"]; !ok {
		t.Error("expected playwright entry")
	}
}

func TestDeployMCP_Idempotent(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "mcp.json")

	playwrightData := []byte(`{"mcpServers":{"playwright":{"command":"npx","args":["@playwright/mcp"]}}}`)

	if err := DeployMCP(context.Background(), configPath, "playwright", playwrightData); err != nil {
		t.Fatal(err)
	}
	if err := DeployMCP(context.Background(), configPath, "playwright", playwrightData); err != nil {
		t.Fatal(err)
	}

	var cfg MCPServers
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}

	count := 0
	for name := range cfg.Servers {
		if name == "playwright" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 playwright entry, got %d", count)
	}
}

func TestRemoveMCP(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "mcp.json")

	playwrightData := []byte(`{"mcpServers":{"playwright":{"command":"npx","args":["@playwright/mcp"]}}}`)
	if err := DeployMCP(context.Background(), configPath, "playwright", playwrightData); err != nil {
		t.Fatal(err)
	}

	mariadbData := []byte(`{"mcpServers":{"mariadb":{"command":"npx","args":["@anthropic/mcp-server-mysql"]}}}`)
	if err := DeployMCP(context.Background(), configPath, "mariadb", mariadbData); err != nil {
		t.Fatal(err)
	}

	if err := RemoveMCP(context.Background(), configPath, "playwright"); err != nil {
		t.Fatalf("RemoveMCP() error = %v", err)
	}

	var cfg MCPServers
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}

	if _, ok := cfg.Servers["playwright"]; ok {
		t.Error("expected playwright entry to be removed")
	}
	if _, ok := cfg.Servers["mariadb"]; !ok {
		t.Error("expected mariadb entry to be preserved")
	}
}

func TestRemoveMCP_MissingName(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "mcp.json")

	initial := []byte(`{"mcpServers":{"existing":{"command":"echo","args":["hello"]}}}`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveMCP(context.Background(), configPath, "nonexistent"); err != nil {
		t.Fatalf("RemoveMCP() error = %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(initial) {
		t.Errorf("config was modified: got %q, want %q", string(data), string(initial))
	}
}

func TestRemoveMCP_CreatesBackup(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "mcp.json")

	playwrightData := []byte(`{"mcpServers":{"playwright":{"command":"npx","args":["@playwright/mcp"]}}}`)
	if err := DeployMCP(context.Background(), configPath, "playwright", playwrightData); err != nil {
		t.Fatal(err)
	}

	if err := RemoveMCP(context.Background(), configPath, "playwright"); err != nil {
		t.Fatalf("RemoveMCP() error = %v", err)
	}

	backupPath := configPath + ".bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("expected backup to exist after remove")
	}
}

func TestEmbeddedTemplatesValid(t *testing.T) {
	entries, err := fs.ReadDir(jkit.McpFS, "templates/mcp")
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) == 0 {
		t.Fatal("no embedded MCP templates found")
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			data, err := jkit.McpFS.ReadFile("templates/mcp/" + entry.Name())
			if err != nil {
				t.Fatalf("reading %s: %v", entry.Name(), err)
			}
			var cfg MCPServers
			if err := json.Unmarshal(data, &cfg); err != nil {
				t.Fatalf("invalid JSON in %s: %v", entry.Name(), err)
			}
			if cfg.Servers == nil || len(cfg.Servers) == 0 {
				t.Errorf("%s: missing mcpServers key or empty", entry.Name())
			}
		})
	}
}
