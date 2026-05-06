package generator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_ComponentType(t *testing.T) {
	data := NewExtensionData("Blog", "Alebak", TypeComponent,
		WithDescription("A test component"),
		WithAuthor("John Doe"),
		WithJoomlaVersion("5.3"),
	)
	targetDir := t.TempDir()

	files, err := Generate(context.Background(), data, targetDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Component should produce 7 files (manifest + admin services + admin controller +
	// admin extension + admin test + site dispatcher + site test)
	if len(files) < 5 {
		t.Fatalf("Generate() returned %d files for component, want >= 5", len(files))
	}

	// Verify manifest XML — now under administrator/components/com_blog/
	adminDir := filepath.Join(targetDir, "administrator", "components", "com_blog")
	manifestPath := filepath.Join(adminDir, "name.xml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatalf("name.xml not created at %s", manifestPath)
	}
	content, _ := os.ReadFile(manifestPath)
	if !strings.Contains(string(content), "com_blog") {
		t.Errorf("manifest should contain 'com_blog', got: %s", string(content))
	}
	if !strings.Contains(string(content), `type="component"`) {
		t.Errorf("manifest should have type='component', got: %s", string(content))
	}

	// Verify admin service provider exists
	providerPath := filepath.Join(adminDir, "services", "provider.php")
	if _, err := os.Stat(providerPath); os.IsNotExist(err) {
		t.Error("administrator/components/com_blog/services/provider.php not created")
	}

	// Verify site dispatcher exists
	siteDir := filepath.Join(targetDir, "components", "com_blog")
	dispPath := filepath.Join(siteDir, "src", "Dispatcher", "Dispatcher.php")
	if _, err := os.Stat(dispPath); os.IsNotExist(err) {
		t.Error("components/com_blog/src/Dispatcher/Dispatcher.php not created")
	}

	// Verify PHP namespace rendered
	phpContent, _ := os.ReadFile(providerPath)
	if !strings.Contains(string(phpContent), `\\Alebak\Component\Blog`) {
		t.Errorf("provider.php should contain namespace Alebak\\Component\\Blog, got: %s", string(phpContent))
	}
}

func TestGenerate_ModuleType(t *testing.T) {
	data := NewExtensionData("MyModule", "Alebak", TypeModule,
		WithDescription("A test module"),
		WithVersion("2.0.0"),
	)
	targetDir := t.TempDir()

	files, err := Generate(context.Background(), data, targetDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(files) < 3 {
		t.Fatalf("Generate() returned %d files for module, want >= 3", len(files))
	}

	// Verify manifest XML — under modules/mod_mymodule/
	moduleDir := filepath.Join(targetDir, "modules", "mod_mymodule")
	manifestPath := filepath.Join(moduleDir, "mod_name.xml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatal("mod_name.xml not created")
	}
	content, _ := os.ReadFile(manifestPath)
	if !strings.Contains(string(content), "mod_mymodule") {
		t.Errorf("manifest should contain 'mod_mymodule', got: %s", string(content))
	}

	// Verify services directory was created
	svcPath := filepath.Join(moduleDir, "services")
	info, err := os.Stat(svcPath)
	if err != nil {
		t.Fatalf("services directory error: %v", err)
	}
	if !info.IsDir() {
		t.Error("services is not a directory")
	}

	// Verify PHP namespace (module template uses {{ .Namespace }} directly)
	svcProvider, _ := os.ReadFile(filepath.Join(moduleDir, "services", "provider.php"))
	if !strings.Contains(string(svcProvider), `Alebak\Module\MyModule`) {
		t.Errorf("provider.php should contain Alebak\\Module\\MyModule, got: %s", string(svcProvider))
	}
}

func TestGenerate_PluginType(t *testing.T) {
	data := NewExtensionData("Auth", "Alebak", TypePlugin,
		WithGroup("user"),
		WithDescription("A test plugin"),
	)
	targetDir := t.TempDir()

	files, err := Generate(context.Background(), data, targetDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(files) < 3 {
		t.Fatalf("Generate() returned %d files for plugin, want >= 3", len(files))
	}

	// Verify manifest — under plugins/user/auth/
	pluginDir := filepath.Join(targetDir, "plugins", "user", "auth")
	manifestPath := filepath.Join(pluginDir, "name.xml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatal("name.xml not created")
	}
	content, _ := os.ReadFile(manifestPath)
	if !strings.Contains(string(content), "plg_user_auth") {
		t.Errorf("manifest should contain 'plg_user_auth', got: %s", string(content))
	}

	// Verify provider contains the plugin class name and package reference
	svcProvider, _ := os.ReadFile(filepath.Join(pluginDir, "services", "provider.php"))
	if !strings.Contains(string(svcProvider), `\PlgUserAuth`) {
		t.Errorf("provider.php should contain \\PlgUserAuth, got: %s", string(svcProvider))
	}
	if !strings.Contains(string(svcProvider), `plg_user_auth`) {
		t.Errorf("provider.php should contain plg_user_auth, got: %s", string(svcProvider))
	}
}

func TestGenerate_LibraryType(t *testing.T) {
	data := NewExtensionData("Foom", "Alebak", TypeLibrary,
		WithDescription("A test library"),
	)
	targetDir := t.TempDir()

	files, err := Generate(context.Background(), data, targetDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(files) < 3 {
		t.Fatalf("Generate() returned %d files for library, want >= 3", len(files))
	}

	// Verify manifest — under libraries/Alebak/foom/
	libDir := filepath.Join(targetDir, "libraries", "Alebak", "foom")
	manifestPath := filepath.Join(libDir, "lib_name.xml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatal("lib_name.xml not created")
	}
	content, _ := os.ReadFile(manifestPath)
	if !strings.Contains(string(content), "lib_foom") {
		t.Errorf("manifest should contain 'lib_foom', got: %s", string(content))
	}
	if !strings.Contains(string(content), "libraryname") {
		t.Errorf("manifest should contain 'libraryname', got: %s", string(content))
	}
}

func TestGenerate_TemplateType(t *testing.T) {
	data := NewExtensionData("Cassiopeia", "Alebak", TypeTemplate,
		WithDescription("A test template"),
	)
	targetDir := t.TempDir()

	files, err := Generate(context.Background(), data, targetDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(files) < 4 {
		t.Fatalf("Generate() returned %d files for template, want >= 4", len(files))
	}

	// Verify manifest — under templates/tpl_cassiopeia/
	templateDir := filepath.Join(targetDir, "templates", "tpl_cassiopeia")
	manifestPath := filepath.Join(templateDir, "templateDetails.xml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatal("templateDetails.xml not created")
	}
	content, _ := os.ReadFile(manifestPath)
	if !strings.Contains(string(content), "tpl_cassiopeia") {
		t.Errorf("manifest should contain 'tpl_cassiopeia', got: %s", string(content))
	}
}

func TestGenerate_PackageType(t *testing.T) {
	data := NewExtensionData("All", "Alebak", TypePackage,
		WithDescription("A test package"),
	)
	targetDir := t.TempDir()

	files, err := Generate(context.Background(), data, targetDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Generate() returned %d files for package, want 1", len(files))
	}

	// Verify manifest
	manifestPath := filepath.Join(targetDir, "packages", "pkg_name.xml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatal("pkg_name.xml not created")
	}
	content, _ := os.ReadFile(manifestPath)
	if !strings.Contains(string(content), "pkg_all") {
		t.Errorf("manifest should contain 'pkg_all', got: %s", string(content))
	}
	if !strings.Contains(string(content), "packagename") {
		t.Errorf("manifest should contain 'packagename', got: %s", string(content))
	}
	if !strings.Contains(string(content), `type="package"`) {
		t.Errorf("manifest should have type='package', got: %s", string(content))
	}
}

func TestGenerate_AllTypesDoNotError(t *testing.T) {
	types := []struct {
		name string
		data ExtensionData
	}{
		{"component", NewExtensionData("Test", "Alebak", TypeComponent)},
		{"module", NewExtensionData("Test", "Alebak", TypeModule)},
		{"plugin", NewExtensionData("Test", "Alebak", TypePlugin, WithGroup("system"))},
		{"template", NewExtensionData("Test", "Alebak", TypeTemplate)},
		{"library", NewExtensionData("Test", "Alebak", TypeLibrary)},
		{"package", NewExtensionData("Test", "Alebak", TypePackage)},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			targetDir := t.TempDir()
			files, err := Generate(context.Background(), tt.data, targetDir)
			if err != nil {
				t.Fatalf("Generate(%s) error = %v", tt.name, err)
			}
			if len(files) == 0 {
				t.Errorf("Generate(%s) returned 0 files", tt.name)
			}
		})
	}
}

// TestGenerate_TemplateHTML validates template output has HTML5 structure
func TestGenerate_TemplateHTML(t *testing.T) {
	data := NewExtensionData("MyTemplate", "Alebak", TypeTemplate)
	targetDir := t.TempDir()

	_, err := Generate(context.Background(), data, targetDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// index.php should exist and contain Joomla boilerplate
	tmplDir := filepath.Join(targetDir, "templates", "tpl_mytemplate")
	indexPath := filepath.Join(tmplDir, "index.php")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatal("index.php not created")
	}
	content, _ := os.ReadFile(indexPath)
	if !strings.Contains(string(content), "<!DOCTYPE html>") {
		t.Error("index.php should contain DOCTYPE html")
	}
	if !strings.Contains(string(content), "<jdoc:include") {
		t.Error("index.php should contain jdoc include statements")
	}
}
