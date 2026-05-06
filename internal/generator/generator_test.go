package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestGenerateFromFS_HappyPath(t *testing.T) {
	fsys := fstest.MapFS{
		"component/name.xml.tmpl": &fstest.MapFile{
			Data: []byte(`<extension type="component"><name>{{ .FullName }}</name></extension>`),
		},
		"component/site/src/Dispatcher/Dispatcher.php.tmpl": &fstest.MapFile{
			Data: []byte(`<?php namespace {{ .Namespace }}; class Dispatcher {}`),
		},
	}

	data := ExtensionData{Name: "Blog", Vendor: "Alebak", Type: TypeComponent}
	targetDir := t.TempDir()

	files, err := generateFromFS(context.Background(), fsys, "component", data, targetDir)
	if err != nil {
		t.Fatalf("generateFromFS() error = %v", err)
	}

	if len(files) != 2 {
		t.Errorf("generateFromFS() returned %d files, want 2", len(files))
	}

	// Check name.xml was rendered
	xmlPath := filepath.Join(targetDir, "name.xml")
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		t.Fatal("name.xml was not created")
	}
	content, _ := os.ReadFile(xmlPath)
	want := `<extension type="component"><name>com_blog</name></extension>`
	if string(content) != want {
		t.Errorf("name.xml content = %q, want %q", string(content), want)
	}

	// Check Dispatcher.php was rendered
	dispPath := filepath.Join(targetDir, "site", "src", "Dispatcher", "Dispatcher.php")
	if _, err := os.Stat(dispPath); os.IsNotExist(err) {
		t.Fatal("Dispatcher.php was not created")
	}
	content, _ = os.ReadFile(dispPath)
	want = `<?php namespace Alebak\Component\Blog; class Dispatcher {}`
	if string(content) != want {
		t.Errorf("Dispatcher.php content = %q, want %q", string(content), want)
	}
}

