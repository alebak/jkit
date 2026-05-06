# Tasks: Extension Generator

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~1050 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

```
Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High
```

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Foundation + Templates | PR 1 (main) | data.go, detect.go, embed, all 6 type .tmpl files (~700 lines, mostly low-cog templates) |
| 2 | Generator Engine + Registry | PR 2 (main) | generator.go, registry.go, walk/render/rollback, yaml dep (~300 lines) |
| 3 | CLI Wiring + Build + Integration | PR 3 (main) | create.go, build.go, integration tests (~200 lines) |

## Phase 1: Infrastructure

- [x] 1.1 `internal/generator/data.go` — `ExtensionData` + `Prefix`/`FullName`/`ClassName`/`Namespace` helpers + table-driven tests
- [x] 1.2 `internal/generator/detect.go` — `DetectJoomlaProject` cascade (CLI→config) + temp-dir tests
- [x] 1.3 `embed_assets.go` — add `ExtensionsFS embed.FS` with `//go:embed templates/extensions/module, templates/extensions/plugin, templates/extensions/library`
- [x] 1.4 `go.mod` — add `gopkg.in/yaml.v3` dependency

## Phase 2: Generator Engine

- [x] 2.1 `internal/generator/generator.go` — `Generate(ctx, data, targetDir)` walking embed.FS, `.tmpl` parse+render, `.raw` copy, suffix stripping, rollback on error + MapFS tests
- [x] 2.2 `internal/generator/registry.go` — `extensions.jkit.yaml` atomic read/write (`Add`, `Get`, `List`) + roundtrip test

## Phase 3: Templates (6 types)

- [x] 3.1 `templates/extensions/component/` — 7 `.tmpl` files (manifest, admin services/, admin Controller/, admin Extension/, admin Tests/, site Dispatcher/, site Tests/)
- [x] 3.2 `templates/extensions/module/` — 5 `.tmpl` files (manifest, entry, services/provider, tmpl/default, Tests/)
- [x] 3.3 `templates/extensions/plugin/` — 4 `.tmpl` files (manifest, entry, services/provider, Tests/)
- [x] 3.4 `templates/extensions/template/` — 6 `.tmpl` files (manifest, index, component, error, offline, Tests/)
- [x] 3.5 `templates/extensions/library/` — 4 `.tmpl` files (manifest, src/NameLibrary, services/provider, Tests/)
- [ ] 3.6 `templates/extensions/package/` — 1 `.tmpl` file (manifest)
- [x] 3.7 Remove `.gitkeep` from module, plugin, library, component, and template dirs

## Phase 4: CLI Wiring

- [ ] 4.1 `cmd/jkit/create.go` — add `--name`, `--vendor`, `--joomla-version`, `--plugin-group`, `--force` flags; parse args → `ExtensionData` → `Generate()`; TTY prompt vs non-TTY reject
- [ ] 4.2 `cmd/jkit/build.go` — read registry, `Build(ctx, name)` via `archive/zip`, write to `builds/`

## Phase 5: Integration

- [x] 5.1 Generator unit tests: MapFS walk-render, syntax error rollback, .raw copy
- [ ] 5.2 CLI tests: `create component` outputs correct tree, `build unknown` returns exit 1
- [ ] 5.3 Integration smoke test: all 6 real template types render without error
