# Changelog

## [Unreleased]

## [0.1.0] - 2026-02-05

### Added

- Initial release of JobCLI
- Concurrent scraping across LinkedIn, Indeed, Glassdoor, ZipRecruiter, and Google Jobs
- TLS fingerprinting via `tls-client` to reduce blocking
- Proxy rotation with temporary bans on 403/429 responses
- Multiple output formats: table, CSV, TSV, JSON, and Markdown
- Config and proxies stored in user config directory
- Search filters: location, remote, job type, country, and more
- Proxy health checking command
- Environment variable support for configuration
