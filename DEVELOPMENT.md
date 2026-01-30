# Development Guide - humble-cli Go Version

This guide provides essential information for developers working on the humble-cli Go implementation.

## Quick Start

```bash
# Clone and navigate
cd humble-cli-go

# Install dependencies
go mod download

# Build
go build -o humble-cli ./cmd/humble-cli

# Run tests
go test ./...

# Run the binary
./humble-cli --help
```

## Project Structure

```
humble-cli-go/
├── cmd/humble-cli/           # CLI entry point
│   └── main.go               # Cobra commands, flags, main()
├── internal/                 # Private packages
│   ├── api/                  # Humble Bundle API client
│   ├── commands/             # Command implementations
│   ├── config/               # Configuration management
│   ├── download/             # Download logic with retry
│   ├── keymatch/             # Fuzzy key matching
│   ├── models/               # Data structures
│   └── util/                 # Utilities (bytes, ranges, tables, etc.)
├── docs/                     # User-facing documentation
├── Makefile                  # Build automation
└── *.md                      # Documentation files
```

## Common Development Tasks

### Building

```bash
# Development build
make build
# or
go build -o humble-cli ./cmd/humble-cli

# Release build (optimized, stripped)
go build -ldflags="-s -w" -o humble-cli ./cmd/humble-cli

# All platforms
make build-all
```

### Testing

```bash
# Run all tests
make test
# or
go test ./...

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/util/...

# Verbose output
go test -v ./...

# Run specific test
go test -v -run TestHumbleTime ./internal/models/...
```

### Code Quality

```bash
# Format code
go fmt ./...
# or
make fmt

# Run linter (requires golangci-lint)
make lint

# Check for issues
go vet ./...
```

### Running

```bash
# Run without building
go run ./cmd/humble-cli --help

# Build and run
make run

# Run specific command
./humble-cli list --help
```

## Key Implementation Details

### 1. Custom Time Handling

Humble Bundle API returns timestamps without timezone: `2021-04-05T20:01:30.481166`

**Solution:** Custom `HumbleTime` type in `internal/models/bundle.go`

```go
type HumbleTime struct {
    time.Time
}

func (ht *HumbleTime) UnmarshalJSON(data []byte) error {
    // Tries multiple formats
}
```

Access the underlying time:
```go
bundle.Created.Time.Format("2006-01-02")
```

### 2. Concurrent API Fetching

API fetches bundle data in batches of 10 using goroutines:

```go
// internal/api/humble.go
resultChan := make(chan bundleResult, len(chunks))
for _, chunk := range chunks {
    go func(keys []string) {
        bundles, err := api.readBundlesData(keys)
        resultChan <- bundleResult{bundles, err}
    }(chunk)
}
```

### 3. Download Retry Logic

Downloads retry up to 3 times with 5-second delays:

```go
// internal/download/download.go
const (
    retryCount = 3
    retryDelay = 5 * time.Second
)
```

Supports resume via HTTP Range headers.

### 4. Partial JSON Deserialization

Custom unmarshaling skips malformed products (Go equivalent of Rust's `VecSkipError`):

```go
// internal/models/bundle.go
func (b *Bundle) UnmarshalJSON(data []byte) error {
    for _, raw := range aux.RawProducts {
        var p Product
        if err := json.Unmarshal(raw, &p); err == nil {
            b.Products = append(b.Products, p)
        }
        // Silently skip malformed products
    }
}
```

### 5. Error Message Mapping

User-friendly error messages for common HTTP errors:

```go
// internal/commands/commands.go
func HandleHTTPErrors(err error) error {
    if strings.Contains(errMsg, "401") {
        return fmt.Errorf("Unauthorized request (401). Is the session key correct?")
    }
    if strings.Contains(errMsg, "404") {
        return fmt.Errorf("Bundle not found (404). Is the bundle key correct?")
    }
}
```

## Adding New Features

### Adding a New Command

1. **Define command in** `cmd/humble-cli/main.go`:
```go
var myNewCmd = &cobra.Command{
    Use:   "my-command",
    Short: "Description",
    RunE: func(cmd *cobra.Command, args []string) error {
        return commands.MyNewCommand()
    },
}

func init() {
    rootCmd.AddCommand(myNewCmd)
}
```

2. **Implement logic in** `internal/commands/commands.go`:
```go
func MyNewCommand() error {
    // Implementation
}
```

3. **Add tests in** `internal/commands/commands_test.go`

### Adding New Flags

```go
// In init() function
myCmd.Flags().StringP("my-flag", "m", "default", "Description")

// In RunE function
value, _ := cmd.Flags().GetString("my-flag")
```

### Adding New Models

1. Define struct in `internal/models/`
2. Add JSON tags: `json:"field_name"`
3. Implement helper methods
4. Add tests in `*_test.go`

## Testing Guidelines

### Unit Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    Type
        expected Type
        wantErr  bool
    }{
        {"description", input, expected, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionName(tt.input)

            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error but got none")
                }
                return
            }

            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }

            if result != tt.expected {
                t.Errorf("got %v; want %v", result, tt.expected)
            }
        })
    }
}
```

### Test Coverage

Check coverage for a package:
```bash
go test -coverprofile=coverage.out ./internal/util
go tool cover -html=coverage.out
```

## Debugging

### Common Issues

1. **Time parsing errors:**
   - Check `HumbleTime` formats in `internal/models/bundle.go`
   - Verify API response format hasn't changed

2. **Concurrent fetching errors:**
   - Check goroutine error handling in `internal/api/humble.go`
   - Verify channel buffering

3. **Download failures:**
   - Check retry logic in `internal/download/download.go`
   - Verify HTTP Range header support

### Debug Logging

Add debug prints:
```go
import "log"

log.Printf("Debug: value = %+v\n", value)
```

## Code Style

Follow standard Go conventions:
- Run `go fmt` before committing
- Use meaningful variable names
- Document public functions
- Keep functions focused and small
- Error messages start with lowercase
- Use `gofmt`, `go vet`, `golangci-lint`

## Dependencies

Update dependencies:
```bash
go get -u github.com/spf13/cobra
go mod tidy
```

Check for vulnerabilities:
```bash
go list -json -m all | nancy sleuth
```

## Compatibility with Rust Version

When adding features, ensure behavioral compatibility:

1. **CLI Interface:** Commands, flags, and output must match
2. **Config Storage:** Use `~/.humble-cli-key`
3. **API Behavior:** Maintain concurrent fetching, error messages
4. **Download Logic:** Keep retry count, delays, resume support
5. **Output Format:** Tables, CSV mode must match exactly

## Performance Considerations

- Use goroutines for I/O-bound operations
- Buffer channels appropriately
- Reuse HTTP clients when possible
- Stream large downloads instead of loading into memory
- Use progress bars for long operations

## Release Process

1. Update `CHANGELOG.md`
2. Run full test suite: `go test ./...`
3. Build all platforms: `make build-all`
4. Tag release: `git tag v1.0.0`
5. Push tag: `git push origin v1.0.0`
6. Create GitHub release with binaries

## Getting Help

- **Documentation:** See `README.md`, `MIGRATION.md`
- **Rust Reference:** Compare with `../src/` for behavior
- **Go Documentation:** https://pkg.go.dev
- **Issues:** Check `CHANGELOG.md` for known issues

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new features
4. Ensure all tests pass
5. Format code with `go fmt`
6. Submit pull request

---

**Last Updated:** January 30, 2026
**Go Version:** 1.21+
**Status:** Production Ready
