# humble-cli

The missing command-line interface for downloading your Humble Bundle purchases!

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
cd humble-cli
go build -o humble-cli ./cmd/humble-cli
```

## 🚀 Usage

To start, go to the [Humble Bundle website][hb-site] and log in. Humble moved
login to `auth.humblebundle.com` (Ory Kratos), so the session cookie is now
named `ory_session_<suffix>` (no longer `_simpleauth_sess`). It is an HttpOnly
cookie, so read it from the browser's developer tools rather than from
`document.cookie`.

See this guide on how to find a cookie value in your browser: [Chrome][guide-chrome], [Firefox][guide-firefox], [Safari][guide-safari].

Store the cookie as `name=value`:

```sh
humble-cli auth 'ory_session_abc123=<value>'
```

A bare value is still accepted and sent as `_simpleauth_sess`.

> Humble removed its JSON API (`/api/v1/user/order`, `/api/v1/orders`,
> `/api/v1/order`) in November 2026. This version reads the library and bundle
> pages instead, and calls the site's `getSignedUrlsForEntitlements` server
> action for download URLs whose signed form is not already in the page.
> Server action ids change with every Humble deployment; they are looked up in
> the page's JavaScript chunks at runtime.
>
> Consequences: the library page carries only id and title, so `list` (table
> and `--claimed`/size fields) and `search` fetch a one-request summary per
> bundle first (~7-9 s for 86 bundles); `list --field key` stays a single
> request. `list-choices` reads the current month's products from the
> membership page payload; the reveal state per key is shown on
> `/home/keys`, not in that payload.

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
- Go 1.25 or later

### Building
```sh
go build -o humble-cli ./cmd/humble-cli
```

### Running Tests
```sh
go test ./...
```

[guide-chrome]: https://github.com/smbl64/humble-cli/blob/master/docs/session-key-chrome.md
[guide-firefox]: https://github.com/smbl64/humble-cli/blob/master/docs/session-key-firefox.md
[guide-safari]: https://github.com/smbl64/humble-cli/blob/master/docs/session-key-safari.md
[hb-site]: https://www.humblebundle.com/

