package generator

import (
	"testing"
)

func TestPrefix_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		extType  ExtensionType
		expected string
	}{
		{"component", TypeComponent, "com_"},
		{"module", TypeModule, "mod_"},
		{"plugin", TypePlugin, "plg_"},
		{"template", TypeTemplate, "tpl_"},
		{"library", TypeLibrary, "lib_"},
		{"package", TypePackage, "pkg_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ExtensionData{Name: "test", Type: tt.extType}
			if got := d.Prefix(); got != tt.expected {
				t.Errorf("Prefix() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestPrefix_UnknownType(t *testing.T) {
	d := ExtensionData{Name: "test", Type: "unknown"}
	if got := d.Prefix(); got != "" {
		t.Errorf("Prefix() for unknown type = %q, want empty string", got)
	}
}

func TestFullName_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		data     ExtensionData
		expected string
	}{
		{
			name:     "component",
			data:     ExtensionData{Name: "Blog", Type: TypeComponent},
			expected: "com_blog",
		},
		{
			name:     "module",
			data:     ExtensionData{Name: "MyModule", Type: TypeModule},
			expected: "mod_mymodule",
		},
		{
			name:     "plugin",
			data:     ExtensionData{Name: "Auth", Type: TypePlugin, Group: "user"},
			expected: "plg_auth",
		},
		{
			name:     "template",
			data:     ExtensionData{Name: "Cassiopeia", Type: TypeTemplate},
			expected: "tpl_cassiopeia",
		},
		{
			name:     "library",
			data:     ExtensionData{Name: "Foom", Type: TypeLibrary},
			expected: "lib_foom",
		},
		{
			name:     "package",
			data:     ExtensionData{Name: "All", Type: TypePackage},
			expected: "pkg_all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.FullName(); got != tt.expected {
				t.Errorf("FullName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFullName_CaseHandling(t *testing.T) {
	tests := []struct {
		name     string
		data     ExtensionData
		expected string
	}{
		{
			name:     "PascalCase name becomes lowercase",
			data:     ExtensionData{Name: "MyComponent", Type: TypeComponent},
			expected: "com_mycomponent",
		},
		{
			name:     "already lowercase stays",
			data:     ExtensionData{Name: "mycomponent", Type: TypeComponent},
			expected: "com_mycomponent",
		},
		{
			name:     "mixed case normalizes",
			data:     ExtensionData{Name: "MyCOMPONENT", Type: TypeComponent},
			expected: "com_mycomponent",
		},
		{
			name:     "spaces become underscores",
			data:     ExtensionData{Name: "My Component", Type: TypeComponent},
			expected: "com_my_component",
		},
		{
			name:     "hyphens become underscores",
			data:     ExtensionData{Name: "my-component", Type: TypeModule},
			expected: "mod_my_component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.FullName(); got != tt.expected {
				t.Errorf("FullName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestClassName_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		data     ExtensionData
		expected string
	}{
		{
			name:     "simple name",
			data:     ExtensionData{Name: "Blog", Type: TypeComponent},
			expected: "Blog",
		},
		{
			name:     "lowercase input",
			data:     ExtensionData{Name: "mycomponent", Type: TypeComponent},
			expected: "Mycomponent",
		},
		{
			name:     "snake_case to PascalCase",
			data:     ExtensionData{Name: "content_history", Type: TypeModule},
			expected: "ContentHistory",
		},
		{
			name:     "with underscores",
			data:     ExtensionData{Name: "my_module", Type: TypeModule},
			expected: "MyModule",
		},
		{
			name:     "already PascalCase",
			data:     ExtensionData{Name: "ContentHistory", Type: TypeModule},
			expected: "ContentHistory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.ClassName(); got != tt.expected {
				t.Errorf("ClassName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNamespace_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		data     ExtensionData
		expected string
	}{
		{
			name:     "component",
			data:     ExtensionData{Name: "Blog", Vendor: "Alebak", Type: TypeComponent},
			expected: `Alebak\Component\Blog`,
		},
		{
			name:     "module",
			data:     ExtensionData{Name: "MyModule", Vendor: "Alebak", Type: TypeModule},
			expected: `Alebak\Module\MyModule`,
		},
		{
			name:     "plugin with group",
			data:     ExtensionData{Name: "Auth", Vendor: "Alebak", Type: TypePlugin, Group: "user"},
			expected: `Alebak\Plugin\User\Auth`,
		},
		{
			name:     "template",
			data:     ExtensionData{Name: "Cassiopeia", Vendor: "Alebak", Type: TypeTemplate},
			expected: `Alebak\Template\Cassiopeia`,
		},
		{
			name:     "library",
			data:     ExtensionData{Name: "Foom", Vendor: "Alebak", Type: TypeLibrary},
			expected: `Alebak\Library\Foom`,
		},
		{
			name:     "package",
			data:     ExtensionData{Name: "All", Vendor: "Alebak", Type: TypePackage},
			expected: `Alebak\Package\All`,
		},
		{
			name:     "with numbers in name",
			data:     ExtensionData{Name: "Joomla4", Vendor: "MyVendor", Type: TypeComponent},
			expected: `MyVendor\Component\Joomla4`,
		},
		{
			name:     "namespace with underscores in class name",
			data:     ExtensionData{Name: "content_history", Vendor: "MyVendor", Type: TypeComponent},
			expected: `MyVendor\Component\ContentHistory`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.Namespace(); got != tt.expected {
				t.Errorf("Namespace() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestShortNamespace_Component(t *testing.T) {
	tests := []struct {
		name     string
		data     ExtensionData
		expected string
	}{
		{
			name:     "component administrator",
			data:     ExtensionData{Name: "Blog", Vendor: "Alebak", Type: TypeComponent, Group: "administrator"},
			expected: `Administrator`,
		},
		{
			name:     "component site",
			data:     ExtensionData{Name: "Blog", Vendor: "Alebak", Type: TypeComponent, Group: "site"},
			expected: `Site`,
		},
		{
			name:     "component empty group defaults to site",
			data:     ExtensionData{Name: "Blog", Vendor: "Alebak", Type: TypeComponent},
			expected: `Site`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.ShortNamespace(); got != tt.expected {
				t.Errorf("ShortNamespace() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestShortNamespace_NonComponent(t *testing.T) {
	tests := []struct {
		name string
		data ExtensionData
	}{
		{"module", ExtensionData{Name: "M", Vendor: "V", Type: TypeModule}},
		{"plugin", ExtensionData{Name: "P", Vendor: "V", Type: TypePlugin, Group: "sys"}},
		{"template", ExtensionData{Name: "T", Vendor: "V", Type: TypeTemplate}},
		{"library", ExtensionData{Name: "L", Vendor: "V", Type: TypeLibrary}},
		{"package", ExtensionData{Name: "K", Vendor: "V", Type: TypePackage}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.ShortNamespace(); got != "" {
				t.Errorf("ShortNamespace() for %s = %q, want empty string", tt.name, got)
			}
		})
	}
}

func TestGroupPascal(t *testing.T) {
	tests := []struct {
		name     string
		group    string
		expected string
	}{
		{"single word", "system", "System"},
		{"snake_case", "user", "User"},
		{"PascalCase input", "Content", "Content"},
		{"multi-word group", "quick_icon", "QuickIcon"},
		{"empty group", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ExtensionData{Name: "Test", Type: TypePlugin, Group: tt.group}
			if got := d.GroupPascal(); got != tt.expected {
				t.Errorf("GroupPascal() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestPluginName_Plugin(t *testing.T) {
	tests := []struct {
		name     string
		data     ExtensionData
		expected string
	}{
		{
			name:     "plugin with group",
			data:     ExtensionData{Name: "Auth", Type: TypePlugin, Group: "user"},
			expected: "plg_user_auth",
		},
		{
			name:     "plugin with system group",
			data:     ExtensionData{Name: "MyPlugin", Type: TypePlugin, Group: "system"},
			expected: "plg_system_myplugin",
		},
		{
			name:     "plugin with content group",
			data:     ExtensionData{Name: "CustomArticle", Type: TypePlugin, Group: "content"},
			expected: "plg_content_customarticle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.PluginName(); got != tt.expected {
				t.Errorf("PluginName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestPluginName_NonPlugin(t *testing.T) {
	tests := []struct {
		name string
		data ExtensionData
	}{
		{"component", ExtensionData{Name: "C", Type: TypeComponent}},
		{"module", ExtensionData{Name: "M", Type: TypeModule}},
		{"library", ExtensionData{Name: "L", Type: TypeLibrary}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.PluginName(); got != "" {
				t.Errorf("PluginName() for %s = %q, want empty string", tt.name, got)
			}
		})
	}
}

func TestPluginName_PluginNoGroup(t *testing.T) {
	d := ExtensionData{Name: "Auth", Type: TypePlugin}
	if got := d.PluginName(); got != "" {
		t.Errorf("PluginName() with empty group = %q, want empty string", got)
	}
}

func TestNewExtensionData(t *testing.T) {
	tests := []struct {
		name     string
		data     ExtensionData
		expected ExtensionData
	}{
		{
			name: "minimal component",
			data: NewExtensionData("Blog", "Alebak", TypeComponent),
			expected: ExtensionData{
				Name:    "Blog",
				Vendor:  "Alebak",
				Type:    TypeComponent,
				Version: "1.0.0",
				Year:    "2026",
			},
		},
		{
			name: "plugin with group",
			data: NewExtensionData("Auth", "MyVendor", TypePlugin, WithGroup("user")),
			expected: ExtensionData{
				Name:    "Auth",
				Vendor:  "MyVendor",
				Type:    TypePlugin,
				Group:   "user",
				Version: "1.0.0",
				Year:    "2026",
			},
		},
		{
			name: "module with all options",
			data: NewExtensionData("MyModule", "Vendor", TypeModule,
				WithDescription("A test module"),
				WithVersion("2.0.0"),
				WithAuthor("John Doe"),
				WithYear("2025"),
				WithJoomlaVersion("6"),
			),
			expected: ExtensionData{
				Name:          "MyModule",
				Vendor:        "Vendor",
				Type:          TypeModule,
				Description:   "A test module",
				Version:       "2.0.0",
				Author:        "John Doe",
				Year:          "2025",
				JoomlaVersion: "6",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.Name != tt.expected.Name {
				t.Errorf("Name = %q, want %q", tt.data.Name, tt.expected.Name)
			}
			if tt.data.Vendor != tt.expected.Vendor {
				t.Errorf("Vendor = %q, want %q", tt.data.Vendor, tt.expected.Vendor)
			}
			if tt.data.Type != tt.expected.Type {
				t.Errorf("Type = %q, want %q", tt.data.Type, tt.expected.Type)
			}
			if tt.data.Group != tt.expected.Group {
				t.Errorf("Group = %q, want %q", tt.data.Group, tt.expected.Group)
			}
			if tt.data.Description != tt.expected.Description {
				t.Errorf("Description = %q, want %q", tt.data.Description, tt.expected.Description)
			}
			if tt.data.Version != tt.expected.Version {
				t.Errorf("Version = %q, want %q", tt.data.Version, tt.expected.Version)
			}
			if tt.data.Author != tt.expected.Author {
				t.Errorf("Author = %q, want %q", tt.data.Author, tt.expected.Author)
			}
			if tt.data.Year != tt.expected.Year {
				t.Errorf("Year = %q, want %q", tt.data.Year, tt.expected.Year)
			}
			if tt.data.JoomlaVersion != tt.expected.JoomlaVersion {
				t.Errorf("JoomlaVersion = %q, want %q", tt.data.JoomlaVersion, tt.expected.JoomlaVersion)
			}
		})
	}
}

func TestNewExtensionData_UnknownType(t *testing.T) {
	data := NewExtensionData("Test", "Vendor", "unknown")
	if data.Type != "unknown" {
		t.Errorf("Type = %q, want 'unknown'", data.Type)
	}
}
