package initpkg

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alebak/jkit"
	"github.com/alebak/jkit/internal/agents"
	"github.com/alebak/jkit/internal/devcontainer"
	"github.com/alebak/jkit/internal/generator"
	"github.com/alebak/jkit/internal/mcp"
)

// Orchestrate runs the full init pipeline: DEVC → AGNT → EXTG → MCPS.
// Fails fast on first error; performs best-effort cleanup on failure.
func Orchestrate(ctx context.Context, cfg InitConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine current directory: %w", err)
	}

	// Step 0: Overwrite guard
	if err := checkOverwriteGuard(filepath.Join(cwd, ".devcontainer"), cfg.Force); err != nil {
		return err
	}

	// Track created files for rollback
	var createdFiles []string
	trackCleanup := func(paths ...string) {
		createdFiles = append(createdFiles, paths...)
	}

	// Step 1: DEVC — Render devcontainer templates
	devcontainerDir := filepath.Join(cwd, ".devcontainer")
	if err := os.MkdirAll(devcontainerDir, 0755); err != nil {
		return fmt.Errorf("creating .devcontainer directory: %w", err)
	}
	trackCleanup(devcontainerDir)

	data := cfg.ToDevcontainerData()
	devcontainerFiles := []string{
		"devcontainer.json",
		"Dockerfile",
		"docker-compose.yml",
		".env",
		".env.example",
		"post-create.sh",
		"php-custom.ini",
	}

	for _, name := range devcontainerFiles {
		if err := ctx.Err(); err != nil {
			return rollback(createdFiles, err)
		}

		var buf bytes.Buffer
		if err := devcontainer.Render(ctx, &buf, name, data); err != nil {
			return rollback(createdFiles, fmt.Errorf("DEVC render %s: %w", name, err))
		}

		path := filepath.Join(devcontainerDir, name)
		if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
			return rollback(createdFiles, fmt.Errorf("DEVC write %s: %w", name, err))
		}
		trackCleanup(path)
	}

	// Step 2: AGNT — Deploy skills for each selected agent
	if len(cfg.Agents) > 0 {
		for _, agentName := range cfg.Agents {
			if err := ctx.Err(); err != nil {
				return rollback(createdFiles, err)
			}

			skillDir, err := agents.SkillDirFor(ctx, agentName)
			if err != nil {
				return rollback(createdFiles, fmt.Errorf("AGNT resolve %s: %w", agentName, err))
			}

			if err := agents.DeploySkill(ctx, cwd, skillDir, "prd-creator-joomla"); err != nil {
				return rollback(createdFiles, fmt.Errorf("AGNT deploy %s: %w", agentName, err))
			}
		}
	}

	// Step 3: EXTG — Generate default component extension
	extData := generator.NewExtensionData(cfg.ProjectName, "jkit", generator.TypeComponent)
	if _, err := generator.Generate(ctx, extData, cwd); err != nil {
		return rollback(createdFiles, fmt.Errorf("EXTG generate: %w", err))
	}

	// Step 4: MCPS — Deploy MCP for playwright + mariadb to first agent's config
	if len(cfg.Agents) > 0 {
		firstAgent := cfg.Agents[0]
		configPath, err := mcp.MCPConfigPathFor(ctx, firstAgent)
		if err != nil {
			return rollback(createdFiles, fmt.Errorf("MCPS config path: %w", err))
		}

		mcpServers := []string{"playwright", "mariadb"}
		for _, mcpName := range mcpServers {
			if err := ctx.Err(); err != nil {
				return rollback(createdFiles, err)
			}

			templateData, err := jkit.McpFS.ReadFile(filepath.Join("templates", "mcp", mcpName+".json"))
			if err != nil {
				return rollback(createdFiles, fmt.Errorf("MCPS read template %s: %w", mcpName, err))
			}

			if err := mcp.DeployMCP(ctx, configPath, mcpName, templateData); err != nil {
				return rollback(createdFiles, fmt.Errorf("MCPS deploy %s: %w", mcpName, err))
			}
		}
	}

	// Step 5: Create .gitignore and builds/ directory
	gitignorePath := filepath.Join(cwd, ".gitignore")
	if err := appendGitignore(gitignorePath); err != nil {
		return rollback(createdFiles, fmt.Errorf("creating .gitignore: %w", err))
	}

	buildsDir := filepath.Join(cwd, "builds")
	if err := os.MkdirAll(buildsDir, 0755); err != nil {
		return rollback(createdFiles, fmt.Errorf("creating builds/ directory: %w", err))
	}

	return nil
}

// checkOverwriteGuard checks if a .devcontainer directory exists and returns
// an error if it does and Force is false.
func checkOverwriteGuard(devcontainerDir string, force bool) error {
	_, err := os.Stat(devcontainerDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("checking .devcontainer/: %w", err)
	}
	if !force {
		return fmt.Errorf("use --force to overwrite existing .devcontainer/")
	}
	return nil
}

// rollback performs best-effort cleanup of created files and returns the
// original error. It removes files first, then directories.
func rollback(files []string, originalErr error) error {
	// Remove files in reverse order (last created, first removed)
	for i := len(files) - 1; i >= 0; i-- {
		_ = os.RemoveAll(files[i]) // best-effort cleanup
	}
	return originalErr
}

// appendGitignore adds a .env entry to .gitignore if not already present.
func appendGitignore(path string) error {
	entry := ".env\n"

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(path, []byte(entry), 0644)
		}
		return fmt.Errorf("reading .gitignore: %w", err)
	}

	// Check if .env entry already exists (exact line match)
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == ".env" {
			return nil
		}
	}

	// Append to existing file
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening .gitignore: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("writing .gitignore: %w", err)
	}
	return nil
}

// extractZipBytes extracts a zip archive into the target directory.
func extractZipBytes(data []byte, targetDir string) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}

	for _, f := range r.File {
		// Prevent Zip Slip
		path := filepath.Join(targetDir, f.Name)
		if !filepath.HasPrefix(path, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", path, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", path, err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("opening %s in zip: %w", f.Name, err)
		}

		out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("creating %s: %w", path, err)
		}

		if _, err := io.Copy(out, rc); err != nil {
			rc.Close()
			out.Close()
			return fmt.Errorf("writing %s: %w", path, err)
		}

		rc.Close()
		out.Close()
	}

	return nil
}
