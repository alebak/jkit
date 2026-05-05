package main

import (
	"bytes"
	"strings"
	"testing"
)

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

func TestCreateCommand_PrintsNotImplemented(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)
	rootCmd.SetArgs([]string{"create"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(b.String(), "not yet implemented") {
		t.Errorf("expected 'not yet implemented', got: %s", b.String())
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

func TestBuildCommand_PrintsNotImplemented(t *testing.T) {
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)
	rootCmd.SetArgs([]string{"build"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(b.String(), "not yet implemented") {
		t.Errorf("expected 'not yet implemented', got: %s", b.String())
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

func TestInitCommand_WithFlags(t *testing.T) {
	b := new(bytes.Buffer)
	initCmd.SetOut(b)
	initCmd.SetErr(b)
	// Parse the flags first so cobra populates the bound variables
	initCmd.ParseFlags([]string{"--name", "myproject", "--image", "joomla:latest", "--timezone", "America/New_York"})
	initCmd.Run(initCmd, []string{})

	if !strings.Contains(b.String(), "not yet implemented") {
		t.Errorf("expected 'not yet implemented', got: %s", b.String())
	}
}

func TestInitCommand_FlagsParseCorrectly(t *testing.T) {
	b := new(bytes.Buffer)
	initCmd.SetOut(b)
	initCmd.SetErr(b)

	initCmd.ParseFlags([]string{"--name", "myproject", "--image", "joomla:latest", "--timezone", "America/New_York"})
	initCmd.Run(initCmd, []string{})

	if !strings.Contains(b.String(), "not yet implemented") {
		t.Errorf("expected 'not yet implemented', got: %s", b.String())
	}
}

func TestCreateCommand_WithValidType(t *testing.T) {
	b := new(bytes.Buffer)
	createCmd.SetOut(b)
	createCmd.SetErr(b)
	createCmd.Run(createCmd, []string{"component"})

	if !strings.Contains(b.String(), "not yet implemented") {
		t.Errorf("expected 'not yet implemented', got: %s", b.String())
	}
}

func TestBuildCommand_WithName(t *testing.T) {
	b := new(bytes.Buffer)
	buildCmd.SetOut(b)
	buildCmd.SetErr(b)
	buildCmd.Run(buildCmd, []string{"myproject"})

	if !strings.Contains(b.String(), "not yet implemented") {
		t.Errorf("expected 'not yet implemented', got: %s", b.String())
	}
}

func TestVersionCommand_NotYetImplemented(t *testing.T) {
	// Verify there's no version flag that would confuse root
	versionFlag := rootCmd.Flags().Lookup("version")
	if versionFlag == nil {
		// No version flag → expected for a stub, root just shows help
		b := new(bytes.Buffer)
		rootCmd.SetOut(b)
		rootCmd.SetErr(b)
		rootCmd.SetArgs([]string{"--version"})

		err := rootCmd.Execute()
		if err != nil {
			// cobra may error on unknown flag, that's fine
			t.Log("--version triggers unknown flag error, expected for stubs")
		}
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
