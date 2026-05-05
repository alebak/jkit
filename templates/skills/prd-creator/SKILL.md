# Skill: PRD Creator

## Description
Creates a Product Requirements Document (PRD.md) for a Joomla project or extension. The PRD defines the product vision, target audience, use cases, technical requirements, and acceptance criteria.

## When to Use
- At the start of a new Joomla project
- When defining requirements for a new extension (component, module, plugin)
- When scoping a new feature for an existing project
- When a stakeholder requests formal product documentation

## Instructions

1. Ask the user for the project/extension name and type
2. Determine the target Joomla version (5.x or 6.x)
3. Define the core problem and target audience
4. List key features with priority (P0/P1/P2)
5. Define technical requirements:
   - PHP version compatibility
   - Database schema needs
   - Joomla API compatibility
   - Frontend assets (if any)
6. Define acceptance criteria for each feature
7. Write to `PRD.md` in the project root

## Output
A `PRD.md` file with:
- **Product Vision** — one-sentence purpose
- **Target Audience** — who uses it
- **User Stories** — what users need to do
- **Functional Requirements** — what the system must do (numbered: FR-01, FR-02…)
- **Technical Requirements** — stack constraints (numbered: TR-01, TR-02…)
- **Out of Scope** — explicitly what won't be built
- **Success Metrics** — how to know it's working
