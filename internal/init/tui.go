package initpkg

import (
	"context"
	"fmt"
	"os"

	"github.com/alebak/jkit"
	"github.com/alebak/jkit/internal/agents"
	"github.com/charmbracelet/huh"
)

// RunInteractive launches the huh TUI and returns the collected configuration.
// If the session is not a TTY, it returns an error explaining that
// parameterized mode should be used instead.
func RunInteractive(ctx context.Context) (InitConfig, error) {
	if err := ctx.Err(); err != nil {
		return InitConfig{}, err
	}

	// Detect TTY — if not a terminal, error out
	fi, err := os.Stdout.Stat()
	if err != nil || (fi.Mode()&os.ModeCharDevice) == 0 {
		return InitConfig{}, fmt.Errorf("not a terminal; use parameterized mode with --name or --image flags")
	}

	// Start with defaults
	cfg := DefaultInitConfig()

	// Load images (filesystem → cache → remote → defaults per DD-03)
	images, err := LoadImages()
	if err != nil {
		return InitConfig{}, fmt.Errorf("loading images: %w", err)
	}

	// Get available agents
	availableAgents, err := agents.ListAvailable(ctx, jkit.AgentsFS)
	if err != nil {
		return InitConfig{}, fmt.Errorf("loading agents: %w", err)
	}

	// Step 1: Project name
	nameInput := huh.NewInput().
		Title("Project name").
		Prompt("? ").
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("project name cannot be empty")
			}
			return nil
		}).
		Value(&cfg.ProjectName)

	// Step 2: Joomla image selection
	var imageOptions []huh.Option[ImageEntry]
	for _, img := range images {
		imageOptions = append(imageOptions, huh.NewOption(
			fmt.Sprintf("%s — %s", img.Tag, img.Description),
			img,
		))
	}
	var selectedImage ImageEntry
	if len(images) > 0 {
		selectedImage = images[0]
		cfg.JoomlaImage = selectedImage.Tag
	}

	imageSelect := huh.NewSelect[ImageEntry]().
		Title("Joomla image").
		Options(imageOptions...).
		Value(&selectedImage)

	// Step 3: Agent multi-select
	var selectedAgents []string
	var agentOptions []huh.Option[string]
	for _, a := range availableAgents {
		agentOptions = append(agentOptions, huh.NewOption(a, a))
	}

	agentSelect := huh.NewMultiSelect[string]().
		Title("AI agents").
		Options(agentOptions...).
		Value(&selectedAgents)

	// Step 4: Timezone
	timezoneInput := huh.NewInput().
		Title("Timezone").
		Prompt("? ").
		Value(&cfg.Timezone)

	// Step 5: Quickstart path
	quickstartInput := huh.NewInput().
		Title("Quickstart .zip path (optional)").
		Prompt("? ").
		Value(&cfg.Quickstart)

	// Step 6: Overwrite confirmation (checked before final confirm)
	// If .devcontainer/ exists and Force is false, prompt for overwrite.
	// We handle this by checking after the form completes.

	// Step 7: Final confirmation
	var confirmed bool
	confirm := huh.NewConfirm().
		Title("Create project?").
		Affirmative("Yes").
		Negative("No").
		Value(&confirmed)

	// Build the form group
	form := huh.NewForm(
		huh.NewGroup(nameInput),
		huh.NewGroup(imageSelect),
		huh.NewGroup(agentSelect),
		huh.NewGroup(timezoneInput),
		huh.NewGroup(quickstartInput),
		huh.NewGroup(confirm),
	)

	if err := form.RunWithContext(ctx); err != nil {
		return InitConfig{}, fmt.Errorf("form cancelled: %w", err)
	}

	// Apply selections
	cfg.JoomlaImage = selectedImage.Tag
	cfg.Agents = selectedAgents

	// Final confirmation check
	if !confirmed {
		return InitConfig{}, fmt.Errorf("aborted by user")
	}

	// Overwrite check
	if _, err := os.Stat(".devcontainer"); err == nil {
		var overwrite bool
		confirmOverwrite := huh.NewConfirm().
			Title("Overwrite existing .devcontainer/?").
			Affirmative("Yes").
			Negative("No").
			Value(&overwrite)

		overwriteForm := huh.NewForm(huh.NewGroup(confirmOverwrite))
		if err := overwriteForm.RunWithContext(ctx); err != nil {
			return InitConfig{}, fmt.Errorf("overwrite check cancelled: %w", err)
		}
		if !overwrite {
			return InitConfig{}, fmt.Errorf("aborted by user")
		}
	}

	return cfg, nil
}
