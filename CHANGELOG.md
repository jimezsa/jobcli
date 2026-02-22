# Changelog

## [0.2.1] - 2026-02-22

### Added

- Added comma-separated multi-query support for `jobcli search` and site commands (`linkedin`, `indeed`, `glassdoor`, `ziprecruiter`, `stepstone`)
- Added `--query-file` support for `search` and site commands to load queries from JSON (`["backend","platform"]` or `{"job_titles":["backend","platform"]}`)
- Added full `--query-file` profile support with optional `search_options` and `global_options`, enabling one-command runs like `jobcli search --query-file queries.json`
- Added multi-query parsing/validation: split and trim tokens, remove empties, case-insensitive dedupe, empty-query error, and max 10 queries
- Added cross-query merge/dedupe flow using seen-key normalization (`title + company`, fallback URL) plus per-query limiting before merge
- Added regression coverage in `internal/cmd/search_test.go` for query parsing, per-query limit behavior, and multi-query seen/update workflows
- Added `docs/multi-search-spec.md` with implementation and behavior details
- Added `skills/jobcli-job-search/scripts/job_discriminator.py` for LLM-based YES/NO job filtering with confidence gating and concurrent workers
- Added default student/internship exclusions in `skills/jobcli-cv-summary/SKILL.md` to reduce irrelevant persona matches

### Changed

- Changed `--limit` semantics to "maximum results per query" (instead of final merged output size)
- Changed `--query-file` behavior to support defaults precedence (`CLI flags > query-file defaults > env/config defaults`) with schema validation (unknown fields, enum constraints, `json/plain` conflict)
- Changed CLI search surface by removing `google` as a selectable direct site command and registry target
- Changed GitHub release workflow to publish notes directly from the matching `CHANGELOG.md` version section
- Updated command overview examples and docs (`README.md`, `docs/usage.md`) including seen-update flow, multi-query examples, and expanded flag tables
- Updated skills docs and search flow to use a binary persona filter pipeline with `jobcli seen update` merge steps and enumerated output formatting
- Updated job discriminator defaults/behavior (model/API setup, timeout/tokens/workers, progress logging, and more permissive matching guidance)

### Fixed

- Fixed environment loading in the pirate motivator script by reading project-root `.env`
- Fixed job discriminator throughput by increasing default parallel worker count for faster batch processing

## [0.2.0] - 2026-02-14

### Added

- Added `description` to the normalized `Job` model and JSON output
- Added LinkedIn job-detail description fetch (`jobs-guest` job posting endpoint) with fallback to card snippet
- Added Stepstone job-detail description extraction with selector + JSON-LD fallback
- Added GitHub release workflow step to auto-generate and apply release notes for tagged releases
- Added regression tests for LinkedIn/Stepstone snippet and description parsing plus seen-key fallback behavior

### Changed

- Improved LinkedIn and Stepstone snippet parsing to avoid metadata/location text being used as job summary
- Updated seen-job keying to prefer normalized `title + company` and fallback to URL when title/company is missing

### Fixed

- Fixed `--new-only` dropping valid jobs when providers omit `company` by using URL-based fallback keys in seen diff/merge
- Fixed empty or misleading snippet values for LinkedIn/Stepstone job cards caused by overly broad card-text extraction

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
