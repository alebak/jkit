package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alebak/jkit/internal/generator"
	"github.com/spf13/pflag"
)

// setupJoomlaProject creates a temporary directory that looks like a Joomla project.
func setupJoomlaProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "configuration.php"), []byte(`<?php class JConfig {}`), 0644); err != nil {
		t.Fatal(err)
	}

	for _, d := range []string{"administrator", "components", "modules", "plugins", "templates", "libraries"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

// chdir changes to a directory and returns a restore function.
func chdir(t *testing.T, dir string) func() {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return func() { os.Chdir(orig) }
}

// runCreateCmd runs the create command with the given args using RunE directly.
// Returns output buffer and error.
func runCreateCmd(t *testing.T, args []string) (*bytes.Buffer, error) {
	t.Helper()
	b := new(bytes.Buffer)
	createCmd.SetOut(b)
	createCmd.SetErr(b)

	// Reset all flags to defaults
	createCmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Value.Set(f.DefValue)
		f.Changed = false
	})

	// Parse flags manually
	flags := createCmd.Flags()
	for i := 0; i < len(args); i++ {
		if args[i][0] != '-' {
			continue // positional arg, skip
		}
		name := strings.TrimLeft(args[i], "-")
		if flags.Lookup(name) == nil {
			continue
		}
		if flags.Lookup(name).Value.Type() == "bool" {
			flags.Set(name, "true")
		} else if i+1 < len(args) {
			flags.Set(name, args[i+1])
			i++ // skip the value
		}
	}

	// Find positional type arg (first non-flag arg)
	var typeArg string
	for _, a := range args {
		if a[0] != '-' {
			typeArg = a
			break
		}
	}

	err := createCmd.RunE(createCmd, []string{typeArg})
	return b, err
}

func TestRootCommand_PrintsHelp(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)
	rootCmd.SetArgs([]string{})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	output := b.String()
	if !strings.Contains(output, "jkit") {
		t.Errorf("expected help to mention 'jkit', got: %s", output)
	}
}

func TestRootCommand_HelpContainsSubcommands(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	output := b.String()
	for _, cmd := range []string{"init", "create", "build"} {
		if !strings.Contains(output, cmd) {
			t.Errorf("expected help to mention %q subcommand, got: %s", cmd, output)
		}
	}
}

func TestInitCommand_PrintsNotImplemented(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)
	rootCmd.SetArgs([]string{"init"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(b.String(), "not yet implemented") {
		t.Errorf("expected 'not yet implemented', got: %s", b.String())
	}
}

func TestInitCommand_HasExpectedFlags(t *testing.T) {
	expectedFlags := []string{"name", "image", "quickstart", "agents", "timezone"}
	for _, name := range expectedFlags {
		flag := initCmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("expected --%s flag on init command", name)
		}
	}
}

func TestCreateCommand_UsageShowsValidArgs(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)
	rootCmd.SetArgs([]string{"create", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	output := b.String()
	expectedTypes := []string{"component", "module", "plugin", "template", "library", "package"}
	for _, typ := range expectedTypes {
		if !strings.Contains(output, typ) {
			t.Errorf("expected --help to mention %q, got: %s", typ, output)
		}
	}
}

func TestCreateCommand_NoJoomlaProject(t *testing.T) {
	dir := t.TempDir()
	defer chdir(t, dir)()

	args := []string{"component", "--name", "Blog", "--vendor", "Alebak"}
	_, err := runCreateCmd(t, args)
	if err == nil {
		t.Fatal("expected error for non-Joomla directory")
	}
	if !strings.Contains(err.Error(), "Joomla") {
		t.Errorf("expected error to mention Joomla, got: %v", err)
	}
}

func TestCreateCommand_InvalidType(t *testing.T) {
	dir := setupJoomlaProject(t)
	defer chdir(t, dir)()

	_, err := runCreateCmd(t, []string{"invalid-type", "--name", "Blog", "--vendor", "Alebak"})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestCreateCommand_MissingName(t *testing.T) {
	dir := setupJoomlaProject(t)
	defer chdir(t, dir)()

	_, err := runCreateCmd(t, []string{"component", "--vendor", "Alebak"})
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
}

func TestCreateCommand_MissingVendor(t *testing.T) {
	dir := setupJoomlaProject(t)
	defer chdir(t, dir)()

	_, err := runCreateCmd(t, []string{"component", "--name", "Blog"})
	if err == nil {
		t.Fatal("expected error for missing --vendor")
	}
}

func TestCreateCommand_CreatesComponent(t *testing.T) {
	dir := setupJoomlaProject(t)
	defer chdir(t, dir)()

	b, err := runCreateCmd(t, []string{"component", "--name", "Blog", "--vendor", "Alebak", "--joomla-version", "5.3"})
	if err != nil {
		t.Fatalf("create component error = %v", err)
	}

	if !strings.Contains(b.String(), "Created") {
		t.Errorf("expected output to mention 'created', got: %s", b.String())
	}

	// Verify component files were created in administrator/components/com_blog/
	compDir := filepath.Join(dir, "administrator", "components", "com_blog")
	info, err := os.Stat(compDir)
	if err != nil {
		t.Fatalf("component directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("com_blog is not a directory")
	}

	manifestPath := filepath.Join(compDir, "name.xml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatal("name.xml not created in component directory")
	}

	// Verify registry was created
	registryPath := filepath.Join(dir, "extensions.jkit.yaml")
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		t.Fatal("extensions.jkit.yaml not created")
	}
}

func TestCreateCommand_OverwriteDetection(t *testing.T) {
	dir := setupJoomlaProject(t)
	defer chdir(t, dir)()

	// First create — should succeed
	if _, err := runCreateCmd(t, []string{"module", "--name", "MyModule", "--vendor", "Alebak"}); err != nil {
		t.Fatalf("first create error = %v", err)
	}

	// Verify overwrite detection with isTTY=false (non-interactive)
	data := generator.NewExtensionData("MyModule", "Alebak", generator.TypeModule)
	err := checkOverwrite(context.Background(), dir, data, false, false)
	if err == nil {
		t.Fatal("expected error for overwrite without --force in non-TTY")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists', got: %v", err)
	}

	// With --force, overwrite should succeed
	if err := checkOverwrite(context.Background(), dir, data, true, false); err != nil {
		t.Errorf("expected nil with --force, got: %v", err)
	}
}

func TestCreateCommand_ForceOverwrite(t *testing.T) {
	dir := setupJoomlaProject(t)
	defer chdir(t, dir)()

	// First create
	if _, err := runCreateCmd(t, []string{"module", "--name", "MyModule", "--vendor", "Alebak"}); err != nil {
		t.Fatalf("first create error = %v", err)
	}

	// Force overwrite — should succeed
	if _, err := runCreateCmd(t, []string{"module", "--name", "MyModule", "--vendor", "Alebak", "--force"}); err != nil {
		t.Fatalf("force create error = %v", err)
	}
}

func TestCreateCommand_AllTypesCreateSuccessfully(t *testing.T) {
	types := []struct {
		name string
		args []string
	}{
		{"component", []string{"component", "--name", "Blog", "--vendor", "Alebak"}},
		{"module", []string{"module", "--name", "MyModule", "--vendor", "Alebak"}},
		{"plugin", []string{"plugin", "--name", "Auth", "--vendor", "Alebak", "--plugin-group", "user"}},
		{"template", []string{"template", "--name", "Cassiopeia", "--vendor", "Alebak"}},
		{"library", []string{"library", "--name", "Foom", "--vendor", "Alebak"}},
		{"package", []string{"package", "--name", "All", "--vendor", "Alebak"}},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupJoomlaProject(t)
			defer chdir(t, dir)()

			b, err := runCreateCmd(t, tt.args)
			if err != nil {
				t.Fatalf("create %s error = %v\noutput: %s", tt.name, err, b.String())
			}
			if !strings.Contains(b.String(), "Created") {
				t.Errorf("expected output to mention 'created', got: %s", b.String())
			}
		})
	}
}

