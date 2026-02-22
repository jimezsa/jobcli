# Multi Search Feature Spec

## Objective

Implement multi-query search support for `jobcli search` (and site commands that
share `runSearch`) so users can run one command with multiple keywords, either
from CLI input or a JSON file with job titles.

CLI example:

```bash
jobcli search "software engineer, hardware engineer, data scientic" --location "Munich, Germany"
```

JSON file example:

```bash
jobcli search --query-file queries.json --location "Munich, Germany"
```

The feature must preserve existing single-query behavior while adding predictable
and documented behavior for comma-separated query lists and file-based query
input.

## Scope

In scope:

- Parse a single positional `<query>` argument into multiple queries when comma-separated.
- Load job-title queries from a JSON file via a CLI flag.
- Merge query input from positional arg and JSON file into one normalized query list.
- Execute searches for each parsed query using existing scraper flow.
- Merge results into one output stream.
- Keep compatibility with existing output and seen-history flags.

Out of scope (v1):

- Boolean query syntax (`AND`, `OR`, parentheses).
- Weighted ranking across query relevance.
- Query-specific filters in a single command (different location/limit per query).
- Nested/complex JSON query objects beyond the supported schema.

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

New optional flag (for `search` and site commands):

- `--query-file <path>`: load job-title queries from JSON.

Input precedence and merge behavior:

- If only positional `<query>` is provided, use positional queries.
- If only `--query-file` is provided, use file queries.
- If both are provided, concatenate (positional first, file second), then dedupe.
- If neither provides at least one valid query, fail.

## Query Input Rules (v2)

Input:

- Raw positional argument string from `SearchCmd.Query` or `SiteCmd.Query`.
- Optional JSON file pointed to by `--query-file`.

Rules:

1. Parse positional input:
   - Split on comma (`,`).
   - Trim whitespace per part.
   - Drop empty entries.
2. Parse JSON file input (if present):
   - Accept either a top-level string array:
     - `["software engineer", "data scientist"]`
   - Or an object containing `job_titles` string array:
     - `{"job_titles":["software engineer","data scientist"]}`
   - Trim whitespace and drop empty entries.
3. Merge positional + file queries (positional first).
4. Deduplicate queries case-insensitively while preserving first-seen order.

Validation:

- If final query list is empty, return: `at least one non-empty query is required`.
- Limit maximum queries to `10` to avoid accidental overload.
- If query count exceeds `10`, return: `too many queries: max 10`.
- If `--query-file` path is unreadable, return a file-read error with the path.
- If JSON is invalid or schema is unsupported, return a clear schema-validation error.

Examples:

- `"software engineer"` -> `["software engineer"]`
- `"software engineer, hardware engineer"` -> `["software engineer","hardware engineer"]`
- `"software engineer, , Data Scientist"` -> `["software engineer","Data Scientist"]`
- `"Backend,backend, BACKEND"` -> `["Backend"]`
- `--query-file queries.json` with `{"job_titles":["Backend","backend","SRE"]}` -> `["Backend","SRE"]`
- positional `"Data Engineer"` + file `{"job_titles":["data engineer","ML Engineer"]}` -> `["Data Engineer","ML Engineer"]`

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

`--limit` is the max number of rows fetched per query:

- Apply limit to each query result set before cross-query merge/dedup.
- If `--limit <= 0`, fetch/output all merged results.

With multiple queries, final merged output may exceed `--limit`.

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
- Query-file read/parse/validation errors fail fast before network calls.
- Spinner remains one spinner for the overall command execution.

## Implementation Plan

### 1) Add query-file flag to search inputs

- Extend search/site options with:
  - `--query-file <path>` (string path to JSON)
- Keep existing positional `<query>` argument for backward compatibility.

### 2) Add JSON query loader helper

Add helper in `internal/cmd/search.go` (or `internal/cmd/search_queries.go`):

- `loadQueriesFromJSON(path string) ([]string, error)`

Responsibilities:

- read file content
- parse supported schemas:
  - top-level array of strings
  - object with `job_titles` array
