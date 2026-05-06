# Tasks: jkit init Interactive TUI

## Execution Plan

7 tasks, ~400 lines total, 1 PR.

## Files

| # | File | Action | Est. Lines |
|---|------|--------|------------|
| 1 | `internal/init/config.go` | Create | 80 | ✅ |
| 2 | `internal/init/tui.go` | Create | 90 | ✅ |
| 3 | `internal/init/orchestrate.go` | Create | 100 | ✅ |
| 4 | `internal/init/init_test.go` | Create | 80 | ✅ |
| 5 | `cmd/jkit/init.go` | Rewrite | 70 | ✅ |
| 6 | `go.mod` / `go.sum` | Update | auto | ✅ |

---

### [x] Task 1: InitConfig + images.yaml parsing

**File**: `internal/init/config.go`

**What**:
- `ImageEntry` struct with `Tag` and `Description` (yaml tags)
- `ParseImagesYAML(fsys fs.FS) ([]ImageEntry, error)` — reads and parses `images.yaml`
  - Expects top-level key `images:` containing array
  - Returns error on malformed YAML or empty list
- `InitConfig` struct with 6 fields (`ProjectName`, `JoomlaImage`, `Agents []string`, `Quickstart`, `Timezone`, `Force`)
- `DefaultInitConfig() InitConfig` — sets `JoomlaImage` to `"joomla:6.1-php8.4-apache"`, `Timezone` to `"UTC"`, others zero
- `ToDevcontainerData() devcontainer.DevcontainerData` — maps fields:
  - `ProjectName` → `DevcontainerData.ProjectName`
  - `JoomlaImage` → `DevcontainerData.JoomlaImage`
  - `Timezone` → `DevcontainerData.Timezone`
  - `Agents` → `DevcontainerData.SelectedAgents`
  - All other DevcontainerData fields use `DefaultDevcontainerData()` values
- Quickstart detection helper: `detectQuickstart(dir string) (string, error)` — globs `*.zip` in dir, returns error if 0 or >1 found

**References**: R-INIT-TUI-04, R-INIT-TUI-05

---

### [x] Task 2: TUI forms with huh

**File**: `internal/init/tui.go`

**What**:
- `RunInteractive(ctx context.Context) (InitConfig, error)` function
- 6 huh form steps chained as groups:
  1. `huh.NewInput()` — "Project name" (required, non-empty validation via `huh.WithInputValidate`)
  2. `huh.NewSelect[ImageEntry]()` — "Joomla image" with options from `ParseImagesYAML(jkit.DevcontainerFS)`
  3. `huh.NewMultiSelect[string]()` — "AI agents" with options from `agents.ListAvailable(ctx, jkit.AgentsFS)`
  4. `huh.NewInput()` — "Timezone" with default `"UTC"`
  5. `huh.NewInput()` — "Quickstart .zip path" (optional)
  6. Overwrite check before final confirm:
     - If `.devcontainer/` exists and `!Force`: `huh.NewConfirm()` — "Overwrite existing .devcontainer/?"
     - If denied: return `fmt.Errorf("aborted by user")`
  7. `huh.NewConfirm()` — "Create project?"
     - If denied: return `fmt.Errorf("aborted by user")`
- Returns fully populated `InitConfig`

**References**: R-INIT-TUI-01, R-INIT-TUI-06

---

### [x] Task 3: Orchestration engine

**File**: `internal/init/orchestrate.go`

**What**:
- `Orchestrate(ctx context.Context, cfg InitConfig) error`
- Fail-fast sequential pipeline:
  1. **Overwrite guard**: `os.Stat(".devcontainer/")` exists AND `!cfg.Force` → return error "use --force to overwrite existing .devcontainer/"
  2. **DEVC**: Create `.devcontainer/` dir, render 7 files (devcontainer.json, Dockerfile, docker-compose.yml, .env, .env.example, post-create.sh, php-custom.ini). Each: `devcontainer.Render(ctx, &buf, name, data)` then `os.WriteFile(path, buf.Bytes(), 0644)`. Track created files for rollback.
  3. **AGNT**: For each agent in `cfg.Agents`:
     - `agents.SkillDirFor(ctx, agent)` → skillDir
     - `agents.DeploySkill(ctx, cwd, skillDir, "prd-creator")`
  4. **EXTG**: Generate default component:
     - `generator.NewExtensionData(cfg.ProjectName, "jkit", generator.TypeComponent)`
     - `generator.Generate(ctx, data, "builds")`
  5. **MCPS**: If `len(cfg.Agents) > 0`:
     - Target first agent: `mcp.MCPConfigPathFor(ctx, cfg.Agents[0])` → configPath
     - `mcp.DeployMCP(ctx, configPath, "playwright", ...)` with embedded MCP template for playwright
     - `mcp.DeployMCP(ctx, configPath, "mariadb", ...)` with embedded MCP template for mariadb
     - MCP template data loaded from `jkit.MCPFS.ReadFile("templates/mcp/playwright.json")` etc.
  6. **Final**: `os.MkdirAll("builds", 0755)`
