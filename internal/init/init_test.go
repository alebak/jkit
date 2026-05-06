package initpkg

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/alebak/jkit/internal/devcontainer"
)

func TestParseImagesYAML_Valid(t *testing.T) {
	fsys := fstest.MapFS{
		"images.yaml": &fstest.MapFile{
			Data: []byte(`images:
  - tag: joomla:6.1-php8.4-apache
    description: "Joomla 6.1 with PHP 8.4"
  - tag: joomla:5.3-php8.3-apache
    description: "Joomla 5.3 with PHP 8.3"
`),
		},
	}

	images, err := ParseImagesYAML(fsys)
	if err != nil {
		t.Fatalf("ParseImagesYAML() error = %v", err)
	}

	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}

	if images[0].Tag != "joomla:6.1-php8.4-apache" {
		t.Errorf("images[0].Tag = %q, want %q", images[0].Tag, "joomla:6.1-php8.4-apache")
	}
	if images[0].Description != "Joomla 6.1 with PHP 8.4" {
		t.Errorf("images[0].Description = %q, want %q", images[0].Description, "Joomla 6.1 with PHP 8.4")
	}
	if images[1].Tag != "joomla:5.3-php8.3-apache" {
		t.Errorf("images[1].Tag = %q, want %q", images[1].Tag, "joomla:5.3-php8.3-apache")
	}
}

func TestParseImagesYAML_EmptyList(t *testing.T) {
	fsys := fstest.MapFS{
		"images.yaml": &fstest.MapFile{
			Data: []byte(`images: []`),
		},
	}

	_, err := ParseImagesYAML(fsys)
	if err == nil {
		t.Fatal("expected error for empty images list, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected error to mention 'empty', got: %v", err)
	}
}

func TestParseImagesYAML_Malformed(t *testing.T) {
	fsys := fstest.MapFS{
		"images.yaml": &fstest.MapFile{
			Data: []byte(`images:
  - tag: "unclosed`),
		},
	}

	_, err := ParseImagesYAML(fsys)
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

func TestParseImagesYAML_MissingKey(t *testing.T) {
	fsys := fstest.MapFS{
		"images.yaml": &fstest.MapFile{
			Data: []byte(`other_key: []`),
		},
	}

	_, err := ParseImagesYAML(fsys)
	if err == nil {
		t.Fatal("expected error for missing 'images' key, got nil")
	}
}

func TestParseImagesYAML_FileNotFound(t *testing.T) {
	fsys := fstest.MapFS{}

	_, err := ParseImagesYAML(fsys)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestDefaultInitConfig(t *testing.T) {
	cfg := DefaultInitConfig()

	if cfg.ProjectName != "" {
		t.Errorf("ProjectName = %q, want empty", cfg.ProjectName)
	}
	if cfg.JoomlaImage != "joomla:6.1-php8.4-apache" {
		t.Errorf("JoomlaImage = %q, want %q", cfg.JoomlaImage, "joomla:6.1-php8.4-apache")
	}
	if cfg.Timezone != "UTC" {
		t.Errorf("Timezone = %q, want %q", cfg.Timezone, "UTC")
	}
	if cfg.Agents != nil {
		t.Errorf("Agents = %v, want nil", cfg.Agents)
	}
	if cfg.Quickstart != "" {
		t.Errorf("Quickstart = %q, want empty", cfg.Quickstart)
	}
	if cfg.Force {
		t.Errorf("Force = true, want false")
	}
}

func TestInitConfigToDevcontainerData(t *testing.T) {
	cfg := InitConfig{
		ProjectName: "myproject",
		JoomlaImage: "joomla:6.1-php8.4-apache",
		Timezone:    "America/New_York",
		Agents:      []string{"claude", "opencode"},
	}

	data := cfg.ToDevcontainerData()

	if data.ProjectName != "myproject" {
		t.Errorf("ProjectName = %q, want %q", data.ProjectName, "myproject")
	}
	if data.JoomlaImage != "joomla:6.1-php8.4-apache" {
		t.Errorf("JoomlaImage = %q, want %q", data.JoomlaImage, "joomla:6.1-php8.4-apache")
	}
	if data.Timezone != "America/New_York" {
		t.Errorf("Timezone = %q, want %q", data.Timezone, "America/New_York")
	}
	if len(data.SelectedAgents) != 2 || data.SelectedAgents[0] != "claude" {
		t.Errorf("SelectedAgents = %v, want [claude opencode]", data.SelectedAgents)
	}
}

func TestInitConfigToDevcontainerData_DefaultsPreserved(t *testing.T) {
	cfg := InitConfig{
		ProjectName: "test",
	}

	data := cfg.ToDevcontainerData()
	defaults := devcontainer.DefaultDevcontainerData()

	if data.DBUser != defaults.DBUser {
		t.Errorf("DBUser = %q, want default %q", data.DBUser, defaults.DBUser)
	}
	if data.AdminUser != defaults.AdminUser {
		t.Errorf("AdminUser = %q, want default %q", data.AdminUser, defaults.AdminUser)
	}
	if data.AdminPassword != defaults.AdminPassword {
		t.Errorf("AdminPassword = %q, want default %q", data.AdminPassword, defaults.AdminPassword)
	}
	if data.SiteName != defaults.SiteName {
		t.Errorf("SiteName = %q, want default %q", data.SiteName, defaults.SiteName)
	}
}

func TestDetectQuickstart_SingleZip(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "quickstart.zip"), []byte("fake zip"), 0644); err != nil {
		t.Fatal(err)
	}

	path, err := detectQuickstart(dir)
	if err != nil {
		t.Fatalf("detectQuickstart() error = %v", err)
	}
	if !strings.HasSuffix(path, "quickstart.zip") {
		t.Errorf("path = %q, want ...quickstart.zip", path)
	}
}

