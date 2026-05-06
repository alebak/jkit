package devcontainer

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestRender_ValidOutput(t *testing.T) {
	var buf bytes.Buffer
	data := DefaultDevcontainerData()
	data.ProjectName = "TestProject"

	err := Render(context.Background(), &buf, "devcontainer.json", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()

	// Verify output contains the project name
	if !strings.Contains(output, "TestProject") {
		t.Errorf("output should contain project name %q", "TestProject")
	}

	// Verify output is valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("output should be valid JSON: %v", err)
	}

	// Verify the name field
	name, ok := result["name"].(string)
	if !ok || name != "TestProject" {
		t.Errorf("expected name 'TestProject', got %v", result["name"])
	}
}

func TestRender_UnknownName(t *testing.T) {
	var buf bytes.Buffer
	err := Render(context.Background(), &buf, "nonexistent.txt", DefaultDevcontainerData())
	if err == nil {
		t.Error("expected error for unknown template, got nil")
	}
}

func TestRender_PhpCustomIniRaw(t *testing.T) {
	var buf bytes.Buffer
	err := Render(context.Background(), &buf, "php-custom.ini", DefaultDevcontainerData())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()

	// Should contain Xdebug config content
	if !strings.Contains(output, "xdebug.mode=debug") {
		t.Errorf("expected xdebug config in output")
	}

	// Should NOT contain Go template markers (raw copy, not processed)
	if strings.Contains(output, "{{") {
		t.Errorf("php-custom.ini should not contain template markers")
	}

	// Should contain all original content
	if !strings.Contains(output, "zend_extension=xdebug") {
		t.Errorf("expected zend_extension=xdebug in output")
	}
	if !strings.Contains(output, "upload_max_filesize") {
		t.Errorf("expected upload_max_filesize in output")
	}
}

