# Exploration: Project Foundation

**Change:** project-foundation
**Date:** 2026-05-05
**Author:** sdd-explore agent

---

## Current State

The JKit project is at **absolute greenfield** on the Go side:

- **No Go files exist** — no `go.mod`, no `go.sum`, no `.go` source files anywhere.
- **Go 1.24.13** is available in the dev environment.
- **`git remote`** confirms `github.com/alebak/jkit` (DD-07).
- **SDD init** has been run — project conventions, testing capabilities, and tech stack are documented in engram.
- **Only 2 commits** exist: the initial structure and a minor chore.

### Templates (templates/devcontainer/) — 7 files, all with hardcoded values

The templates currently contain **real credentials from the author's "El Repuestazo" project**. Every file needs Go template placeholders.

| File | Current State | Hardcoded Values |
|---|---|---|
| `devcontainer.json` | Raw JSON | `"name": "elrepuestazo.com"`, specific ports, PHP VSCode extensions |
| `docker-compose.yml` | 4 services | `TZ: UTC` on mail service, `mariadb:11.4.10`, `axllent/mailpit` |
| `Dockerfile` | Dockerfile | `FROM joomla:6.1-php8.4-apache` |
| `.env` | Real credentials | ALL values — site name, admin user, passwords, DB creds, etc. |
| `.env.example` | Example credentials | All values, but safe examples |
| `php-custom.ini` | Static config | Xdebug settings — no placeholders needed |
| `post-create.sh` | Bash script | Agent installs for Claude Code, OpenCode, gentle-ai |

### .devcontainer/ (JKit's own dev environment) — 4 files

These are for developing JKit itself (Go 1.24-trixie). NOT templates for generated projects. Must NOT be touched by DEVC.

### Missing directories

The following directories from the PRD architecture diagram do NOT exist yet:
- `cmd/jkit/` — CLI entry point
- `internal/init/`, `internal/generator/`, `internal/agents/`, `internal/mcp/`
- `templates/extensions/` — extension skeletons (6 types)
- `templates/agents/` — agent installation snippets
- `templates/skills/prd-creator/` — prd-creator skill
- `images.yaml` — curated Joomla image list
- `scripts/` — install.sh

---

## Affected Areas

| Path | Why Affected |
|---|---|
| `go.mod` | **Must be created.** Module path: `github.com/alebak/jkit`. Go 1.24. |
| `cmd/jkit/main.go` | **Must be created.** Root cobra command. Entry point. |
| `cmd/jkit/init.go` | **Must be created.** `jkit init` command with interactive + parameterized modes. |
| `internal/init/` | **Must be created.** Core init orchestration logic — calls DEVC, AGNT, EXTG, MCPS in order. |
| `internal/init/renderer.go` | **Must be created.** Template rendering engine for DEVC (reads embedded templates, executes Go templates, writes output). |
| `templates/devcontainer/*` | **7 files must be modified.** Replace all hardcoded values with `{{.GoTemplate}}` placeholders. |
| `openspec/changes/project-foundation/` | Change tracking directory created by this exploration. |
| `internal/generator/` | Placeholder only in this change — actual implementation deferred. |
| `internal/agents/` | Placeholder only in this change — actual implementation deferred. |
| `internal/mcp/` | Placeholder only in this change — actual implementation deferred. |

---

## Approaches

### 1. Template Engine Architecture: Go `html/template` vs `text/template`

| Aspect | Decision |
|---|---|
| **Choice** | `text/template` |
| **Why** | We're generating config files (JSON, YAML, INI, bash, Dockerfile, .env), NOT HTML. `text/template` won't escape valid characters that `html/template` would. |
| **Key types** | `template.FuncMap` for utility funcs (e.g., `lower`, `default`, `envfile` for .env escaping) |
| **Embed** | `embed.FS` via `//go:embed templates/devcontainer/*` — separate patterns for dotfiles |

**Template data struct:**
```go
type DevcontainerData struct {
    ProjectName    string
    JoomlaImage    string
    Timezone       string
    SiteName       string
    AdminUser      string
    AdminUsername  string
    AdminPassword  string
    AdminEmail     string
    DBUser         string
    DBPassword     string
    DBName         string
    DBPrefix       string
    RootPassword   string
    Extensions     []string  // VSCode extensions
}
```

### 2. Dotfile Handling in go:embed

**Problem:** `//go:embed templates/devcontainer/*` does NOT match `.env` or `.env.example` (dotfiles).

**Solution:** Use explicit patterns:
```go
//go:embed templates/devcontainer/devcontainer.json
//go:embed templates/devcontainer/docker-compose.yml
//go:embed templates/devcontainer/Dockerfile
//go:embed templates/devcontainer/.env
//go:embed templates/devcontainer/.env.example
//go:embed templates/devcontainer/php-custom.ini
//go:embed templates/devcontainer/post-create.sh
var devcontainerTemplates embed.FS
```

