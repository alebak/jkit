# CLI Commands Specification

## Purpose

Define the CLI entry points, module initialization, and internal package stubs for JKit. This change delivers compile-ready scaffolding — subcommands print "not yet implemented".

## Requirements

### R-CLI-01: Module Initialization

The Go module MUST be `github.com/alebak/jkit` targeting Go 1.24, compilable and vet-clean.

#### Scenario: Module compiles cleanly
- GIVEN a fresh checkout at `/workspaces/jkit`
- WHEN `go build ./...` runs
- THEN it exits with code 0

#### Scenario: Go vet passes
- GIVEN a successful build
- WHEN `go vet ./...` runs
- THEN it exits with code 0

#### Scenario: Dependencies resolve
- GIVEN `go.mod` and initial imports
- WHEN `go mod tidy` runs
- THEN `go.sum` is created and no errors occur

### R-CLI-02: Cobra Command Stubs

The CLI MUST define a root cobra command with three subcommands: `init`, `create`, `build`. Each subcommand MUST print "not yet implemented" and exit 0.

#### Scenario: Root command produces no output by default
- GIVEN the built `jkit` binary
- WHEN invoked with no arguments
- THEN it prints cobra help text (no error)
- AND exit code is 0

#### Scenario: Init subcommand stub
- GIVEN `jkit init`
- WHEN executed
- THEN output contains "not yet implemented"
- AND exit code is 0

#### Scenario: Create subcommand stub
- GIVEN `jkit create`
- WHEN executed
- THEN output contains "not yet implemented"

#### Scenario: Build subcommand stub
- GIVEN `jkit build`
- WHEN executed
- THEN output contains "not yet implemented"

### R-CLI-03: Internal Package Stubs

Three internal packages SHALL compile: `internal/generator`, `internal/agents`, `internal/mcp`.

#### Scenario: Internal packages compile
- GIVEN stub `.go` files in each package
- WHEN `go build ./internal/...` runs
- THEN it exits with code 0

### R-CLI-04: Gitignore

The `.gitignore` MUST contain `.atl/`.

#### Scenario: .gitignore excludes .atl
- GIVEN the project `.gitignore`
- WHEN inspected
- THEN it contains a line matching `.atl/`

### Coverage

| Spec | PRD Req |
|------|---------|
| R-CLI-01 | DD-07 (module path) |
| R-CLI-02 | R-INIT-03 (entry point) |
| R-CLI-03 | PRD §5 (internal structure) |
| R-CLI-04 | R-DEVC-06 (gitignore) |
