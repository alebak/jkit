# Delta for devcontainer-init

## MODIFIED Requirements

### R-DEVC-07: Extension Template Files

`templates/extensions/` SHALL contain 6 subdirectories (`component`, `module`, `plugin`, `template`, `library`, `package`) populated with `.tmpl` skeleton files. `embed_assets.go` MUST add `//go:embed templates/extensions/**/*` directive.
(Previously: 6 empty directories with `.gitkeep` only)

#### Scenario: Component templates embedded
- GIVEN the repository
- WHEN `templates/extensions/component/` is inspected
- THEN it contains `.tmpl` files for manifest, services/provider, src/, tmpl/
- AND the `//go:embed` directive in `embed_assets.go` includes `templates/extensions/**/*`

#### Scenario: All 6 types have templates
- GIVEN the embedded filesystem
- WHEN listing each type subdirectory
- THEN each contains at least one `.tmpl` file

#### Scenario: `.gitkeep` removed from populated dirs
- GIVEN a directory that now contains `.tmpl` files
- WHEN inspected
- THEN `.gitkeep` is absent
