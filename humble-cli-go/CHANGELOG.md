# Changelog

All notable changes to the humble-cli Go port will be documented in this file.

## [Unreleased]

### Fixed
- Fixed time parsing error when listing bundles. Humble Bundle API returns timestamps in format `2021-04-05T20:01:30.481166` without timezone information, which is now correctly handled with a custom `HumbleTime` type.

### Added
- Custom `HumbleTime` type that handles multiple datetime formats:
  - Humble Bundle format with microseconds: `2006-01-02T15:04:05.999999`
  - Humble Bundle format without microseconds: `2006-01-02T15:04:05`
  - RFC3339 format: `2006-01-02T15:04:05Z07:00`
- Comprehensive tests for time parsing and bundle unmarshaling

### Technical Details
The error occurred because Go's default `time.Time` JSON unmarshaling expects RFC3339 format with timezone (e.g., `2021-04-05T20:01:30Z`), but Humble Bundle API returns timestamps without timezone information. The custom `HumbleTime` type tries multiple formats and handles this gracefully.

## [Initial Release]

### Added
- Complete port of humble-cli from Rust to Go
- All commands: auth, list, list-choices, details, search, download, bulk-download, completion
- All flags and options from Rust version
- Concurrent API fetching
- Download retry and resume functionality
- Progress bars for downloads
- Shell completion support (bash, zsh, fish, powershell)
- Comprehensive test suite
- Build automation with Makefile