Then iterate via `fs.ReadDir` on the embedded filesystem. Access file content with `devcontainerTemplates.ReadFile(path)`.

### 3. Cobra Command Structure for `jkit init`

```
jkit
├── init                          # interactive or parameterized
│   ├── --name, -n                # project name (required)
│   ├── --image, -i               # Joomla image tag
│   ├── --quickstart, -q          # path to quickstart .zip
│   ├── --agents, -a              # comma-separated agents
│   └── --timezone, -tz           # timezone
├── create                        # parent command
│   ├── component                 # jkit create component
│   ├── module                    # jkit create module
│   ├── plugin                    # jkit create plugin
│   ├── template                  # jkit create template
│   ├── library                   # jkit create library
│   └── package                   # jkit create package
├── agents                        # parent command (future)
│   ├── add                       # jkit agents add [name]
│   └── remove                    # jkit agents remove [name]
├── mcp                           # parent command (future)
│   └── add                       # jkit mcp add [name]
└── build                         # jkit build [name] (future)
```

**Approach for flags vs positional args:**
- `jkit init` with NO flags → interactive TUI mode with huh
- `jkit init --name foo ...` → parameterized mode, validate required flags
- `jkit create component` → positional arg for type, flags for name
- Keep it simple: flags for init (R-INIT-01 through R-INIT-04), positional for create (R-INIT-10)

### 4. .env Template Handling

Both `.env` and `.env.example` are Go templates. Approach:
- **`.env` template**: Rendered with user-provided values (real credentials). Template variables everywhere. Output has `{{.AdminPassword}}` etc.
- **`.env.example` template**: Rendered with SANITIZED example values (safe for git). Template has the same structure but defaults won't expose real creds.
- **`.gitignore`**: Generated/updated by DEVC to include `.env` (R-DEVC-06). Add `.env` entry.

### 5. post-create.sh Generation (Two Approaches)

**Approach A — Concatenation then Optional Template (Recommended)**
- Individual agent scripts (`templates/agents/*.sh`) are plain bash snippets, NOT Go templates
- post-create.sh = static header + concatenated agent snippets
- If post-create.sh itself needs {{.ProjectName}} (for echo messages, log prefixes), template it AFTER concatenation
- Pro: Agent snippets stay independent, can be tested standalone
- Con: Two-pass processing

**Approach B — Everything is a template**
- Each agent .sh is a Go template too
- post-create.sh header is a template
- Single render pass
- Pro: Single pass
- Con: Agent snippets become coupled to Go template syntax, harder to test independently

**Recommendation:** Approach A — the post-create.sh HEADER is a Go template (for {{.ProjectName}}), agent snippets are plain bash. Two-pass: (1) concatenate header + agent blobs, (2) run through text/template.

### 6. TUI Framework for Interactive Mode (huh)

Huh (Charmbracelet) is the right choice per DD-01. It maps cleanly to the init flow:

| Step | Huh Component | PRD Requirement |
|---|---|---|
| Project name | `huh.NewInput().Title("Project name")` | R-INIT-01 |
| Joomla image | `huh.NewSelect()` from parsed images.yaml | R-DEVC-07, R-DEVC-08 |
| Or manual image | `huh.NewInput()` with validation | R-DEVC-08 |
| Agents | `huh.NewMultiSelect()` | R-AGNT-02, R-INIT-09 |
| Timezone | `huh.NewInput().Value(time.Now().Location().String())` | R-DEVC-02 |
| Detect quickstart | Auto-scan for *.zip, then `huh.NewConfirm()` | R-INIT-04 |
| Overwrite confirm | `huh.NewConfirm()` for existing .devcontainer/ | R-DEVC-10, R-INIT-08 |

**Dependencies needed:**
- `github.com/spf13/cobra` — CLI framework
- `github.com/charmbracelet/huh` — TUI forms
- `github.com/charmbracelet/bubbletea` — transitive dep of huh
- `gopkg.in/yaml.v3` — parse images.yaml
- All stdlib for template, embed, fs, flag parsing

---

## Template Placeholder Map (Complete)

### devcontainer.json
| Current Value | Template Variable | Notes |
|---|---|---|
| `"name": "elrepuestazo.com"` | `{{.ProjectName}}` | R-DEVC-02 |
| VSCode extensions list | `{{.VSCodeExtensions}}` | Slice, R-DEVC-12 |

### docker-compose.yml
| Current Value | Template Variable | Notes |
|---|---|---|
| `TZ: UTC` (mail service) | `{{.Timezone}}` | R-DEVC-02 |

