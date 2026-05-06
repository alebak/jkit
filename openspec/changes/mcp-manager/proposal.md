# Proposal: MCP Manager

## Intent

JKit generates devcontainers for Joomla projects that need MCP servers (Playwright, MariaDB) for AI agents. Currently `internal/mcp/` is a stub with zero implementation — no CLI commands, no config deployment, no templates. Agents can't discover, deploy, or remove MCP configs.

## Scope

### In Scope
- `jkit mcp list|add|remove` subcommands mirroring `jkit agents` pattern
- Playwright MCP JSON template (`templates/mcp/playwright.json`)
- MariaDB MCP JSON template (`templates/mcp/mariadb.json`)
- JSON merge deployment to agent MCP configs (not symlinks — agents expect single `mcp.json`)
- Xdebug trace/profile in `php-custom.ini` (config change, not MCP — per PA-02)
- MCP installation section in `post-create.sh` for init flow

### Out of Scope
- gentle-ai MCP config (delegated per DD-04 — gentle-ai discovers its own config)
- Custom MCPs for third-party extensions (future work)

## Capabilities

### New Capabilities
- `mcp-management`: MCP server lifecycle — discover built-in MCPs from embedded templates, deploy/remove per-agent configs, JSON merge strategy

### Modified Capabilities
- `cli-commands`: New `jkit mcp list|add|remove` subcommands with flag contracts
- `devcontainer-init`: MCP config deployment phase in init flow; `php-custom.ini` gains `xdebug.mode=debug,trace,profile`; `post-create.sh` gains MCP section

## Approach

Follow AGNT pattern exactly:

- **`internal/mcp/mcp.go`** — `ListAvailable()`, `MCPConfigPathFor()`, `DeployMCP()`, `RemoveMCP()` mirroring `internal/agents/`
- **`internal/mcp/config.go`** — MCP JSON structs (`MCPServers`, `MCPEntry`), read/merge/write helpers for `mcp.json`
- **`cmd/jkit/mcp.go`** — Cobra subcommands: `jkit mcp list`, `jkit mcp add <name>`, `jkit mcp remove <name>`
- **`templates/mcp/`** — Embedded JSON templates (playwright.json, mariadb.json)
- **JSON merge**: Deploy reads target `mcp.json`, merges entries from `.jkit/agents/mcp/<name>.json`, writes back. Remove reverses. No symlinks.
- **`embed_assets.go`** — Add `//go:embed templates/mcp/*.json`
- **post-create.sh** — New `# --- mcp installations ---` section after agent install

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/mcp/mcp.go` | Modified | Full implementation replacing stub |
| `internal/mcp/config.go` | New | JSON structs, merge/read/write |
| `cmd/jkit/mcp.go` | New | `jkit mcp list|add|remove` |
| `templates/mcp/` | New | playwright.json, mariadb.json |
| `embed_assets.go` | Modified | Add templates/mcp/*.json embed |
| `templates/devcontainer/php-custom.ini` | Modified | `xdebug.mode=debug,trace,profile` |
| `templates/devcontainer/post-create.sh` | Modified | MCP install section |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| JSON merge corrupts agent's mcp.json | Low | Backup before write; validate JSON after merge |
| Agent path resolution fails | Low | `MCPConfigPathFor()` uses `$HOME`, env var overrides, `os.UserHomeDir()` fallback |
| MariaDB MCP fails if db service not running | Medium | Config is valid JSON; runtime failure handled by agent |

## Rollback Plan

1. **CLI**: `jkit mcp remove <name>` removes entry and restores from backup
2. **Manual**: `rm -rf .jkit/agents/mcp/` removes all deployed configs
3. **Init revert**: Re-run `jkit init` without MCP flag

## Dependencies

- `internal/agents/` package (AGNT pattern reference — already implemented)
- Node.js in devcontainer (already in Dockerfile — needed for `npx @playwright/mcp`)
- MariaDB `db` service in docker-compose (already exists)

## Success Criteria

- [ ] `jkit mcp list` shows `playwright` and `mariadb` as available
- [ ] `jkit mcp add playwright` merges JSON into agent's `mcp.json` without overwriting existing entries
- [ ] `jkit mcp remove playwright` removes entry and leaves others intact
- [ ] `jkit init` generates `post-create.sh` that deploys selected MCPs
- [ ] `php-custom.ini` contains `xdebug.mode=debug,trace,profile`
- [ ] All existing specs pass unchanged (backward compatible)