func TestDetectQuickstart_NoZip(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := detectQuickstart(dir)
	if err == nil {
		t.Fatal("expected error for no zip files, got nil")
	}
	if !strings.Contains(err.Error(), "no quickstart") {
		t.Errorf("expected 'no quickstart' in error, got: %v", err)
	}
}

func TestDetectQuickstart_MultipleZips(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.zip", "b.zip"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("fake"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	_, err := detectQuickstart(dir)
	if err == nil {
		t.Fatal("expected error for multiple zip files, got nil")
	}
	if !strings.Contains(err.Error(), "multiple") && !strings.Contains(err.Error(), "more than one") {
		t.Errorf("expected 'multiple' or 'more than one' in error, got: %v", err)
	}
}

func TestOrchestrateOverwriteCheck_WithoutForce(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".devcontainer"), 0755); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultInitConfig()
	cfg.ProjectName = "test"
	cfg.Force = false

	err = Orchestrate(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for existing .devcontainer/ without Force, got nil")
	}
	if !strings.Contains(err.Error(), "force") {
		t.Errorf("expected error to mention '--force', got: %v", err)
	}
}

func TestOrchestrateOverwriteCheck_WithForce(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".devcontainer"), 0755); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultInitConfig()
	cfg.ProjectName = "test"
	cfg.Force = true
	cfg.Agents = nil // no agents to avoid side effects on local machine

	// The pipeline should succeed (embedded templates work from any CWD)
	// or fail on some other step — the key is that it passes the overwrite guard
	// and the error (if any) is NOT about --force.
	err = Orchestrate(context.Background(), cfg)
	if err != nil && strings.Contains(err.Error(), "force") {
		t.Errorf("with Force=true, error should not be about --force, got: %v", err)
	}
	// Note: if err == nil, the pipeline completed successfully,
	// which is also valid — just means all steps worked.
}

// TestOrchestrateFailFast uses context cancellation to verify that
// when DEVC fails due to cancellation, subsequent steps are not run.
func TestOrchestrateFailFast_CancelledContext(t *testing.T) {
	dir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultInitConfig()
	cfg.ProjectName = "test"
	cfg.Agents = []string{"claude"}

	// Cancel the context so Orchestrate returns immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = Orchestrate(ctx, cfg)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	// The error should mention context cancellation
	if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "canceled") {
		t.Errorf("expected context-related error, got: %v", err)
	}

	// Verify no filesystem changes were made
	if _, statErr := os.Stat(filepath.Join(dir, ".devcontainer")); statErr == nil {
		t.Error("expected .devcontainer/ to NOT exist (cancelled before DEVC)")
	}
	if _, statErr := os.Stat(filepath.Join(dir, ".jkit")); statErr == nil {
		t.Error("expected .jkit/ to NOT exist (AGNT should not have run)")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "builds")); statErr == nil {
		t.Error("expected builds/ to NOT exist (EXTG should not have run)")
	}
}

