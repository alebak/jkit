## Exploration: MCP Manager

### Current State

**MCP is a stub.** `internal/mcp/mcp.go` has only a package doc — zero implementation. The MCP package is imported in `cmd/jkit/stubs_test.go` (blank import for compile check) but has no usable code.

**Dev Container** already provides the infrastructure MCPs need:
- MariaDB `db` service in `docker-compose.yml` (port 3306, hostname `db`)
- Node.js installed in Dockerfile (needed for Playwright MCP via npx)
- Xdebug installed but only in `debug` mode (`php-custom.ini`: `xdebug.mode=debug`)

**No MCP configs exist anywhere.** No agent MCP config files are generated in `post-create.sh`, no `.jkit/agents/mcp/` directory exists, no CLI commands for MCP management.

### Existing Patterns (AGNT — to follow exactly)

The `AGNT` component provides the blueprint for MCP:

| Pattern | AGNT (agents) | MCPS (proposed) |
|---------|---------------|-----------------|
| Package | `internal/agents/` | `internal/mcp/` |
| Deploy root | `.jkit/agents/skills/` | `.jkit/agents/mcp/` |
| Symlink target | Agent skill dirs (`~/.claude/skills/`, etc.) | Agent MCP config dirs (`~/.claude/mcp.json`, etc.) |
| CLI commands | `jkit agents list\|add\|remove` | `jkit mcp list\|add\|remove` |
| Discovery | FS-driven from `templates/agents/*.sh` | TBD — could be embedded templates or a list |
| Config/state source | Post-create.sh delimiter comments | TBD — MCP registry or `.jkit/` YAML |

**Key AGNT patterns to mirror:**
- `SkillDirFor(agentName)` → `MCPConfigPathFor(agentName)` — resolves per-agent MCP config path with env var overrides
- `DeploySkill()` → `DeployMCP()` — write config to `.jkit/agents/mcp/`, symlink to agent's MCP config dir
- File-based marker parsing → Config file parsing (MCP configs are JSON, not marker-delimited bash)
- `agents.go` + `markers.go` → `mcp.go` + `config.go` (or similar)

### Affected Areas

- `internal/mcp/mcp.go` — **modify**: full implementation replacing stub
- `cmd/jkit/mcp.go` — **create**: `jkit mcp list|add|remove` subcommands (mirrors `agents.go` exactly)
- `cmd/jkit/main.go` — **modify**: auto-load? No — `init()` in sibling file registers subcommand automatically
- `embed_assets.go` — **modify**: add `//go:embed templates/mcp/*.json` (or similar) for built-in MCP templates
- `templates/mcp/` — **create**: directory with built-in MCP config templates (playwright.json, mariadb.json)
- `templates/devcontainer/php-custom.ini` — **modify**: add `xdebug.mode=debug,trace,profile` for R-MCPS-03
- `templates/devcontainer/post-create.sh` — **modify**: add MCP installation snippets (npx playwright, etc.)
- `templates/devcontainer/docker-compose.yml` — **potential modify**: no changes needed — MariaDB already exposed on port 3306
- `openspec/specs/cli-commands/spec.md` — **modify**: new spec section for MCP commands
- `openspec/specs/devcontainer-init/spec.md` — **modify**: add MCP phase to init flow

### MCP Technical Details

#### How MCP Configs Work Per Agent

| Agent | Config file location | Format |
|-------|---------------------|--------|
| Claude Code | `~/.claude/mcp.json` | `{"mcpServers": {"name": {"command": "...", "args": [...]}}}` |
| OpenCode | `~/.config/opencode/mcp.json` | Same JSON structure |
| Gemini CLI | `~/.gemini/mcp.json` | Same JSON structure |

All agents use the same JSON schema: `{"mcpServers": {"<name>": {"command": "...", "args": [...], "env": {...}}}}`. This means a single config template works for all agents — only the target path differs.

#### Built-in MCP Configs

**Playwright MCP** (R-MCPS-01):
```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["@playwright/mcp"]
    }
  }
}
```
Runs via `npx` — requires Node.js (already installed via Dockerfile). Playwright browsers are installed on first run.

**MariaDB MCP** (R-MCPS-02):
```json
{
  "mcpServers": {
    "mariadb": {
      "command": "npx",
      "args": ["-y", "@anthropic/mcp-server-mysql"],
      "env": {
        "MYSQL_HOST": "db",
        "MYSQL_PORT": "3306",
        "MYSQL_USER": "${JOOMLA_DB_USER}",
        "MYSQL_PASSWORD": "${JOOMLA_DB_PASSWORD}",
        "MYSQL_DATABASE": "${JOOMLA_DB_NAME}"
      }
    }
  }
}
```
Connects to the MariaDB `db` service in docker-compose. Environment variables from `.env` file. There may be a more specific MariaDB MCP, but the MySQL MCP works since MariaDB is wire-compatible.

