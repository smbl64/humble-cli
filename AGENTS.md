# AGENTS.md

This document is for automated agents and contributors who will work on the humble-cli repository. It documents the repository layout, developer-facing commands, patterns, and important gotchas discovered in the codebase. Only facts observed in the repository are included.

## Project Overview

**Language:** Go 1.21+
**Purpose:** Command-line tool to interact with Humble Bundle purchases: list bundles, show details, search products, download items, and manage a session key for authentication.

## Repository Structure

```
.
├── cmd/
│   └── humble-cli/          # CLI entry point (main.go with Cobra framework)
├── internal/
│   ├── api/                 # Humble Bundle API client
│   ├── commands/            # Command implementations
│   ├── config/              # Configuration management
│   ├── download/            # File download with retry
│   ├── keymatch/            # Fuzzy key matching
│   ├── models/              # Data structures (Bundle, Product, etc.)
│   └── util/                # Utility functions
├── docs/                    # Browser session key extraction guides
├── .github/workflows/       # CI/CD workflows
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
├── Makefile                 # Build automation
├── README.md                # User-facing documentation
├── MIGRATION.md             # Rust-to-Go migration technical details
├── CHANGELOG.md             # Version history
├── DEVELOPMENT.md           # Development guide
└── QUICK_REFERENCE.md       # Quick command reference
```

## Essential Commands

### Building
```bash
# Debug build
go build -o humble-cli ./cmd/humble-cli

# Release build (optimized, smaller binary)
go build -ldflags="-s -w" -o humble-cli ./cmd/humble-cli

# Build for all platforms (using Makefile)
make build-all
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests with race detector
go test -race ./...
```

### Development
```bash
# Install to $GOPATH/bin
go install ./cmd/humble-cli

# Format code
go fmt ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Download dependencies
go mod download
go mod tidy

# Clean build artifacts
make clean
```

## How Authentication Works

- Session key is stored in: `~/.humble-cli-key` (plain text file)
- Set via: `humble-cli auth "<SESSION-KEY>"`
- The session key is the `_simpleauth_sess` cookie from humblebundle.com
- See `docs/session-key-*.md` for browser-specific extraction guides

## Code Patterns and Conventions

### CLI Framework
- Uses **Cobra** for command-line interface
- Commands defined in `cmd/humble-cli/main.go`
- Command implementations in `internal/commands/commands.go`
- Flags and subcommands follow Cobra conventions

### Error Handling
- Standard Go error patterns: `if err != nil`
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- User-facing errors mapped in `internal/commands/commands.go`:
  - HTTP 401 → "Is the session key correct?"
  - HTTP 404 → "Is the bundle key correct?"

### Concurrency
- Uses goroutines and channels (no external async framework)
- Bundle fetching: concurrent batches of 10 keys
- Download retry: 3 attempts with 5-second delay

### Time Handling
- Custom `HumbleTime` type in `internal/models/bundle.go`
- Handles Humble Bundle's datetime format (no timezone)
- Tries multiple formats: microseconds, seconds, RFC3339

### JSON Unmarshaling
- Custom `UnmarshalJSON` on `Bundle` type
- Silently skips malformed products (partial deserialization)
- Equivalent to Rust's `serde_with(VecSkipError)`

### Testing
- Test files alongside source: `*_test.go`
- Table-driven tests for utilities
- No mocks/fakes - tests use realistic data structures

## Dependencies

All dependencies use permissive licenses (MIT/BSD/Apache 2.0):
- `github.com/spf13/cobra` v1.8.0 - CLI framework
- `github.com/PuerkitoBio/goquery` v1.9.0 - HTML parsing
- `github.com/schollz/progressbar/v3` v3.14.0 - Progress bars
- `github.com/olekukonko/tablewriter` v0.0.5 - Table formatting

## CI/CD Workflows

### tests.yml
- Runs on: ubuntu-latest, windows-latest, macos-latest
- Go versions: 1.21, 1.22, 1.23
- Commands: `go test ./...` and `go test -race ./...`
- Triggers: pushes/PRs to master, changes in cmd/, internal/, or Go files

### release.yml
- Triggers: tags matching `v[0-9]+.*`
- Builds binaries for: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
- Creates GitHub release with CHANGELOG.md
- Uploads `.tar.gz` (Unix) and `.zip` (Windows) archives

### automerge.yml
- Auto-merges minor dependabot updates
- Requires `DEPENDABOT_AUTO_MERGE` secret

## Important Gotchas

### Session Key Required
- Most commands require `~/.humble-cli-key` to exist
- Use `humble-cli auth "<KEY>"` to set it up
- If missing or invalid, commands will fail with clear error messages

### Time Parsing
- Humble Bundle API returns timestamps without timezone
- `HumbleTime` type handles this by trying multiple formats
- Always use `HumbleTime` for bundle/product timestamps

### Partial JSON Parsing
- Bundle unmarshaling skips malformed products silently
- This prevents one bad product from breaking the entire bundle
- Check logs if products seem missing

### Download Behavior
- Downloads create directories: `bundle-name/product-name/`
- Filenames are sanitized: invalid chars replaced with `_`
- Progress bars show current/total bytes
- Resume supported via HTTP Range headers

### CSV Output
- `list --field key --field name` produces CSV-like output
- Valid fields: `key`, `name`, `size`, `claimed`
- Useful for scripting and automation

## When Making Changes

1. **Run tests after edits**: `go test ./...`
2. **Format code**: `go fmt ./...`
3. **Update CHANGELOG.md** when fixing bugs or adding features
4. **Check for warnings**: `go build ./...`
5. **Preserve behavioral parity**: User-facing behavior should match expectations
6. **Update documentation**: Keep README.md and AGENTS.md in sync

## Representative Files to Read Before Making Changes

- `cmd/humble-cli/main.go` - CLI definitions and entry point
- `internal/commands/commands.go` - All command implementations
- `internal/api/humble.go` - HTTP API interactions
- `internal/models/bundle.go` - Data structures and time handling
- `internal/config/config.go` - Session key storage
- `.github/workflows/tests.yml` - CI test strategy
- `Makefile` - Build automation

## Naming Conventions

- Directories: lowercase (cmd, internal, docs)
- Files: lowercase with underscores only in test files (`*_test.go`)
- Packages: lowercase, single word (api, config, util)
- Exported functions/types: CamelCase (HumanizeBytes, Bundle)
- Unexported functions/types: camelCase (sanitizeFilename, downloadFile)
- Constants: CamelCase or ALL_CAPS for package-level

## What's NOT in This Repository

- No Makefile linter config (could be added)
- No Docker configuration
- No Homebrew formula (maintained separately if it exists)
- No shell completions in repo (generated at runtime via Cobra)

---

## Legacy: Rust Version (Archived)

This project was originally written in Rust and has been rewritten in Go. The Rust version (v0.20.0 and earlier) is preserved in git history for reference.

### Key Differences from Rust Version
- **CLI Framework**: Cobra (Go) instead of clap (Rust)
- **Concurrency**: Goroutines instead of tokio async/await
- **Error Handling**: Standard Go errors instead of anyhow::Error
- **JSON**: Custom UnmarshalJSON instead of serde
- **HTTP**: net/http instead of reqwest
- **Testing**: Go testing instead of cargo test

### Why Go?
- Easier cross-platform distribution (static binaries)
- Simpler contribution path (wider Go adoption)
- No runtime dependencies
- Faster compilation
- Smaller final binary size

### Accessing Rust Version
- Git history contains all Rust code
- Last Rust release: v0.20.0
- Rust source was in: `src/`, `tests/`, `Cargo.toml`

---

**END OF AGENTS.MD**
