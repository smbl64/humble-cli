# Quick Reference - Continue Working on humble-cli Go

This is a quick reference guide to help you (or other agents) continue working on this project.

## 📍 Location

```
/Users/mohammad/Projects/humble-cli/humble-cli-go/
```

## ⚡ Quick Commands

```bash
# Navigate to project
cd /Users/mohammad/Projects/humble-cli/humble-cli-go

# Build
go build -o humble-cli ./cmd/humble-cli

# Test
go test ./...

# Run
./humble-cli --help

# Clean & Rebuild
rm humble-cli && go build -o humble-cli ./cmd/humble-cli
```

## 🎯 Current Status

**✅ COMPLETE AND WORKING**

- All 8 commands implemented
- All tests passing
- Binary compiles successfully
- Time parsing bug fixed
- Documentation complete

## 📁 Key Files to Know

### Where the Action Is
- `cmd/humble-cli/main.go` - Add new commands here
- `internal/commands/commands.go` - Add command logic here
- `internal/models/bundle.go` - Data structures (includes HumbleTime fix)
- `internal/api/humble.go` - API client

### Tests
- `internal/*/\*_test.go` - All test files
- Run with: `go test ./...`

### Documentation
- `README.md` - User guide
- `MIGRATION.md` - Technical details
- `DEVELOPMENT.md` - Developer guide
- `CLAUDE.md` (parent dir) - Agent instructions

## 🔧 Recent Fixes

### Time Parsing Bug (FIXED)
**Problem:** `parsing time "2021-04-05T20:01:30.481166" as "2006-01-02T15:04:05Z07:00"`

**Solution:** Custom `HumbleTime` type in `internal/models/bundle.go`

**Files Changed:**
- `internal/models/bundle.go` - Added HumbleTime
- `internal/api/humble.go` - Updated sorting
- `internal/commands/commands.go` - Updated date formatting
- `internal/models/bundle_test.go` - Added tests

## 🧪 Testing

```bash
# All tests
go test ./...

# Specific package
go test ./internal/util/...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Verbose
go test -v ./...
```

## 🏗️ Project Structure

```
humble-cli-go/
├── cmd/humble-cli/main.go       # CLI entry (Cobra)
├── internal/
│   ├── api/                     # Humble API client
│   ├── commands/                # Command implementations
│   ├── config/                  # ~/.humble-cli-key
│   ├── download/                # Download with retry
│   ├── keymatch/                # Fuzzy key matching
│   ├── models/                  # Data structures
│   └── util/                    # Utilities
├── docs/                        # Browser guides
└── *.md                         # Documentation
```

## 🎨 Common Tasks

### Add a New Command

1. In `cmd/humble-cli/main.go`:
```go
var newCmd = &cobra.Command{
    Use:   "new-command",
    Short: "Description",
    RunE: func(cmd *cobra.Command, args []string) error {
        return commands.NewCommand()
    },
}

func init() {
    rootCmd.AddCommand(newCmd)
}
```

2. In `internal/commands/commands.go`:
```go
func NewCommand() error {
    // Implementation
}
```

### Add Tests

Create `*_test.go` file in same package:
```go
func TestMyFunction(t *testing.T) {
    // Test implementation
}
```

### Debug API Issues

Check `internal/api/humble.go`:
- API endpoints
- Concurrent fetching
- Error handling

### Debug Time Issues

Check `internal/models/bundle.go`:
- `HumbleTime` type
- Format strings
- UnmarshalJSON implementation

## 📊 Dependencies

```
github.com/spf13/cobra v1.8.0              # CLI framework
github.com/PuerkitoBio/goquery v1.9.0      # HTML parsing
github.com/schollz/progressbar/v3 v3.14.0  # Progress bars
github.com/olekukonko/tablewriter v0.0.5   # Tables
```

Update: `go get -u <package>` then `go mod tidy`

## 🐛 Known Issues

**None currently!**

All known issues have been resolved:
- ✅ Time parsing fixed with HumbleTime
- ✅ All tests passing
- ✅ Binary compiles successfully

## 🔍 Useful Searches

Find where something is used:
```bash
# Find all uses of a function
grep -r "FunctionName" internal/

# Find all time.Time accesses
grep -r "\.Created\." internal/

# Find all API calls
grep -r "api\." internal/commands/
```

## 💡 Tips for Agents

1. **Always run tests** after changes: `go test ./...`
2. **Format code** before committing: `go fmt ./...`
3. **Check CLAUDE.md** for detailed patterns
4. **Match Rust behavior** - compare with `../src/`
5. **HumbleTime gotcha** - Access time with `.Time` property

## 🚀 Next Session Checklist

When you (or another agent) continue:

1. ✅ Navigate to: `/Users/mohammad/Projects/humble-cli/humble-cli-go`
2. ✅ Check current state: `go test ./...`
3. ✅ Review recent changes: `git log --oneline -10`
4. ✅ Read CLAUDE.md for context
5. ✅ Make changes
6. ✅ Run tests: `go test ./...`
7. ✅ Build: `go build -o humble-cli ./cmd/humble-cli`
8. ✅ Test manually: `./humble-cli <command>`
9. ✅ Update CHANGELOG.md if fixing bugs
10. ✅ Update this file if needed

## 📚 Documentation Files

- **README.md** - User installation & usage
- **MIGRATION.md** - Technical migration details
- **IMPLEMENTATION_COMPLETE.md** - Final status
- **DEVELOPMENT.md** - Developer guide (detailed)
- **CHANGELOG.md** - Version history
- **SESSION_SUMMARY.md** - What was done
- **QUICK_REFERENCE.md** - This file
- **../CLAUDE.md** - Agent instructions (includes Go section)

## ✅ Verification Commands

Run these to verify everything works:

```bash
# All tests pass
go test ./...

# Binary compiles
go build -o humble-cli ./cmd/humble-cli

# Binary works
./humble-cli --version
./humble-cli --help

# No compilation errors
go build ./...

# Dependencies okay
go mod verify
```

Expected output: All ✅

---

**Last Updated:** January 30, 2026
**Status:** Production Ready
**All Tests:** Passing
**Ready For:** Distribution or further development