### Dockerfile
| Current Value | Template Variable | Notes |
|---|---|---|
| `FROM joomla:6.1-php8.4-apache` | `{{.JoomlaImage}}` | R-DEVC-02 |
| `mariadb-client` | (static, keep) | Needed for DB access |

### .env (ALL values are template variables)
| Current Value | Template Variable | Default |
|---|---|---|
| `JOOMLA_SITE_NAME` | `{{.SiteName}}` | required, no default |
| `JOOMLA_ADMIN_USER` | `{{.AdminUser}}` | "superdev" per R-INIT-02 |
| `JOOMLA_ADMIN_USERNAME` | `{{.AdminUsername}}` | "superdev" per R-INIT-02 |
| `JOOMLA_ADMIN_PASSWORD` | `{{.AdminPassword}}` | "superpassword" per R-INIT-02 |
| `JOOMLA_ADMIN_EMAIL` | `{{.AdminEmail}}` | "admin@example.com" |
| `JOOMLA_DB_USER` | `{{.DBUser}}` | auto from project name |
| `JOOMLA_DB_PASSWORD` | `{{.DBPassword}}` | auto-generated or user-specified |
| `JOOMLA_DB_NAME` | `{{.DBName}}` | auto from project name |
| `JOOMLA_DB_PREFIX` | `{{.DBPrefix}}` | auto-generated 6-char random |
| `JOOMLA_DB_SMTP_HOST` | `{{.SMTPHost}}` | "mail:1025" (static) |
| `ROOT_PASSWORD` | `{{.RootPassword}}` | "dev" (static, low risk) |
| `TZ` | `{{.Timezone}}` | "UTC" |

### .env.example
Same structure as .env but rendered with EXAMPLE values (safe for git).

### php-custom.ini
**No template variables needed.** Entirely static Xdebug config.

### post-create.sh
| Section | Approach |
|---|---|
| Header (shebang, PATH, npm prefix) | Go template for {{.ProjectName}} |
| Agent installations | Concatenated from templates/agents/*.sh per user selection |
| gentle-ai always | Always included per R-AGNT-01 |

---

## Risks

| ID | Risk | Impact | Mitigation |
|---|---|---|---|
| R1 | **Dotfiles in go:embed** | `.env` and `.env.example` won't be embedded by wildcard patterns | Use explicit `//go:embed` directives for each dotfile |
| R2 | **post-create.sh duality** | Two-pass rendering (concatenate + template) adds complexity | Document clearly; keep template vars in header ONLY |
| R3 | **images.yaml doesn't exist** | R-DEVC-07 references a file that isn't created yet | Create a minimal images.yaml with initial curated images |
| R4 | **templates/agents/ don't exist** | AGNT component can't generate post-create.sh | Create stub .sh files (minimal, working) |
| R5 | **No openspec directory before this exploration** | First change needs artifact tracking | Bootstrap openspec/changes/ structure as part of this change |
| R6 | **Current .env has real credentials** | If committed as-is, real creds are in git history | Replace with template variables BEFORE first binary commit |
| R7 | **Module path must be right the first time** | Changing module path later breaks all import paths | Already confirmed: `github.com/alebak/jkit` (DD-07, git remote) |

---

## Recommendations

### Immediate (this change: project-foundation)

1. **Create `go.mod`** with `module github.com/alebak/jkit`, Go 1.24
2. **Create `cmd/jkit/main.go`** with root cobra command
3. **Create `cmd/jkit/init.go`** with `jkit init` command (stub: just prints "not yet implemented" for now)
4. **Create `internal/init/renderer.go`** with the DEVC template rendering logic (the core of this change)
5. **Convert 7 templates** in `templates/devcontainer/` — replace hardcoded values with Go template variables
6. **Create internal package stubs** for `generator`, `agents`, `mcp` (empty packages, just `package generator` etc.)
7. **Create a minimal `images.yaml`** with 3-4 curated Joomla Apache images
8. **Create `templates/agents/`** with stub agent installation snippets
9. **DO NOT touch** `.devcontainer/` (JKit's own dev environment) — it stays as-is

### For the design phase, focus on:

- The exact shape of the `DevcontainerData` struct (what fields, types, defaults)
- How `images.yaml` is parsed (remote fetch + local cache per DD-03 and R-DEVC-13)
- How huh forms map to each field
- File output: `os.WriteFile` with `os.ModePerm`? Or `os.Create` with specific perms?
- The renderer interface: `func RenderDevcontainer(w io.Writer, data DevcontainerData) error`

---

## Ready for Proposal

**Yes.** The codebase has been fully investigated, templates have been read, and the approach is clear. The orchestrator should proceed to `sdd-propose` with these findings as input. No further information is needed from the user or codebase.
