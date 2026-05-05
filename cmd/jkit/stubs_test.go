package main

import (
	"testing"

	_ "github.com/alebak/jkit/internal/agents"
	_ "github.com/alebak/jkit/internal/generator"
	_ "github.com/alebak/jkit/internal/mcp"
)

func TestStubPackagesCompile(t *testing.T) {
	// Compilation test: if this test runs without build errors,
	// all three internal stub packages compile and are importable.
	// No runtime assertions needed — the blank imports above are the test.
	t.Log("All internal stub packages compile successfully: generator, agents, mcp")
}
