# Code Review Rules — JKit

## Go

- Module: `github.com/alebak/jkit`
- Format with `gofmt` before committing
- Follow standard Go naming conventions (camelCase for unexported, PascalCase for exported)
- Always handle errors explicitly — never ignore with `_` unless justified
- Prefer returning errors over panicking
- Use interfaces to decouple packages and enable testing
- No global mutable state
- Use `context.Context` as first argument in functions that do I/O or can be cancelled
- Embed static files with `//go:embed` — never read from disk at runtime for templates

## Project structure

- CLI commands live in `cmd/jkit/`
- Business logic lives in `internal/` — never expose internal packages
- Templates live in `templates/` and are embedded at compile time
- One responsibility per package — no god packages

## Tests

- Every exported function in `internal/` MUST have at least one test
- Table-driven tests preferred
- No mocking of the filesystem — use `os.DirFS` or `testing/fstest` instead

## Commits

- Follow Conventional Commits: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`
- One logical change per commit
- Commit messages in English

## General

- No hardcoded paths — use variables or flags
- No `fmt.Println` in library code — use structured logging or return errors
- Comments in English; user-facing messages in the language of the interface