- trim/filter empty entries
- return deterministic ordered list or descriptive validation errors

### 3) Unify query-source parsing

Refactor parsing flow to produce one final query list from both sources:

- `parseQueries(raw string)` for positional input
- `loadQueriesFromJSON(path string)` for file input (optional)
- `mergeAndNormalizeQueries(positional, fromFile []string) ([]string, error)`:
  - positional first, then file
  - case-insensitive dedupe preserving first-seen
  - max query enforcement
  - empty-final-list enforcement

### 4) Keep execution pipeline unchanged after query list creation

After final query list is built:

- keep current per-query execution flow
- keep per-query `--limit` behavior
- keep merge/dedupe/sort behavior
- keep seen-history behavior unchanged

### 5) Keep site commands compatible

Site commands already use the shared search path; they should gain
`--query-file` behavior with no separate implementation path.

### 6) Docs updates (after implementation)

Update:

- `README.md` quick start + flags notes
- `docs/usage.md` search examples and multi-query notes
- `docs/multi-search-spec.md` (this file) as the source of truth for JSON schema

Include examples for:

- positional-only input
- file-only input
- positional + file merge input

## Test Plan

Unit tests in `internal/cmd/search_test.go`:

- `parseQueries` single query.
- `parseQueries` multi query with spaces.
- `parseQueries` empty/blank tokens removed.
- `parseQueries` case-insensitive dedupe preserves first token.
- `parseQueries` max-query validation.
- `parseQueries` empty input validation.
- `loadQueriesFromJSON` with top-level array schema.
- `loadQueriesFromJSON` with `job_titles` object schema.
- `loadQueriesFromJSON` invalid JSON and unsupported schema.
- query-source merge test (positional + file) with order-preserving dedupe.
- query-source merge test enforcing max-10 across combined sources.

Behavioral tests in `internal/cmd/search_test.go`:

- Merged multi-query result set is deduplicated via seen-key behavior.
- `--new-only` on multi-query output returns unseen jobs only.
- `--seen-update` updates history once with merged unseen jobs.
- `--limit` applied per query before merge/dedup.
- file-based query source triggers multi-query execution with identical semantics.

Regression tests:

- Single-query path produces same output as before for equivalent fixtures.
- Existing positional multi-query behavior remains unchanged when `--query-file`
  is not used.

## Acceptance Criteria

1. `jobcli search --query-file queries.json` executes searches for all valid job
   titles from the file.
2. `jobcli search "a,b" --query-file queries.json` merges both sources,
   preserves first-seen order, and deduplicates case-insensitively.
3. Combined output removes duplicates across query overlaps.
4. Seen workflow flags (`--seen`, `--new-only`, `--new-out`, `--seen-update`)
   work unchanged on merged results.
5. Single-query positional behavior remains backward compatible.
6. Invalid/missing/unsupported query file formats fail fast with clear errors.
7. Usage docs include at least one `--query-file` example and semantics note for
   combined-source query handling.

---

## V3 Plan: Query File As Full Search Profile (No Code Yet)

## Objective (v3)

Allow this:

```bash
jobcli search --query-file queries.json
```

to behave the same as:

```bash
jobcli search --query-file queries.json --location "Munich, Germany" --limit 5 --hours 28 \
  --seen jobs_seen.json --seen-update --new-only --json --output jobs_new.json
```

by storing search defaults in `queries.json`.

## Proposed JSON Schema (Backward Compatible)

Continue supporting existing query-only formats:

- `["backend","platform"]`
- `{"job_titles":["backend","platform"]}`

Add an extended object format:

```json
{
  "job_titles": [
    "software engineer",
    "hardware engineer"
  ],
  "search_options": {
    "location": "Munich, Germany",
    "country": "de",
    "sites": "all",
    "limit": 5,
    "offset": 0,
    "remote": false,
    "job_type": "",
    "hours": 28,
    "format": "",
    "links": "full",
    "output": "jobs_new.json",
    "proxies": "",
    "seen": "jobs_seen.json",
    "new_only": true,
    "new_out": "",
    "seen_update": true
  },
  "global_options": {
    "json": true,
    "plain": false,
    "color": "auto",
    "verbose": false
  }
}
```