// Test quickstart extraction by providing a real .zip file.
func TestDetectQuickstart_Integration(t *testing.T) {
	dir := t.TempDir()

	// Create a valid .zip file
	zipPath := filepath.Join(dir, "starter.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	// Write minimal zip header (PK\x03\x04) + end record so it's a valid zip
	f.Write([]byte("PK\x03\x04\x14\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"))
	f.Close()

	path, err := detectQuickstart(dir)
	if err != nil {
		t.Fatalf("detectQuickstart() error = %v", err)
	}
	if path != zipPath {
		t.Errorf("path = %q, want %q", path, zipPath)
	}
}

// Triangulation: appendGitignore creates new file with .env entry
func TestAppendGitignore_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")

	// Clean up any leftovers from other tests
	if err := appendGitignore(path); err != nil {
		t.Fatalf("appendGitignore() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != ".env\n" {
		t.Errorf("expected '.env\\n', got %q", string(data))
	}
}

// Triangulation: appendGitignore does not duplicate .env
func TestAppendGitignore_ExistingDotEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")

	// Create with .env already present
	if err := os.WriteFile(path, []byte(".env\nnode_modules/\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := appendGitignore(path); err != nil {
		t.Fatalf("appendGitignore() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != ".env\nnode_modules/\n" {
		t.Errorf("file should not be modified, got %q", string(data))
	}
}

// Triangulation: appendGitignore appends to existing file
func TestAppendGitignore_Appends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")

	if err := os.WriteFile(path, []byte("node_modules/\nbuilds/\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := appendGitignore(path); err != nil {
		t.Fatalf("appendGitignore() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), ".env") {
		t.Errorf("expected .env to be appended, got %q", string(data))
	}
}

// Test that Orchestrate with empty agents skips AGNT and MCPS steps.
func TestOrchestrateNoAgents(t *testing.T) {
	dir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultInitConfig()
	cfg.ProjectName = "test"
	cfg.Agents = []string{} // explicit empty

	err = Orchestrate(context.Background(), cfg)
	if err != nil {
		// If error, it should not be agent-related
		if strings.Contains(err.Error(), "agent") || strings.Contains(err.Error(), "mcp") {
			t.Errorf("with empty agents, error should not be agent/MCP-related, got: %v", err)
		}
	} else {
		// Pipeline succeeded — verify DEVC files were created (proving DEVC ran)
		devcDir := filepath.Join(dir, ".devcontainer")
		if _, statErr := os.Stat(devcDir); statErr != nil {
			t.Errorf("expected .devcontainer/ to exist, got: %v", statErr)
		}
		// Verify AGNT/MCPS dirs were NOT created
		if _, statErr := os.Stat(filepath.Join(dir, ".jkit")); statErr == nil {
			t.Error("expected .jkit/ to NOT exist (AGNT should be skipped)")
		}
	}
}

// Additional triangulation: edge cases for ParseImagesYAML
func TestParseImagesYAML_NonYAMLContent(t *testing.T) {
	fsys := fstest.MapFS{
		"images.yaml": &fstest.MapFile{
			Data: []byte(`this is not yaml: {{{`),
		},
	}

	_, err := ParseImagesYAML(fsys)
	if err == nil {
		t.Fatal("expected error for non-YAML content, got nil")
	}
}

func TestParseImagesYAML_EmptyFile(t *testing.T) {
	fsys := fstest.MapFS{
		"images.yaml": &fstest.MapFile{
			Data: []byte(``),
		},
	}

	_, err := ParseImagesYAML(fsys)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
}

// Triangulation: verify ToDevcontainerData preserves zero-value fields from defaults
func TestInitConfigToDevcontainerData_EmptyMigratesFromDefaults(t *testing.T) {
	cfg := InitConfig{
		ProjectName: "test",
		JoomlaImage: "",
		Timezone:    "",
	}

	data := cfg.ToDevcontainerData()
	defaults := devcontainer.DefaultDevcontainerData()

	// Empty JoomlaImage should use default
	if data.JoomlaImage != defaults.JoomlaImage {
		t.Errorf("JoomlaImage = %q, want default %q", data.JoomlaImage, defaults.JoomlaImage)
	}
	// Empty Timezone should use default
	if data.Timezone != defaults.Timezone {
		t.Errorf("Timezone = %q, want default %q", data.Timezone, defaults.Timezone)
	}
}

// Triangulation: detectQuickstart with non-zip files only
func TestDetectQuickstart_NonZipFilesOnly(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"readme.txt", "file.tar.gz", "notes.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	_, err := detectQuickstart(dir)
	if err == nil {
		t.Fatal("expected error for no .zip files, got nil")
	}
}

// Triangulation: overwrite guard with various force values
func TestInitConfigForceGuard(t *testing.T) {
	t.Run("nonexistent dir returns nil", func(t *testing.T) {
		dir := t.TempDir()
		err := checkOverwriteGuard(filepath.Join(dir, "nonexistent"), false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("existing dir without force", func(t *testing.T) {
		dir := t.TempDir()
		devcDir := filepath.Join(dir, ".devcontainer")
		if err := os.MkdirAll(devcDir, 0755); err != nil {
			t.Fatal(err)
		}
		err := checkOverwriteGuard(devcDir, false)
		if err == nil {
			t.Fatal("expected error for existing dir without force")
		}
	})

	t.Run("existing dir with force", func(t *testing.T) {
		dir := t.TempDir()
		devcDir := filepath.Join(dir, ".devcontainer")
		if err := os.MkdirAll(devcDir, 0755); err != nil {
			t.Fatal(err)
		}
		err := checkOverwriteGuard(devcDir, true)
		if err != nil {
			t.Fatalf("expected nil with force, got: %v", err)
		}
	})
}
