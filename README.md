[![Build status](https://github.com/smbl64/humble-cli/actions/workflows/tests.yml/badge.svg)](https://github.com/smbl64/humble-cli/actions/workflows/tests.yml)
![GitHub](https://img.shields.io/github/license/smbl64/humble-cli)
![GitHub release (with filter)](https://img.shields.io/github/v/release/smbl64/humble-cli)

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
**Option 1:** Download the binaries in the [Releases][releases] page. Windows, macOS and Linux are supported.

**Option 2:** Install it via `cargo`:

```sh
cargo install humble-cli
```

## 🚀 Usage 

To start, go to the [Humble Bundle website][hb-site] and log in. Then find the cookie value for `_simpleauth_sess`. This is required to interact with Humble Bundle API. 

See this guide on how to find the cookie value for your browser: [Chrome][guide-chrome], [Firefox][guide-firefox], [Safari][guide-safari].

Use `humble-cli auth "<YOUR SESSION KEY>"` to store the authentication key locally for other subcommands.

After that you will have access to the following sub-commands:

```
$ humble-cli --help
The missing Humble Bundle CLI

USAGE:
    humble-cli <SUBCOMMAND>

OPTIONS:
    -h, --help       Print help information
    -V, --version    Print version information

SUBCOMMANDS:
    auth            Set the authentication session key
    completion      Generate shell completions
    details         Print details of a certain bundle [aliases: info]
    download        Selectively download items from a bundle [aliases: d]
    help            Print this message or the help of the given subcommand(s)
    list            List all your purchased bundles [aliases: ls]
    list-choices    List your current Humble Choices
    search          Search through all bundle products for keywords

Note: `humble-cli -h` prints a short and concise overview while `humble-cli --help` gives all
details.
```

[releases]: https://github.com/smbl64/humble-cli/releases
[hb-site]: https://www.humblebundle.com/
[guide-chrome]: https://github.com/smbl64/humble-cli/blob/master/docs/session-key-chrome.md
[guide-firefox]: https://github.com/smbl64/humble-cli/blob/master/docs/session-key-firefox.md
[guide-safari]: https://github.com/smbl64/humble-cli/blob/master/docs/session-key-safari.md

