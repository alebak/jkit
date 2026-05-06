# Init TUI Specification

## Purpose

Define the interactive and parameterized initialization flow for JKit. This is the main product entry point — it collects user inputs (TUI or flags) and orchestrates devcontainer setup, agent deployment, extension generation, and MCP configuration.

## Requirements

### R-INIT-TUI-01: Interactive Form

The `jkit init` command SHOULD launch an interactive TUI when no flags are provided and stdout is a TTY. The TUI MUST use huh forms to collect: project name, Joomla image, AI agents, timezone, and optional quickstart path.

#### Scenario: TUI collects project name
- GIVEN a TTY session
- WHEN `jkit init` is invoked with no flags
- THEN a text input for project name is presented
- AND the input MUST be non-empty to proceed

#### Scenario: TUI shows Joomla images from images.yaml
- GIVEN `images.yaml` at repo root
- WHEN the image selection step renders
- THEN it shows all entries as selectable options
- AND each entry displays its tag and description

#### Scenario: TUI shows available agents
- GIVEN the embedded `templates/agents/` directory has `.sh` files
- WHEN the agent multi-select step renders
- THEN all agents discovered from the embedded FS are shown
- AND none are pre-selected

#### Scenario: TUI defaults timezone to UTC
- GIVEN the timezone input step
- WHEN rendered
- THEN the default value is "UTC"

#### Scenario: TUI prompts for overwrite if .devcontainer/ exists
- GIVEN a CWD with an existing `.devcontainer/` directory
- WHEN the form reaches the confirmation step
- THEN a confirm prompt asks "Overwrite existing .devcontainer/?"
- AND if denied, the process exits with no changes

### R-INIT-TUI-02: Parameterized Mode

The `jkit init` command MUST accept CLI flags to bypass the TUI entirely. When any of `--name`, `--image`, or `--quickstart` is set, the command MUST run in parameterized mode without interactive prompts.

#### Scenario: Parameterized mode with all flags
- GIVEN a non-TTY environment
- WHEN `jkit init --name myproject --image joomla:6.1-php8.4-apache --agents claude,opencode --timezone America/New_York` is invoked
- THEN the init proceeds without TUI prompts
- AND all provided values are used

#### Scenario: Parameterized mode with minimal flags
- GIVEN `jkit init --name myproject`
- WHEN invoked
- THEN default image (`joomla:6.1-php8.4-apache`), timezone (`UTC`), and no agents are used

#### Scenario: Parameterized mode requires --name
- GIVEN `jkit init --image joomla:6.1-php8.4-apache` without `--name`
- WHEN invoked
- THEN it returns an error

#### Scenario: Non-TTY without flags fails clearly
- GIVEN a non-TTY environment (e.g. piped stdout)
- WHEN `jkit init` is invoked with no flags
- THEN it returns an error explaining flags are required in non-TTY mode

### R-INIT-TUI-03: Orchestration Order

The orchestration MUST execute components in this order: DEVC Render → AGNT DeploySkill → EXTG Generate → MCPS DeployMCP. Each step MUST fail fast — if any step errors, subsequent steps MUST NOT run.

#### Scenario: Orchestration runs in order
- GIVEN a valid InitConfig
- WHEN Orchestrate() is called
- THEN DEVC renders `.devcontainer/` first
- THEN AGNT deploys skills for selected agents
- THEN EXTG generates the default component
- THEN MCPS deploys MCPs for the first selected agent
- AND the `builds/` directory is created last

#### Scenario: Error in DEVC stops all further steps
- GIVEN a scenario where DEVC Render fails
- WHEN Orchestrate() is called
- THEN it returns an error
- AND AGNT/EXTG/MCPS steps never execute

#### Scenario: No agents skips AGNT and MCPS
- GIVEN an InitConfig with `Agents: []` (empty)
- WHEN Orchestrate() is called
- THEN DEVC renders
- THEN AGNT is skipped
- THEN EXTG generates
- THEN MCPS is skipped (no target agent)
- AND no error is returned

### R-INIT-TUI-04: Defaults

Defaults SHALL follow R-INIT-02: admin username `superdev`, admin password `superpassword`, timezone `UTC`. All 3 agents (claude, opencode, gemini) SHALL be available for selection.

#### Scenario: Defaults match DevcontainerData
- GIVEN an InitConfig with only ProjectName set
- WHEN converted to DevcontainerData
- THEN AdminUser is "superdev"
- AND AdminPassword is "superpassword"
- AND Timezone is "UTC"

### R-INIT-TUI-05: Quickstart Detection

When `--quickstart` is provided (or quickstart path entered in TUI), the init flow MUST detect `.zip` files in CWD. If exactly one `.zip` file exists, use it. If multiple exist, return an error asking the user to specify.

#### Scenario: Quickstart with single .zip
- GIVEN a CWD with exactly one `quickstart.zip`
- WHEN `jkit init --quickstart quickstart.zip` is run
- THEN the zip is extracted as the project base

#### Scenario: Quickstart with no .zip
- GIVEN a CWD with no `.zip` files
- WHEN `jkit init --quickstart` is run
- THEN it returns an error "no quickstart .zip found in current directory"

### R-INIT-TUI-06: Overwrite Protection

If `.devcontainer/` already exists in the target directory and `--force` is not set, the init MUST prompt (TUI) or error (parameterized) before overwriting. With `--force`, overwrite proceeds without confirmation.

#### Scenario: Overwrite prompt in TUI
- GIVEN existing `.devcontainer/`
- WHEN running interactively without `--force`
- THEN a confirm prompt is shown
- AND if confirmed, files are overwritten
- AND if denied, process exits

#### Scenario: Force flag bypasses prompt
- GIVEN existing `.devcontainer/`
- WHEN `jkit init --name X --force` is run
- THEN files are overwritten without confirmation

#### Scenario: Parameterized mode without force errors
- GIVEN existing `.devcontainer/`
- WHEN `jkit init --name X` is run (parameterized, no `--force`)
- THEN it returns an error "use --force to overwrite existing .devcontainer/"

### R-INIT-TUI-07: Error Handling

All errors in the init flow MUST be user-facing messages (not stack traces). Partial cleanup MUST happen on failure — created files/directories SHOULD be removed on error rollback.

#### Scenario: TUI error is user-friendly
- GIVEN a form validation error (empty project name)
- WHEN the form is submitted
- THEN the error is shown inline in the form
- AND no filesystem changes occur

#### Scenario: Orchestration rollback on error
- GIVEN Orchestrate() has created `.devcontainer/` files
- WHEN the AGNT step fails
- THEN previously created files are best-effort cleaned up
- AND the error is returned to the user

#### Scenario: Invalid flag value returns error
- GIVEN `jkit init --name myproject --agents nonexistent`
- WHEN invoked
- THEN it returns an error "unknown agent: nonexistent"

### Coverage

| Spec | PRD Req |
|------|---------|
| R-INIT-TUI-01 | R-INIT-01, R-INIT-03, R-INIT-05 |
| R-INIT-TUI-02 | R-INIT-03, R-INIT-06 |
| R-INIT-TUI-03 | R-INIT-01, R-INIT-04 |
| R-INIT-TUI-04 | R-INIT-02 |
| R-INIT-TUI-05 | R-INIT-04 |
| R-INIT-TUI-06 | R-INIT-08, R-DEVC-10 |
| R-INIT-TUI-07 | R-DEVC-11 |
