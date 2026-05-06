package jkit

import "embed"

//go:embed templates/devcontainer/*
var DevcontainerFS embed.FS

//go:embed templates/agents/*.sh
var AgentsFS embed.FS

//go:embed templates/skills/prd-creator/*
var SkillsFS embed.FS

//go:embed templates/extensions/module
//go:embed templates/extensions/plugin
//go:embed templates/extensions/library
//go:embed templates/extensions/component
//go:embed templates/extensions/template
//go:embed templates/extensions/package
var ExtensionsFS embed.FS

//go:embed templates/mcp/*.json
var McpFS embed.FS

//go:embed images.yaml
var ImagesYAML []byte
