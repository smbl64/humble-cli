# humble-cli
The missing CLI for downloading your Humble Bundle purchases!

[![Build status](https://github.com/smbl64/humble-cli/actions/workflows/tests.yml/badge.svg)](https://github.com/smbl64/humble-cli/actions/workflows/tests.yml)

## Install
You can find the binaries in the [Releases][releases] page. Windows, macOS and Linux are supported.

## Usage

To start, go to the [Humble Bundle website][hb-site] and log in. Then find the cookie value for `_simpleauth_sess`. This is required to interact with Humble Bundle API. 

See this guide on how to find the cookie value for your browser: [Chrome][guide-chrome], [Firefox][guide-firefox], [Safari][guide-safari].

Use `humble-cli auth <YOUR SESSION KEY>` to store the authentication key locally for other subcommands.

After that you will have access to the following sub-commands:

```
$ humble-cli --help

humble-cli 0.3.1
The missing Humble Bundle CLI

USAGE:
    humble-cli <SUBCOMMAND>

OPTIONS:
    -h, --help       Print help information
    -V, --version    Print version information

SUBCOMMANDS:
    auth        Set the authentication session key
    details     Print details of a certain bundle
    download    Download all items in a bundle
    help        Print this message or the help of the given subcommand(s)
    list        List all purchased bundles

Note: `humble-cli -h` prints a short and concise overview while `humble-cli --help` gives all
details.
```

[releases]: https://github.com/smbl64/humble-cli/releases
[hb-site]: https://www.humblebundle.com/
[guide-chrome]: https://github.com/smbl64/humble-cli/blob/master/docs/session-key-chrome.md
[guide-firefox]: https://github.com/smbl64/humble-cli/blob/master/docs/session-key-firefox.md
[guide-safari]: https://github.com/smbl64/humble-cli/blob/master/docs/session-key-safari.md

