# Design: Project Foundation

## Technical Approach

Foundation change: establish the Go module, cobra CLI scaffold, template rendering engine for devcontainer generation, and all static assets (agent snippets, extension stubs, skill files, images.yaml). Subcommands are stubs that print "not yet implemented". This is the bedrock all future changes build on.

**Module**: `github.com/alebak/jkit` → Go 1.24  
**CLI**: cobra — root + 3 subcommands (init, create, build)  
**Renderer**: `text/template` with `ParseFS` against embedded templates  
**Templates**: 6 of 7 files become Go templates; `php-custom.ini` is static (passed through raw)

---

## Architecture Decisions

### Decision: Template Embedding — directory + explicit patterns

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Wildcard `*` per file | Simpler embed but misses dotfiles (.env, .env.example) | ❌ |
| Directory embed + explicit ParseFS patterns | One embed directive; patterns are explicit and reliable | ✅ |

**Rationale**: `//go:embed templates/devcontainer` embeds all files recursively including dotfiles (Go 1.24). `template.ParseFS` then receives 7 explicit relative paths — no glob ambiguity, no missed dotfiles. Agents directory uses `templates/agents/*.sh` since those are NOT templates and need no parsing.

```go
//go:embed templates/devcontainer
var templateFS embed.FS

patterns := []string{
    "templates/devcontainer/devcontainer.json",
    "templates/devcontainer/Dockerfile",
    "templates/devcontainer/docker-compose.yml",
    "templates/devcontainer/.env",
    "templates/devcontainer/.env.example",
    "templates/devcontainer/php-custom.ini",
    "templates/devcontainer/post-create.sh",
}
tmpl := template.Must(template.New("devcontainer").ParseFS(templateFS, patterns...))
```

### Decision: DevcontainerData — full .env coverage

**Rationale**: The `.env` and `.env.example` files have 12 variables. Storing defaults for all of them in the struct means the renderer needs zero external config beyond user-provided `ProjectName`. Defaults per R-INIT-02 (superdev/superpassword).

```go
type DevcontainerData struct {
    ProjectName    string   // JOOMLA_SITE_NAME
    JoomlaImage    string   // Docker image tag
    Timezone       string   // TZ (default "UTC")
    AdminUser      string   // defaults "superdev"
    AdminUsername  string   // defaults "superdev"
    AdminPassword  string   // defaults "superpassword"
    AdminEmail     string   // defaults "admin@example.com"
    DBUser         string   // derived from ProjectName
    DBPassword     string   // defaults "joomla"
    DBName         string   // derived from ProjectName
    DBPrefix       string   // derived from ProjectName
    SMTPHost       string   // defaults "mail:1025"
    RootPassword   string   // defaults "dev"
    VSCodeExtensions []string // defaults: xdebug, intelephense, prettier
}
```

### Decision: post-create.sh — two-pass (header + agent bash)

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Single template with all agents | Bloated; agents are not Go templates | ❌ |
| Two-pass: template header + raw bash concatenation | Clean separation; agents stay plain bash | ✅ |

**Pass 1**: Render Go template header — installs gentle-ai unconditionally (R-AGNT-01), sets up npm prefix, `{{.ProjectName}}` in echo messages.  
**Pass 2**: Read and concatenate `<name>.sh` from `templates/agents/` for each selected agent. Snippets are raw bash, embedded via `//go:embed templates/agents/*.sh`.

### Decision: php-custom.ini — static passthrough

**Rationale**: No Go template directives — Xdebug config is environment-agnostic. Parsing it as a template is harmless but wasteful. The renderer copies it byte-for-byte from embedded FS. Achieved by omitting it from template ParseFS and using `fs.ReadFile` instead.

### Decision: CLI external deps — cobra only (no huh in foundation)

**Rationale**: `jkit init/create/build` are stubs — they print "not yet implemented". The `huh` dependency is needed only when interactive mode is implemented (future change). Foundation keeps deps minimal: `cobra` + `yaml.v3`.

---

## Package Dependency Diagram

```
cmd/jkit/main.go
    └── cmd/jkit/root.go
            ├── cmd/jkit/init.go
            │       └── internal/init
            │               ├── templates.go  (//go:embed)
            │               ├── devcontainer.go  (DevcontainerData + defaults)
            │               └── renderer.go  (Render func)
            ├── cmd/jkit/create.go
            └── cmd/jkit/build.go

internal/
    ├── init/        ← full implementation
    ├── generator/   ← stub only
    ├── agents/      ← stub only
    └── mcp/         ← stub only

External deps:
    github.com/spf13/cobra     (CLI)
    gopkg.in/yaml.v3           (images.yaml parsing)
```

---

## Data Flow

```
1. jkit init (no args)
    │
    ▼
2. cobra dispatches → init.go  ──►  prints "not yet implemented"
   (future: launches huh TUI)

─────────────────────────────────────────

1. internal/init/renderer.Render(w, name, data)
    │
    ├── Load embedded templates (init once via sync.Once)
    │
    ├── name == "php-custom.ini"
    │       └── fs.ReadFile(templateFS, "templates/devcontainer/php-custom.ini")
    │           → w.Write(rawBytes)
    │
    └── name != "php-custom.ini"
            └── tmpl.ExecuteTemplate(w, "templates/devcontainer/"+name, data)
                → substituted output on w

─────────────────────────────────────────

1. post-create.sh generation
    │
    ├── Render template header (gentle-ai install, npm prefix, {{.ProjectName}})
    │
    ├── For each selected agent:
    │     └── Read templates/agents/<name>.sh → append to output
    │
    └── Write combined bash to w
```

