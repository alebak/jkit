## Exploration: Agent Management

### Current State

JKit currently has a basic agent scaffolding:

- **3 agent templates**: `templates/agents/{claude,opencode,gemini}.sh` — simple bash snippets that check/install each agent binary. Embedded via `//go:embed` in `embed_assets.go` as `AgentsFS`.
- **`DevcontainerData.SelectedAgents []string`**: field in the struct, defaults to all 3 agents. When empty, all agents are included (backward compat).
- **`renderPostCreate()`** in `internal/devcontainer/renderer.go`: renders the `post-create.sh` Go template header (which always installs `gentle-ai`), then concatenates agent bash snippets filtered by `SelectedAgents`.
- **`internal/agents/agents.go`**: empty stub package (package doc only).
- **`cmd/jkit/init.go`**: has `--agents` flag (string slice) but the command is still "not yet implemented".
- **`prd-creator` skill**: embedded at `templates/skills/prd-creator/SKILL.md`. Currently only stored, never deployed.

### How gentle-ai Handles Skills

gentle-ai ships 14 built-in skills (SDD + Foundation) embedded in its binary. It distributes these to per-agent skill directories (e.g., `~/.claude/skills/`, `~/.config/opencode/skills/`, `~/.cursor/skills/`, etc.) during `gentle-ai install` or `gentle-ai sync`.

