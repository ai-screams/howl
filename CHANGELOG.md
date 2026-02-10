# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Gitleaks secret scanning to all CI workflows
- Weekly security vulnerability scanning (govulncheck)
- Modular reusable workflows for CI/CD pipeline
- Auto-release with semantic versioning (svu)
- Configuration system: individual metric toggles and priority ordering (Phase 1c)
- Interactive preset selection via `/howl:configure` skill
- Preset system for customizable statusline display (Phase 1b)

### Fixed

- CI/CD workflow failures: govulncheck, gitleaks, release-build
- Upgrade Go to 1.24.13 for TLS/x509 security patches

### Changed

- Optimize CI workflows for 40% faster builds
- Plugin version pin removed for auto-latest updates

## [1.1.0] - 2026-02-08

### Added

- Support M suffix for 1M+ context window display
- Account email display for multi-account identification
- Claude Code plugin structure (Phase 1a)

### Fixed

- Migrate golangci-lint config to v2 schema
- Upgrade golangci-lint-action v6 to v7

### Changed

- Code cleanup and stdlib usage improvements

## [1.0.0] - 2026-02-07

### Added

- Core data types and stdin JSON parsing
- Derived metrics calculations (cost velocity, cache efficiency, API wait ratio)
- ANSI rendering with adaptive layout (2-4 lines normal, 2 lines danger mode)
- Git status display integration
- OAuth usage quota tracking (5h/7d)
- Transcript parsing for tools/agents display
- Release automation with GoReleaser and GitHub Actions
- Comprehensive unit tests (78% coverage)

### Changed

- Extract magic numbers to constants
- Remove deprecated and unused render functions

[Unreleased]: https://github.com/ai-screams/Howl/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/ai-screams/Howl/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/ai-screams/Howl/releases/tag/v1.0.0
