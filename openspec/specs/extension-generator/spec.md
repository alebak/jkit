# Extension Generator Specification

## Purpose

Define the generator engine, CLI scaffolding, zip packaging, and extension registry for creating Joomla 5+ extension skeletons.

## Requirements

### R-EXTG-01: Generator Engine

The generator MUST walk embedded `.tmpl` files, render them with `text/template` using an `ExtensionData` struct, and strip the `.tmpl` suffix on write. Files with `.raw` extension MUST be copied verbatim. Directories MUST be created as needed.

#### Scenario: Render component template tree
- GIVEN `templates/extensions/component/` embedded with `.tmpl` files
- WHEN `generator.Create(ctx, "component", data)` is called with valid `ExtensionData`
- THEN all `.tmpl` files are rendered to the target Joomla project with `.tmpl` stripped
- AND `.raw` files are copied verbatim

#### Scenario: Invalid template returns error
- GIVEN an embedded `.tmpl` file with a syntax error
- WHEN the generator attempts to render it
- THEN it returns an error
- AND no files are written

#### Scenario: Joomla naming conventions applied
- GIVEN `ExtensionData{Name: "Blog", Type: "component"}`
- WHEN `Create` is called
- THEN output dirs use prefix `com_` (component), `mod_` (module), `plg_` (plugin), `tpl_` (template), `lib_` (library), `pkg_` (package)

### R-EXTG-02: CLI Create

The `jkit create [type]` command MUST accept exactly one positional arg matching a valid type. Flags: `--name`, `--vendor`, `--joomla-version`, `--plugin-group`, `--force`. The command MUST detect the Joomla project root via cascade (try CLI: `cli/joomla.php list`, fallback to `configuration.php` + dir structure).

| Flag | Type | Default | Required | Applies to |
|------|------|---------|----------|------------|
| `--name` | string | — | yes | All types |
| `--vendor` | string | prompt | no | All types |
| `--joomla-version` | string | `5` | no | All types |
| `--plugin-group` | string | `system` | no | plugin only |
| `--force` | bool | false | no | All types |

#### Scenario: Create with all flags
- GIVEN a Joomla project at the current directory
- WHEN `jkit create component --name=Blog --vendor=Alebak --joomla-version=5` runs
- THEN `administrator/components/com_blog/` and `components/com_blog/` skeleton exists
- AND `extensions.jkit.yaml` is created/updated

#### Scenario: Invalid type rejected
- GIVEN `jkit create invalid_type`
- WHEN run
- THEN error is printed listing valid types
- AND exit code is 1

#### Scenario: Missing --name
- GIVEN `jkit create component` without `--name`
- WHEN run in non-interactive mode (stdin not tty)
- THEN error is printed
- AND exit code is 1

#### Scenario: Joomla project not found
- GIVEN a directory without `cli/joomla.php` or `configuration.php`
- WHEN `jkit create component --name=Blog` runs
- THEN error is printed: "Joomla project not found"
- AND exit code is 1

### R-EXTG-03: Overwrite Protection

The system MUST check if the target extension directory exists. If it does:
- In TTY mode: prompt `[y/N]` before overwriting (interactive)
- In non-TTY mode: auto-reject unless `--force` is set

#### Scenario: Force overwrite existing extension
- GIVEN `com_blog/` already exists
- WHEN `jkit create component --name=Blog --force` runs
- THEN the extension is overwritten without prompting

#### Scenario: Non-TTY rejects overwrite
- GIVEN `com_blog/` already exists and stdin is not a TTY
- WHEN `jkit create component --name=Blog` runs (no `--force`)
- THEN error is printed: "already exists, use --force to overwrite"
- AND exit code is 1

### R-EXTG-04: CLI Build

The `jkit build [name]` command MUST read the extension registry for the named extension, create a `.zip` archive in `builds/` with correct Joomla directory structure, and update the registry with the build timestamp.

