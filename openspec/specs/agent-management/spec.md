# Agent Management Specification

## Purpose

Manage AI agent installation for devcontainer projects and deploy project-specific skills to agent skill directories.

## Requirements

### R-AGNT-01: Agent Discovery

Available agents MUST be discovered from embedded `templates/agents/*.sh` (AgentsFS). Agent name SHALL be the filename minus `.sh`.

#### Scenario: Discover available agents
- GIVEN AgentsFS with `claude.sh`, `opencode.sh`, `gemini.sh`
- WHEN `ListAvailable()` is called
- THEN it returns `["claude", "opencode", "gemini"]`

#### Scenario: Empty AgentsFS
- GIVEN embedded AgentsFS is empty
- WHEN `ListAvailable()` is called
- THEN it returns an empty slice

### R-AGNT-02: Agent Listing

`jkit agents list` MUST show all available agents and mark which are installed (present in post-create.sh agent delimiter comments).

#### Scenario: Installed vs available
- GIVEN post-create.sh marks `claude` and `opencode` but not `gemini`
- WHEN `jkit agents list` runs
- THEN output shows `claude` and `opencode` as installed, `gemini` as available

#### Scenario: No post-create.sh exists
- GIVEN no `.devcontainer/post-create.sh`
- WHEN `jkit agents list` runs
- THEN all agents shown as available with a hint to run `jkit init`

### R-AGNT-03: Agent Add

`jkit agents add [name...]` MUST validate each name against available agents. If all valid, it SHALL regenerate post-create.sh via `renderPostCreate()` with `SelectedAgents` set to the given names.

#### Scenario: Add valid agents
- GIVEN a project with existing post-create.sh installing all agents
- WHEN `jkit agents add claude opencode` runs
- THEN post-create.sh is regenerated with only `claude` and `opencode`
- AND delimiter `# --- agent:claude ---` appears in output

#### Scenario: Invalid agent name
- GIVEN no agent named "invalid" exists in AgentsFS
- WHEN `jkit agents add invalid` runs
- THEN exit code is non-zero
- AND error message names "invalid" as unknown

#### Scenario: No post-create.sh
- GIVEN no `.devcontainer/post-create.sh` exists
- WHEN `jkit agents add claude` runs
- THEN exit code is non-zero
- AND error message tells user to run `jkit init` first

#### Scenario: No arguments
- WHEN `jkit agents add` runs with no arguments
- THEN usage help is printed listing valid agent names
- AND exit code is non-zero

### R-AGNT-04: Agent Remove

`jkit agents remove [name...]` MUST parse post-create.sh agent-delimiter comments, subtract the named agents, and regenerate with the remainder.

#### Scenario: Remove installed agent
- GIVEN post-create.sh selects `claude`, `opencode`, `gemini`
- WHEN `jkit agents remove gemini` runs
- THEN post-create.sh is regenerated with only `claude` and `opencode`

#### Scenario: Remove agent not installed
- GIVEN post-create.sh selects only `claude`
- WHEN `jkit agents remove gemini` runs
- THEN post-create.sh is unchanged
- AND a warning states `gemini` was not installed

#### Scenario: No post-create.sh
- GIVEN no `.devcontainer/post-create.sh` exists
- WHEN `jkit agents remove claude` runs
- THEN exit code is non-zero
- AND error message tells user to run `jkit init` first

#### Scenario: Remove all agents leaves gentle-ai
- GIVEN post-create.sh selects only `claude`
- WHEN `jkit agents remove claude` runs
- THEN post-create.sh has no agent sections
- AND gentle-ai install section is preserved

### R-AGNT-05: Agent Markers in Output

Rendered post-create.sh MUST include `# --- agent:<name> ---` and `# --- end agent:<name> ---` delimiter comments around each agent's bash snippet for machine-parseability.

#### Scenario: Markers present for all selected agents
- GIVEN `SelectedAgents: ["claude"]`
- WHEN post-create.sh is rendered
- THEN output contains `# --- agent:claude ---` and `# --- end agent:claude ---`

### R-AGNT-06: Skill Deployment

`jkit init` MUST deploy the prd-creator skill for each selected agent. It SHALL copy `templates/skills/prd-creator/SKILL.md` to `.jkit/agents/skills/prd-creator/` and create a symlink from each agent's skill directory to that location.

Agent-to-directory mapping:
- Claude Code: `~/.claude/skills/`
- OpenCode: `~/.config/opencode/skills/`
- Gemini CLI: `~/.gemini/skills/`

#### Scenario: Deploy for selected agents
- GIVEN `jkit init` with agents `claude` and `gemini`
- WHEN init completes
- THEN `.jkit/agents/skills/prd-creator/SKILL.md` exists
- AND `~/.claude/skills/prd-creator` is a symlink to `.jkit/agents/skills/prd-creator`
- AND `~/.gemini/skills/prd-creator` is a symlink to `.jkit/agents/skills/prd-creator`
- AND no symlink for opencode exists

#### Scenario: Existing symlink is recreated
- GIVEN `~/.claude/skills/prd-creator` already exists
- WHEN `jkit init` runs
- THEN the symlink is recreated without error

## Coverage

| Req | Description |
|-----|-------------|
| R-AGNT-01 | Agent discovery from embedded FS |
| R-AGNT-02 | `jkit agents list` |
| R-AGNT-03 | `jkit agents add` |
| R-AGNT-04 | `jkit agents remove` |
| R-AGNT-05 | Agent delimiter markers |
| R-AGNT-06 | Skill deployment in init |
