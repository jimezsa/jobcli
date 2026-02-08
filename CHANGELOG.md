# Changelog

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
