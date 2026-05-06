# Tasks: MCP Manager

**PRs**: 1 (single PR, ~300 lines, low risk)
**Delivery strategy**: single-pr
**Decision needed before apply**: No
**Chained PRs recommended**: No
**400-line budget risk**: Low

## Task 1: Foundation — MCP discovery + config path mapping

**File**: `internal/mcp/mcp.go`

- [x] 1.1 Implement `ListAvailable(ctx, fsys)` — reads `templates/mcp/*.json` from embedded FS, strips `.json` extension, sorts. Same pattern as `agents.ListAvailable` but scans `.json` instead of `.sh`.
- [x] 1.2 Implement `MCPConfigPathFor(ctx, agentName)` — switch on `claude`/`opencode`/`gemini` with env var overrides (`CLAUDE_MCP_CONFIG`, `OPENCODE_MCP_CONFIG`, `GEMINI_MCP_CONFIG`), falls back to `$HOME/.<agent>/mcp.json` / `$HOME/.config/<agent>/mcp.json`.

## Task 2: Core — DeployMCP + RemoveMCP with JSON merge

**File**: `internal/mcp/config.go`

- [x] 2.1 Define `MCPServers` (map `mcpServers`) and `MCPEntry` (command + args) structs with JSON tags.
- [x] 2.2 Implement `readMCPConfig(path)` → `MCPServers` — read + unmarshal; return empty struct if file missing.
- [x] 2.3 Implement `writeMCPConfig(path, cfg)` — marshal with indent, write to path.
- [x] 2.4 Implement `DeployMCP(ctx, configPath, name, data)` — read existing, merge entry into map, backup → `.bak`, write merged.
- [x] 2.5 Implement `RemoveMCP(ctx, configPath, name)` — read existing, delete key from map, backup → `.bak`, write. No-op if key missing (warn but exit 0).

## Task 3: Templates — playwright.json + mariadb.json + embed

- [x] 3.1 Create `templates/mcp/playwright.json` — `npx @playwright/mcp` entry.
- [x] 3.2 Create `templates/mcp/mariadb.json` — `npx -y @anthropic/mcp-server-mariadb` entry.
- [x] 3.3 Modify `embed_assets.go` — add `//go:embed templates/mcp/*.json` + `var McpFS embed.FS`.

## Task 4: CLI — jkit mcp list|add|remove subcommands

**File**: `cmd/jkit/mcp.go`

- [x] 4.1 `mcpCmd` — `Use: "mcp"`, `Short: "Manage MCP server installations"`. Registers `list`, `add`, `remove` subcommands. `init()` adds to `rootCmd`.
- [x] 4.2 `mcpListCmd` — calls `mcp.ListAvailable(ctx, jkit.McpFS)`, prints each name, exit 0 on empty.
- [x] 4.3 `mcpAddCmd` — validates name against available, resolves `MCPConfigPathFor(agent)` (with `--agent` flag, default `opencode`), reads embedded template, calls `mcp.DeployMCP`. Flags: `--agent` (default "opencode").
- [x] 4.4 `mcpRemoveCmd` — resolves config path, calls `mcp.RemoveMCP`. Same `--agent` flag.

## Task 5: Xdebug — update php-custom.ini

**File**: `templates/devcontainer/php-custom.ini`

- [x] 5.1 Change `xdebug.mode=debug` to `xdebug.mode=debug,trace,profile`.

## Task 6: Tests + verification

**Files**: `internal/mcp/mcp_test.go` (new), `cmd/jkit/mcp_test.go` (new)

- [x] 6.1 Write `TestListAvailable` — fstest.MapFS with `.json` files, mirrors `agents_test.go` pattern.
- [x] 6.2 Write `TestMCPConfigPathFor` — env var override, default paths, unknown agent error.
- [x] 6.3 Write `TestDeployMCP` — temp dir, deploy playwright, verify JSON content + `.bak` created.
- [x] 6.4 Write `TestRemoveMCP` — deploy then remove, verify entry gone and others intact.
- [x] 6.5 Write `TestDeployMCP_Idempotent` — deploy same MCP twice, no duplicate entry.
- [x] 6.6 Write `TestRemoveMCP_MissingName` — remove nonexistent, no error, file unmodified.
- [x] 6.7 Write `TestEmbeddedTemplatesValid` — iterate `McpFS`, parse each as `MCPServers`, verify top-level `mcpServers` key.
- [x] 6.8 Run `go test ./internal/mcp/...` — all pass.
- [x] 6.9 Run `go vet ./...` — clean.
- [x] 6.10 Run `go build ./cmd/jkit/...` — builds clean.
