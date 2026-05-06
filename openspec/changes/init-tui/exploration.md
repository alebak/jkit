## Exploration: jkit init TUI + Orchestration

### Current State

`jkit init` is a stub that deploys skills (via `--agents` flag) and prints "not yet implemented". Flags are already registered (`--name`, `--image`, `--quickstart`, `--agents`, `--timezone`) but their values are never used. All four infrastructure components are fully working and tested independently:

| Component | Status | Entry Point |
|-----------|--------|-------------|
| **DEVC** — `internal/devcontainer/renderer.go` | ✅ `Render()` writes 7 template files to `io.Writer` | `devcontainer.Render(ctx, w, name, data)` |
| **AGNT** — `internal/agents/agents.go` | ✅ `DeploySkill()` copies prd-creator + symlinks per agent | `agents.DeploySkill(ctx, projectDir, skillDir, skillName)` |
| **EXTG** — `internal/generator/generator.go` | ✅ `Generate()` walks embedded templates and renders | `generator.Generate(ctx, data, targetDir)` |
| **MCPS** — `internal/mcp/config.go` | ✅ `DeployMCP()` merges entries into per-agent mcp.json | `mcp.DeployMCP(ctx, configPath, name, templateData)` |

Dependencies present: `github.com/spf13/cobra`, `gopkg.in/yaml.v3`. No `huh` or `bubbletea` yet. `images.yaml` exists at repo root with 4 curated Joomla Apache images.

### Affected Areas

- `cmd/jkit/init.go` — **Main target**: replace stub with TUI (no flags) or parameterized (flags) orchestration
- `go.mod` — New dependency: `github.com/charmbracelet/huh` (plus transitive deps: `bubbletea`, `lipgloss`, etc.)
- `internal/devcontainer/devcontainer.go` — May need `ImageEntry` struct or images loading function for TUI
- `openspec/changes/init-tui/exploration.md` — This file
- (No other existing files need modification — all four infra components are consumed as-is)

### Approaches

1. **All-in-one in init.go RunE** — Implement everything in `cmd/jkit/init.go`: detect flags vs TTY, collect inputs, call DEVC → AGNT → EXTG → MCPS
   - Pros: Single file change pattern (matches create.go, agents.go); straightforward control flow
   - Cons: `init.go` becomes ~200+ lines; TUI and orchestration logic mixed; harder to test
   - Effort: Medium

2. **Extract to internal/init package** — New `internal/init/` package with `Run()` or `Orchestrate()`, init.go just parses flags and delegates
   - Pros: Separates concerns (CLI vs orchestration vs TUI); testable without cobra; aligns with PRD §5 which shows `internal/init/`
   - Cons: New package; more files; slightly more complexity
   - Effort: Medium

3. **Direct orchestration in cmd — delegate TUI to a helper** — Keep orchestration in init.go RunE, extract TUI form to `internal/init/tui.go` and `internal/init/images.go`
   - Pros: TUI forms isolatable; orchestration stays visible in the CLI layer; balanced complexity
   - Cons: Still some mixing of concerns
   - Effort: Medium

### Recommendation

**Approach 2 — Extract to `internal/init/`**. Rationale:

- The PRD already anticipates `internal/init/` in its architecture diagram (§5). Keeping parity with the documented structure reduces future confusion.
- Orchestrating 4 independent components (DEVC, AGNT, EXTG, MCPS) with file I/O, overwrite checks, and error handling is non-trivial — it deserves its own package.
- The TUI form logic (huh forms for project name, image selection, agent multiselect, confirm) is cleanly isolatable from the orchestration flow.
- Testing `init.go` via cobra RunE is painful (flag munging, output capture). Testing an `internal/init` package with direct function calls is significantly easier.

**Architecture:**

```
cmd/jkit/init.go
  └─► parses flags or detects TTY
      ├── TTY mode:     init.RunInteractive(ctx, cwd) → collects InitConfig
      └── Flag mode:    init.RunParameterized(ctx, cwd, flags) → builds InitConfig
          └─► init.Orchestrate(ctx, cwd, config) → error
               ├── 1. DEVC: devcontainer.Render() × 7 files
               ├── 2. AGNT: agents.DeploySkill() per selected agent
               ├── 3. EXTG: generator.Generate() with default 'component'
               ├── 4. MCPS: mcp.DeployMCP() for playwright + mariadb
               └── 5. Create builds/ directory
```

**Key Decisions:**

| Decision | Choice | Rationale |
|----------|--------|-----------|
| TUI library | `huh` | Per DD-01. Already evaluated in foundation exploration. Bubbletea as fallback only if `huh` cannot express the flow. |
| Image loading | Parse `images.yaml` at runtime with `gopkg.in/yaml.v3` | Reuse existing dep. The foundation exploration speculated remote fetch + cache, but for MVP local file is sufficient. |
| Overwrite check | `Confirm` prompt in TUI; `--force` flag in parameterized | Per R-INIT-08 / R-DEVC-10. Must check `.devcontainer/` dir exists before generating. |
| Default extension | Generate a `component` named after the project | Per PRD — every project needs an initial extension, component is the most common and complex. |
| Default agents | Prompt via MultiSelect (none pre-selected) | Per R-INIT-09 / R-AGNT-07. User chooses. In parameterized mode, `--agents` is optional. |
| MCPS deployment | Default to all available MCPs (playwright, mariadb) | Per PRD these are "by default" in §3. R-MCPS-06 says don't install unrequested, but init is the explicit request. |
| Timezone | Default `UTC` | Per existing `DefaultDevcontainerData()` and current `--timezone` flag default. |

