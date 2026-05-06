package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectJoomlaProject_ConfigurationPhp(t *testing.T) {
	// Create a temp directory with configuration.php
	dir := t.TempDir()

	// Create a minimal configuration.php
	configContent := []byte(`<?php
class JConfig {
	public $db = "joomla";
}
`)
	if err := os.WriteFile(filepath.Join(dir, "configuration.php"), configContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Create some standard Joomla directories
	for _, d := range []string{"administrator", "components", "modules", "plugins", "templates", "libraries", "includes", "cli", "language", "media", "tmp", "cache", "logs"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	got, err := DetectJoomlaProject(context.Background(), dir)
	if err != nil {
		t.Fatalf("DetectJoomlaProject() error = %v", err)
	}
	if !got {
		t.Errorf("DetectJoomlaProject() = %v, want true", got)
	}
}

func TestDetectJoomlaProject_NotJoomla(t *testing.T) {
	// Create a temp directory with no Joomla files
	dir := t.TempDir()

	got, err := DetectJoomlaProject(context.Background(), dir)
	if err != nil {
		t.Fatalf("DetectJoomlaProject() error = %v", err)
	}
	if got {
		t.Errorf("DetectJoomlaProject() = %v, want false for empty dir", got)
	}
}

func TestDetectJoomlaProject_PartialDirectories(t *testing.T) {
	// configuration.php exists but only partial Joomla structure
	dir := t.TempDir()

	configContent := []byte(`<?php
class JConfig {
	public $db = "joomla";
}
`)
	if err := os.WriteFile(filepath.Join(dir, "configuration.php"), configContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Only create a few Joomla dirs (not all required)
	for _, d := range []string{"administrator", "components", "modules"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	got, err := DetectJoomlaProject(context.Background(), dir)
	if err != nil {
		t.Fatalf("DetectJoomlaProject() error = %v", err)
	}
	if !got {
		t.Errorf("DetectJoomlaProject() = %v, want true with partial dirs + configuration.php", got)
	}
}

func TestDetectJoomlaProject_NonExistentDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nonexistent")

	got, err := DetectJoomlaProject(context.Background(), dir)
	if err == nil {
		t.Error("DetectJoomlaProject() expected error for non-existent dir, got nil")
	}
	if got {
		t.Errorf("DetectJoomlaProject() = %v, want false for non-existent dir", got)
	}
}
