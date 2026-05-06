package devcontainer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"text/template"

	"github.com/alebak/jkit"
	"github.com/alebak/jkit/internal/agents"
)

// Render reads the named template from the embedded filesystem, processes
// it according to its type, and writes the result to w.
//
// Special handling:
//   - php-custom.ini: copied as raw bytes (no template processing)
//   - post-create.sh: rendered as Go template, then agent bash snippets
//     from templates/agents/ are concatenated at the bottom
//   - All other files: parsed as Go templates with the provided data
func Render(ctx context.Context, w io.Writer, name string, data any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	switch name {
	case "php-custom.ini":
		return renderRaw(ctx, w, "templates/devcontainer/"+name)
	case "post-create.sh":
		return renderPostCreate(ctx, w, data)
	default:
		return renderTemplate(ctx, w, name, data)
	}
}

// renderRaw copies a file from the embedded FS as raw bytes.
func renderRaw(ctx context.Context, w io.Writer, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	content, err := jkit.DevcontainerFS.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = w.Write(content)
	return err
}

// renderTemplate parses and executes a Go template from the embedded FS.
func renderTemplate(ctx context.Context, w io.Writer, name string, data any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	funcMap := template.FuncMap{
		"json": func(v any) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("json template func: %w", err)
			}
			return string(b), nil
		},
	}
	tmpl, err := template.New(name).Funcs(funcMap).ParseFS(
		jkit.DevcontainerFS,
		"templates/devcontainer/"+name,
	)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}

// renderPostCreate renders the post-create.sh template header and then
// concatenates agent bash snippets for the selected agents, wrapping
// each with # --- agent:<name> --- / # --- end agent:<name> --- markers.
func renderPostCreate(ctx context.Context, w io.Writer, data any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Render the Go template header
	tmpl, err := template.New("post-create.sh").ParseFS(
		jkit.DevcontainerFS,
		"templates/devcontainer/post-create.sh",
	)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(w, data); err != nil {
		return err
	}

	// Determine selected agents:
	//   - nil SelectedAgents means "include all" (default)
	//   - empty slice means "include none" (explicit removal)
	//   - non-empty slice means "include only these"
	var selected map[string]bool
	if d, ok := data.(DevcontainerData); ok && d.SelectedAgents != nil {
		selected = make(map[string]bool)
		for _, a := range d.SelectedAgents {
			selected[strings.ToLower(a)] = true
		}
	}

	entries, err := fs.ReadDir(jkit.AgentsFS, "templates/agents")
	if err != nil {
		return err
	}

	// Collect matching agent names
	var agentNames []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sh") {
			continue
		}
		agentName := strings.ToLower(strings.TrimSuffix(entry.Name(), ".sh"))
		if selected != nil && !selected[agentName] {
			continue
		}
		agentNames = append(agentNames, agentName)
	}

	// Write the agent section header (only if there are agents)
	if len(agentNames) > 0 {
		if err := agents.WriteAgentMarkers(ctx, w, agentNames); err != nil {
			return err
		}
	}

	// Write each agent section wrapped in markers
	for _, agentName := range agentNames {
		if err := ctx.Err(); err != nil {
			return err
		}

		// Start marker
		if _, err := fmt.Fprintf(w, agents.AgentStartFmt, agentName); err != nil {
			return err
		}

		// Agent bash content
		content, err := jkit.AgentsFS.ReadFile("templates/agents/" + agentName + ".sh")
		if err != nil {
			return err
		}
		if _, err := w.Write(content); err != nil {
			return err
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}

		// End marker
		if _, err := fmt.Fprintf(w, agents.AgentEndFmt, agentName); err != nil {
			return err
		}
	}

	return nil
}