**Crucially**: gentle-ai does NOT support third-party/plugin skills in its distribution model. External skills (like JKit's prd-creator) must be placed by the tool that owns them. The "JKit delegates skill placement to gentle-ai" (DD-04) means JKit relies on gentle-ai for the general skill infrastructure, but JKit-specific skills must be placed by JKit itself.

### Affected Areas

- `internal/agents/agents.go` — currently empty stub, needs real implementation
- `cmd/jkit/agents.go` (new) — `jkit agents list|add|remove` subcommand
- `internal/devcontainer/renderer.go` — might need `RenderAgentsOnly()` or similar for re-generation
- `internal/devcontainer/devcontainer.go` — DevcontainerData struct might need config file integration
- `templates/agents/*.sh` — existing patterns, may need a listing mechanism
- `embed_assets.go` — add `//go:embed templates/agents/*.sh` if not already (it's there)
- `openspec/specs/cli-commands/spec.md` — needs new spec section for agent commands
- `openspec/specs/devcontainer-init/spec.md` — may need updates for agent config persistence
- `templates/skills/prd-creator/SKILL.md` — needs deployment mechanism in post-create.sh or init flow

### Config Persistence Options

We need to persist agent selection for R-AGNT-09 (add/remove in existing projects).

**Option A: Extend `extensions.jkit.yaml` with `agents:` field**

Reuse the existing YAML registry pattern (atomic write via temp file + rename). Same file stores both extensions and agents.

```
# extensions.jkit.yaml
extensions: []
agents: ["claude", "opencode"]
```

- Pros: Reuses proven pattern, single config file, atomic writes, no new file to discover
- Cons: Mixed responsibility (extensions + agents in one file)
- Effort: Low

**Option B: New `agents.jkit.yaml`**

Dedicated config file for agent selection only.

- Pros: Clean separation of concerns
- Cons: New file, needs read/write logic duplication or shared helper
- Effort: Low-Medium

**Option C: No config — only `jkit init --agents`**

Agents only set during project creation, stored only in the generated `post-create.sh`.

- Pros: Simplest, no persistence mechanism needed
- Cons: Violates R-AGNT-09 (can't add/remove in existing projects)
- Effort: Trivial

### Recommendation: Option A

The `extensions.jkit.yaml` file already has read/write infrastructure in `internal/generator/registry.go`. Extending it with an `agents` field is the pragmatic choice — one config file, one read/write pattern, and the same atomic write guarantees. The agents field is simple: a `[]string` of agent names.

### Agent Commands

**`jkit agents list`**:
- Available agents: discovered from embedded `templates/agents/*.sh` (via `AgentsFS`)
- Installed agents: read from `extensions.jkit.yaml`
- Output: both lists side by side (or installed with a marker)

**`jkit agents add <name>`**:
- Validate agent exists in embedded FS
- Add to config
- Regenerate `.devcontainer/post-create.sh`
- Output: success message

**`jkit agents remove <name>`**:
- Validate agent is in config
- Remove from config
- Regenerate `.devcontainer/post-create.sh`
- Output: success message

**`jkit agents add --help` / remove --help** should list valid agent names.

### Post-Create.sh Regeneration Flow

When agents change, the flow is:

1. Read current `extensions.jkit.yaml`
2. Update `agents` list
3. Write updated config (atomic write)
4. Read current `.devcontainer/` files (or construct a DevcontainerData from defaults + config)
5. Update `SelectedAgents` in DevcontainerData
6. Re-render `post-create.sh` with the renderer
7. Write back to `.devcontainer/post-create.sh`

This requires the `agents` command to have access to:
- The renderer (internal/devcontainer)
- The config file (internal/generator registry pattern)
- The project root directory

### Skill Deployment (prd-creator)

The prd-creator skill is Joomla-specific. Three approaches:

**Approach 1: Post-create.sh deploys it**
After agent installs, the post-create.sh copies the skill from JKit's embedded FS to each agent's skills directory (e.g., `~/.claude/skills/prd-creator/SKILL.md` for Claude Code, `~/.config/opencode/skills/prd-creator/SKILL.md` for OpenCode, etc.).
- Pros: Automatic, works for new projects
- Cons: Needs per-agent directory knowledge (coupling), adds complexity to bash scripts

**Approach 2: `jkit init` deploys it**
During the init wizard, JKit copies the skill to a project-level location (e.g., `.claude/skills/prd-creator/SKILL.md` in the project root) for each selected agent.
- Pros: Controlled by Go code, can map agent names to directories
- Cons: Per-agent directory mapping needed

**Approach 3: User manual placement via `jkit agents`**
The `jkit agents add` command could deploy the skill alongside the agent install.
- Pros: Unified flow, only when needed
- Cons: Mixed responsibility (agent install vs skill deploy)

**Recommendation**: Approach 2 (deploy during `jkit init`). The init wizard already knows which agents are selected and can map agent names to their skill directory paths. This keeps the bash scripts simple and the logic in Go where it's testable. The skill paths for each agent are known:
- Claude Code: `~/.claude/skills/prd-creator/SKILL.md`
- OpenCode: `~/.config/opencode/skills/prd-creator/SKILL.md`
- Gemini CLI: `~/.gemini/skills/prd-creator/SKILL.md`

### Risks

1. **gentle-ai skill path changes**: If gentle-ai changes its skill directory layout, JKit's per-agent hardcoded paths break. Mitigation: document the mapping in one place, version it.
2. **Config file responsibility creep**: Adding agents to `extensions.jkit.yaml` mixes concerns. Mitigation: keep the agents field simple (just `[]string`), document clearly.
3. **Regeneration needs full DevcontainerData**: The current `Render()` needs a complete `DevcontainerData`. When regenerating just the agents, we need the original data. Mitigation: store minimal restoring data in the config, or read existing `.devcontainer/` files.
4. **No interactive menu in CLI**: `jkit agents list` should show available vs installed, but there's no TUI. Mitigation: simple table output is sufficient for an MVP.
5. **No gentle-ai skills in JKit scope**: gentle-ai manages 14 built-in skills. JKit should NOT try to duplicate or override these. The prd-creator is the only JKit-specific skill.
6. **New agents need templates**: Adding a new agent (e.g., `cursor`, `windsurf`) requires a new `templates/agents/<name>.sh` file. The agent commands auto-discover from the embedded FS, so no code changes needed — just add the `.sh` file.

### Ready for Proposal

Yes. All key questions are answered:

| Question | Answer |
|----------|--------|
| Where does gentle-ai expect skills? | gentle-ai manages its own 14 skills. External skills (prd-creator) must be placed by JKit. |
| Should `jkit agents list` show available vs installed? | Yes — available from embedded FS, installed from config |
| How to track installed agents? | `extensions.jkit.yaml` — extend the existing registry file with an `agents: []` field |
| What should `jkit agents add claude` do? | Validate → update config → regenerate post-create.sh → print success |
