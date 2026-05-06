# Delta for CLI Commands

## MODIFIED Requirements

### Requirement: R-CLI-02: Cobra Command Stubs

The CLI MUST define a root cobra command with subcommands: `init`, `create`, `build`, `agents`. The `init`, `create`, `build` subcommands MUST print "not yet implemented" and exit 0. The `agents` subcommand MUST define `list`, `add`, and `remove` sub-subcommands.
(Previously: three subcommands: init, create, build, all stubs)

#### Scenario: Root command produces no output by default
- GIVEN the built `jkit` binary
- WHEN invoked with no arguments
- THEN it prints cobra help text (no error)
- AND exit code is 0

#### Scenario: Init subcommand stub
- GIVEN `jkit init`
- WHEN executed
- THEN output contains "not yet implemented"
- AND exit code is 0

#### Scenario: Create subcommand stub
- GIVEN `jkit create`
- WHEN executed
- THEN output contains "not yet implemented"

#### Scenario: Build subcommand stub
- GIVEN `jkit build`
- WHEN executed
- THEN output contains "not yet implemented"

#### Scenario: Agents list subcommand
- GIVEN the built `jkit` binary
- WHEN `jkit agents list` executes
- THEN exit code is 0
- AND output lists available agent names

#### Scenario: Agents add subcommand
- GIVEN the built `jkit` binary
- WHEN `jkit agents add` executes
- THEN it processes the request (success or error)
- AND exit code reflects the result

#### Scenario: Agents remove subcommand
- GIVEN the built `jkit` binary
- WHEN `jkit agents remove` executes
- THEN it processes the request (success or error)
- AND exit code reflects the result