- **Rollback**: On any step failure, best-effort remove created files/dirs (tracked during DEVC step)

**References**: R-INIT-TUI-03

---

### [x] Task 4: Overwrite check + quickstart detection (in orchestrate)

**Note**: This is part of Task 3's orchestrate.go but deserves explicit testing attention.

**What**:
- Quickstart support in orchestrate: if `cfg.Quickstart != ""`, extract the `.zip` into CWD before DEVC step
  - Use `archive/zip` to extract
  - If `cfg.Quickstart == ""` AND quickstart flag was a bare boolean (no path), call `detectQuickstart()`
- Overwrite guard behavior:
  - Interactive: handled in TUI (Task 2), but orchestrate ALSO checks `--force`
  - Parameterized: orchestrate returns error if `.devcontainer/` exists and `!Force`
- Rollback cleanup on error: tracked file list + `os.RemoveAll` on `.devcontainer/` + builds/

**References**: R-INIT-TUI-05, R-INIT-TUI-06, R-INIT-TUI-07

---

### [x] Task 5: Wire cmd/jkit/init.go

**File**: `cmd/jkit/init.go` — rewrite

**What**:
- Keep existing `initCmd` and `init()` registration
- Register flags: `--name`, `--image`, `--agents`, `--timezone`, `--quickstart`, `--force`
- `RunE` logic:
  1. If `--name` is set OR any flag is explicitly changed → parameterized mode:
     - Validate `--name` is non-empty → error if empty
     - Build `InitConfig` from flags (`DefaultInitConfig()` + overrides)
     - Optional: validate `--agents` values against `agents.ListAvailable()` → error "unknown agent"
     - Call `init.Orchestrate(ctx, cfg)`
  2. Else if `os.IsTerminal(os.Stdout.Fd())` → TUI mode:
     - Call `init.RunInteractive(ctx)` → cfg
     - Call `init.Orchestrate(ctx, cfg)`
  3. Else → error: "run with --name or --image flags in non-TTY mode"
- All errors printed via `cmd.PrintErrln` or returned from `RunE`

**References**: R-INIT-TUI-02, R-INIT-TUI-07

---

### [x] Task 6: Add huh dependency

**File**: `go.mod`, `go.sum`

**What**:
- `go get github.com/charmbracelet/huh@latest`
- `go mod tidy` to update go.sum

---

### [x] Task 7: Tests

**File**: `internal/init/init_test.go`

**What**:
- **`TestParseImagesYAML`**: Table-driven with `fstest.MapFS`. Test valid YAML, empty list, malformed YAML, missing key.
- **`TestDefaultInitConfig`**: Verify defaults: empty ProjectName, "joomla:6.1-php8.4-apache" image, "UTC" timezone, nil agents, empty quickstart, false force.
- **`TestInitConfigToDevcontainerData`**: Full InitConfig → verify all DevcontainerData fields mapped correctly + non-overridden fields use defaults.
- **`TestDetectQuickstart`**: With temp dirs — single zip returns path, no zip returns error, multiple zips returns error.
- **`TestOrchestrateOverwriteCheck`**: Create temp dir with `.devcontainer/`, call orchestrate without Force → expect error. With Force → expect no error (will fail on Render since no templates, but that's OK — verify it passes the guard).
- **`TestOrchestrateFailFast`** (if feasible): Verify that when DEVC fails, AGNT is not called.

**References**: All R-INIT-TUI requirements

---

## Ordering

1. Task 6 (go.mod)
2. Task 1 (config.go)
3. Task 2 (tui.go)
4. Task 3 (orchestrate.go)
5. Task 4 (overwrite + quickstart, part of orchestrate.go)
6. Task 5 (cmd/jkit/init.go)
7. Task 7 (tests)
