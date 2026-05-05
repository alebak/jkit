# Tasks: Project Foundation

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~612 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (~267) → PR 2 (~345) |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Module, CLI stubs, internal stubs, static assets | PR 1 | Base = main. ~267 lines. Cobra-only, no huh. |
| 2 | Renderer engine, template conversion, tests | PR 2 | Base = main. ~345 lines. Depends on module from PR 1. |

## Phase 1: Infrastructure

- [x] 1.1 Create `go.mod` — module `github.com/alebak/jkit`, Go 1.24
- [x] 1.2 Create `.gitignore` — add `.atl/`, `.env`, `builds/`, `/jkit`
- [x] 1.3 Create stub packages: `internal/generator/generator.go`, `internal/agents/agents.go`, `internal/mcp/mcp.go`

## Phase 2: Core — Template Engine

- [ ] 2.1 Create `internal/init/devcontainer.go` — `DevcontainerData` struct + `DefaultDevcontainerData()`
- [ ] 2.2 Create `internal/init/templates.go` — `//go:embed templates/devcontainer` + explicit ParseFS patterns
- [ ] 2.3 Create `internal/init/renderer.go` — `Render(io.Writer, string, any) error` with static passthrough for `php-custom.ini`
- [ ] 2.4 Create `internal/init/renderer_test.go` — table-driven: valid output, unknown name error, no cred leaks, all 7 non-empty

## Phase 3: Templates — Devcontainer Conversion

- [ ] 3.1 Convert `devcontainer.json` — `"elrepuestazo.com"` → `{{.ProjectName}}`
- [ ] 3.2 Convert `Dockerfile` — `FROM joomla:6.1-php8.4-apache` → `FROM {{.JoomlaImage}}`
- [ ] 3.3 Convert `.env` — all 12 values → `{{.Variable}}`, creds → defaults per R-INIT-02
- [ ] 3.4 Convert `.env.example` — same structure, placeholder values
- [ ] 3.5 Convert `post-create.sh` — template header with `{{.ProjectName}}` + agent concatenation chain
- [ ] 3.6 Convert `docker-compose.yml` — `TZ: UTC` → `{{.Timezone}}`

## Phase 4: CLI — Cobra Commands

- [x] 4.1 Create `cmd/jkit/main.go` — root cobra command + entry point
- [x] 4.2 Create `cmd/jkit/init.go` — stub with flags (--name, --image, --quickstart, --agents, --timezone)
- [x] 4.3 Create `cmd/jkit/create.go` — stub with ValidArgs (component, module, plugin, template, library, package)
- [x] 4.4 Create `cmd/jkit/build.go` — stub with args
- [x] 4.5 Wire all commands to root (init() registration in each file)

## Phase 5: Static Assets

- [x] 5.1 Create `images.yaml` — 4 curated Joomla Apache images (6.1-php8.4, 5.3-php8.3, 5.3-php8.4, 6.1-php8.5)
- [x] 5.2 Create `templates/agents/` — `claude.sh`, `opencode.sh`, `gemini.sh`
- [x] 5.3 Create `templates/extensions/` — 6 dirs: component, module, plugin, template, library, package
- [x] 5.4 Create `templates/skills/prd-creator/SKILL.md`

## Phase 6: Integration & Verification

- [x] 6.1 Run `go mod tidy` — verify clean go.sum
- [x] 6.2 Run `go build ./...` + `go vet ./...` — zero errors
- [x] 6.3 Run `go test ./...` — all pass
- [ ] 6.4 Verify no "El Repuestazo" or "development2026" remain in any template (deferred to PR 2 — templates not yet converted)

## Dependency Order

Phase 1 → Phase 2 → Phase 3 → Phase 4 | Phase 5 (independent) → Phase 6
Phase 1 must complete first (module must exist). Phase 5 is fully independent. Phase 6 needs everything.
