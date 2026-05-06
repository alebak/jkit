# Proposal: jkit init Interactive TUI

## Intent

Replace the `jkit init` stub with a full interactive TUI (huh) + parameterized mode that orchestrates Devcontainer setup, agent skill deployment, extension generation, and MCP configuration. This is P0 — the main product entry point.

## Scope

### In Scope
- Interactive TUI with huh: project name, Joomla image, AI agents, timezone, quickstart path
- Parameterized mode: direct flag input (`--name`, `--image`, `--agents`, `--timezone`, `--quickstart`, `--force`)
- Orchestration engine: DEVC Render × 7 → AGNT DeploySkill per agent → EXTG Generate component → MCPS DeployMCP (playwright + mariadb)
- `images.yaml` parsing via `gopkg.in/yaml.v3` for image selection
- Quickstart `.zip` auto-detection with error on ambiguity
- Overwrite protection: TUI prompt or `--force` flag
- New `internal/init/` package (orchestrate.go, tui.go, config.go)
- `cmd/jkit/init.go` rewrite: detect TTY vs flags, dispatch

### Out of Scope
- Package type grouping (R-EXTG-07)
- `images.yaml` caching (R-DEVC-13)
- Remote image registry fetching
- Agent config persistence (no state file)

## Capabilities

### New Capabilities
- `init-tui`: Interactive project initialization flow with huh forms, orchestration, and parameterized fallback

### Modified Capabilities
- `cli-commands`: `jkit init` changes from stub to full implementation with TUI/flag dispatch
- `devcontainer-init`: No spec-level change — orchestration calls existing `Render()` unchanged

## Approach

New `internal/init/` package with three files:
- **config.go**: `InitConfig` struct, `ImageEntry`, `LoadImages()` from `images.yaml`
- **tui.go**: `RunInteractive()` → collects `InitConfig` via huh forms (5 steps)
- **orchestrate.go**: `Orchestrate()` → DEVC (7 files) → AGNT (per agent) → EXTG (component) → MCPS (playwright+mariadb) → builds/

`cmd/jkit/init.go`: detect TTY or flags, call `RunInteractive()` or `RunParameterized()`, then `Orchestrate()`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/jkit/init.go` | Rewrite | Flag parsing, TTY detection, dispatch to internal/init |
| `internal/init/` | New | orchestrate.go, tui.go, config.go (+ test) |
| `go.mod` / `go.sum` | Update | Add `github.com/charmbracelet/huh` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|-------------|
| huh dependency tree adds bloat | Med | Acceptable for CLI tool; monitor in PR review |
| Quickstart ZIP ambiguous (>1 .zip) | Low | Error: "multiple .zip files found" |
| TUI vs TTY detection flaky | Low | Use `os.Stdout` file mode + `os.IsTerminal` |

## Rollback Plan

Revert `cmd/jkit/init.go`, `rm -rf internal/init/`, `git checkout go.mod go.sum`. Previous stub is minimal — no data loss.

## Dependencies

- `github.com/charmbracelet/huh` (new — pull in as explicit dep)
- Existing: `internal/devcontainer`, `internal/agents`, `internal/generator`, `internal/mcp`
- `gopkg.in/yaml.v3` (already transitive dep via internal/generator)

## Success Criteria

- [ ] `jkit init` with no flags launches huh TUI in TTY session
- [ ] `jkit init --name X --image Y` runs parameterized without TUI
- [ ] `Orchestrate()` creates `.devcontainer/` with 7 files + `builds/` dir
- [ ] Selected agents get prd-creator skill deployed
- [ ] Default component generated for project name
- [ ] MCPs deployed for first selected agent (or opencode fallback)
- [ ] All existing tests pass