Notes:

- `output` is the canonical JSON field for `--output/--out/--file`.
- `sites` in file applies only to `search`; site commands ignore it.
- Unknown fields should fail validation with a clear error.

## Option Precedence Rules

Final option values should be resolved in this order:

1. Explicit CLI args/flags (highest priority).
2. `query-file` defaults (`search_options`, `global_options`).
3. Environment variables (`JOBCLI_*`).
4. Config file defaults.
5. Built-in defaults.

Examples:

- CLI `--limit 20` overrides file `search_options.limit=5`.
- CLI `--new-only` overrides file `search_options.new_only=false`.
- If CLI omits `--location`, file `search_options.location` is used.

## Functional Rules

1. Query resolution:
   - Keep current behavior (positional queries + file `job_titles` merge/dedupe).
   - If positional query is absent, use `job_titles` from file.
2. Search options:
   - Apply `search_options` defaults before running `runSearch`.
   - Keep existing validation constraints (`--new-only` requires `--seen`, path conflict checks, etc.).
3. Global options:
   - Apply `global_options` before output mode resolution.
   - Preserve current `--json` + `--plain` mutual exclusion.
4. Output behavior:
   - File-driven JSON mode must still keep stdout/stderr compatibility rules unchanged.
5. Backward compatibility:
   - Existing array/object query-file formats continue to work unchanged.

## Validation Rules (v3)

- `job_titles` (if present) must be string array.
- `search_options.limit`, `offset`, `hours` must be numeric.
- Enum fields must match existing CLI enums:
  - `job_type`: `fulltime|parttime|contract|internship|""`
  - `format`: `csv|json|md|""`
  - `links`: `short|full`
  - `color`: `auto|always|never`
- Reject invalid `global_options` combinations (`json=true` and `plain=true`).
- Reject invalid schema with explicit field-level errors.

## Implementation Plan (v3)

### 1) Introduce profile parser for `--query-file`

- Expand query-file decoding to support:
  - current query-only schemas
  - new extended object with `search_options` and `global_options`
- Return one structured result:
  - `queries []string`
  - optional option defaults for search/global layers

### 2) Add option-merge layer before `runSearch`

- Resolve global defaults into CLI runtime context.
- Resolve search defaults into `SearchOptions`.
- Keep CLI flags as highest priority.

### 3) Preserve current behavior paths

- Existing commands with explicit flags must keep current behavior.
- Existing minimal query files must continue to behave exactly as today.

### 4) Add strict schema/enum validation

- Fail fast on bad types, unknown keys, invalid enum values.
- Keep errors specific and actionable (field name + expected type/value).

### 5) Documentation updates

- Update `README.md` with:
  - minimal query file example
  - full profile query file example
  - precedence table (CLI vs file vs env)
- Update `docs/usage.md` with file-profile examples.
- Keep this spec as source of truth.

## Test Plan (v3)

- Parsing tests:
  - minimal array schema still works
  - minimal `job_titles` object still works
  - full profile schema parses and maps correctly
  - invalid field types/enums fail with clear errors
- Precedence tests:
  - CLI overrides query-file defaults
  - query-file overrides env/config defaults
- Behavior tests:
  - `jobcli search --query-file queries.json` equals the long explicit command behavior
  - seen workflow (`seen`, `new_only`, `seen_update`, `output`) matches explicit flags
  - global output mode (`json/plain/color/verbose`) matches explicit flags
- Regression tests:
  - existing non-profile query-file usage remains unchanged
  - existing positional-only and positional+file query merge remains unchanged

## Acceptance Criteria (v3)

1. Running `jobcli search --query-file queries.json` with full profile options
   yields the same behavior as passing those flags explicitly on CLI.
2. CLI flags still override values provided by `queries.json`.
3. Existing query-only JSON formats remain fully supported.
4. Invalid profile schema/values fail fast with clear field-level errors.
5. Docs include both minimal and full-profile query-file examples plus
   precedence rules.
