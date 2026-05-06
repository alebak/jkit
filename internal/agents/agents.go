// Package agents manages AI agent configurations, discovery, and skill deployment.
package agents

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alebak/jkit"
)

// SkillDirFor returns the skill directory path for the given agent name.
// Callers should use this to resolve the target before calling DeploySkill.
// Environment variables override the default paths:
//
//	CLAUDE_SKILLS    — override for Claude Code
//	OPENCODE_SKILLS  — override for OpenCode
//	GEMINI_SKILLS    — override for Gemini CLI
func SkillDirFor(ctx context.Context, agentName string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	switch agentName {
	case "claude":
		if dir := os.Getenv("CLAUDE_SKILLS"); dir != "" {
			return dir, nil
		}
		return filepath.Join(home, ".claude", "skills"), nil
	case "opencode":
		if dir := os.Getenv("OPENCODE_SKILLS"); dir != "" {
			return dir, nil
		}
		return filepath.Join(home, ".config", "opencode", "skills"), nil
	case "gemini":
		if dir := os.Getenv("GEMINI_SKILLS"); dir != "" {
			return dir, nil
		}
		return filepath.Join(home, ".gemini", "skills"), nil
	default:
		return "", fmt.Errorf("unknown agent: %s", agentName)
	}
}

// ListAvailable returns agent names discovered from .sh files in
// the templates/agents/ directory of the given filesystem. Names are
// the filenames without the .sh extension, sorted alphabetically.
func ListAvailable(ctx context.Context, fsys fs.FS) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := fs.ReadDir(fsys, "templates/agents")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading agents directory: %w", err)
	}

	var agents []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sh") {
			continue
		}
		agents = append(agents, strings.TrimSuffix(entry.Name(), ".sh"))
	}
	sort.Strings(agents)
	return agents, nil
}

// DeploySkill deploys the named skill into skillDir.
// It copies the embedded SKILL.md to .jkit/agents/skills/{skillName}/
// in the project directory and creates a symlink from skillDir to that location.
func DeploySkill(ctx context.Context, projectDir, skillDir, skillName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Create .jkit/agents/skills/{skillName}/ in the project directory
	jkitSkillDir := filepath.Join(projectDir, ".jkit", "agents", "skills", skillName)
	if err := os.MkdirAll(jkitSkillDir, 0755); err != nil {
		return fmt.Errorf("creating skill directory: %w", err)
	}

	// Copy SKILL.md from embedded assets
	skillContent, err := jkit.SkillsFS.ReadFile("templates/skills/" + skillName + "/SKILL.md")
	if err != nil {
		return fmt.Errorf("reading embedded skill: %w", err)
	}

	targetPath := filepath.Join(jkitSkillDir, "SKILL.md")
	if err := os.WriteFile(targetPath, skillContent, 0644); err != nil {
		return fmt.Errorf("writing SKILL.md: %w", err)
	}

	// Create or recreate the symlink from the agent's skill dir
	symlinkPath := filepath.Join(skillDir, skillName)
	if err := os.Remove(symlinkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing existing symlink: %w", err)
	}
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("creating agent skill directory: %w", err)
	}

	rel, err := filepath.Rel(skillDir, jkitSkillDir)
	if err != nil {
		return fmt.Errorf("computing relative path: %w", err)
	}
	if err := os.Symlink(rel, symlinkPath); err != nil {
		return fmt.Errorf("creating symlink: %w", err)
	}

	return nil
}
