# Multi Search Feature Spec

## Objective

Implement multi-query search support for `jobcli search` (and site commands that
share `runSearch`) so users can run one command with multiple keywords, e.g.:

```bash
jobcli search "software engineer, hardware engineer, data scientic" --location "Munich, Germany"
```

The feature must preserve existing single-query behavior while adding predictable,
documented behavior for comma-separated query lists.

## Scope

In scope:

- Parse a single positional `<query>` argument into multiple queries when comma-separated.
- Execute searches for each parsed query using existing scraper flow.
- Merge results into one output stream.
- Keep compatibility with existing output and seen-history flags.

Out of scope (v1):

- Boolean query syntax (`AND`, `OR`, parentheses).
- Weighted ranking across query relevance.
- Query-specific filters in a single command (different location/limit per query).

## Command Surface

No new command is required.

Existing commands continue to work:

- `jobcli search <query> ...`
- `jobcli linkedin <query> ...`
- `jobcli indeed <query> ...`
- `jobcli glassdoor <query> ...`
- `jobcli ziprecruiter <query> ...`
- `jobcli stepstone <query> ...`

Multi-search is triggered by a comma-separated `<query>` value.

## Query Parsing Rules (v1)

Input:

- Raw positional argument string from `SearchCmd.Query` or `SiteCmd.Query`.

Rules:

1. Split on comma (`,`).
2. Trim whitespace per part.
3. Drop empty entries.
4. Deduplicate queries case-insensitively while preserving first-seen order.

Validation:

- If final query list is empty, return: `at least one non-empty query is required`.
- Limit maximum queries to `10` to avoid accidental overload.
- If query count exceeds `10`, return: `too many queries: max 10`.

Examples:

- `"software engineer"` -> `["software engineer"]`
- `"software engineer, hardware engineer"` -> `["software engineer","hardware engineer"]`
- `"software engineer, , Data Scientist"` -> `["software engineer","Data Scientist"]`
- `"Backend,backend, BACKEND"` -> `["Backend"]`

## Functional Behavior

### Execution model

- Parse query list from the existing positional query argument.
- For each query:
  - Build `models.SearchParams` with the same shared flags (`location`, `country`,
    `hours`, etc.) and `Query=<current query>`.
  - Execute the existing scraper pipeline.
- Combine all query results into a single `[]models.Job`.

### Deduplication

Deduplicate combined results across all queries using `seen.Key(job)`:

- Primary key: normalized `title + company`
- Fallback key: normalized `url`

Keep first occurrence and discard later duplicates.

Rationale:

- Prevent repeated listings when the same job matches multiple keywords.
- Reuse an existing normalized key path already used by seen-history logic.

### Sorting

After merging and deduplication, keep existing sorting behavior:

- stable sort by `Site` (case-insensitive), same as current `runScrapers` output.

### Limit semantics

`--limit` remains the max number of rows in final output for the command:

- Apply limit after cross-query merge/dedup.
- If `--limit <= 0`, output all merged results.

This keeps user-visible meaning of `--limit` consistent with current command-level
expectation (max rows returned).

### Seen flags behavior

Current behavior is preserved and applied to merged results:

- `--seen`: diff merged results (`A`) against seen history (`B`).
- `--new-only`: output only unseen (`C = A - B`).
- `--new-out`: write unseen JSON to file.
- `--seen-update`: merge newly discovered unseen jobs into `--seen`.

No semantic changes are required for seen workflow; only input `A` becomes a
multi-query aggregate.

## Error and Logging Behavior

- Scraper failures keep current behavior:
  - command continues with successful sites.
  - failures are shown in verbose mode via existing `reportScraperFailures`.
- Query parsing errors fail fast before network calls.
- Spinner remains one spinner for the overall command execution.

## Implementation Plan

### 1) Query parsing helper

Add helper in `internal/cmd/search.go` (or `internal/cmd/search_queries.go`):

- `parseQueries(raw string) ([]string, error)`

Responsibilities:

- split/trim/filter/dedupe
- enforce max query count
- return deterministic ordered list

### 2) Refactor search execution path

Refactor `runSearch` to:

- Parse query list once.
- Build shared runtime dependencies once per command:
  - proxies
  - rotator
  - scraper registry
  - selected scrapers
- Loop through parsed queries and execute scraper searches.
- Merge + dedupe accumulated jobs.
- Apply existing output/seen/update flow once at the end.

Suggested extraction:

- `runScrapersForQuery(selected []scraper.Scraper, base models.SearchParams, query string) ([]models.Job, []scraperFailure, error)`
- `mergeUniqueJobs(existing []models.Job, incoming []models.Job) []models.Job`

### 3) Keep site commands compatible

No command-surface change is required because `SiteCmd` already calls `runSearch`.
After refactor, site commands automatically support comma-separated query input.

### 4) Docs updates (after implementation)

Update:

- `README.md` quick start + flags notes
- `docs/usage.md` search examples and multi-query notes

Include one explicit example showing comma-separated queries.

## Test Plan

Unit tests in `internal/cmd/search_test.go`:

- `parseQueries` single query.
- `parseQueries` multi query with spaces.
- `parseQueries` empty/blank tokens removed.
- `parseQueries` case-insensitive dedupe preserves first token.
- `parseQueries` max-query validation.
- `parseQueries` empty input validation.

Behavioral tests in `internal/cmd/search_test.go`:

- Merged multi-query result set is deduplicated via seen-key behavior.
- `--new-only` on multi-query output returns unseen jobs only.
- `--seen-update` updates history once with merged unseen jobs.
- `--limit` applied after merge/dedup.

Regression tests:

- Single-query path produces same output as before for equivalent fixtures.

## Acceptance Criteria

1. `jobcli search "a,b,c"` executes searches for 3 queries in one command.
2. Combined output removes duplicates across query overlaps.
3. Seen workflow flags (`--seen`, `--new-only`, `--new-out`, `--seen-update`)
   work unchanged on merged results.
4. Single-query behavior remains backward compatible.
5. Usage docs include at least one multi-query example and semantics note for
   `--limit`.