func TestGenerateFromFS_StripsTmplSuffix(t *testing.T) {
	fsys := fstest.MapFS{
		"test/hello.php.tmpl": &fstest.MapFile{
			Data: []byte(`Hello {{ .Name }}`),
		},
	}

	data := ExtensionData{Name: "World", Type: TypeModule}
	targetDir := t.TempDir()

	files, err := generateFromFS(context.Background(), fsys, "test", data, targetDir)
	if err != nil {
		t.Fatalf("generateFromFS() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}

	// Output file should not have .tmpl suffix
	wantPath := filepath.Join(targetDir, "hello.php")
	if files[0] != wantPath {
		t.Errorf("returned path = %q, want %q", files[0], wantPath)
	}
	if _, err := os.Stat(wantPath); os.IsNotExist(err) {
		t.Errorf("hello.php was not created (got .tmpl suffix)")
	}
	// Ensure .tmpl file does NOT exist
	tmplPath := filepath.Join(targetDir, "hello.php.tmpl")
	if _, err := os.Stat(tmplPath); err == nil {
		t.Errorf("hello.php.tmpl should not exist on disk")
	}
}

func TestGenerateFromFS_RawFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"test/config.ini.raw": &fstest.MapFile{
			Data: []byte(`setting=value`),
		},
	}

	data := ExtensionData{Name: "Test", Type: TypeModule}
	targetDir := t.TempDir()

	files, err := generateFromFS(context.Background(), fsys, "test", data, targetDir)
	if err != nil {
		t.Fatalf("generateFromFS() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}

	wantPath := filepath.Join(targetDir, "config.ini")
	if files[0] != wantPath {
		t.Errorf("returned path = %q, want %q", files[0], wantPath)
	}

	content, _ := os.ReadFile(wantPath)
	if string(content) != "setting=value" {
		t.Errorf("config.ini content = %q, want %q", string(content), "setting=value")
	}
}

func TestGenerateFromFS_TemplateSyntaxError(t *testing.T) {
	fsys := fstest.MapFS{
		"test/bad.tmpl": &fstest.MapFile{
			Data: []byte(`{{ .MissingField }}`),
		},
		"test/good.tmpl": &fstest.MapFile{
			Data: []byte(`{{ .Name }}`),
		},
	}

	data := ExtensionData{Name: "Test", Type: TypeModule}
	targetDir := t.TempDir()

	_, err := generateFromFS(context.Background(), fsys, "test", data, targetDir)
	if err == nil {
		t.Fatal("expected error for missing template field, got nil")
	}

	// Verify rollback: good.tmpl should NOT exist on disk
	goodPath := filepath.Join(targetDir, "good")
	if _, err := os.Stat(goodPath); err == nil {
		t.Errorf("good file was not rolled back after error")
	}
}

func TestGenerateFromFS_NonTmplFilesAreSkipped(t *testing.T) {
	fsys := fstest.MapFS{
		"test/readme.txt": &fstest.MapFile{
			Data: []byte(`readme`),
		},
		"test/code.php.tmpl": &fstest.MapFile{
			Data: []byte(`<?php echo '{{ .Name }}';`),
		},
	}

	data := ExtensionData{Name: "Test", Type: TypeModule}
	targetDir := t.TempDir()

	files, err := generateFromFS(context.Background(), fsys, "test", data, targetDir)
	if err != nil {
		t.Fatalf("generateFromFS() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("got %d files, want 1 (only .tmpl files should be processed)", len(files))
	}

	// readme.txt should not be created
	if _, err := os.Stat(filepath.Join(targetDir, "readme.txt")); err == nil {
		t.Errorf("readme.txt should not exist (non-template file)")
	}

	// code.php should be created
	if _, err := os.Stat(filepath.Join(targetDir, "code.php")); os.IsNotExist(err) {
		t.Errorf("code.php should exist")
	}
}

func TestGenerateFromFS_NestedDirectoryStructure(t *testing.T) {
	fsys := fstest.MapFS{
		"test/level1/level2/file.txt.tmpl": &fstest.MapFile{
			Data: []byte(`deep {{ .Name }}`),
		},
		"test/level1/file.txt.tmpl": &fstest.MapFile{
			Data: []byte(`shallow {{ .Name }}`),
		},
	}

	data := ExtensionData{Name: "Nest", Type: TypeModule}
	targetDir := t.TempDir()

	files, err := generateFromFS(context.Background(), fsys, "test", data, targetDir)
	if err != nil {
		t.Fatalf("generateFromFS() error = %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}

	// Check nested file
	nestedPath := filepath.Join(targetDir, "level1", "level2", "file.txt")
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Errorf("nested file not created: %s", nestedPath)
	}
	content, _ := os.ReadFile(nestedPath)
	if string(content) != "deep Nest" {
		t.Errorf("nested content = %q, want %q", string(content), "deep Nest")
	}

	// Check shallow file
	shallowPath := filepath.Join(targetDir, "level1", "file.txt")
	if _, err := os.Stat(shallowPath); os.IsNotExist(err) {
		t.Errorf("shallow file not created: %s", shallowPath)
	}
	content, _ = os.ReadFile(shallowPath)
	if string(content) != "shallow Nest" {
		t.Errorf("shallow content = %q, want %q", string(content), "shallow Nest")
	}
}

func TestGenerateFromFS_AllTemplateFields(t *testing.T) {
	fsys := fstest.MapFS{
		"test/output.txt.tmpl": &fstest.MapFile{
			Data: []byte(`{{ .Prefix }}|{{ .FullName }}|{{ .ClassName }}|{{ .Namespace }}|{{ .GroupPascal }}|{{ .PluginName }}`),
		},
	}

	data := ExtensionData{Name: "Auth", Vendor: "MyVendor", Type: TypePlugin, Group: "user"}
	targetDir := t.TempDir()

	files, err := generateFromFS(context.Background(), fsys, "test", data, targetDir)
	if err != nil {
		t.Fatalf("generateFromFS() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}

	content, _ := os.ReadFile(filepath.Join(targetDir, "output.txt"))
	want := "plg_|plg_auth|Auth|MyVendor\\Plugin\\User\\Auth|User|plg_user_auth"
	if string(content) != want {
		t.Errorf("output = %q, want %q", string(content), want)
	}
}

func TestGenerateFromFS_EmptySubDir(t *testing.T) {
	fsys := fstest.MapFS{}
	data := ExtensionData{Name: "Empty", Type: TypeModule}
	targetDir := t.TempDir()

	files, err := generateFromFS(context.Background(), fsys, "nonexistent", data, targetDir)
	if err != nil {
		t.Fatalf("generateFromFS() error = %v", err)
	}
	if len(files) != 0 {
		t.Errorf("got %d files, want 0", len(files))
	}
}
