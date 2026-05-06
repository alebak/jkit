package agents

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"
)

const defaultSkillName = "prd-creator"

func TestListAvailable(t *testing.T) {
	tests := []struct {
		name    string
		fsys    fs.FS
		want    []string
		wantErr bool
	}{
		{
			name: "all agents from .sh files",
			fsys: fstest.MapFS{
				"templates/agents/claude.sh":   &fstest.MapFile{Data: []byte("echo claude")},
				"templates/agents/opencode.sh": &fstest.MapFile{Data: []byte("echo opencode")},
				"templates/agents/gemini.sh":   &fstest.MapFile{Data: []byte("echo gemini")},
			},
			want: []string{"claude", "gemini", "opencode"},
		},
		{
			name: "non-sh files ignored",
			fsys: fstest.MapFS{
				"templates/agents/claude.sh":  &fstest.MapFile{Data: []byte("echo claude")},
				"templates/agents/readme.txt": &fstest.MapFile{Data: []byte("readme")},
				"templates/agents/notes.md":   &fstest.MapFile{Data: []byte("notes")},
			},
			want: []string{"claude"},
		},
		{
			name: "subdirectories ignored",
			fsys: fstest.MapFS{
				"templates/agents/claude.sh":     &fstest.MapFile{Data: []byte("echo claude")},
				"templates/agents/backup/old.sh": &fstest.MapFile{Data: []byte("old")},
			},
			want: []string{"claude"},
		},
		{
			name: "empty filesystem",
			fsys: fstest.MapFS{},
			want: nil,
		},
		{
			name: "only .txt files",
			fsys: fstest.MapFS{
				"templates/agents/list.txt": &fstest.MapFile{Data: []byte("agents list")},
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

func TestParsePostCreateMarkers(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name: "all agents present",
			content: `#!/bin/bash
set -e

# gentle-ai install section
echo "gentle-ai..."

# --- Agent installations ---
# --- agent:claude ---
echo "install claude"
# --- end agent:claude ---
# --- agent:opencode ---
echo "install opencode"
# --- end agent:opencode ---
# --- agent:gemini ---
echo "install gemini"
# --- end agent:gemini ---
`,
			want: []string{"claude", "opencode", "gemini"},
		},
		{
			name: "some agents present",
			content: `#!/bin/bash
set -e

# --- Agent installations ---
# --- agent:claude ---
echo "install claude"
# --- end agent:claude ---
# --- agent:opencode ---
echo "install opencode"
# --- end agent:opencode ---
`,
			want: []string{"claude", "opencode"},
		},
		{
			name: "no markers present",
			content: `#!/bin/bash
set -e
echo "hello"
`,
			want: nil,
		},
		{
			name:    "empty file",
			content: "",
			want:    nil,
		},
		{
			name: "marker-like but not exact",
			content: `# agent:claude
# --- agent:claude --
# --- agent:claude --
# -- agent:claude ---
`,
			want: nil,
		},
		{
			name: "duplicate markers deduplicated",
			content: `# --- agent:claude ---
echo "install"
# --- end agent:claude ---
# --- agent:claude ---
echo "install again"
# --- end agent:claude ---
`,
			want: []string{"claude"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "post-create.sh")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			got, err := ParsePostCreateMarkers(context.Background(), path)
			if err != nil {
				t.Fatalf("ParsePostCreateMarkers() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePostCreateMarkers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePostCreateMarkers_FileNotFound(t *testing.T) {
	_, err := ParsePostCreateMarkers(context.Background(), "/nonexistent/path/post-create.sh")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestDeploySkill(t *testing.T) {
	projectDir := t.TempDir()
	skillDir := filepath.Join(t.TempDir(), "claude", "skills")

	// Override agent dirs for test isolation
	// oldDirs removed: AgentSkillDirs
	// AgentSkillDirs removed; pass skillDir directly
	// cleanup: AgentSkillDirs removed

	err := DeploySkill(context.Background(), projectDir, skillDir, defaultSkillName)
	if err != nil {
		t.Fatalf("DeploySkill() error = %v", err)
	}

	// Verify SKILL.md was copied
	skillPath := filepath.Join(projectDir, ".jkit", "agents", "skills", defaultSkillName, "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Errorf("SKILL.md not created at %s", skillPath)
	}

	// Verify symlink was created
	symlinkPath := filepath.Join(skillDir, defaultSkillName)
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("symlink not created: %v", err)
	}

	// Resolve relative symlink target to absolute
	absTarget := filepath.Clean(filepath.Join(skillDir, target))
	expectedTarget := filepath.Clean(filepath.Join(projectDir, ".jkit", "agents", "skills", defaultSkillName))
	if absTarget != expectedTarget {
		t.Errorf("symlink resolves to %s, want %s", absTarget, expectedTarget)
	}
}

func TestDeploySkill_ExistingSymlink(t *testing.T) {
	projectDir := t.TempDir()
	skillDir := filepath.Join(t.TempDir(), "claude", "skills")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a symlink that points elsewhere first
	existingLink := filepath.Join(skillDir, defaultSkillName)
	if err := os.Symlink("/tmp/nonexistent", existingLink); err != nil {
		t.Fatal(err)
	}

	// oldDirs removed: AgentSkillDirs
	// AgentSkillDirs removed; pass skillDir directly
	// cleanup: AgentSkillDirs removed

	err := DeploySkill(context.Background(), projectDir, skillDir, defaultSkillName)
	if err != nil {
		t.Fatalf("DeploySkill() error = %v", err)
	}

	// Verify symlink now points to the jkit skill dir
	target, err := os.Readlink(existingLink)
	if err != nil {
		t.Fatalf("symlink not found: %v", err)
	}
	absTarget := filepath.Clean(filepath.Join(skillDir, target))
	expectedTarget := filepath.Clean(filepath.Join(projectDir, ".jkit", "agents", "skills", defaultSkillName))
	if absTarget != expectedTarget {
		t.Errorf("symlink resolves to %s, want %s", absTarget, expectedTarget)
	}
}

func TestDeploySkill_UnknownAgent(t *testing.T) {
	// SkillDirFor is the agent validation point — DeploySkill accepts any skillDir.
	if _, err := SkillDirFor(context.Background(), "nonexistent-agent"); err == nil {
		t.Error("expected error for unknown agent")
	}

	// DeploySkill with a non-existent ancestor dir should still error.
	err := DeploySkill(context.Background(), t.TempDir(), "/nonexistent/path/skills", defaultSkillName)
	if err == nil {
		t.Error("expected error for deploy to non-existent skill directory")
	}
}

func TestWriteAgentMarkers(t *testing.T) {
	tests := []struct {
		name   string
		agents []string
		want   string
	}{
		{
			name:   "with agents",
			agents: []string{"claude", "opencode"},
			want:   "\n# --- Agent installations ---\n",
		},
		{
			name:   "empty list",
			agents: []string{},
			want:   "",
		},
		{
			name:   "nil",
			agents: nil,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf testBuffer
			err := WriteAgentMarkers(context.Background(), &buf, tt.agents)
			if err != nil {
				t.Fatalf("WriteAgentMarkers() error = %v", err)
			}
			if buf.String() != tt.want {
				t.Errorf("WriteAgentMarkers() = %q, want %q", buf.String(), tt.want)
			}
		})
	}
}

// testBuffer is a simple io.Writer that captures writes to a string.
type testBuffer struct {
	data []byte
}

func (b *testBuffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *testBuffer) String() string {
	return string(b.data)
}
