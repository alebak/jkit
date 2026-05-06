# Exploration: Extension Generator (EXTG)

## Current State

The JKit codebase is a Go CLI with three cobra subcommands (`init`, `create`, `build`) — all printing "not yet implemented". The `internal/generator/` package exists as a stub (just a doc comment). The `templates/extensions/` directory has 6 empty subdirectories (component, module, plugin, template, library, package) each with only a `.gitkeep`.

The **devcontainer renderer** (`internal/devcontainer/`) establishes the pattern this feature must follow:
- Go templates via `text/template` embedded with `//go:embed`
- Templates in `templates/` directory tree
- A data struct (e.g., `DevcontainerData`) passed to render functions
- Render functions write to `io.Writer`
- No `embed` directive exists yet for `templates/extensions/`

**PRD defines** EXTG requirements R-EXTG-01 through R-EXTG-11 and DD-05 (hybrid approach: Go templates for structure, gentle-ai for implementation).

## Affected Areas

| File | Why affected |
|------|-------------|
| `internal/generator/generator.go` | Stub — needs full implementation |
| `internal/generator/extensions.go` | New file — extension type definitions, render logic, zip packaging |
| `internal/generator/renderer.go` | New file — template rendering for extension skeletons |
| `internal/generator/generator_test.go` | New file — table-driven tests |
| `embed_assets.go` | Needs `//go:embed templates/extensions/*` directive added |
| `cmd/jkit/create.go` | Stub — needs to wire into generator |
| `cmd/jkit/build.go` | Stub — needs zip packaging logic |
| `templates/extensions/component/` | New Go template files for component skeleton |
| `templates/extensions/module/` | New Go template files for module skeleton |
| `templates/extensions/plugin/` | New Go template files for plugin skeleton |
| `templates/extensions/template/` | New Go template files for template skeleton |
| `templates/extensions/library/` | New Go template files for library skeleton |
| `templates/extensions/package/` | New Go template files for package skeleton |

## Approaches

### 1. Go Templates (recommended — consistent with existing pattern)

Each extension type has a directory tree of Go template files under `templates/extensions/{type}/`. The generator walks the embedded FS, processes templates with `text/template`, and writes to the target project directory.

**How it works:**
- `embed_assets.go` gets `//go:embed templates/extensions/**/*` 
- `internal/generator/extensions.go` defines `ExtensionType` enum, `ExtensionData` struct
- `internal/generator/renderer.go` handles template rendering (mirrors `internal/devcontainer/renderer.go`)
- `cmd/jkit/create.go` parses args → builds `ExtensionData` → calls generator
- `cmd/jkit/build.go` scans project for extension manifests → zips them into `builds/`

**Template files per type:**
- **Component**: `com_component.xml` (manifest), `services/provider.php`, `src/Extension/Component.php`, `src/Controller/DisplayController.php`, `admin/services/provider.php`, `admin/src/Extension/Component.php`, `admin/src/Controller/DisplayController.php`, `tmpl/default.php`
- **Module**: `mod_module.xml` (manifest), `mod_module.php`, `tmpl/default.php`, `services/provider.php`
- **Plugin**: `plg_type_name.xml` (manifest), `name.php`, `services/provider.php`
- **Template**: `templateDetails.xml` (manifest), `index.php`, `component.php`, `error.php`, `offline.php`
- **Library**: `lib_library.xml` (manifest), `src/Library.php`
- **Package**: `pkg_package.xml` (manifest — references other extensions)

| Pros | Cons |
|------|------|
| Matches existing devcontainer renderer pattern exactly | Each extension type needs ~5-15 template files |
| Templates visible and editable in source tree | Template syntax (`{{.Var}}`) in XML/PHP files adds some noise |
| No recompile needed to view/modify templates | — |
| Variables fully customizable per type | — |

**Effort**: High (6 extension types × 5-15 templates each + renderer + CLI wiring + tests + zip packaging)

### 2. Programmatic Go Structs

All file generation happens in Go code — no template files. A Go struct defines each extension type's directory tree, and the generator creates files programmatically.

**How it works:**
- Go code defines each file's content as string literals or `strings.Builder`
- No `templates/extensions/` files needed beyond the `.gitkeep` stubs
- Manifest XML and PHP files constructed in Go code

| Pros | Cons |
|------|------|
| No template files to manage | Deviates from established embed/template pattern |
| Easier to test in pure Go | Harder to review XML/PHP content (buried in Go strings) |
| No template parsing overhead | PHP developers can't see/modify skeleton files easily |
| — | Template content changes require recompile |
| — | Go string literals for XML/PHP are ugly with escaping |

**Effort**: Medium (less files, but more complex Go code)

### 3. Static Files + Go Templates for Manifests Only

PHP stubs are static files (no template processing), only XML manifests are Go templates (need variable substitution).

**How it works:**
- Manifests XML → Go templates (need extension name, version, etc.)
- PHP files → static files with `// TODO: gentle-ai implement` placeholders
- Generator copies PHP files as-is, renders XML manifests

| Pros | Cons |
|------|------|
| Simple — most files are just copied | Doesn't support namespace injection in PHP files |
| Low template maintenance | PHP class names must be hardcoded in static stubs |
| — | Less flexible for PSR-4 namespace declaration |
| — | Breaks R-EXTG-04 (correct PSR-4 namespaces) |

**Effort**: Medium (static copies + manifest templates)

## Recommendation

