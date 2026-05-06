package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestMCPCommand_List(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"mcp", "list"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Available MCP servers:") {
		t.Errorf("expected 'Available MCP servers:', got: %s", output)
	}
	if !strings.Contains(output, "playwright") {
		t.Errorf("expected playwright in list, got: %s", output)
	}
	if !strings.Contains(output, "mariadb") {
		t.Errorf("expected mariadb in list, got: %s", output)
	}
}

func TestMCPCommand_AddNoArgs(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"mcp", "add"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for no args")
	}
}

func TestMCPCommand_RemoveNoArgs(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"mcp", "remove"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for no args")
	}
}

func TestMCPCommand_AddInvalidMCP(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"mcp", "add", "nonexistent"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid MCP name")
	}
	if !strings.Contains(err.Error(), "unknown MCP") {
		t.Errorf("expected error to mention 'unknown MCP', got: %v", err)
	}
}