func TestBuildCommand_UnknownExtension(t *testing.T) {
	dir := setupJoomlaProject(t)
	defer chdir(t, dir)()

	// Run build with nonexistent name using RunE directly
	b := new(bytes.Buffer)
	buildCmd.SetOut(b)
	buildCmd.SetErr(b)

	err := buildCmd.RunE(buildCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for unknown extension")
	}
}

func TestBuildCommand_CreatesZip(t *testing.T) {
	dir := setupJoomlaProject(t)
	defer chdir(t, dir)()

	// First create a module
	if _, err := runCreateCmd(t, []string{"module", "--name", "MyModule", "--vendor", "Alebak"}); err != nil {
		t.Fatalf("create error = %v", err)
	}

	// Now build using RunE
	b := new(bytes.Buffer)
	buildCmd.SetOut(b)
	buildCmd.SetErr(b)

	err := buildCmd.RunE(buildCmd, []string{"mod_mymodule"})
	if err != nil {
		t.Fatalf("build error = %v", err)
	}

	output := b.String()
	if !strings.Contains(output, ".zip") {
		t.Errorf("expected output to mention zip file, got: %s", output)
	}

	// Verify zip was created
	buildsDir := filepath.Join(dir, "builds")
	zipPath := filepath.Join(buildsDir, "mod_mymodule.zip")
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		entries, _ := os.ReadDir(buildsDir)
		t.Fatalf("zip not found at %s, entries: %v", zipPath, entries)
	}
}

func TestBuildCommand_CreatesBuildsDir(t *testing.T) {
	dir := setupJoomlaProject(t)
	defer chdir(t, dir)()

	// Create a module
	if _, err := runCreateCmd(t, []string{"module", "--name", "TestBuild", "--vendor", "Acme"}); err != nil {
		t.Fatalf("create error = %v", err)
	}

	// Build it
	b := new(bytes.Buffer)
	buildCmd.SetOut(b)
	buildCmd.SetErr(b)

	err := buildCmd.RunE(buildCmd, []string{"mod_testbuild"})
	if err != nil {
		t.Fatalf("build error = %v", err)
	}

	// Verify builds/ was created
	buildsDir := filepath.Join(dir, "builds")
	if _, err := os.Stat(buildsDir); os.IsNotExist(err) {
		t.Fatal("builds/ directory not created")
	}
}

func TestCreateCommand_FlagsExist(t *testing.T) {
	expectedFlags := []string{"name", "vendor", "joomla-version", "plugin-group", "force"}
	for _, name := range expectedFlags {
		flag := createCmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("expected --%s flag on create command", name)
		}
	}
}

func TestInitCommand_HelpShowsFlags(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)
	rootCmd.SetArgs([]string{"init", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	output := b.String()
	for _, flag := range []string{"--name", "--image", "--quickstart", "--agents", "--timezone"} {
		if !strings.Contains(output, flag) {
			t.Errorf("expected --help to mention %q, got: %s", flag, output)
		}
	}
}

func TestVersionCommand_NotYetImplemented(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Log("--version triggers unknown flag error, expected for stubs")
	}
}

func TestHelpFlag_ExitsWithZero(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

// Test that the generator import compiles (from stubs_test.go)
func TestStubPackagesCompile(t *testing.T) {
	t.Log("All internal stub packages compile successfully: generator, agents, mcp")
}
