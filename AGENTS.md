# AGENTS.md

This file provides guidance to coding agents when working with code in this repository.

## Commands

```bash
# Build all packages
go build ./...

# Run all tests
go test ./...

# Run tests for a single sub-package
go test ./php/...
go test ./ruby/...
go test ./java/...

# Run a single named test
go test ./php/... -run TestFormatPhp/d_leading_zero
```

## Architecture

`gotime` is a Go module (`github.com/ducng99/gotime`) with no external dependencies. It lets callers format a `time.Time` value using another language's date format string syntax.

Each supported language lives in its own sub-package under the repo root (e.g. `php/`, `ruby/`, `java/`). Every sub-package is self-contained and follows the same pattern:

1. **Wrapper type** — `type Time time.Time` (named after the package, not "GoTime" or anything else).
2. **`Format` method** — `func (t *Time) Format(format string) string` walks the format string and dispatches each token via an unexported per-character function.
3. **`Parse` function** (optional) — `func Parse(format, value string) (Time, error)` converts a language format string to a Go layout and calls `time.Parse`. Tokens that produce non-reversible output return an error.
4. **Predefined format constants** — mirror the target language's named format constants (e.g. `DateTimeATOM`, `DateTimeRFC3339`).

### Token parsing differences per language

- **PHP** (`php/`): single-rune tokens; backslash escapes a literal character (e.g. `\T`).
- **Ruby** (`ruby/`): `%`-prefixed directives with optional flags (`-`, `_`, `0`, `^`, `#`), optional width, and optional case modifier (`E`/`O`).
- **Java** (`java/`): repeated identical letters form a specifier (count changes output width); single-quoted text is literal; `''` produces a literal quote.

Adding a new language means creating a new directory and two files: `<lang>/<lang>.go` and `<lang>/<lang>_test.go`. Use the `gotime-add-language` skill for guidance on this — it covers the correct parser shape, Go time layout tricks, common pitfalls, and what to test.

## Commit style

Use [Conventional Commits](https://www.conventionalcommits.org/): `type: message` (e.g. `feat:`, `fix:`, `refactor!:`). No `Co-Authored-By` trailers.