**Option 1: Go Templates** — because:
1. It mirrors the exact pattern proven in `internal/devcontainer/renderer.go` — consistent codebase, reviewer-friendly.
2. DD-05 specifically calls for template-generated structure with AI-generated implementation. This delivers structure templates now, AI fills PHP logic later.
3. Templates are visible in the repo — Joomla developers can inspect/modify skeleton XML/PHP without touching Go code.
4. The existing `embed_assets.go` pattern makes adding `templates/extensions/` trivial.
5. `text/template` is already a dependency (used by devcontainer), no new deps needed.

### Variables needed in ExtensionData

| Variable | Example | Used in |
|----------|---------|---------|
| `Name` | `MyComponent` | Directory names, class names, namespaces |
| `Prefix` | `com_`, `mod_`, `plg_`, `tpl_`, `lib_`, `pkg_` | Directory name, manifest filename |
| `FullName` | `com_mycomponent` | Manifest filename, extension identity |
| `Vendor` | `Alebak` | PSR-4 namespace root |
| `Description` | `My component for...` | Manifest description |
| `Version` | `1.0.0` | Manifest version |
| `JoomlaVersion` | `5` | Manifest min Joomla version |
| `Author` | from config | Manifest author |
| `ExtensionType` | `component` | Type-specific rendering logic |
| `PluginGroup` | `content` (plugin only) | Plugin group in manifest |
| `Namespace` | `Alebak\Component\MyComponent\Site` | Full PSR-4 namespace |
| `NamespaceAdmin` | `Alebak\Component\MyComponent\Administrator` | Admin namespace |
| `ClassName` | `MyComponent` | Class names derived from Name |
| `Year` | `2026` | Copyright headers |

### Template file structure for each extension type

```
templates/extensions/
├── component/
│   ├── administrator/
│   │   ├── services/
│   │   │   └── provider.php.tmpl
│   │   └── src/
│   │       └── Extension/
│   │           └── Component.php.tmpl
│   ├── src/
│   │   └── Controller/
│   │       └── DisplayController.php.tmpl
│   ├── tmpl/
│   │   └── default.php.tmpl
│   └── manifest.xml.tmpl                    → com_component.xml
├── module/
│   ├── services/
│   │   └── provider.php.tmpl
│   ├── tmpl/
│   │   └── default.php.tmpl
│   ├── mod_module.php.tmpl
│   └── manifest.xml.tmpl                    → mod_module.xml
├── plugin/
│   ├── services/
│   │   └── provider.php.tmpl
│   ├── name.php.tmpl
│   └── manifest.xml.tmpl                    → plg_type_name.xml
├── template/
│   ├── index.php.tmpl
│   ├── component.php.tmpl
│   ├── error.php.tmpl
│   ├── offline.php.tmpl
│   └── manifest.xml.tmpl                    → templateDetails.xml
├── library/
│   ├── src/
│   │   └── Library.php.tmpl
│   └── manifest.xml.tmpl                    → lib_library.xml
└── package/
    └── manifest.xml.tmpl                    → pkg_package.xml
```

## Risks

| ID | Risk | Mitigation |
|----|------|------------|
| RK-EXTG-01 | **Joomla 5/6 namespace conventions may shift** between releases | Base on current Joomla 5.2+ conventions; make template-driven so updates just change `.tmpl` files |
| RK-EXTG-02 | **PSR-4 autoloader paths require precise directory alignment** | Generate PHP files with exact class-to-file mapping per Joomla conventions |
| RK-EXTG-03 | **Service provider pattern complexity** for components | Start with the minimal `services/provider.php` that Joomla requires; AI fills dispatcher/subscriber logic |
| RK-EXTG-04 | **Manifest XML schema varies** per extension type | Each type gets its own manifest template, matched against Joomla 5.x XSD |
| RK-EXTG-05 | **`jkit build` needs to detect existing extensions** in project | Scan for XML manifest files matching extension patterns; requires convention (`com_*.xml`, `mod_*.xml`, etc.) |
| RK-EXTG-06 | **Template .tmpl extension adds new file type** to the project pattern | All devcontainer templates use their original extension (`.json`, `.yml`, `.sh`, `.ini`), but `.tmpl` avoids confusion with actual `.php`/`.xml` that editors would syntax-check. Consistent with Go conventions. |
| RK-EXTG-07 | **No `--vendor` flag currently exists** in the CLI | The PRD assumes a vendor namespace but doesn't specify one. Could default to project name, or add `--vendor` flag to `jkit create`. |

### Key Open Questions for Proposal Phase

1. **Vendor namespace**: Where does the PSR-4 vendor prefix come from? A `--vendor` flag on `jkit create`? A config file? Default to project name?
2. **Project root detection**: How does `jkit create` know where the Joomla project root is? Current working directory? Expect Joomla directory structure?
3. **`jkit build` discovery**: Should it scan the project for manifest XML files, or should extensions be registered in a config file?
4. **Joomla version target**: Min Joomla version for manifest `<extension>` tag? 5.0? 5.1?
5. **Test framework conventions**: PHPUnit? Pest? Directory structure under `tests/` mirroring source?

## Ready for Proposal

**Yes** — the codebase patterns are clear, the devcontainer renderer provides a proven template, and all 6 extension types have well-documented Joomla 5/6 structures. The hybrid approach (DD-05) maps cleanly to Option 1.

The proposal phase should address the 5 open questions above, particularly the vendor namespace and project root detection.
