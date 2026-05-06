# Design: MCP Manager

## Technical Approach

Mirror the `internal/agents/` pattern exactly: embedded JSON templates, `MCPConfigPathFor()` (analogous to `SkillDirFor()`), `ListAvailable()`, `DeployMCP()`, `RemoveMCP()`. JSON merge strategy reads target `mcp.json`, adds/removes entries, creates `.bak` before writing. CLI mirrors `cmd/jkit/agents.go` with `jkit mcp list|add|remove`.

## Architecture Decisions

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Symlink vs JSON merge | Symlinks simpler but agents expect single flat mcp.json | **JSON merge** — read → merge → backup → write |
| Per-agent config map | Hardcoded switch vs map | **Switch** — mirrors `SkillDirFor()` exactly, env var overrides for each agent |
| Validation on add | Validate template exists before deploy | **Pre-check** — `ListAvailable()` on embedded McpFS, error if not found |
| Backup naming | `mcp.json.bak` (overwrites) vs timestamped | **`.bak` overwrite** — simpler, sufficient for rollback per spec R-MCPS-04 |

## Data Flow

```
CLI (cmd/jkit/mcp.go)
  │
  ├── list → internal/mcp.ListAvailable(McpFS)→ prints names from templates/mcp/*.json
  │
  ├── add <name>
  │     ├── Validate name via ListAvailable(...)
  │     ├── Read embedded templates/mcp/<name>.json
  │     ├── Resolve MCPConfigPathFor(agent)
  │     ├── Read existing mcp.json (or init empty)
  │     ├── Backup → mcp.json.bak
  │     ├── JSON merge entry into mcpServers map
  │     └── Write merged mcp.json
  │
  └── remove <name>
        ├── Resolve MCPConfigPathFor(agent)
        ├── Read existing mcp.json
        ├── Backup → mcp.json.bak
        ├── Delete entry from mcpServers map
        └── Write mcp.json (or error if missing)
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/mcp/mcp.go` | Modify | Replace stub with full `ListAvailable`, `MCPConfigPathFor`, `DeployMCP`, `RemoveMCP` |
| `internal/mcp/config.go` | Create | `MCPServers`, `MCPEntry` structs; `readMCPConfig`, `writeMCPConfig`, `mergeMCPEntry`, `removeMCPEntry` |
| `cmd/jkit/mcp.go` | Create | Cobra `mcpCmd` + `listCmd`, `addCmd`, `removeCmd` — mirrors `agents.go` |
| `templates/mcp/playwright.json` | Create | `{"mcpServers":{"playwright":{"command":"npx","args":["@playwright/mcp"]}}}` |
| `templates/mcp/mariadb.json` | Create | `{"mcpServers":{"mariadb":{"command":"npx","args":["-y","@anthropic/mcp-server-mariadb","..."]}}}` |
| `embed_assets.go` | Modify | Add `//go:embed templates/mcp/*.json` + `var McpFS embed.FS` |
| `templates/devcontainer/php-custom.ini` | Modify | Change `xdebug.mode=debug` to `xdebug.mode=debug,trace,profile` |

## Interfaces / Contracts

```go
// Package mcp provides MCP server config lifecycle for AI agent tooling.
package mcp

type MCPServers struct {
    Servers map[string]MCPEntry `json:"mcpServers"`
}

type MCPEntry struct {
    Command string   `json:"command"`
    Args    []string `json:"args"`
}

func ListAvailable(ctx context.Context, fsys fs.FS) ([]string, error)
func MCPConfigPathFor(ctx context.Context, agentName string) (string, error)
func DeployMCP(ctx context.Context, configPath, name string, data []byte) error
func RemoveMCP(ctx context.Context, configPath, name string) error
```

`MCPConfigPathFor` agent map:
- `claude` → `~/.claude/mcp.json` (env: `CLAUDE_MCP_CONFIG`)
- `opencode` → `~/.config/opencode/mcp.json` (env: `OPENCODE_MCP_CONFIG`)
- `gemini` → `~/.gemini/mcp.json` (env: `GEMINI_MCP_CONFIG`)

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | `ListAvailable` | `fstest.MapFS` with `.json` files — same pattern as `agents_test.go` |
| Unit | `MCPConfigPathFor` | Table-driven with env var overrides, `HOME` override via `t.TempDir()` |
| Unit | `DeployMCP` / `RemoveMCP` | Temp dirs, verify JSON content, verify `.bak` created, verify idempotency |
| Unit | JSON merge edge cases | Corrupt template, duplicate entry, remove missing entry, empty existing config |
| Unit | JSON validity | Verify all embedded templates parse as valid `MCPServers` |

## Migration / Rollout

No migration required. Existing projects with `php-custom.ini` will pick up the new xdebug mode on next `jkit init`. The MCP manager is additive — no existing behavior changes.

## Open Questions

None.
