# humble-cli (Go Version)

The missing command-line interface for downloading your Humble Bundle purchases!

**Note:** This is a Go rewrite of the original [Rust version](https://github.com/smbl64/humble-cli), providing 1:1 feature parity with improved performance and easier cross-platform distribution.

## ✨ Features
- List all your Humble Bundle purchases
- List entries in a bundle, their file formats, and file size
- Download items in a bundle separately, and optionally filter them with
    - file format (e.g., EPUB, PDF)
    - file size
- Easily see which of your bundles have unclaimed keys
- Check your Humble Bundle Choices in current and previous months
- Search through all your purchases for a specific product

## 🔧 Install

### Option 1: Download Pre-built Binaries
Download the binaries from the Releases page. Windows, macOS, and Linux are supported.

### Option 2: Install via Go
```sh
go install github.com/smbl64/humble-cli/cmd/humble-cli@latest
```

### Option 3: Build from Source
```sh
git clone https://github.com/smbl64/humble-cli.git
cd humble-cli/humble-cli-go
go build -o humble-cli ./cmd/humble-cli
```

## 🚀 Usage

To start, go to the [Humble Bundle website][hb-site] and log in. Then find the cookie value for `_simpleauth_sess`. This is required to interact with the Humble Bundle API.

See this guide on how to find the cookie value for your browser: [Chrome][guide-chrome], [Firefox][guide-firefox], [Safari][guide-safari].

Use `humble-cli auth "<YOUR SESSION KEY>"` to store the authentication key locally for other subcommands.

After that you will have access to the following sub-commands:

```
$ humble-cli --help
Command-line tool to interact with Humble Bundle purchases: list bundles, show details, search products, and download items.

Usage:
  humble-cli [command]

Available Commands:
  auth          Set the authentication session key
  bulk-download Download items from multiple bundles
  completion    Generate shell completions
  details       Print details of a certain bundle
  download      Selectively download items from a bundle
  help          Help about any command
  list          List all your purchased bundles
  list-choices  List your current Humble Choices
  search        Search through all bundle products for keywords

Flags:
  -h, --help      help for humble-cli
  -v, --version   version for humble-cli

Use "humble-cli [command] --help" for more information about a command.
```

## 📝 Examples

### List all bundles
```sh
humble-cli list
```

### List bundles with specific fields (CSV output)
```sh
humble-cli list --field key --field name
```

### Filter by claimed status
```sh
humble-cli list --claimed no
```

### Show bundle details
```sh
humble-cli details <BUNDLE-KEY>
```

### Search for products
```sh
humble-cli search "civilization" --mode any
```

### Download a bundle
```sh
# Download specific formats
humble-cli download <BUNDLE-KEY> -f pdf -f epub

# Download with size limit
humble-cli download <BUNDLE-KEY> -s 100MB

# Download specific items
humble-cli download <BUNDLE-KEY> -i 1,3,5-10

# Download torrents only
humble-cli download <BUNDLE-KEY> -t
```

### Bulk download
```sh
# Create a file with bundle keys (one per line)
humble-cli list --field key > bundles.txt

# Download all bundles
humble-cli bulk-download bundles.txt -f pdf
```

## 🔑 Shell Completion

Generate shell completions for your preferred shell:

```sh
# Bash
source <(humble-cli completion bash)

# Zsh
humble-cli completion zsh > "${fpath[1]}/_humble-cli"

# Fish
humble-cli completion fish | source

# PowerShell
humble-cli completion powershell | Out-String | Invoke-Expression
```

## 🛠️ Development

### Prerequisites
- Go 1.21 or later

### Building
```sh
go build -o humble-cli ./cmd/humble-cli
```

### Running Tests
```sh
go test ./...
```

## 📄 License

Same license as the original Rust version.

## 🙏 Credits

This is a Go port of the excellent [humble-cli](https://github.com/smbl64/humble-cli) Rust project. All credit for the original design and implementation goes to the original author.

## ⚠️ Disclaimer

This tool is not affiliated with or endorsed by Humble Bundle, Inc.

[hb-site]: https://www.humblebundle.com/
[guide-chrome]: docs/session-key-chrome.md
[guide-firefox]: docs/session-key-firefox.md
[guide-safari]: docs/session-key-safari.md
