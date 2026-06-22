## ADDED Requirements

### Requirement: Interactive command palette

`lufy-ai` SHALL provide an interactive command palette when invoked without arguments in an interactive terminal.

#### Scenario: TTY without args opens palette
- **WHEN** the user runs `lufy-ai` without arguments in an interactive terminal
- **THEN** the CLI shows a Bubble Tea command palette with command names and descriptions

#### Scenario: Non-TTY preserves help behavior
- **WHEN** `lufy-ai` is invoked without arguments in a non-interactive environment
- **THEN** the CLI preserves the existing help/usage behavior and does not block waiting for input

#### Scenario: Command selection shows parameters
- **WHEN** the user selects a command from the palette
- **THEN** the palette shows editable parameters with descriptions, defaults and allowed choices when applicable

#### Scenario: Confirm executes generated args
- **WHEN** the user confirms a command selection
- **THEN** the CLI executes the equivalent generated `lufy-ai <command> ...` args through the existing command dispatcher

### Requirement: Declarative command registry

The palette SHALL use a declarative registry for commands and parameters instead of hardcoding per-command UI behavior.

#### Scenario: Boolean flags are explicit
- **WHEN** a command has a boolean flag such as `--dry-run`, `--yes`, `--json` or `--verbose`
- **THEN** the registry marks it as boolean and the palette lets the user toggle it

#### Scenario: Choices are constrained
- **WHEN** a command has constrained values such as `--scope` or `--tool`
- **THEN** the registry exposes choices and the generated args include only the selected choice

#### Scenario: Text values are editable
- **WHEN** a command has a text value such as `--target`, `--to`, `--base` or positional path/query
- **THEN** the palette lets the user type the value and includes it in generated args when non-empty or required

### Requirement: Scriptable commands remain stable

Existing commands and flags SHALL keep their current behavior for direct invocations.

#### Scenario: Direct command bypasses palette
- **WHEN** the user runs `lufy-ai setup --dry-run` or any other command with arguments
- **THEN** the CLI executes that command directly without opening the palette
