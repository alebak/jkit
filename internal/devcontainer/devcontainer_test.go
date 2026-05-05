package devcontainer

import (
	"reflect"
	"testing"
)

func TestDefaultDevcontainerData(t *testing.T) {
	data := DefaultDevcontainerData()

	// Verify type is correct (struct with expected fields)
	if data.ProjectName != "" {
		t.Errorf("expected empty ProjectName, got %q", data.ProjectName)
	}

	if data.JoomlaImage != "joomla:6.1-php8.4-apache" {
		t.Errorf("expected joomla:6.1-php8.4-apache, got %q", data.JoomlaImage)
	}

	if data.Timezone != "UTC" {
		t.Errorf("expected UTC, got %q", data.Timezone)
	}

	// R-INIT-02 defaults: superdev/superpassword
	if data.AdminUser != "superdev" {
		t.Errorf("expected superdev, got %q", data.AdminUser)
	}
	if data.AdminUsername != "superdev" {
		t.Errorf("expected superdev, got %q", data.AdminUsername)
	}
	if data.AdminPassword != "superpassword" {
		t.Errorf("expected superpassword, got %q", data.AdminPassword)
	}
	if data.AdminEmail != "admin@example.com" {
		t.Errorf("expected admin@example.com, got %q", data.AdminEmail)
	}

	// DB defaults
	if data.DBUser != "joomla" {
		t.Errorf("expected joomla, got %q", data.DBUser)
	}
	if data.DBPassword != "joomla" {
		t.Errorf("expected joomla, got %q", data.DBPassword)
	}
	if data.DBName != "joomla" {
		t.Errorf("expected joomla, got %q", data.DBName)
	}
	if data.DBPrefix != "joom_" {
		t.Errorf("expected joom_, got %q", data.DBPrefix)
	}

	// Infrastructure defaults
	if data.SMTPHost != "mail:1025" {
		t.Errorf("expected mail:1025, got %q", data.SMTPHost)
	}
	if data.RootPassword != "dev" {
		t.Errorf("expected dev, got %q", data.RootPassword)
	}

	// VSCode extensions - must have defaults
	expectedExtensions := []string{"xdebug.php-debug", "bmewburn.vscode-intelephense-client", "esbenp.prettier-vscode"}
	if !reflect.DeepEqual(data.VSCodeExtensions, expectedExtensions) {
		t.Errorf("expected %v, got %v", expectedExtensions, data.VSCodeExtensions)
	}

	// SelectedAgents - must default to all 3
	expectedAgents := []string{"claude", "opencode", "gemini"}
	if !reflect.DeepEqual(data.SelectedAgents, expectedAgents) {
		t.Errorf("expected %v, got %v", expectedAgents, data.SelectedAgents)
	}
}

func TestDevcontainerDataStructFields(t *testing.T) {
	// Verify the struct has all 16 fields by constructing one
	data := DevcontainerData{
		ProjectName:      "test-project",
		JoomlaImage:      "joomla:test",
		Timezone:         "America/New_York",
		VSCodeExtensions: []string{"ext1"},
		SelectedAgents:   []string{"claude", "opencode"},
		DBUser:           "dbuser",
		DBPassword:       "dbpass",
		DBName:           "dbname",
		DBPrefix:         "pref_",
		SiteName:         "My Site",
		AdminUser:        "admin",
		AdminUsername:    "adminuser",
		AdminPassword:    "adminpass",
		AdminEmail:       "admin@test.com",
		SMTPHost:         "mail:1025",
		RootPassword:     "rootpass",
	}

	if data.ProjectName != "test-project" {
		t.Errorf("ProjectName field mismatch")
	}
	if data.JoomlaImage != "joomla:test" {
		t.Errorf("JoomlaImage field mismatch")
	}
	if data.Timezone != "America/New_York" {
		t.Errorf("Timezone field mismatch")
	}
	if len(data.VSCodeExtensions) != 1 || data.VSCodeExtensions[0] != "ext1" {
		t.Errorf("VSCodeExtensions field mismatch")
	}
	if len(data.SelectedAgents) != 2 || data.SelectedAgents[0] != "claude" {
		t.Errorf("SelectedAgents field mismatch")
	}
	if data.DBUser != "dbuser" {
		t.Errorf("DBUser field mismatch")
	}
	if data.DBPassword != "dbpass" {
		t.Errorf("DBPassword field mismatch")
	}
	if data.DBName != "dbname" {
		t.Errorf("DBName field mismatch")
	}
	if data.DBPrefix != "pref_" {
		t.Errorf("DBPrefix field mismatch")
	}
	if data.SiteName != "My Site" {
		t.Errorf("SiteName field mismatch")
	}
	if data.AdminUser != "admin" {
		t.Errorf("AdminUser field mismatch")
	}
	if data.AdminUsername != "adminuser" {
		t.Errorf("AdminUsername field mismatch")
	}
	if data.AdminPassword != "adminpass" {
		t.Errorf("AdminPassword field mismatch")
	}
	if data.AdminEmail != "admin@test.com" {
		t.Errorf("AdminEmail field mismatch")
	}
	if data.SMTPHost != "mail:1025" {
		t.Errorf("SMTPHost field mismatch")
	}
	if data.RootPassword != "rootpass" {
		t.Errorf("RootPassword field mismatch")
	}
}
