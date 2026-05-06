# Design: Extension Generator

## Technical Approach

Mirror the proven `internal/devcontainer/` pattern: `text/template` + `go:embed` with a walk-and-render engine. The generator walks an embedded `.tmpl` file tree, renders each file against `ExtensionData`, and strips `.tmpl` on write. Key divergence: renders a full directory tree at once (`Generate(ctx, data, targetDir)`) instead of single files.

## Architecture Decisions

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| D1 | Template naming | `.tmpl` suffix, `.raw` for verbatim | Distinguishes template vs raw at FS level. `.raw` mirrors devcontainer's hardcoded php-custom.ini exception. |
| D2 | Embed strategy | Single `//go:embed templates/extensions/*` → `ExtensionsFS` | Simpler than per-type vars. Type selection via path prefix. |
| D3 | Render approach | Walk embed.FS subdir → os.MkdirAll + WriteFile | Templates visible in repo, editable without recompile. Matches devcontainer pattern. |
| D4 | Registry format | YAML + atomic rename | Matches `images.yaml` convention. Atomic write prevents corruption. |
| D5 | Joomla detection | Cascade: `cli/joomla.php list` → `configuration.php` + dirs | CLI is authority for Joomla 5+; config fallback covers projects without CLI. |
| D6 | Overwrite | `--force` / TTY prompt / non-TTY reject | Safe default. Non-TTY reject protects CI. |
| D7 | Vendor | `--vendor` flag first, interactive fallback, `unknown` default | Per-extension vendor matches Joomla PSR-4 namespace conventions. |

## Package Dependencies

```
create.go → generator{pkg} → embed_assets.go (ExtensionsFS) → templates/extensions/{type}/
build.go  → generator{pkg} → extensions.jkit.yaml + builds/*.zip
```

No new external deps — `text/template`, `embed`, `archive/zip`, `io/fs` are stdlib. Cobra already in `go.mod`.

## Data Flow

```
jkit create component --name=Blog --vendor=Alebak
→ create.go: parse flags → ExtensionData{Name:"Blog", Vendor:"Alebak", Type:"component"}
→ detect.go: DetectJoomlaProject(".") → targetDir
→ generator.go: Generate(ctx, data, targetDir)
  → resolveFS("templates/extensions/component")
  → walk embedded tree:
    .tmpl → parse+render → write(strip .tmpl)
    .raw  → copy verbatim
  → registry.Add() → atomic write extensions.jkit.yaml
→ Output: "✅ Created component 'com_blog' (12 files)"
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/generator/generator.go` | Replace | Stub → `Generate()`/`Build()` entry points |
| `internal/generator/data.go` | Create | `ExtensionData` + `Prefix()`, `FullName()`, `ClassName()`, `Namespace()`, `PluginName()` |
| `internal/generator/registry.go` | Create | `extensions.jkit.yaml` atomic read/write |
| `internal/generator/detect.go` | Create | `DetectJoomlaProject(dir)` cascade |
| `internal/generator/*_test.go` (×4) | Create | Table-driven tests per file |
| `embed_assets.go` | Modify | Add `ExtensionsFS embed.FS` with `//go:embed templates/extensions/*` |
| `cmd/jkit/create.go` | Modify | Wire flags → ExtensionData → generator.Create |
| `cmd/jkit/build.go` | Modify | Wire args → registry → zip → builds/ |
| `templates/extensions/{6 types}/` | Populate | `.tmpl` skeleton files replacing `.gitkeep` |

## Interfaces

```go
type ExtensionData struct {
    Name, Vendor, JoomlaVersion, Description, Version, Author, Year string
    Type ExtensionType // component|module|plugin|template|library|package
    Group string // plugin only
}
func (d) Prefix() string    // "com_", "mod_"...
func (d) FullName() string  // "com_blog"
func (d) ClassName() string // "Blog"
func (d) Namespace() string // "Alebak\Component\Blog"

func Generate(ctx context.Context, data ExtensionData, targetDir string) ([]string, error)
func Build(ctx context.Context, name string, targetDir string) (string, error)
func DetectJoomlaProject(dir string) (bool, error)
type ExtEntry struct { Name, Type, Vendor, Path, Version, BuiltAt, CreatedAt string }
type Registry struct { Extensions []ExtEntry }
```

Template vars: `{{.Name}}`, `{{.Vendor}}`, `{{.JoomlaVersion}}`, `{{.Version}}`, `{{.Year}}`, `{{.FullName}}`, `{{.ClassName}}`, `{{.Namespace}}`, `{{.Prefix}}`, `{{.Group}}` (plugin), `{{.Description}}`, `{{.Author}}`.

## Error Handling & Testing

**Errors**: Template syntax → rollback (delete created dirs). Missing `--name` in non-TTY → error. Unknown type → error + list valid. Registry write → atomic rename. Duplicate without `--force` + non-TTY → `ErrAlreadyExists`.

**Testing**: `testing/fstest.MapFS` for walk-render unit tests (no real embedded FS). Table-driven for helper methods. Temp dirs for detect.go. SetOut capture for CLI error paths. One integration smoke test renders all 6 real template types.

## Open Questions

- None. All decisions documented.