---

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `go.mod` | Create | Module `github.com/alebak/jkit`, Go 1.24, deps: cobra, yaml.v3 |
| `cmd/jkit/main.go` | Create | Entry point, calls `root.Execute()` |
| `cmd/jkit/root.go` | Create | Cobra root command with `jkit` |
| `cmd/jkit/init.go` | Create | Subcommand: prints "not yet implemented" |
| `cmd/jkit/create.go` | Create | Subcommand: prints "not yet implemented" |
| `cmd/jkit/build.go` | Create | Subcommand: prints "not yet implemented" |
| `internal/init/templates.go` | Create | `//go:embed` directives + template FS variable |
| `internal/init/devcontainer.go` | Create | `DevcontainerData` struct + `DefaultDevcontainerData()` |
| `internal/init/renderer.go` | Create | Public `Render(io.Writer, string, any) error` func |
| `internal/init/renderer_test.go` | Create | Table-driven tests per AGENTS.md |
| `internal/generator/generator.go` | Create | Package stub with doc comment |
| `internal/agents/agents.go` | Create | Package stub with doc comment |
| `internal/mcp/mcp.go` | Create | Package stub with doc comment |
| `images.yaml` | Create | 4 curated Joomla Apache images |
| `templates/agents/claude.sh` | Create | Claude Code install bash snippet |
| `templates/agents/opencode.sh` | Create | OpenCode install bash snippet |
| `templates/agents/gemini.sh` | Create | Gemini CLI install bash snippet |
| `templates/extensions/{6}/.gitkeep` | Create | Empty dir stubs (component, module, plugin, template, library, package) |
| `templates/skills/prd-creator/skill.md` | Create | PRD creator skill content |
| `templates/devcontainer/devcontainer.json` | Modify | `"elrepuestazo.com"` → `{{.ProjectName}}` |
| `templates/devcontainer/Dockerfile` | Modify | Hardcoded image → `{{.JoomlaImage}}` |
| `templates/devcontainer/docker-compose.yml` | Modify | Keep `env_file:` + env var refs; no template changes needed (env vars served by .env) |
| `templates/devcontainer/.env` | Modify | All values → `{{.Variable}}` placeholders |
| `templates/devcontainer/.env.example` | Modify | Same structure as .env with `{{.Variable}}` |
| `templates/devcontainer/php-custom.ini` | No change | Static file, no template vars |
| `templates/devcontainer/post-create.sh` | Modify | Remove hardcoded agent installs; add `{{.ProjectName}}` header |
| `.gitignore` | Create | Add `.atl/` |

---

## Interfaces / Contracts

### internal/init — Renderer

```go
package init

import "io"

// DevcontainerData holds all values for devcontainer template rendering.
// Zero-value defaults are set via DefaultDevcontainerData().
type DevcontainerData struct {
    ProjectName       string
    JoomlaImage       string
    Timezone          string
    AdminUser         string
    AdminUsername     string
    AdminPassword     string
    AdminEmail        string
    DBUser            string
    DBPassword        string
    DBName            string
    DBPrefix          string
    SMTPHost          string
    RootPassword      string
    VSCodeExtensions  []string
    SelectedAgents    []string // agent names for post-create.sh
}

// DefaultDevcontainerData returns a DevcontainerData with safe defaults.
// Only ProjectName must be overwritten by the caller.
func DefaultDevcontainerData() DevcontainerData { ... }

// Render executes the named template against data and writes to w.
// If name is "php-custom.ini", the file is copied verbatim (no template processing).
// Returns error if the template name does not exist.
func Render(w io.Writer, name string, data any) error { ... }
```

### images.yaml schema

```yaml
images:
  - tag: string
    description: string
```

Parsed at init-time into:
```go
type ImageEntry struct {
    Tag         string `yaml:"tag"`
    Description string `yaml:"description"`
}
type ImagesConfig struct {
    Images []ImageEntry `yaml:"images"`
}
```

---

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | `Render(w, name, data)` | Table-driven: valid templates → correct output; unknown name → error; php-custom.ini → raw bytes; all 7 templates → non-empty; no "El Repuestazo" leaks |
| Unit | `DefaultDevcontainerData()` | Verify defaults: superdev/superpassword, UTC timezone, proper derived DB fields |
| Unit | Cobra stubs | Execute each subcommand, verify "not yet implemented" in stdout + exit 0 |
| Build | `go build ./...` | Must pass with no errors |
| Lint | `go vet ./...` | Must pass with no errors |
| Dependencies | `go mod tidy` | Must produce clean go.sum |

All tests use `testing/fstest` for filesystem operations per AGENTS.md — no mocking of the filesystem.

---

## Migration / Rollout

No migration required — this is the first compilable state of the project. After this change, `go build ./...` will produce a working `jkit` binary for the first time.

---

## Open Questions

- [ ] Confirm exact set of YAML tags for `images.yaml` (joomla:6.1-php8.4-apache, 5.3-php8.3-apache, 5.3-php8.4-apache, 6.1-php8.5-apache — verify all exist on Docker Hub)
- [ ] Decide whether `opencode` install snippet uses `npm isntall -g` or `curl | bash` (Claude Code already migrated; OpenCode may follow)
