package jkit

import "embed"

//go:embed templates/devcontainer/*
//go:embed templates/devcontainer/.env
//go:embed templates/devcontainer/.env.example
var DevcontainerFS embed.FS

//go:embed templates/agents/*.sh
var AgentsFS embed.FS

//go:embed templates/skills/prd-creator/*
var SkillsFS embed.FS
