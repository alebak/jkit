# Design: jkit init Interactive TUI

## Technical Approach

New `internal/init/` package (config.go, tui.go, orchestrate.go) that provides the interactive and parameterized init flow. `cmd/jkit/init.go` detects TTY vs flags, dispatches to `RunInteractive()` or builds `InitConfig` from flags, then calls `Orchestrate()`.

The existing `devcontainer.Render()` writes to `io.Writer` — orchestration wraps it with per-file `os.Create()` to persist 7 devcontainer files. AGNT, EXTG, MCPS steps call existing package APIs directly.

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|-------------|-----------|
| Package location | `internal/init/` | Keep in cmd/ | Matches existing pattern (devcontainer, agents, etc.); keeps CLI thin |
| Form library | `huh` | `survey`, `promptui` | Charm ecosystem; native Go, active maintenance, themable |
| TTY detection | `os.IsTerminal(os.Stdout.Fd())` | Detect by flag presence | Flag presence is sufficient: if `--name` given, skip TUI regardless of TTY |
| Overwrite check | Check `os.Stat(".devcontainer/")` before orchestration | Always overwrite | R-INIT-TUI-06 requires prompt/error on existing dir |
| Quickstart detection | `filepath.Glob("*.zip")` in CWD | Require explicit path | R-INIT-TUI-05: single zip auto-detection, multiple → error |
| InitConfig → DevcontainerData | Fill fields from InitConfig, rest from defaults | Full manual mapping | `DefaultDevcontainerData()` provides all Joomla defaults; override only what user specified |
| Orchestration fail-fast | Sequential with early return on first error | Parallel fan-out | R-INIT-TUI-03: order matters (DEVC must finish before AGNT needs the dir) |

## Data Flow

```
TTY + no flags ──► RunInteractive() ──► InitConfig
                                             │
Flag(s) present ──► flag parsing ────────────┘
                                             │
                                             ▼
                                     Orchestrate(cfg)
                                      │
                          ┌───────────┼───────────┐
                          ▼           ▼           ▼
                     DEVC Render   AGNT Deploy  EXTG Generate
                      .devcontainer/  .jkit/agents/   builds/src/
                          │
                          ▼
                     MCPS DeployMCP
                      ~/.config/opencode/mcp.json
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/init/config.go` | Create | `InitConfig` struct, `ImageEntry`, `ParseImagesYAML()`, `DefaultInitConfig()`, `ToDevcontainerData()` |
| `internal/init/tui.go` | Create | `RunInteractive()` with huh forms (5 steps + overwrite confirm) |
| `internal/init/orchestrate.go` | Create | `Orchestrate()` — DEVC→AGNT→EXTG→MCPS pipeline with rollback |
| `internal/init/init_test.go` | Create | Table-driven tests for config parsing, orchestration, TUI not tested directly |
| `cmd/jkit/init.go` | Rewrite | Flag registration, TTY detection, dispatch to TUI or parameterized mode |
| `go.mod` / `go.sum` | Update | Add `github.com/charmbracelet/huh` dependency |

## Interfaces / Contracts

```go
// internal/init/config.go

type ImageEntry struct {
    Tag         string `yaml:"tag"`
    Description string `yaml:"description"`
}

type InitConfig struct {
    ProjectName string
    JoomlaImage string
    Agents      []string
    Quickstart  string   // path to .zip, empty if none
    Timezone    string
    Force       bool
}

// ParseImagesYAML reads and parses images.yaml from the given filesystem.
func ParseImagesYAML(fsys fs.FS) ([]ImageEntry, error)

// DefaultInitConfig returns InitConfig with defaults (empty ProjectName).
func DefaultInitConfig() InitConfig

// ToDevcontainerData maps InitConfig fields + defaults to DevcontainerData.
func (c InitConfig) ToDevcontainerData() devcontainer.DevcontainerData
```

```go
// internal/init/tui.go

// RunInteractive launches the huh TUI and returns the collected config.
func RunInteractive(ctx context.Context) (InitConfig, error)
```

```go
// internal/init/orchestrate.go

// Orchestrate runs the full init pipeline: DEVC → AGNT → EXTG → MCPS.
// Fails fast on first error; best-effort rollback on failure.
func Orchestrate(ctx context.Context, cfg InitConfig) error
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `ParseImagesYAML()` | Table-driven with test FS (fstest.MapFS) |
| Unit | `InitConfig.ToDevcontainerData()` | Verify field mapping and defaults |
| Unit | `Orchestrate()` overwrite check | Mock filesystem with temp dir |
| Unit | Overwrite prompt decision | Verify error when `.devcontainer/` exists without Force |
| CLI | Flag dispatch | Verify TUI vs parameterized dispatch logic |
| CLI | Flag validation | Missing `--name` in parameterized mode returns error |

## Migration / Rollout

No migration required. Existing `jkit init` is a stub with no persistent state.

## Open Questions

None.
