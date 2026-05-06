# Delta for Devcontainer Init

## ADDED Requirements

### R-DEVC-10: MCP Init Phase

The `jkit init` flow MUST run an MCP deployment phase after the AGNT (agent installation) phase. This SHALL generate `# --- mcp installations ---` section in `post-create.sh` that installs selected MCP servers via `npx` commands.

#### Scenario: post-create.sh contains MCP section

- GIVEN `jkit init` has completed the AGNT phase
- WHEN the rendered `post-create.sh` is inspected
- THEN it contains a `# --- mcp installations ---` comment delimiter
- AND the section includes `npx @playwright/mcp` (or equivalent) for selected MCPs

#### Scenario: Init without MCP flag skips MCP phase

- GIVEN `jkit init` is called without an MCP selection flag
- WHEN the init flow runs
- THEN the MCP phase is skipped
- AND `post-create.sh` contains no `mcp installations` section

### R-DEVC-11: Xdebug Trace and Profile Mode

The `php-custom.ini` template MUST set `xdebug.mode=debug,trace,profile`. The existing `xdebug.mode=debug` SHALL be replaced.

#### Scenario: php-custom.ini has trace and profile

- GIVEN the embedded `php-custom.ini` template
- WHEN inspected
- THEN `xdebug.mode` equals `debug,trace,profile`

## MODIFIED Requirements

### R-DEVC-02: Template Placeholders

Seven template files SHALL use `{{.ProjectName}}`, `{{.JoomlaImage}}`, and `{{.Timezone}}` in appropriate locations. `post-create.sh` gains a `# --- mcp installations ---` section placeholder.
(Previously: post-create.sh had agent install section only)

| File | Substitutions | Notes |
|------|--------------|-------|
| `devcontainer.json` | `{{.ProjectName}}` | Replaces hardcoded "elrepuestazo.com" |
| `Dockerfile` | `{{.JoomlaImage}}` | Replaces hardcoded version |
| `docker-compose.yml` | `{{.JoomlaImage}}` | Dockerfile build context (indirect) |
| `.env` | All fields | Generated from `DevcontainerData` defaults |
| `.env.example` | All fields | Same structure, placeholder values only |
| `php-custom.ini` | None (static) | Xdebug config: `xdebug.mode=debug,trace,profile` |
| `post-create.sh` | `{{.ProjectName}}` | Header context for agent installs + MCP section |

#### Scenario: php-custom.ini includes xdebug trace/profile

- GIVEN the rendered `php-custom.ini`
- WHEN inspected
- THEN `xdebug.mode` contains `debug,trace,profile`

## Coverage Update

| Spec | Source |
|------|--------|
| R-DEVC-10 | Proposal: MCP init phase in post-create.sh |
| R-DEVC-11 | Proposal: php-custom.ini xdebug mode |
