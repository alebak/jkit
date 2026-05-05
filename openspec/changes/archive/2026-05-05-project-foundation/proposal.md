# Proposal: Project Foundation

## Intent

JKit has no executable code, no module, and hardcoded credentials in templates. This change builds the foundation: Go module, CLI entry point, template rendering engine, and asset conversion. Without it, nothing else can ship.

## Scope

### In Scope
1. `go.mod` — module `github.com/alebak/jkit`, Go 1.24
2. `cmd/jkit/` — cobra root + 3 stubs (init, create, build)
3. `internal/init/renderer.go` — DEVC template engine via `text/template`
4. Convert 7 `templates/devcontainer/` files to Go templates (replace "El Repuestazo" creds)
5. `templates/agents/` — 3 bash snippets (claude, opencode, gemini)
6. `images.yaml` — 3-4 curated Joomla Apache images
7. `templates/extensions/` — 6 empty stubs
8. `templates/skills/prd-creator/` — skill content
9. Internal stubs: `generator`, `agents`, `mcp`
10. `.gitignore` add `.atl/`

### Out of Scope
Full EXT, AGNT, MCP components; `jkit init/create/build` actual execution

## Capabilities

### New
- `cli-commands`: Cobra CLI scaffolding
- `devcontainer-init`: Template rendering for Joomla devcontainer generation

### Modified
None

## Approach

**CLI**: Cobra root + one file per subcommand. Stubs print "not yet implemented".
**Renderer**: `text/template` with `FuncMap`. Explicit `//go:embed` per file (dotfiles won't match wildcard).
**Templates**: Replace all `{{.Variable}}` placeholders via `DevcontainerData` struct.
**post-create.sh**: Two-pass — templated header + concatenated agent bash.
**images.yaml**: Static YAML parsed at runtime.
**Agents**: Standalone `.sh` files, plain bash (not templates).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `go.mod` | New | Module init |
| `cmd/jkit/` | New | main.go + 3 stubs |
| `internal/init/` | New | renderer.go + engine |
| `internal/{generator,agents,mcp}/` | New | Package stubs |
| `templates/devcontainer/` | Modified | 7 files → template vars |
| `templates/agents/` | New | 3 bash snippets |
| `images.yaml` | New | Curated image list |
| `.gitignore` | Modified | Add `.atl/` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Dotfiles not embedded by wildcard | High | Explicit `//go:embed` per file |
| Real creds in git history | High | Replace before first commit |
| Wrong module path | Low | Verified via `git remote` |

## Rollback Plan

Before first commit: `git checkout -- templates/devcontainer/` to restore files. `rm -rf go.mod cmd/ internal/ templates/agents/ templates/extensions/ templates/skills/ images.yaml`. No dependencies depend on this.

## Dependencies

None — this is the first change.

## Success Criteria

- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` succeeds
- [ ] `go mod tidy` completes cleanly
- [ ] All 7 templates render valid output
- [ ] No "El Repuestazo" credentials remain in templates
- [ ] `go test ./...` passes (tests compile)
- [ ] `.atl/` in `.gitignore`
