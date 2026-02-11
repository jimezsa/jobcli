# Changelog

## [0.1.2] - 2026-02-11

### Added

- Added `jobcli seen diff` to compute unseen jobs (`A - B`) from JSON arrays using normalized `title + company`
- Added `jobcli seen update` to merge newly accepted jobs into seen history without duplicates
- Added `internal/seen` package with reusable compare/merge logic and JSON IO helpers
- Added search/site flags: `--seen`, `--new-only`, and `--new-out` for one-step unseen filtering workflows
- Added `--seen-update` to `search` and site commands to automatically merge newly discovered unseen jobs into the `--seen` history JSON after a successful run
- Added unit tests for seen normalization, diff/merge behavior, idempotency, and file IO

### Changed

- Updated top-level docs (`README.md`, `AGENTS.md`) for the new seen-jobs command group and workflow
- Clarified `--new-out` vs `--output` behavior in docs (`README.md`, `docs/usage.md`)

### Fixed

- Fixed `search`/site file output format resolution so `--out` now respects `--json` and `--plain` (instead of always defaulting to CSV)
- Added regression test coverage for `resolveFormat` with file output paths

## [0.1.1] - 2026-02-07

### Changed

- Improved CLI help text and command descriptions for clarity
- Added SKILL.md and SKILL-jobsearch.md for AI-assisted job search workflows
- Added AGENTS.md project structure guide for LLM and contributor onboarding

## [0.1.0] - 2026-02-06

Initial public release of `jobcli` Fast, single-binary job aggregation CLI written in Go. Scrapes multiple sites in parallel and exports results to table, CSV, TSV, JSON, or Markdown.

### Added

- Initial release of JobCLI
- Concurrent scraping across LinkedIn, Indeed, Glassdoor, ZipRecruiter, Google Jobs, and Stepstone
- TLS fingerprinting via `tls-client` to reduce blocking
- Proxy rotation with temporary bans on 403/429 responses
- Multiple output formats: table, CSV, TSV, JSON, and Markdown
- Config and proxies stored in user config directory
- Search filters: location, remote, job type, country, and more
- Proxy health checking command
- Environment variable support for configuration
- GoReleaser configuration for automated cross-platform builds
- GitHub Actions release workflow with Homebrew tap support
