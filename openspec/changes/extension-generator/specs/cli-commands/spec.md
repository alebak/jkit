# Delta for cli-commands

## MODIFIED Requirements

### R-CLI-02: Create and Build Commands

The CLI MUST define a root cobra command with three subcommands: `init`, `create`, `build`. The `create` subcommand SHALL scaffold Joomla extension skeletons. The `build` subcommand SHALL package extensions as `.zip` archives. Both SHALL accept extension-specific flags via cobra persistent flags.
(Previously: All subcommands printed "not yet implemented")

#### Scenario: Create component with flags
- GIVEN the built `jkit` binary in a Joomla project root
- WHEN `jkit create component --name=Blog --vendor=Alebak --joomla-version=5` is executed
- THEN exit code is 0
- AND extension skeleton is written to the project

#### Scenario: Build existing extension
- GIVEN an extension `com_blog` registered in `extensions.jkit.yaml`
- WHEN `jkit build com_blog` is executed
- THEN exit code is 0
- AND `builds/com_blog-1.0.0.zip` is created

#### Scenario: Unknown build target
- GIVEN no extension named `com_unknown` exists
- WHEN `jkit build com_unknown` is executed
- THEN error output mentions "not found"
- AND exit code is 1

#### Scenario: Invalid create type
- GIVEN `jkit create unknown`
- WHEN executed
- THEN error output lists valid types
- AND exit code is 1

#### Scenario: Init subcommand unchanged
- GIVEN `jkit init`
- WHEN executed (stub still pending implementation)
- THEN output contains "not yet implemented"
