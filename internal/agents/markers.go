package agents

import (
	"context"
	"io"
	"os"
	"regexp"
)

const (
	// AgentSectionPrefix is the marker line denoting the start of the agent installation section.
	AgentSectionPrefix = "# --- Agent installations ---\n"

	// AgentStartFmt is a format string for the start marker of an agent section.
	// Use fmt.Sprintf(AgentStartFmt, agentName) to produce "# --- agent:<name> ---\n".
	AgentStartFmt = "# --- agent:%s ---\n"

	// AgentEndFmt is a format string for the end marker of an agent section.
	// Use fmt.Sprintf(AgentEndFmt, agentName) to produce "# --- end agent:<name> ---\n".
	AgentEndFmt = "# --- end agent:%s ---\n"
)

// markerRe matches agent start markers of the form "# --- agent:<name> ---".
var markerRe = regexp.MustCompile(`# --- agent:(\w+) ---`)

// ParsePostCreateMarkers reads the file at postCreatePath and extracts agent names
// from # --- agent:<name> --- markers. Duplicate names are returned only once,
// in order of first appearance.
func ParsePostCreateMarkers(ctx context.Context, postCreatePath string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(postCreatePath)
	if err != nil {
		return nil, err
	}
	return parseMarkers(string(data)), nil
}

// parseMarkers extracts agent names from marker lines in content.
func parseMarkers(content string) []string {
	matches := markerRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	var result []string
	for _, m := range matches {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}
	return result
}

// WriteAgentMarkers writes the AgentSectionPrefix header line to w.
// The agents parameter is reserved for future use (e.g., writing per-agent
// start markers). It is currently only used to decide whether to write
// anything at all: if agents is empty or nil, nothing is written.
func WriteAgentMarkers(ctx context.Context, w io.Writer, agents []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(agents) == 0 {
		return nil
	}
	_, err := io.WriteString(w, "\n"+AgentSectionPrefix)
	return err
}
