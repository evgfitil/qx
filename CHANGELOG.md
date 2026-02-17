# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.3] - 2026-02-17

### Fixed

- Long input text now wraps instead of disappearing when it exceeds terminal width

## [0.7.2] - 2026-02-13

### Fixed

- Multiline command corruption on repeated hotkey invocation

## [0.7.1] - 2026-02-13

### Changed

- LLM prompt prefers single tool capabilities over pipe chains

## [0.7.0] - 2026-02-07

### Added

- Stdin pipe support for context-aware command generation
- Post-selection action menu: execute, copy to clipboard, or quit

## [0.6.1] - 2026-02-06

### Fixed

- Startup errors now display inside TUI instead of corrupting prompt theme

## [0.6.0] - 2026-02-06

### Added

- Fish shell integration

## [0.5.0] - 2026-02-03

### Added

- Prompt restoration on cancel: pressing Esc now restores the query text to the command line for easy editing

## [0.4.0] - 2026-01-31

### Added

- Security guard for input/output sanitization

## [0.3.2] - 2026-01-30

### Fixed

- Homebrew tap configuration (separate repository, branch, token)

## [0.3.1] - 2026-01-30

### Added

- Homebrew tap support
- Shell integration caveats after brew install
- API key configuration via config file

## [0.3.0] - 2026-01-29

### Added

- CI workflow and pre-commit configuration

### Removed

- ElizaProvider in favor of raw OpenAI-compatible endpoint

## [0.2.0] - 2026-01-29

### Added

- `--query` flag for inline shell editing

## [0.1.2] - 2026-01-29

### Fixed

- Module path to match GitHub repository

## [0.1.1] - 2026-01-29

### Added

- Command formatting with line breaks at pipe operators

## [0.1.0] - 2026-01-25

### Added

- Initial release
- Interactive TUI with bubbletea
- fzf-style picker for command selection
- LLM integration with OpenAI-compatible providers
- Shell integration (bash, zsh)
- Goreleaser and release workflow

[Unreleased]: https://github.com/evgfitil/qx/compare/v0.7.3...HEAD
[0.7.3]: https://github.com/evgfitil/qx/compare/v0.7.2...v0.7.3
[0.7.2]: https://github.com/evgfitil/qx/compare/v0.7.1...v0.7.2
[0.7.1]: https://github.com/evgfitil/qx/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/evgfitil/qx/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/evgfitil/qx/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/evgfitil/qx/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/evgfitil/qx/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/evgfitil/qx/compare/v0.3.2...v0.4.0
[0.3.2]: https://github.com/evgfitil/qx/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/evgfitil/qx/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/evgfitil/qx/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/evgfitil/qx/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/evgfitil/qx/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/evgfitil/qx/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/evgfitil/qx/releases/tag/v0.1.0
