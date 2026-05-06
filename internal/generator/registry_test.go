package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReadRegistry_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	reg, err := ReadRegistry(context.Background(), dir)
	if err != nil {
		t.Fatalf("ReadRegistry() error = %v", err)
	}
	if reg == nil {
		t.Fatal("ReadRegistry() returned nil, expected empty Registry")
	}
	if len(reg.Extensions) != 0 {
		t.Errorf("Extensions = %d, want 0", len(reg.Extensions))
	}
}

func TestReadRegistry_ValidFile(t *testing.T) {
	dir := t.TempDir()
	yamlContent := []byte(`extensions:
    - name: com_blog
      type: component
      version: 1.0.0
    - name: mod_mymodule
      type: module
      version: 2.0.0
`)
	if err := os.WriteFile(filepath.Join(dir, "extensions.jkit.yaml"), yamlContent, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	reg, err := ReadRegistry(context.Background(), dir)
	if err != nil {
		t.Fatalf("ReadRegistry() error = %v", err)
	}
	if reg == nil {
		t.Fatal("ReadRegistry() returned nil")
	}
	if len(reg.Extensions) != 2 {
		t.Fatalf("Extensions = %d, want 2", len(reg.Extensions))
	}
	if reg.Extensions[0].Name != "com_blog" {
		t.Errorf("Extensions[0].Name = %q, want %q", reg.Extensions[0].Name, "com_blog")
	}
	if reg.Extensions[0].Type != "component" {
		t.Errorf("Extensions[0].Type = %q, want %q", reg.Extensions[0].Type, "component")
	}
	if reg.Extensions[0].Version != "1.0.0" {
		t.Errorf("Extensions[0].Version = %q, want %q", reg.Extensions[0].Version, "1.0.0")
	}
	if reg.Extensions[1].Name != "mod_mymodule" {
		t.Errorf("Extensions[1].Name = %q, want %q", reg.Extensions[1].Name, "mod_mymodule")
	}
}

func TestReadRegistry_EmptyYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "extensions.jkit.yaml"), []byte("extensions: []\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	reg, err := ReadRegistry(context.Background(), dir)
	if err != nil {
		t.Fatalf("ReadRegistry() error = %v", err)
	}
	if reg == nil {
		t.Fatal("ReadRegistry() returned nil")
	}
	if len(reg.Extensions) != 0 {
		t.Errorf("Extensions = %d, want 0", len(reg.Extensions))
	}
}

func TestWriteRegistry_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	entry := RegistryEntry{
		Name:    "com_blog",
		Type:    "component",
		Version: "1.0.0",
	}

	if err := WriteRegistry(context.Background(), dir, entry); err != nil {
		t.Fatalf("WriteRegistry() error = %v", err)
	}

	// Verify file was created
	regPath := filepath.Join(dir, "extensions.jkit.yaml")
	if _, err := os.Stat(regPath); os.IsNotExist(err) {
		t.Fatal("extensions.jkit.yaml was not created")
	}

	// Verify content
	reg, err := ReadRegistry(context.Background(), dir)
	if err != nil {
		t.Fatalf("ReadRegistry() error = %v", err)
	}
	if len(reg.Extensions) != 1 {
		t.Fatalf("Extensions = %d, want 1", len(reg.Extensions))
	}
	if reg.Extensions[0].Name != "com_blog" {
		t.Errorf("Name = %q, want %q", reg.Extensions[0].Name, "com_blog")
	}
	if reg.Extensions[0].Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", reg.Extensions[0].Version, "1.0.0")
	}
}

func TestWriteRegistry_AppendsToExisting(t *testing.T) {
	dir := t.TempDir()

	// Write first entry
	entry1 := RegistryEntry{Name: "com_blog", Type: "component", Version: "1.0.0"}
	if err := WriteRegistry(context.Background(), dir, entry1); err != nil {
		t.Fatalf("WriteRegistry(entry1) error = %v", err)
	}

	// Write second entry
	entry2 := RegistryEntry{Name: "mod_mymodule", Type: "module", Version: "2.0.0"}
	if err := WriteRegistry(context.Background(), dir, entry2); err != nil {
		t.Fatalf("WriteRegistry(entry2) error = %v", err)
	}

	// Verify both entries exist
	reg, err := ReadRegistry(context.Background(), dir)
	if err != nil {
		t.Fatalf("ReadRegistry() error = %v", err)
	}
	if len(reg.Extensions) != 2 {
		t.Fatalf("Extensions = %d, want 2", len(reg.Extensions))
	}
	if reg.Extensions[0].Name != "com_blog" {
		t.Errorf("Extensions[0].Name = %q, want %q", reg.Extensions[0].Name, "com_blog")
	}
	if reg.Extensions[1].Name != "mod_mymodule" {
		t.Errorf("Extensions[1].Name = %q, want %q", reg.Extensions[1].Name, "mod_mymodule")
	}
}