**Note:** The Xdebug trace/profile (R-MCPS-03) is NOT an MCP server — it's a PHP config change. Per PA-02 (PRD), Xdebug MCP was investigated and descarted. Instead, Xdebug is configured in `php-custom.ini` to output trace/profile files, and agents read those files directly.

#### Custom MCPs (R-MCPS-07)

Third-party extensions that provide MCP servers would add their own JSON to `.jkit/agents/mcp/<extension-name>.json`. The `jkit mcp add custom` command would accept a `--json` flag or stdin for the config body.

### Approaches

1. **Follow AGNT pattern exactly** — `internal/mcp/` mirrors `internal/agents/`, with `DeployMCP()`, `MCPConfigPathFor()`, `ListAvailableMCPs()`
   - Pros: Consistent architecture, predictable for contributors, reuses symlink pattern, test patterns already established
   - Cons: MCP configs are JSON not bash, so file operations differ (merge JSON vs concatenate bash snippets)
   - Effort: Medium

2. **Minimal — inline in post-create.sh** — Add MCP config generation directly in the post-create.sh bash script
   - Pros: Fastest to implement, no new package needed
   - Cons: Violates separation of concerns (bash is not a JSON template engine), no `jkit mcp` CLI, no testability, hard to manage
   - Effort: Low (but technical debt)

3. **Full MCP registry** — Similar to `extensions.jkit.yaml`, track MCPs in a YAML registry with merge/remove operations
   - Pros: Full lifecycle management, rollback support, atomic operations
   - Cons: Over-engineering for current needs (3 built-in MCPs), YAML + JSON double format
   - Effort: High

### Recommendation

**Approach 1** — Follow AGNT pattern exactly. The `internal/agents/` package provides a proven template:

```
internal/mcp/
├── mcp.go           — ListAvailable(), MCPConfigPathFor(), DeployMCP(), RemoveMCP()
├── config.go        — MCP JSON structs, merge/read/write helpers
└── mcp_test.go      — table-driven tests mirroring agents_test.go

cmd/jkit/mcp.go     — jkit mcp list|add|remove subcommands
templates/mcp/      — built-in MCP JSON templates (playwright.json, mariadb.json)
```

**Specific design decisions:**

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Config storage | `.jkit/agents/mcp/<name>.json` | Mirrors `.jkit/agents/skills/<name>/` — single source of truth |
| Agent config location | `MCPConfigPathFor(agent)` returns dir containing `mcp.json`, env var overrides | Mirrors `SkillDirFor()` — no hardcoded paths (R-MCPS-05) |
| Activation mechanism | Symlink from agent's mcp.json dir to `.jkit/agents/mcp/` OR merge JSON into agent's mcp.json | Symlink for whole directory, or JSON merge if the agent expects a single mcp.json file |
| Xdebug trace/profile | Modify `php-custom.ini` template: `xdebug.mode=debug,trace,profile` | Not an MCP — PA-02 confirmed. Agents read output files directly. |
| Custom MCPs | `jkit mcp add <name> --json '...'` or `--file path` | Accept JSON from flag or file, deploy to `.jkit/agents/mcp/` |
| Init integration | `jkit init` calls MCPS phase after AGNT phase (R-INIT-06) | MCP configs depend on agents being installed first |

### Key Questions Answered

| Question | Answer |
|----------|--------|
| How does gentle-ai know the MCP config location? | DD-04 delegation: gentle-ai already discovers MCP configs at agent-standard paths (`~/.claude/mcp.json`, `~/.config/opencode/mcp.json`). JKit writes configs to those paths (via symlinks or JSON merge). No code change in gentle-ai needed. |
| Should MCP configs be per-agent or shared? | **Shared** with per-agent activation. Configs live in `.jkit/agents/mcp/` and are symlinked or merged into each agent's MCP config file. Same pattern as skills. |
| What format for Playwright MCP config? | Standard MCP JSON: `{"mcpServers": {"playwright": {"command": "npx", "args": ["@playwright/mcp"]}}}` |
| What format for MariaDB MCP config? | Standard MCP JSON with connection env vars for the `db` docker-compose service. Uses MySQL-compatible MCP server. |
| How to handle Xdebug trace/profile? | **Not an MCP.** Modify `php-custom.ini` to add `trace,profile` to `xdebug.mode`. Agents read `.xt`/`.xg` output files from the project. |
| What about the existing `post-create.sh`? | MCP installation is a NEW section after agent installation. The post-create.sh template gets a `# --- MCP installations ---` section (following the marker pattern from agents). |

### Ready for Proposal

Yes. All requirements from the PRD (R-MCPS-01 through R-MCPS-08) have clear, implementable solutions. The AGNT pattern is well-understood and provides a battle-tested template.

The orchestrator should proceed with:
1. **Proposal**: "MCP Manager — `jkit mcp list|add|remove` commands and devcontainer MCP integration"
2. **Spec**: Delta specs for MCPS requirements (R-MCPS-01 through R-MCPS-08)
3. **Design**: Follow internal/mcp/ pattern, JSON config structure, symlink deployment
4. **Tasks**: 5-8 implementation tasks mirroring the AGNT change structure
