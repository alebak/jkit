package generator

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alebak/jkit"
)

// Generate walks the embedded extension templates for the given ExtensionData.Type,
// renders .tmpl files with text/template against data, copies .raw files verbatim,
// and writes the result to targetDir. Returns the list of created file paths.
func Generate(ctx context.Context, data ExtensionData, targetDir string) ([]string, error) {
	switch data.Type {
	case TypeComponent:
		return generateComponent(ctx, data, targetDir)
	case TypeModule:
		return generateFromFS(ctx, jkit.ExtensionsFS,
			"templates/extensions/module", data,
			filepath.Join(targetDir, "modules", data.FullName()))
	case TypePlugin:
		return generateFromFS(ctx, jkit.ExtensionsFS,
			"templates/extensions/plugin", data,
			filepath.Join(targetDir, "plugins", data.Group, SanitizeName(data.Name)))
	case TypeTemplate:
		return generateFromFS(ctx, jkit.ExtensionsFS,
			"templates/extensions/template", data,
			filepath.Join(targetDir, "templates", data.FullName()))
	case TypeLibrary:
		return generateFromFS(ctx, jkit.ExtensionsFS,
			"templates/extensions/library", data,
			filepath.Join(targetDir, "libraries", data.Vendor, SanitizeName(data.Name)))
	case TypePackage:
		return generateFromFS(ctx, jkit.ExtensionsFS,
			"templates/extensions/package", data,
			filepath.Join(targetDir, "packages"))
	default:
		return nil, fmt.Errorf("unknown extension type: %s", data.Type)
	}
}

// generateComponent handles the special case for components, which span
// two Joomla directory roots: administrator/components/{name}/ and components/{name}/.
func generateComponent(ctx context.Context, data ExtensionData, targetDir string) ([]string, error) {
	baseDir := "templates/extensions/component"
	var allFiles []string

	// Administrator subtree → administrator/components/{fullname}/
	adminPrefix := filepath.Join("administrator", "components", data.FullName())
	adminFiles, err := generateFromFS(ctx, jkit.ExtensionsFS,
		baseDir+"/administrator", data, filepath.Join(targetDir, adminPrefix))
	if err != nil {
		return nil, fmt.Errorf("component administrator: %w", err)
	}
	allFiles = append(allFiles, adminFiles...)

	// Site subtree → components/{fullname}/
	sitePrefix := filepath.Join("components", data.FullName())
	siteFiles, err := generateFromFS(ctx, jkit.ExtensionsFS,
		baseDir+"/site", data, filepath.Join(targetDir, sitePrefix))
	if err != nil {
		return nil, fmt.Errorf("component site: %w", err)
	}
	allFiles = append(allFiles, siteFiles...)

	return allFiles, nil
}

// generateFromFS is the core walk-and-render engine. Walks an fs.FS subdirectory,
// renders .tmpl files with text/template, copies .raw files verbatim, and creates
// the directory tree under targetDir.
func generateFromFS(ctx context.Context, fsys fs.FS, subDir string, data ExtensionData, targetDir string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Check that the subdirectory exists in the FS
	if _, err := fs.ReadDir(fsys, subDir); err != nil {
		return nil, nil // nonexistent subdir is not an error
	}

	var createdFiles []string
	var createdDirs []string

	trackDir := func(dir string) {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			createdDirs = append(createdDirs, dir)
		}
	}

	walkErr := fs.WalkDir(fsys, subDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		// Skip the root directory itself
		if path == subDir {
			return nil
		}

		relPath, err := filepath.Rel(subDir, path)
		if err != nil {
			return fmt.Errorf("relative path for %s: %w", path, err)
		}

		outPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			trackDir(outPath)
			if err := os.MkdirAll(outPath, 0755); err != nil {
				return fmt.Errorf("mkdir %s: %w", outPath, err)
			}
			return nil
		}

		// Determine file type by suffix
		var outputPath string
		isTemplate := strings.HasSuffix(relPath, ".tmpl")
		isRaw := strings.HasSuffix(relPath, ".raw")

		switch {
		case isTemplate:
			outputPath = strings.TrimSuffix(outPath, ".tmpl")
		case isRaw:
			outputPath = strings.TrimSuffix(outPath, ".raw")
		default:
			// Skip non-template, non-raw files (e.g. .gitkeep)
			return nil
		}

		// Ensure output directory exists
		parentDir := filepath.Dir(outputPath)
		trackDir(parentDir)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", parentDir, err)
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		var writeContent []byte

		if isTemplate {
			t, err := template.New(relPath).Option("missingkey=error").Parse(string(content))
			if err != nil {
				return fmt.Errorf("parse template %s: %w", relPath, err)
			}

			var buf strings.Builder
			if err := t.Execute(&buf, data); err != nil {
				return fmt.Errorf("execute template %s: %w", relPath, err)
			}
			writeContent = []byte(buf.String())
		} else {
			// .raw files: copy verbatim
			writeContent = content
		}

		if err := os.WriteFile(outputPath, writeContent, 0644); err != nil {
			return fmt.Errorf("write %s: %w", outputPath, err)
		}

		createdFiles = append(createdFiles, outputPath)
		return nil
	})

	if walkErr != nil {
		// Rollback: best-effort cleanup of created files and directories
		for _, f := range createdFiles {
			_ = os.Remove(f)
		}
		// Rollback: delete created directories in reverse order
		for i := len(createdDirs) - 1; i >= 0; i-- {
			_ = os.Remove(createdDirs[i])
		}
		return nil, walkErr
	}

	return createdFiles, nil
}