**InitConfig struct (proposed):**

```go
type InitConfig struct {
    ProjectName    string
    JoomlaImage    string
    Timezone       string
    Agents         []string
    QuickstartPath string   // empty = no quickstart
    Force          bool     // skip overwrite prompts
}
```

**Orchestrate() flow:**

```
Orchestrate(ctx, cwd, cfg):
  1. if cfg.QuickstartPath != "":
       - Detect .zip, extract as project base
  2. Check .devcontainer/ exists → if yes, prompt overwrite (unless cfg.Force)
  3. Build DevcontainerData from cfg + defaults
  4. Create directories: .devcontainer/, builds/
  5. DEVC: for each of 7 templates → Render(ctx, fileWriter, name, data)
     - Write each to .devcontainer/{name}
  6. AGNT: for each selected agent → DeploySkill(ctx, cwd, skillDir, "prd-creator")
  7. EXTG: generate a default component:
       generator.Generate(ctx, ExtensionData{Name: cfg.ProjectName, Type: component}, cwd)
  8. MCPS: for each default MCP → DeployMCP(ctx, configPath, name, templateData)
       - Target agent: use first selected agent, or opencode as fallback
  9. Success message with instructions
```

**File plan:**

| File | Action | Description |
|------|--------|-------------|
| `internal/init/orchestrate.go` | Create | `Orchestrate()` — calls DEVC, AGNT, EXTG, MCPS in sequence |
| `internal/init/tui.go` | Create | `RunInteractive()` — huh forms collecting all inputs |
| `internal/init/init.go` | Create | `Config` struct, `RunParameterized()`, `ImageEntry` + `LoadImages()` |
| `internal/init/init_test.go` | Create | Unit tests for orchestration flow (temp dir, mock or call real Render) |
| `cmd/jkit/init.go` | **Major rewrite** | Parse flags, detect TTY, delegate to `init.RunInteractive()` or `init.RunParameterized()`, then `init.Orchestrate()` |
| `internal/devcontainer/devcontainer.go` | Minor | May export `DefaultDevcontainerData()` already exists; consider if `ImageEntry` type belongs here or in `internal/init` |
| `go.mod` | Update | Add `github.com/charmbracelet/huh` (and transitive deps) |
| `go.sum` | Update | Auto-generated by `go mod tidy` |

**TUI Form Flow (huh):**

```
Step 1: Input  ── Project name (required, validated non-empty)
Step 2: Select ── Joomla image (from parsed images.yaml, custom option)
Step 3: MultiSelect ── AI agents (from agents.ListAvailable, none pre-selected)
Step 4: Input  ── Timezone (default "UTC")
Step 5: Confirm ── Overwrite existing .devcontainer/ (if exists)
```

**Testing strategy:**

| Layer | What | Approach |
|-------|------|----------|
| Unit | `Orchestrate()` | Temp dir, verify 7 files created in `.devcontainer/`, verify `builds/` dir, verify component generated |
| Unit | `LoadImages()` | Parse real `images.yaml`, verify 4 entries, verify each has `-apache` |
| Unit | Input validation | Empty name → error, invalid timezone → warning, etc. |
| CLI | Flag routing | `--name X --image Y` → parameterized mode, no flags → TTY detection |
| CLI | Overwrite check | Existing `.devcontainer/` without `--force` → error (non-TTY) |

**huh dependency impact:**

`huh` pulls in `bubbletea`, `lipgloss`, and several Charm libraries. This increases binary size but is acceptable for a CLI tool (Go statically compiles). The dependency graph is well-maintained and backwards-compatible. If binary size is a concern, bubbletea direct is lighter but requires more TUI boilerplate — `huh` gives us the Vite-like TUI UX with minimal code.

**Risks:**

- **Testing with huh**: `huh` forms are inherently interactive. Strategies: (a) isolate form logic from execution so tests can bypass forms, (b) use `huh.NewForm().WithProgramOptions(bubbletea.WithInput())` with mock input, (c) accept that TUI tests are integration-level. Recommendation: use strategy (a) — `RunInteractive()` returns `InitConfig`, `Orchestrate()` consumes `InitConfig`.
- **images.yaml parsing**: Currently `gopkg.in/yaml.v3` is a transitive dep (used by `internal/generator/registry.go`). No explicit `yaml` import in `internal/devcontainer`. The image type should live in `internal/init` or a shared location.
- **Quickstart ZIP detection**: Auto-detecting `.zip` in CWD (R-INIT-04) must be a non-breaking scan — if multiple `.zip` files exist, the TUI should let the user choose. For MVP, raise an error if ambiguous.
- **Component naming**: The default component uses the project name. Must sanitize for Joomla conventions (no spaces/hyphens in `com_` name). `generator.SanitizeName()` and `NewExtensionData()` already handle this.

### Ready for Proposal

Yes. All dependencies, data flows, and architecture decisions are clear. The orchestrator should proceed with `sdd-propose`.

**Key point for the user:** The main architecture decision is whether to use `huh` (Charmbracelet — per DD-01, Vite-like UX, adds ~5 dependencies) vs `bubbletea` directly (lighter, more control, more boilerplate). The foundation exploration already landed on `huh` as first choice. No reason to revisit — proceed with `huh`.
