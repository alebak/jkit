package devcontainer

// DevcontainerData holds all configuration values for rendering
// a .devcontainer/ directory for a Joomla project.
type DevcontainerData struct {
	ProjectName      string
	JoomlaImage      string
	Timezone         string
	VSCodeExtensions []string
	SelectedAgents   []string
	DBUser           string
	DBPassword       string
	DBName           string
	DBPrefix         string
	SiteName         string
	AdminUser        string
	AdminUsername    string
	AdminPassword    string
	AdminEmail       string
	SMTPHost         string
	RootPassword     string
}

// DefaultDevcontainerData returns a DevcontainerData populated with
// safe defaults per R-INIT-02 (superdev/superpassword) and common
// infrastructure defaults.
func DefaultDevcontainerData() DevcontainerData {
	return DevcontainerData{
		ProjectName: "",
		JoomlaImage: "joomla:6.1-php8.4-apache",
		Timezone:    "UTC",
		VSCodeExtensions: []string{
			"xdebug.php-debug",
			"bmewburn.vscode-intelephense-client",
			"esbenp.prettier-vscode",
		},
		SelectedAgents: []string{"claude", "opencode", "gemini"},
		DBUser:         "joomla",
		DBPassword:     "joomla",
		DBName:         "joomla",
		DBPrefix:       "joom_",
		SiteName:       "Joomla",
		AdminUser:      "superdev",
		AdminUsername:  "superdev",
		AdminPassword:  "superpassword",
		AdminEmail:     "admin@example.com",
		SMTPHost:       "mail:1025",
		RootPassword:   "dev",
	}
}