func TestWriteRegistry_UpdatesExistingEntry(t *testing.T) {
	dir := t.TempDir()

	// Write initial entry
	entry1 := RegistryEntry{Name: "com_blog", Type: "component", Version: "1.0.0"}
	if err := WriteRegistry(context.Background(), dir, entry1); err != nil {
		t.Fatalf("WriteRegistry(entry1) error = %v", err)
	}

	// Update the same entry with new version
	entryUpdated := RegistryEntry{Name: "com_blog", Type: "component", Version: "2.0.0"}
	if err := WriteRegistry(context.Background(), dir, entryUpdated); err != nil {
		t.Fatalf("WriteRegistry(entryUpdated) error = %v", err)
	}

	// Verify only one entry with updated version
	reg, err := ReadRegistry(context.Background(), dir)
	if err != nil {
		t.Fatalf("ReadRegistry() error = %v", err)
	}
	if len(reg.Extensions) != 1 {
		t.Fatalf("Extensions = %d, want 1", len(reg.Extensions))
	}
	if reg.Extensions[0].Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", reg.Extensions[0].Version, "2.0.0")
	}
}

func TestWriteRegistry_AtomicWrite(t *testing.T) {
	dir := t.TempDir()

	// Write initial entry
	entry := RegistryEntry{Name: "com_blog", Type: "component", Version: "1.0.0"}
	if err := WriteRegistry(context.Background(), dir, entry); err != nil {
		t.Fatalf("WriteRegistry() error = %v", err)
	}

	// Verify the file is the actual file, not a .tmp
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temporary file left behind: %s", e.Name())
		}
	}

	// Verify final file is valid YAML
	regPath := filepath.Join(dir, "extensions.jkit.yaml")
	data, err := os.ReadFile(regPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("extensions.jkit.yaml is empty")
	}
}

func TestReadRegistry_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "extensions.jkit.yaml"), []byte("invalid: yaml: [bad"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := ReadRegistry(context.Background(), dir)
	if err == nil {
		t.Fatal("ReadRegistry() expected error for invalid YAML, got nil")
	}
}

func TestRegistry_Find_Found(t *testing.T) {
	reg := &Registry{
		Extensions: []RegistryEntry{
			{Name: "com_blog", Type: "component", Version: "1.0.0"},
			{Name: "mod_mymodule", Type: "module", Version: "2.0.0"},
		},
	}

	entry := reg.Find("com_blog")
	if entry == nil {
		t.Fatal("Find() returned nil for existing entry")
	}
	if entry.Name != "com_blog" {
		t.Errorf("Find() name = %q, want %q", entry.Name, "com_blog")
	}
	if entry.Type != "component" {
		t.Errorf("Find() type = %q, want %q", entry.Type, "component")
	}
}

func TestRegistry_Find_NotFound(t *testing.T) {
	reg := &Registry{
		Extensions: []RegistryEntry{
			{Name: "com_blog", Type: "component"},
		},
	}

	entry := reg.Find("nonexistent")
	if entry != nil {
		t.Errorf("Find() = %v, want nil for nonexistent entry", entry)
	}
}

func TestRegistry_Find_EmptyRegistry(t *testing.T) {
	reg := &Registry{}
	entry := reg.Find("anything")
	if entry != nil {
		t.Errorf("Find() = %v, want nil for empty registry", entry)
	}
}

func TestWriteRegistry_MultipleDifferentTypes(t *testing.T) {
	dir := t.TempDir()

	entries := []RegistryEntry{
		{Name: "com_blog", Type: "component", Version: "1.0.0"},
		{Name: "mod_mymodule", Type: "module", Version: "1.0.0"},
		{Name: "plg_user_auth", Type: "plugin", Version: "1.0.0"},
		{Name: "tpl_cassiopeia", Type: "template", Version: "1.0.0"},
		{Name: "lib_foom", Type: "library", Version: "1.0.0"},
		{Name: "pkg_all", Type: "package", Version: "1.0.0"},
	}

	for _, e := range entries {
		if err := WriteRegistry(context.Background(), dir, e); err != nil {
			t.Fatalf("WriteRegistry(%s) error = %v", e.Name, err)
		}
	}

	reg, err := ReadRegistry(context.Background(), dir)
	if err != nil {
		t.Fatalf("ReadRegistry() error = %v", err)
	}
	if len(reg.Extensions) != 6 {
		t.Fatalf("Extensions = %d, want 6", len(reg.Extensions))
	}
}
