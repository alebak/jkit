# Delta for Devcontainer Init

## ADDED Requirements

### Requirement: R-DEVC-10: Agent Marker Comments

The renderer MUST wrap each selected agent's bash snippet with machine-parseable delimiter comments: `# --- agent:<name> ---` before and `# --- end agent:<name> ---` after the snippet. The gentle-ai install section MUST NOT have agent markers.

#### Scenario: Markers wrap selected agent snippets
- GIVEN `DevcontainerData{SelectedAgents: ["claude"]}`
- WHEN post-create.sh is rendered
- THEN output contains `# --- agent:claude ---` and `# --- end agent:claude ---`
- AND claude's snippet appears between the markers

#### Scenario: No markers when empty selection
- GIVEN `DevcontainerData{SelectedAgents: []}`
- WHEN post-create.sh is rendered
- THEN no agent delimiter comments appear
- AND output is valid bash
