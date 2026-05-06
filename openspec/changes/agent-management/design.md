# Design: Agent Management

## Technical Approach

Three new CLI subcommands (`list|add|remove`) under `jkit agents`. Agent discovery is FS-driven from `templates/agents/*.sh` — no hardcoded lists. Post-create.sh is the authoritative record of installed agents, using machine-parseable delimiter comments. Skill deployment copies prd-creator to `.jkit/agents/skills/` and creates symlinks to each agent's skill directory. Regeneration always goes through the existing `Render()` path for idempotency (specs R-AGNT-01 through R-AGNT-06).

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|--------------|-----------|
| Agent state source | Post-create.sh delimiter comments | YAML config, JSON state file | No sync needed — post-create.sh is already generated and idempotent. Markers (`# --- agent:name ---`) make it parseable without bash execution. |
| Agent discovery | `fs.ReadDir(templates/agents/*.sh)` | Hardcoded Go map, YAML manifest | Adding an agent = adding a `.sh` file. Zero code changes. Matches existing embedded FS pattern. |
| Regeneration strategy | Full re-render via `Render()` | Sed/awk inline editing | Renderer is already idempotent. Full re-render guarantees template consistency. No risk of corrupted partial edits. |
| Skill deploy strategy | One copy + symlinks per agent | Copy to each agent dir | Single source of truth in `.jkit/`. Updating the skill file updates all agents via symlinks. Matches project-local dotfile convention. |

## Data Flow

```
jkit agents add claude
  ┌─► agents.ListAvailable() ──► AgentsFS ──► ["claude","opencode","gemini"]
  │    Validate "claude" ∈ available
  ├─► Read .devcontainer/post-create.sh
  ├─► agents.ParsePostCreateMarkers() ──► ["claude","opencode"] (current set)
  ├─► Union: input ∪ current = ["claude","opencode"]
  ├─► DevcontainerData{SelectedAgents: ["claude","opencode"]}
  ├─► devcontainer.Render(ctx, w, "post-create.sh", data)
  └─► Write w → .devcontainer/post-create.sh

jkit init (skill deployment)
  └─► agents.DeploySkill(["claude","gemini"])
       ├─► Copy SkillsFS → .jkit/agents/skills/prd-creator/SKILL.md
       ├─► ln -s → ~/.claude/skills/prd-creator
       └─► ln -s → ~/.gemini/skills/prd-creator
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/agents/agents.go` | Modify | `ListAvailable()`, `DeploySkill()`, `agentSkillDirs` map |
| `internal/agents/markers.go` | Create | `ParsePostCreateMarkers()`, marker constants |
| `internal/agents/agents_test.go` | Create | Unit tests for discovery, parsing, validation |
| `cmd/jkit/agents.go` | Create | Cobra `agentsCmd` + `list|add|remove` subcommands |
| `cmd/jkit/init.go` | Modify | Wire `DeploySkill()` after post-create.sh generation |
| `internal/devcontainer/renderer.go` | Modify | Wrap agent snippets with marker delimiters |

## Interfaces / Contracts

```go
// internal/agents/agents.go
func ListAvailable() ([]string, error)
func DeploySkill(agents []string) error

var agentSkillDirs = map[string]string{
    "claude":   filepath.Join(os.Getenv("HOME"), ".claude", "skills"),
    "opencode": filepath.Join(os.Getenv("HOME"), ".config", "opencode", "skills"),
    "gemini":   filepath.Join(os.Getenv("HOME"), ".gemini", "skills"),
}

// internal/agents/markers.go
const AgentSectionPrefix = "# --- Agent installations ---\n"
const AgentStartFmt     = "# --- agent:%s ---\n"
const AgentEndFmt       = "# --- end agent:%s ---\n"

func ParsePostCreateMarkers(content string) []string
```

Marker output format in `renderPostCreate`:
```bash
# --- Agent installations ---
# --- agent:claude ---
# (claude installation script)
# --- end agent:claude ---
```

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | `ListAvailable()` | Table-driven with `testing/fstest.MapFS`: .sh files, non-.sh files, empty FS |
| Unit | `ParsePostCreateMarkers()` | Markers present, absent, partial, empty content, all-removed edge case |
| Unit | `DeploySkill()` | Temp dir: verify copy + symlinks created; existing symlinks recreated |
| Renderer | Agent markers in output | Extend `TestRender_PostCreateSh_AgentFiltering` to check `# --- agent:claude ---` |
| CLI | Subcommand routing | Via `rootCmd.Execute()` — same pattern as `main_test.go` (RunE, buf capture) |

## Migration / Rollout

No migration required. Existing post-create.sh files without agent markers are treated as "no agents installed" — the `gentle-ai` section (in the template header, not agent section) is always preserved per R-AGNT-04.

## Open Questions

None.
