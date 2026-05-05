# Devcontainer Init Specification

## Purpose

Define the template rendering engine and asset structure for generating `.devcontainer/` configuration in generated Joomla projects.

## Requirements

### R-DEVC-01: Template Renderer

The `internal/init` package MUST provide a renderer using `text/template` that accepts a `DevcontainerData` struct and writes to `io.Writer`. Templates MUST be embedded via `//go:embed`.

#### Scenario: Renderer produces valid output
- GIVEN a `DevcontainerData{ProjectName:"test", JoomlaImage:"joomla:6.1-php8.4-apache", Timezone:"America/Argentina/Buenos_Aires"}`
- WHEN `Render(w, "devcontainer.json", data)` is called
- THEN output is valid JSON with project name substituted

#### Scenario: Unknown template returns error
- GIVEN a template name that does not exist
- WHEN `Render` is called
- THEN it returns an error

### R-DEVC-02: Template Placeholders

Seven template files SHALL use `{{.ProjectName}}`, `{{.JoomlaImage}}`, and `{{.Timezone}}` in appropriate locations.

| File | Substitutions | Notes |
|------|--------------|-------|
| `devcontainer.json` | `{{.ProjectName}}` | Replaces hardcoded "elrepuestazo.com" |
| `Dockerfile` | `{{.JoomlaImage}}` | Replaces hardcoded version |
| `docker-compose.yml` | `{{.JoomlaImage}}` | Dockerfile build context (indirect) |
| `.env` | All fields | Generated from `DevcontainerData` defaults |
| `.env.example` | All fields | Same structure, placeholder values only |
| `php-custom.ini` | None (static) | Xdebug config, no substitutions |
| `post-create.sh` | `{{.ProjectName}}` | Header context for agent installs |

#### Scenario: All templates render without error
- GIVEN valid `DevcontainerData`
- WHEN all 7 templates are rendered
- THEN each produces non-empty output

#### Scenario: No hardcoded credentials remain
- GIVEN the rendered output of all templates
- WHEN searched for "El Repuestazo" or "development2026"
- THEN none are found

### R-DEVC-03: DevcontainerData Struct

The struct MUST contain `ProjectName string`, `JoomlaImage string`, `Timezone string`.

#### Scenario: Struct fields map to templates
- GIVEN `DevcontainerData{ProjectName:"test"}`
- WHEN template contains `{{.ProjectName}}`
- THEN output contains "test"

### R-DEVC-04: Default Credentials

Generated `.env` MUST use `superdev`/`superpassword` as admin defaults (R-INIT-02).

#### Scenario: .env has default admin creds
- GIVEN a rendered `.env` from default data
- WHEN inspected
- THEN `JOOMLA_ADMIN_USERNAME` equals "superdev"
- AND `JOOMLA_ADMIN_PASSWORD` equals "superpassword"

### R-DEVC-05: Agent Bash Templates

`templates/agents/` SHALL contain `claude.sh`, `opencode.sh`, `gemini.sh`. These SHALL be plain bash (not Go templates), embedded via `//go:embed`.

#### Scenario: Agent scripts embed and concatenate
- GIVEN the embedded agent files
- WHEN concatenated in order (gentle-ai + selected agents)
- THEN output is valid bash

### R-DEVC-06: images.yaml

`images.yaml` MUST list 3-4 Apache+Debian Joomla images parseable at runtime (R-DEVC-09).

#### Scenario: images.yaml parses correctly
- GIVEN the file
- WHEN parsed as YAML
- THEN it yields a non-empty list of image strings
- AND no image tag lacks `-apache`

### R-DEVC-07: Extension Stubs

`templates/extensions/` SHALL contain 6 empty directories: `component`, `module`, `plugin`, `template`, `library`, `package`.

#### Scenario: Extension stub directories exist
- GIVEN the repository
- WHEN listing `templates/extensions/`
- THEN each of the 6 types is present as a directory

### R-DEVC-08: prd-creator Skill

`templates/skills/prd-creator/` SHALL contain skill files embedded via `//go:embed`.

#### Scenario: Skill embeds at compile time
- GIVEN `go build`
- WHEN the binary inspects its embedded filesystem
- THEN `templates/skills/prd-creator/` is accessible

### R-DEVC-09: post-create.sh Gentle AI Install

The rendered `post-create.sh` MUST install `gentle-ai` unconditionally (R-AGNT-01).

#### Scenario: post-create.sh contains gentle-ai install
- GIVEN a rendered `post-create.sh`
- WHEN inspected
- THEN it contains a command to install `gentle-ai`

### Coverage

| Spec | PRD Req |
|------|---------|
| R-DEVC-01 | R-DEVC-01, R-DEVC-02, DD-01 |
| R-DEVC-02 | R-DEVC-01, R-DEVC-02 |
| R-DEVC-03 | R-DEVC-02 |
| R-DEVC-04 | R-INIT-02 |
| R-DEVC-05 | R-AGNT-05 |
| R-DEVC-06 | R-DEVC-07, R-DEVC-09 |
| R-DEVC-07 | R-EXTG-01 |
| R-DEVC-08 | R-AGNT-04 |
| R-DEVC-09 | R-AGNT-01 |
