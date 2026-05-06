package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alebak/jkit/internal/generator"
	"github.com/spf13/cobra"
)

// checkOverwrite checks if an extension already exists in the project.
// When isTTY is true and force is false, prompts the user interactively.
func checkOverwrite(ctx context.Context, projectRoot string, data generator.ExtensionData, force, isTTY bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var paths []string
	sanitizedName := generator.SanitizeName(data.Name)

	switch data.Type {
	case generator.TypeComponent:
		// Components can exist in both admin and site locations
		paths = append(paths,
			filepath.Join(projectRoot, "administrator", "components", data.FullName()),
			filepath.Join(projectRoot, "components", data.FullName()),
		)
	case generator.TypeModule:
		paths = append(paths, filepath.Join(projectRoot, "modules", data.FullName()))
	case generator.TypePlugin:
		paths = append(paths, filepath.Join(projectRoot, "plugins", data.Group, sanitizedName))
	case generator.TypeTemplate:
		paths = append(paths, filepath.Join(projectRoot, "templates", data.FullName()))
	case generator.TypeLibrary:
		paths = append(paths, filepath.Join(projectRoot, "libraries", data.Vendor, sanitizedName))
	case generator.TypePackage:
		paths = append(paths, filepath.Join(projectRoot, "packages", data.FullName()+".xml"))
	}

	existingPaths := make([]string, 0)
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			existingPaths = append(existingPaths, p)
		}
	}

	if len(existingPaths) > 0 {
		if force {
			return nil
		}
		if !isTTY {
			return fmt.Errorf("extension %s already exists at %s. Use --force to overwrite",
				data.FullName(), strings.Join(existingPaths, ", "))
		}
		// Interactive TTY prompt (best-effort — skip if prompt fails)
		fmt.Printf("Extension %s already exists at %s. Overwrite? [y/N]: ",
			data.FullName(), strings.Join(existingPaths, ", "))
		var response string
		_, _ = fmt.Scanln(&response) // best-effort: scan failure → empty string → treat as "no"
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("aborted by user")
		}
	}

	return nil
}

// isTerminal checks whether stdin is a terminal.
func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

var createCmd = &cobra.Command{
	Use:   "create [type]",
	Short: "Scaffold a new Joomla extension",
	Long: `Scaffold a new Joomla extension of the specified type.
Supported types: component, module, plugin, template, library, package.

Generates extension files from embedded templates into the detected
Joomla project directory.`,
	Args: cobra.ExactArgs(1),
	ValidArgs: []string{
		"component",
		"module",
		"plugin",
		"template",
		"library",
		"package",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		extType := args[0]

		// Validate type
		validTypes := map[string]generator.ExtensionType{
			"component": generator.TypeComponent,
			"module":    generator.TypeModule,
			"plugin":    generator.TypePlugin,
			"template":  generator.TypeTemplate,
			"library":   generator.TypeLibrary,
			"package":   generator.TypePackage,
		}
		extTypeVal, ok := validTypes[extType]
		if !ok {
			return fmt.Errorf("invalid extension type %q. Valid types: component, module, plugin, template, library, package", extType)
		}

		// Parse flags — cobra only errors here when the flag was never
		// registered in init(). All five flags are registered below.
		name, _ := cmd.Flags().GetString("name")
		vendor, _ := cmd.Flags().GetString("vendor")
		joomlaVersion, _ := cmd.Flags().GetString("joomla-version")
		pluginGroup, _ := cmd.Flags().GetString("plugin-group")
		force, _ := cmd.Flags().GetBool("force")

		// Validate required flags
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		if vendor == "" {
			return fmt.Errorf("--vendor is required")
		}

		// Detect Joomla project from current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		isJoomla, err := generator.DetectJoomlaProject(context.Background(), cwd)
		if err != nil {
			return fmt.Errorf("error detecting Joomla project: %w", err)
		}
		if !isJoomla {
			return fmt.Errorf("not a Joomla project: %s. Run jkit init first or ensure configuration.php and standard Joomla directories exist", cwd)
		}

		// Build ExtensionData from flags
		opts := []generator.ExtensionOption{
			generator.WithJoomlaVersion(joomlaVersion),
		}
		if extTypeVal == generator.TypePlugin && pluginGroup != "" {
			opts = append(opts, generator.WithGroup(pluginGroup))
		}
		data := generator.NewExtensionData(name, vendor, extTypeVal, opts...)

		// Check overwrite
		tty := isTerminal()
		if err := checkOverwrite(context.Background(), cwd, data, force, tty); err != nil {
			return err
		}

		// Determine target directory and generate
		target := cwd

		createdFiles, err := generator.Generate(context.Background(), data, target)
		if err != nil {
			return fmt.Errorf("generation failed: %w", err)
		}

		// Register extension in the YAML registry.
		// For plugins, use PluginName (includes group) so the build command
		// can find the correct directory.
		registryName := data.FullName()
		if data.Type == generator.TypePlugin && data.Group != "" {
			registryName = data.PluginName()
		}
		registryEntry := generator.RegistryEntry{
			Name:      registryName,
			Type:      string(data.Type),
			Version:   data.Version,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		if err := generator.WriteRegistry(context.Background(), cwd, registryEntry); err != nil {
			return fmt.Errorf("failed to update registry: %w", err)
		}

		cmd.Printf("✅ Created %s %s (%d files)\n", data.Type, data.FullName(), len(createdFiles))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().String("name", "", "Extension name (e.g. Blog)")
	createCmd.Flags().String("vendor", "", "Vendor/author namespace (e.g. Acme)")
	createCmd.Flags().String("joomla-version", "", "Target Joomla version (e.g. 5.3)")
	createCmd.Flags().String("plugin-group", "", "Plugin group (e.g. system, user, content)")
	createCmd.Flags().Bool("force", false, "Overwrite existing extension without prompting")
}
