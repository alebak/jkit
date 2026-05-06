# Delta for CLI Commands

## ADDED Requirements

### R-CLI-05: MCP Subcommands

The CLI MUST define a `mcp` command group with `list`, `add`, and `remove` subcommands. `jkit mcp list` prints available MCPs. `jkit mcp add <name>` deploys an MCP to the agent config. `jkit mcp remove <name>` removes it.

#### Scenario: mcp list shows built-in MCPs

- GIVEN the embedded MCP templates
- WHEN `jkit mcp list` runs
- THEN output contains "playwright" and "mariadb"
- AND exit code is 0

#### Scenario: mcp add deploys a named MCP

- GIVEN `playwright` is a built-in template
- WHEN `jkit mcp add playwright` runs
- THEN the command succeeds (exit 0)
- AND a `mcpServers.playwright` entry appears in the target config

#### Scenario: mcp remove removes deployed MCP

- GIVEN `playwright` is deployed in `mcp.json`
- WHEN `jkit mcp remove playwright` runs
- THEN the entry is removed
- AND exit code is 0

#### Scenario: mcp add unknown name fails

- GIVEN no template named `bogus` exists
- WHEN `jkit mcp add bogus` runs
- THEN exit code is non-zero
- AND error message includes "unknown MCP"

## MODIFIED Requirements

### R-CLI-02: Cobra Command Stubs

The CLI MUST define a root cobra command with four subcommands: `init`, `create`, `build`, and `mcp`. Each subcommand MUST print "not yet implemented" and exit 0, except `mcp` which SHALL have its own subcommand group.
(Previously: Root command had three subcommands: init, create, build)

#### Scenario: Root command produces no output by default

- GIVEN the built `jkit` binary
- WHEN invoked with no arguments
- THEN it prints cobra help text (no error)
- AND exit code is 0
- AND help mentions `mcp` as a subcommand

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

## Coverage Update

| Spec | Source |
|------|--------|
| R-CLI-05 | Proposal: New `jkit mcp` commands |
