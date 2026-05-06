package generator

import (
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ExtensionType represents the type of a Joomla extension.
type ExtensionType string

const (
	TypeComponent ExtensionType = "component"
	TypeModule    ExtensionType = "module"
	TypePlugin    ExtensionType = "plugin"
	TypeTemplate  ExtensionType = "template"
	TypeLibrary   ExtensionType = "library"
	TypePackage   ExtensionType = "package"
)

// ExtensionData holds all input values for rendering an extension skeleton.
type ExtensionData struct {
	Name          string
	Vendor        string
	Type          ExtensionType
	Group         string // plugin only, e.g. "system", "user", "content"
	JoomlaVersion string
	Description   string
	Version       string
	Author        string
	Year          string
}

// ExtensionOption is a functional option for NewExtensionData.
type ExtensionOption func(*ExtensionData)

// WithGroup sets the plugin group.
func WithGroup(group string) ExtensionOption {
	return func(d *ExtensionData) {
		d.Group = group
	}
}

// WithDescription sets the extension description.
func WithDescription(desc string) ExtensionOption {
	return func(d *ExtensionData) {
		d.Description = desc
	}
}

// WithVersion sets the extension version.
func WithVersion(version string) ExtensionOption {
	return func(d *ExtensionData) {
		d.Version = version
	}
}

// WithAuthor sets the extension author.
func WithAuthor(author string) ExtensionOption {
	return func(d *ExtensionData) {
		d.Author = author
	}
}

// WithYear sets the copyright year.
func WithYear(year string) ExtensionOption {
	return func(d *ExtensionData) {
		d.Year = year
	}
}

// WithJoomlaVersion sets the Joomla version target.
func WithJoomlaVersion(version string) ExtensionOption {
	return func(d *ExtensionData) {
		d.JoomlaVersion = version
	}
}

// NewExtensionData creates an ExtensionData with default values.
func NewExtensionData(name, vendor string, extType ExtensionType, opts ...ExtensionOption) ExtensionData {
	d := ExtensionData{
		Name:    name,
		Vendor:  vendor,
		Type:    extType,
		Version: "1.0.0",
		Year:    strconv.Itoa(time.Now().Year()),
	}
	for _, opt := range opts {
		opt(&d)
	}
	return d
}

// prefixFor returns the Joomla file prefix for the given extension type.
func prefixFor(t ExtensionType) string {
	switch t {
	case TypeComponent:
		return "com_"
	case TypeModule:
		return "mod_"
	case TypePlugin:
		return "plg_"
	case TypeTemplate:
		return "tpl_"
	case TypeLibrary:
		return "lib_"
	case TypePackage:
		return "pkg_"
	default:
		return ""
	}
}

// namespacePartFor returns the PHP namespace segment for the given extension type.
func namespacePartFor(t ExtensionType) string {
	switch t {
	case TypeComponent:
		return "Component"
	case TypeModule:
		return "Module"
	case TypePlugin:
		return "Plugin"
	case TypeTemplate:
		return "Template"
	case TypeLibrary:
		return "Library"
	case TypePackage:
		return "Package"
	default:
		return ""
	}
}

// SanitizeName converts an extension name to a filesystem-safe lowercase form
// by replacing spaces and hyphens with underscores.
func SanitizeName(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return strings.ToLower(name)
}

// toPascalCase converts a PascalCase, snake_case, space-separated, or
// kebab-case string to PascalCase.
func toPascalCase(s string) string {
	// Insert delimiter before uppercase letters (except at position 0)
	// so PascalCase input like "ContentHistory" becomes "Content_History".
	var withDelims strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			withDelims.WriteRune('_')
		}
		withDelims.WriteRune(r)
	}

	words := strings.FieldsFunc(withDelims.String(), func(r rune) bool {
		return r == '_' || r == ' ' || r == '-'
	})
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		runes := []rune(w)
		runes[0] = unicode.ToUpper(runes[0])
		for j := 1; j < len(runes); j++ {
			runes[j] = unicode.ToLower(runes[j])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, "")
}

// Prefix returns the Joomla file prefix for the extension type
// (e.g. "com_", "mod_", "plg_"). Returns empty string for unknown types.
func (d ExtensionData) Prefix() string {
	return prefixFor(d.Type)
}

// FullName returns the full Joomla extension name with prefix
// (e.g. "com_blog", "mod_mymodule").
func (d ExtensionData) FullName() string {
	return d.Prefix() + SanitizeName(d.Name)
}

// ClassName returns the extension name as a PascalCase PHP class name.
// Handles PascalCase, snake_case, kebab-case, and space-separated inputs.
func (d ExtensionData) ClassName() string {
	return toPascalCase(d.Name)
}

// Namespace returns the PHP namespace for the extension
// (e.g. "Alebak\Component\Blog").
func (d ExtensionData) Namespace() string {
	nsPart := namespacePartFor(d.Type)
	className := d.ClassName()

	var groupPart string
	if d.Type == TypePlugin && d.Group != "" {
		groupPart = "\\" + toPascalCase(d.Group)
	}

	return d.Vendor + "\\" + nsPart + groupPart + "\\" + className
}

// ShortNamespace returns the short namespace suffix for components:
// "Administrator" or "Site". Returns empty for non-component types.
func (d ExtensionData) ShortNamespace() string {
	if d.Type != TypeComponent {
		return ""
	}
	switch strings.ToLower(d.Group) {
	case "administrator":
		return "Administrator"
	default:
		return "Site"
	}
}

// GroupPascal returns the Group field as PascalCase (e.g. "System", "User").
// Returns empty string when Group is empty.
func (d ExtensionData) GroupPascal() string {
	if d.Group == "" {
		return ""
	}
	return toPascalCase(d.Group)
}

// PluginName returns the full plugin identifier (e.g. "plg_system_myplugin").
// Returns empty string for non-plugin types or when Group is empty.
func (d ExtensionData) PluginName() string {
	if d.Type != TypePlugin || d.Group == "" {
		return ""
	}
	return "plg_" + SanitizeName(d.Group) + "_" + SanitizeName(d.Name)
}
