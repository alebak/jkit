# MCP Management Specification

## Purpose

Define the lifecycle for MCP (Model Context Protocol) server configs: discovery of built-in MCPs from embedded templates, per-agent deployment and removal via JSON merge, and backup safety.

## Requirements

### R-MCPS-01: List Built-in MCPs

The `jkit mcp list` command MUST enumerate available built-in MCP servers from embedded templates in `templates/mcp/*.json`.

#### Scenario: Lists available MCPs

- GIVEN the embedded MCP templates directory contains `playwright.json` and `mariadb.json`
- WHEN `jkit mcp list` is executed
- THEN output includes "playwright" and "mariadb"
- AND exit code is 0

#### Scenario: Empty templates directory

- GIVEN no MCP templates exist at compile time (empty embed)
- WHEN `jkit mcp list` is executed
- THEN output indicates no MCPs available
- AND exit code is 0

### R-MCPS-02: Add MCP

The `jkit mcp add <name>` command MUST validate the name against built-in templates, read the MCP JSON, and merge it into the target agent's `mcp.json` via JSON merge.

#### Scenario: Add known MCP succeeds

- GIVEN `playwright.json` is a built-in MCP template
- WHEN `jkit mcp add playwright` is executed
- THEN the target agent's `mcp.json` contains a `mcpServers.playwright` entry
- AND existing entries in `mcp.json` are preserved

#### Scenario: Add unknown MCP fails

- GIVEN no template named `unknown` exists
- WHEN `jkit mcp add unknown` is executed
- THEN an error message is printed: "unknown MCP: unknown"
- AND exit code is non-zero

#### Scenario: Add duplicate MCP is idempotent

- GIVEN `playwright` is already deployed in the target `mcp.json`
- WHEN `jkit mcp add playwright` is executed again
- THEN the operation succeeds (idempotent)
- AND no duplicate entry is created

### R-MCPS-03: Remove MCP

The `jkit mcp remove <name>` command MUST remove the named MCP entry from the agent's `mcp.json` and leave all other entries intact.

#### Scenario: Remove existing MCP

- GIVEN target `mcp.json` has `playwright` and `mariadb` entries
- WHEN `jkit mcp remove playwright` is executed
- THEN `mcpServers.playwright` is removed from `mcp.json`
- AND `mcpServers.mariadb` is preserved

#### Scenario: Remove non-existent MCP

- GIVEN target `mcp.json` has no `nonexistent` entry
- WHEN `jkit mcp remove nonexistent` is executed
- THEN the command prints a warning but exits 0
- AND `mcp.json` is not modified

### R-MCPS-04: JSON Merge Safety

The deploy and remove operations MUST create a backup of the target `mcp.json` before writing, and MUST validate JSON after merge.

#### Scenario: Backup created before write

- GIVEN target `mcp.json` exists with content `{"mcpServers":{"old":{"command":"x"}}}`
- WHEN `jkit mcp add playwright` executes
- THEN a backup file `mcp.json.bak` exists with the original content

#### Scenario: Corrupt merge is rejected

- GIVEN a malformed MCP template (invalid JSON)
- WHEN `jkit mcp add playwright` is attempted
- THEN the command errors and exits non-zero
- AND the target `mcp.json` is not modified

### R-MCPS-05: Built-in MCP Templates

`templates/mcp/` SHALL contain `playwright.json` and `mariadb.json` with valid MCP server JSON, embedded via `//go:embed`.

#### Scenario: Templates embed correctly

- GIVEN the Go binary is built
- WHEN embedded filesystem is read
- THEN `templates/mcp/playwright.json` and `templates/mcp/mariadb.json` are accessible
- AND each file is valid JSON with a `mcpServers` top-level key

## Coverage

| Spec | Source |
|------|--------|
| R-MCPS-01 | Proposal: `jkit mcp list` |
| R-MCPS-02 | Proposal: `jkit mcp add` |
| R-MCPS-03 | Proposal: `jkit mcp remove` |
| R-MCPS-04 | Proposal: JSON merge backup |
| R-MCPS-05 | Proposal: templates/mcp/ embed |
