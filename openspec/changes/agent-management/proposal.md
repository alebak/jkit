# Proposal: Agent Management

## Intent

Add `jkit agents list|add|remove` commands and skill deployment so users can select AI agents for their devcontainer and get the prd-creator skill installed — without a config file to maintain.

## Scope

### In Scope
- `jkit agents list` — discover available agents from embedded `AgentsFS`
- `jkit agents add [name...]` — validate, regenerate post-create.sh with selected agents
- `jkit agents remove [name...]` — read current post-create.sh, remove specified agents, regenerate
- Agent auto-discovery from `templates/agents/*.sh` — no hardcoded list
- Post-create.sh regeneration: read/render/write via existing renderer
- Skill deployment: copy prd-creator to `.jkit/agents/skills/` + symlinks in `jkit init`

### Out of Scope
- Agent config persistence (no state file; post-create.sh is source of truth)
- Full `jkit init` TUI implementation (deferred)
- gentle-ai skill management (gentle-ai manages its own 14 skills)
- MCP management

## Capabilities

### New Capabilities
- `agent-management`: Agent listing, add/remove, and skill deployment for selected AI agents

### Modified Capabilities
- `cli-commands`: New `agents` subcommand group (`list|add|remove`)
- `devcontainer-init`: Post-create.sh re-generation flow; agent listing from embedded FS

## Approach

**Agent listing**: Read `templates/agents/*.sh` via `AgentsFS`. Name = filename minus `.sh`. Map names to display labels in Go code.

**`jkit agents add [name...]`**: Validates names against available agents. Regenerates post-create.sh via `renderPostCreate()` with `SelectedAgents` set to given names. Output includes agent-delimiter comments so the file is parseable.

**`jkit agents remove [name...]`**: Reads existing post-create.sh, parses agent-delimiter comments to discover current set, subtracts removed agents, regenerates with remainder.

**Skill deployment**: `jkit init` copies `templates/skills/prd-creator/SKILL.md` → `.jkit/agents/skills/prd-creator/SKILL.md`. Creates symlinks from each agent's skill dir (`~/.claude/skills/prd-creator`, `~/.config/opencode/skills/prd-creator`, `~/.gemini/skills/prd-creator`) to `.jkit/agents/skills/prd-creator/SKILL.md`.

**No config file**: Unlike extensions, agent selection is not persisted in YAML. The post-create.sh content is the authoritative record.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/jkit/agents.go` | New | `jkit agents list|add|remove` commands |
| `internal/agents/agents.go` | Modified | Agent discovery, validation, post-create.sh parsing |
| `internal/devcontainer/renderer.go` | Modified | Add agent marker comments to output |
| `cmd/jkit/init.go` | Modified | Add skill deployment call |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Post-create.sh format change breaks parsing | Low | Use explicit agent-delimiter comments (`# --- agent:xyz ---`) |
| Symlinks fail on Windows | Low | Document Linux/macOS requirement in help text |
| Post-create.sh doesn't exist yet | Low | `add` creates fresh; `remove` errors with "no agents file found" |

## Rollback Plan

Re-run `jkit agents add` with the desired names, or restore previous `post-create.sh` from git.

## Dependencies

- Existing `AgentsFS`, `SkillsFS` embedded filesystems (in `embed_assets.go`)
- Existing `renderPostCreate()` in `internal/devcontainer/renderer.go`
- Cobra CLI framework (already in `go.mod`)

## Success Criteria

- [ ] `jkit agents list` shows available agents from embedded FS
- [ ] `jkit agents add claude opencode` validates and regenerates post-create.sh with only those agents
- [ ] `jkit agents remove gemini` parses current file, removes gemini, rewrites
- [ ] `jkit init` creates `.jkit/agents/skills/prd-creator/SKILL.md` and symlinks
- [ ] All existing tests pass; new tests cover agent discovery and parsing
