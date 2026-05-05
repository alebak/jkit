package devcontainer

// Embedded template filesystems.
//
// The actual //go:embed directives live in the jkit package at the
// module root (embed_assets.go) because //go:embed paths are relative
// to the source file's directory and cannot use ".." to reach the
// templates/ directory from here.
//
// This package accesses the embedded filesystems through the jkit
// package and uses them in the Render function.