func TestRender_PostCreateSh(t *testing.T) {
	var buf bytes.Buffer
	data := DefaultDevcontainerData()
	data.ProjectName = "TestProject"

	err := Render(context.Background(), &buf, "post-create.sh", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()

	// Should contain the project name from template
	if !strings.Contains(output, "TestProject") {
		t.Errorf("post-create.sh should contain project name")
	}

	// Should contain gentle-ai install (header part)
	if !strings.Contains(output, "gentle-ai") {
		t.Errorf("post-create.sh should reference gentle-ai")
	}

	// Should contain agent scripts (concatenated from templates/agents/)
	if !strings.Contains(output, "Claude Code") {
		t.Errorf("post-create.sh should contain Claude Code agent script")
	}
	if !strings.Contains(output, "OpenCode") {
		t.Errorf("post-create.sh should contain OpenCode agent script")
	}
	if !strings.Contains(output, "Gemini CLI") {
		t.Errorf("post-create.sh should contain Gemini CLI agent script")
	}

	// Should contain agent markers for all three agents
	if !strings.Contains(output, "# --- agent:claude ---") {
		t.Errorf("post-create.sh should contain start marker for claude")
	}
	if !strings.Contains(output, "# --- end agent:claude ---") {
		t.Errorf("post-create.sh should contain end marker for claude")
	}
	if !strings.Contains(output, "# --- agent:opencode ---") {
		t.Errorf("post-create.sh should contain start marker for opencode")
	}
	if !strings.Contains(output, "# --- agent:gemini ---") {
		t.Errorf("post-create.sh should contain start marker for gemini")
	}
}

func TestRender_Dockerfile(t *testing.T) {
	var buf bytes.Buffer
	data := DefaultDevcontainerData()
	data.JoomlaImage = "joomla:5.3-php8.3-apache"

	err := Render(context.Background(), &buf, "Dockerfile", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()

	// Verify the FROM line uses the custom image
	if !strings.Contains(output, "FROM joomla:5.3-php8.3-apache") {
		t.Errorf("Dockerfile should contain custom JoomlaImage, got: %s", output)
	}
}

func TestRender_DockerComposeTimezone(t *testing.T) {
	var buf bytes.Buffer
	data := DefaultDevcontainerData()
	data.Timezone = "America/Argentina/Buenos_Aires"

	err := Render(context.Background(), &buf, "docker-compose.yml", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()

	// Verify the TZ line uses the custom timezone
	if !strings.Contains(output, "TZ: America/Argentina/Buenos_Aires") {
		t.Errorf("docker-compose.yml should contain custom timezone, got: %s", output)
	}
}

func TestRender_EnvFile(t *testing.T) {
	var buf bytes.Buffer
	data := DefaultDevcontainerData()

	err := Render(context.Background(), &buf, ".env", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()

	// Verify admin defaults per R-INIT-02
	if !strings.Contains(output, `JOOMLA_ADMIN_USERNAME="superdev"`) {
		t.Errorf(".env should contain superdev username")
	}
	if !strings.Contains(output, `JOOMLA_ADMIN_PASSWORD="superpassword"`) {
		t.Errorf(".env should contain superpassword")
	}

	// Verify no hardcoded credentials from El Repuestazo remain
	if strings.Contains(output, "El Repuestazo") {
		t.Errorf(".env should not contain 'El Repuestazo'")
	}
	if strings.Contains(output, "development2026") {
		t.Errorf(".env should not contain 'development2026'")
	}
	if strings.Contains(output, "aarroyave") {
		t.Errorf(".env should not contain 'aarroyave'")
	}

	// Should reference template variables (not {{.Field}} left unrendered)
	if strings.Contains(output, "{{") {
		t.Errorf(".env should have all template variables rendered, got unrendered markers")
	}
	if strings.Contains(output, "}}") {
		t.Errorf(".env should have all template variables rendered, got unrendered markers")
	}
}

func TestRender_EnvExample(t *testing.T) {
	var buf bytes.Buffer
	data := DefaultDevcontainerData()

	err := Render(context.Background(), &buf, ".env.example", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()

	// Should render successfully with defaults
	if !strings.Contains(output, `JOOMLA_DB_USER="joomla"`) {
		t.Errorf(".env.example should contain default DB user")
	}

	// Should NOT contain raw template markers
	if strings.Contains(output, "{{") {
		t.Errorf(".env.example should have all template variables rendered")
	}
}

func TestRender_PostCreateSh_AgentFiltering(t *testing.T) {
	var buf bytes.Buffer
	data := DefaultDevcontainerData()
	data.ProjectName = "TestProject"
	data.SelectedAgents = []string{"claude"} // only Claude

	err := Render(context.Background(), &buf, "post-create.sh", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()

	// Should contain Claude Code
	if !strings.Contains(output, "Claude Code") {
		t.Errorf("post-create.sh should contain Claude Code when selected")
	}

	// Should NOT contain OpenCode or Gemini when not selected
	if strings.Contains(output, "OpenCode") {
		t.Errorf("post-create.sh should NOT contain OpenCode when not selected")
	}
	if strings.Contains(output, "Gemini CLI") {
		t.Errorf("post-create.sh should NOT contain Gemini CLI when not selected")
	}

	// Should contain agent markers per R-AGNT-05
	if !strings.Contains(output, "# --- agent:claude ---") {
		t.Errorf("post-create.sh should contain start marker for claude")
	}
	if !strings.Contains(output, "# --- end agent:claude ---") {
		t.Errorf("post-create.sh should contain end marker for claude")
	}

	// Should NOT contain markers for unselected agents
	if strings.Contains(output, "# --- agent:opencode ---") {
		t.Errorf("post-create.sh should NOT contain start marker for opencode")
	}
	if strings.Contains(output, "# --- agent:gemini ---") {
		t.Errorf("post-create.sh should NOT contain start marker for gemini")
	}
}

func TestRender_AllTemplatesNonEmpty(t *testing.T) {
	// Verify all 7 templates produce non-empty output
	templates := []string{
		"devcontainer.json",
		"Dockerfile",
		"docker-compose.yml",
		".env",
		".env.example",
		"post-create.sh",
		"php-custom.ini",
	}

	data := DefaultDevcontainerData()
	data.ProjectName = "NonEmptyTest"

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			var buf bytes.Buffer
			err := Render(context.Background(), &buf, tmpl, data)
			if err != nil {
				t.Fatalf("Render(%q) error = %v", tmpl, err)
			}
			if buf.Len() == 0 {
				t.Errorf("Render(%q) produced empty output", tmpl)
			}
		})
	}
}
