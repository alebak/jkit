# Tasks: Agent Management

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~335 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | All 5 phases in one PR | PR 1 → main | ~335 lines, under budget, no split needed |

## Phase 1: Foundation — Agent Discovery + Marker Parsing

- [ ] 1.1 Add `ListAvailable()` to `internal/agents/agents.go` — reads `templates/agents/*.sh` from `jkit.AgentsFS`, returns agent names (filename minus `.sh`)
- [ ] 1.2 Create `internal/agents/markers.go` with marker constants (`AgentSectionPrefix`, `AgentStartFmt`, `AgentEndFmt`) and `ParsePostCreateMarkers(content string) []string` that extracts agent names from `# --- agent:<name> ---` delimiters

## Phase 2: Core — CLI + Skill Deployment

- [ ] 2.1 Create `cmd/jkit/agents.go` with `agentsCmd` + `list` (shows available vs installed), `add [name...]` (validates, re-renders post-create.sh), and `remove [name...]` (parses markers, subtracts, re-renders). Handle error paths per R-AGNT-03/R-AGNT-04.
- [ ] 2.2 Add `DeploySkill(agents []string) error` + `agentSkillDirs` map to `internal/agents/agents.go` — copies `templates/skills/prd-creator/SKILL.md` to `.jkit/agents/skills/prd-creator/` and creates symlinks from each agent's skill directory per R-AGNT-06.

## Phase 3: Integration — Markers + Init Wiring

- [ ] 3.1 Modify `renderPostCreate` in `internal/devcontainer/renderer.go` to wrap each agent snippet with `# --- agent:<name> ---` and `# --- end agent:<name> ---` delimiters per R-AGNT-05
- [ ] 3.2 Wire `DeploySkill()` into `cmd/jkit/init.go` after post-create.sh generation

## Phase 4: Tests + Verification

- [ ] 4.1 Create `internal/agents/agents_test.go` — table-driven tests for `ListAvailable` (`.sh` files, non-`.sh` files, empty FS), `ParsePostCreateMarkers` (present, absent, partial, empty, all-removed), and `DeploySkill` (copy + symlink, existing symlink recreated)
- [ ] 4.2 Extend `TestRender_PostCreateSh_AgentFiltering` in `internal/devcontainer/renderer_test.go` to assert `# --- agent:claude ---` markers in output
- [ ] 4.3 Add CLI tests in `cmd/jkit/main_test.go` for `jkit agents list|add|remove` subcommand routing via `rootCmd.Execute()`
