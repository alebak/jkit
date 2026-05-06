package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alebak/jkit/internal/generator"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [name]",
	Short: "Build a Joomla extension package",
	Long: `Build a Joomla extension into a distributable zip archive.
Reads the extension registry to locate the extension directory,
then creates a .zip file in the builds/ directory.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		// Read the extension registry
		registry, err := generator.ReadRegistry(context.Background(), cwd)
		if err != nil {
			return fmt.Errorf("failed to read registry: %w", err)
		}

		// Find the extension by name
		entry := registry.Find(name)
		if entry == nil {
			return fmt.Errorf("extension %q not found in registry", name)
		}

		// Determine the extension source directory based on type
		sourceDir := sourceDirForEntry(cwd, entry)
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			return fmt.Errorf("extension directory not found: %s", sourceDir)
		}

		// Create builds/ directory
		buildsDir := filepath.Join(cwd, buildsDirName)
		if err := os.MkdirAll(buildsDir, 0755); err != nil {
			return fmt.Errorf("failed to create builds directory: %w", err)
		}

		// Create zip archive
		zipPath := filepath.Join(buildsDir, name+".zip")
		zipFile, err := os.Create(zipPath)
		if err != nil {
			return fmt.Errorf("failed to create zip file: %w", err)
		}

		w := zip.NewWriter(zipFile)

		// Walk the extension directory and add files to the zip
		err = filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Calculate the relative path for the zip entry
			relPath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				return fmt.Errorf("relative path error: %w", err)
			}

			// Skip the root
			if relPath == "." {
				return nil
			}

			if d.IsDir() {
				// Add directory entry (with trailing slash)
				_, err := w.Create(relPath + "/")
				return err
			}

			// Add file entry
			fh := &zip.FileHeader{
				Name:     relPath,
				Method:   zip.Deflate,
				Modified: time.Now(),
			}
			f, err := w.CreateHeader(fh)
			if err != nil {
				return fmt.Errorf("create zip entry %s: %w", relPath, err)
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read %s: %w", path, err)
			}

			if _, err := f.Write(content); err != nil {
				return fmt.Errorf("write %s to zip: %w", relPath, err)
			}

			return nil
		})

		// Close the zip writer (must close before checking errors, or we lose the zip)
		closeErr := w.Close()
		if zcErr := zipFile.Close(); zcErr != nil && closeErr == nil {
			closeErr = zcErr
		}

		if err != nil {
			_ = os.Remove(zipPath) // best-effort cleanup
			return fmt.Errorf("failed to walk extension directory: %w", err)
		}
		if closeErr != nil {
			_ = os.Remove(zipPath) // best-effort cleanup
			return fmt.Errorf("failed to finalize zip: %w", closeErr)
		}

		// Update registry with build timestamp
		entry.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		if err := generator.WriteRegistry(context.Background(), cwd, *entry); err != nil {
			return fmt.Errorf("failed to update registry: %w", err)
		}

		cmd.Printf("✅ Created %s/%s.zip\n", buildsDirName, name)
		return nil
	},
}

const buildsDirName = "builds"

// sourceDirForEntry returns the expected source directory for a registry entry.
func sourceDirForEntry(projectRoot string, entry *generator.RegistryEntry) string {
	switch entry.Type {
	case string(generator.TypeComponent):
		return filepath.Join(projectRoot, "components", entry.Name)
	case string(generator.TypeModule):
		return filepath.Join(projectRoot, "modules", entry.Name)
	case string(generator.TypePlugin):
		// The plugin name format is "plg_group_name" — extract group and name
		parts := strings.SplitN(entry.Name, "_", 3)
		if len(parts) == 3 {
			return filepath.Join(projectRoot, "plugins", parts[1], parts[2])
		}
		return filepath.Join(projectRoot, "plugins", entry.Name)
	case string(generator.TypeTemplate):
		return filepath.Join(projectRoot, "templates", entry.Name)
	case string(generator.TypeLibrary):
		return filepath.Join(projectRoot, "libraries", entry.Name)
	case string(generator.TypePackage):
		return filepath.Join(projectRoot, "packages")
	}
	return filepath.Join(projectRoot, entry.Name)
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