#### Scenario: Build creates valid zip
- GIVEN `com_blog/` extension registered in `extensions.jkit.yaml`
- WHEN `jkit build com_blog` runs
- THEN `builds/com_blog-1.0.0.zip` is created
- AND the zip contains the correct Joomla directory tree

#### Scenario: Unknown extension name
- GIVEN no extension named `com_unknown` in registry
- WHEN `jkit build com_unknown` runs
- THEN error is printed
- AND exit code is 1

### R-EXTG-05: Extension Registry

The extension registry (`extensions.jkit.yaml`, YAML format) MUST be read/written atomically. It SHALL append entries on `create`, update entries on `build`, and never delete entries. Each entry MUST include: name, type, vendor, path, version, built-at (if built).

#### Scenario: Registry written on create
- GIVEN `jkit create component --name=Blog --vendor=Alebak`
- WHEN the command completes
- THEN `extensions.jkit.yaml` contains an entry for `com_blog`
- AND the file is valid YAML

#### Scenario: Registry file is atomic
- GIVEN concurrent writes are not expected in single-user CLI
- WHEN the registry is written
- THEN the write uses atomic rename pattern (write temp, rename)
- AND partial writes never leave a corrupt file

### R-EXTG-06: Test Structure

Every generated extension MUST include `Tests/Unit/ExampleTest.php` stub and `phpunit.xml.dist` in the extension root.

#### Scenario: Tests generated with component
- GIVEN `jkit create component --name=Blog --vendor=Alebak`
- WHEN completed
- THEN `com_blog/Tests/Unit/ExampleTest.php` exists
- AND `com_blog/phpunit.xml.dist` exists
- AND the test file is valid PHP with correct namespace

### R-EXTG-07: Extension Templates

Each extension type MUST produce a valid manifest XML and the required directory structure per Joomla 5+ conventions.

#### Scenario: Component skeleton
- GIVEN `component` with `Name: "Blog", Vendor: "Alebak"`
- WHEN generated
- THEN manifests use `<extension type="component" method="upgrade">` and PSR-4 namespace `Alebak\Component\Blog`

#### Scenario: Module skeleton
- GIVEN `module` with `Name: "Latest", Vendor: "Alebak"`
- WHEN generated
- THEN manifest uses `<extension type="module" ...>` with `client="site"`
- AND files at `modules/mod_latest/`

#### Scenario: Plugin skeleton
- GIVEN `plugin` with `Name: "Maps", Group: "content", Vendor: "Alebak"`
- WHEN generated
- THEN manifest uses `<extension type="plugin" group="content" method="upgrade">`
- AND files at `plugins/content/maps/`

#### Scenario: Template skeleton
- GIVEN `template` with `Name: "Base", Vendor: "Alebak"`
- WHEN generated
- THEN `templateDetails.xml` includes `<extension type="template" client="site">`
- AND files at `templates/tpl_base/` with `index.php`, `component.php`, `error.php`, `offline.php`

#### Scenario: Library skeleton
- GIVEN `library` with `Name: "Utils", Vendor: "Alebak"`
- WHEN generated
- THEN `lib_utils.xml` uses `<extension type="library" method="upgrade">`
- AND files at `libraries/alebak/utils/`

#### Scenario: Package skeleton
- GIVEN `package` with `Name: "Suite", Vendor: "Alebak"`
- WHEN generated
- THEN `pkg_suite.xml` is valid package manifest
- AND files at `packages/pkg_suite/`

### Coverage

| Spec | PRD Req |
|------|---------|
| R-EXTG-01 | R-EXTG-01, R-EXTG-03, R-EXTG-04 |
| R-EXTG-02 | R-EXTG-01, R-EXTG-10 |
| R-EXTG-03 | R-EXTG-09 |
| R-EXTG-04 | R-EXTG-06 |
| R-EXTG-05 | R-EXTG-05 |
| R-EXTG-06 | R-EXTG-11 |
| R-EXTG-07 | R-EXTG-02, R-EXTG-08 |
