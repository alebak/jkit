# Proposal: Extension Generator

## Intent

JKit's `create` and `build` commands are stubs that print "not yet implemented". Joomla developers need to scaffold extensions (component, module, plugin, template, library, package) with correct Joomla 5+ structure and generate deployable `.zip` packages. This change delivers the full generator engine.

## Scope

### In Scope
- Generator engine in `internal/generator/` — walks embedded template trees, applies variables, writes to target project
- Template files for 6 extension types (component, module, plugin, template, library, package)
- `jkit create [type]` with `--name`, `--vendor`, `--joomla-version` flags + interactive prompts
- `jkit build [name]` — packages extension as `.zip` in `builds/`
- Extension registry `extensions.jkit.yaml` for tracking created extensions
- Joomla project detection (CLI cascade: `cli/joomla.php list` → `configuration.php` + dir structure)
- Overwrite protection (prompt before overwriting existing extension)
- Test structure generation (`Tests/Unit/`, `phpunit.xml.dist`, `ExampleTest.php` stub)

### Out of Scope
- Full PHP implementation code (controllers, models, views — filled by gentle-ai)
- Package type grouping for `jkit build` (needs all extensions built first — R-EXTG-07)
- Interactive TUI mode (huh integration — future change)
- Running Joomla CLI inside devcontainer (delegated to Docker, not JKit)

## Capabilities

### New Capabilities
- `extension-generator`: Core generator — walks embedded FS, renders `.tmpl` files with `text/template`, writes extension skeletons. Sub-commands: `create` (scaffold) and `build` (zip packaging).

### Modified Capabilities
- `cli-commands`: R-CLI-02 evolves — `create`/`build` go from stubs to implementation. `create` gains `--name`, `--vendor`, `--joomla-version`. `build` gains extension discovery and zip logic.
- `devcontainer-init`: R-DEVC-07 evolves — `templates/extensions/{type}/` populated with `.tmpl` files; new `//go:embed` directive in `embed_assets.go`.

## Approach

Use Go Templates with `go:embed` — the exact pattern from `internal/devcontainer/renderer.go`.

**Files**:
- `generator.go`: Package entry, public `Create(ctx, type, data) error` / `Build(ctx, name) error`
- `extensions.go`: `ExtensionType` enum (6 types), `ExtensionData` struct, per-type metadata (prefix, dirs, files)
- `renderer.go`: Walks `embed.FS` subtree, renders `.tmpl` with `text/template`, strips `.tmpl` suffix on write
- `create.go`: Parse flags → build `ExtensionData` → call `generator.Create`
- `build.go`: Scan project for manifest XML files → zip matching dirs → write to `builds/`
- `embed_assets.go`: Add `//go:embed templates/extensions/**/*`

**Template convention**: `.tmpl` suffix avoids editor syntax-checking `.php`/`.xml` as real code. Suffix stripped on write (e.g. `manifest.xml.tmpl` → `com_component.xml`).

**Variables** in `ExtensionData`: `Name`, `Vendor`, `Prefix`, `FullName`, `Namespace`, `NamespaceAdmin`, `JoomlaVersion`, `Version`, `Description`, `Author`, `Year`, `PluginGroup` (plugin only).

**Template tree per type**: See `exploration.md` §"Template file structure for each extension type" — 6 trees, 3-9 files each.

**Vendor resolution**: `--vendor` flag first, interactive prompt if empty.

**Project detection**: Try `cli/joomla.php list` → Joomla 5+; fallback `configuration.php` exists + expected dirs.

**Overwrite protection**: Check if target dir exists → prompt `[y/N]` before overwriting.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/generator/generator.go` | Modified | Stub → full entry point |
| `internal/generator/extensions.go` | New | Extension type defs, data structs, metadata |
| `internal/generator/renderer.go` | New | Template rendering, FS walking, `.tmpl` stripping |
| `internal/generator/generator_test.go` | New | Table-driven tests |
| `embed_assets.go` | Modified | Add `//go:embed templates/extensions/**/*` |
| `cmd/jkit/create.go` | Modified | Stub → flag handling, interactive prompts, generator wiring |
| `cmd/jkit/build.go` | Modified | Stub → extension discovery, zip packaging |
| `templates/extensions/{component,module,plugin,template,library,package}/` | Modified | Populated with `.tmpl` skeleton files |
| `extensions.jkit.yaml` | New | Registry in target project (not in repo) |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Template `.tmpl` suffix inconsistent with existing non-templated templates | Low | Devcontainer templates don't need suffix (no `.php`/`.xml` collision). Document in code comments. |
| Manifest XML schema varies per type | Medium | One template per type, matched to Joomla 5.2+ XSD references |
| Vendor namespace design (per-extension vs global) | Low | Resolved: `--vendor` per extension with interactive fallback |
| Multi-task implementation exceeds 400-line review budget | Medium | Split into chained PRs: (1) templates + embed, (2) generator engine + create, (3) build + tests |

## Rollback Plan

1. Revert `cmd/jkit/create.go` and `cmd/jkit/build.go` to stub versions
2. Remove `//go:embed templates/extensions/**/*` from `embed_assets.go`
3. Delete new files in `internal/generator/` (extensions.go, renderer.go, generator_test.go)
4. Restore `internal/generator/generator.go` to stub comment-only
5. Revert `templates/extensions/{type}/` to `.gitkeep` only

All changes are additive — no breaking changes to existing commands or public API.

## Dependencies

- Go stdlib: `text/template`, `embed`, `archive/zip`, `io/fs`, `path/filepath`
- `github.com/spf13/cobra` (already in `go.mod`)
- No new external dependencies

## Success Criteria

- [ ] `go build ./...` compiles cleanly; `go vet ./...` passes
- [ ] `jkit create component --name=Hello --vendor=Alebak` generates valid Joomla 5 component skeleton with correct dirs, manifest, and PSR-4 namespaces
- [ ] `jkit build com_hello` in a project with `com_hello/` produces `builds/com_hello.zip` with correct structure
- [ ] All 6 extension types generate valid manifest XML and correct directory trees
- [ ] Joomla project detection finds project via `cli/joomla.php` or `configuration.php`
- [ ] Overwrite protection prompts `[y/N]` before replacing existing extension dir
- [ ] Generated extension includes `Tests/Unit/ExampleTest.php` and `phpunit.xml.dist`
- [ ] `extensions.jkit.yaml` is created/updated after `jkit create`/`build`
