package devcontainer

import (
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"strings"
	"text/template"

	"github.com/alebak/jkit"
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
		"json": func(v any) string {
			b, err := json.Marshal(v)
			if err != nil {
				return "[]"
			}
			return string(b)
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
// concatenates agent bash snippets for the selected agents.
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

	// Gather agent names to include: all if SelectedAgents is empty
	var selected map[string]bool
	if d, ok := data.(DevcontainerData); ok && len(d.SelectedAgents) > 0 {
		selected = make(map[string]bool)
		for _, a := range d.SelectedAgents {
			selected[strings.ToLower(a)] = true
		}
	}

	entries, err := fs.ReadDir(jkit.AgentsFS, "templates/agents")
	if err != nil {
		return err
	}

	// Write a separator comment before agent snippets
	if _, err := io.WriteString(w, "\n# --- Agent installations ---\n"); err != nil {
		return err
	}

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sh") {
			continue
		}
		// Filter by SelectedAgents if specified
		agentName := strings.TrimSuffix(entry.Name(), ".sh")
		if selected != nil && !selected[agentName] {
			continue
		}
		content, err := jkit.AgentsFS.ReadFile("templates/agents/" + entry.Name())
		if err != nil {
			return err
		}
		if _, err := w.Write(content); err != nil {
			return err
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}

	return nil
}
