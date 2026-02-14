# Seen Jobs Feature Plan

## Objective

Implement a JSON-based "seen jobs" workflow to prevent re-reporting jobs already
processed in previous runs.

- `A.json`: new search results
- `B.json`: historical seen jobs
- `C.json`: unseen jobs (`A - B`)

Primary match rule: compare normalized `title + company` (with URL fallback when title/company is missing).

## Final Command Surface

New command group:

- `jobcli seen diff`
- `jobcli seen update`

Search/site integration flags:

- `--seen` path to `B.json`
- `--new-only` output only unseen jobs
- `--new-out` optional file path to write unseen jobs JSON (`C.json`)

## JSON Contract

All files are JSON arrays compatible with existing `jobcli --json` output
(`[]models.Job`).

Preferred fields for comparison:

- `title`
- `company`

Fallback field (when title/company is missing):

- `url`

All other job fields are preserved when writing `C.json` and updating `B.json`.

## Comparison Rules

Normalized key:

- Primary: `key = normalize(title) + "::" + normalize(company)`
- Fallback: `key = "url::" + normalize(url)`

`normalize()` v1:

1. trim leading/trailing spaces
2. lowercase
3. collapse internal repeated whitespace to one space

Behavior:

- Record is invalid for comparison only if both primary and fallback keys are unavailable.
- `seen diff` skips invalid records and reports count.
- `A` and `B` are deduplicated by key during processing.
- A job is considered seen if its key exists in `B`.

## Command Specs

### `jobcli seen diff`

Purpose: compute unseen jobs list `C`.

Flags:

- `--new` required, path to `A.json`
- `--seen` required, path to `B.json` (missing file treated as empty list)
- `--out` required, output path for `C.json`
- `--stats` optional, print counts

Output:

- Writes `C.json` (JSON array of unseen jobs)
- Stats include:
  - total `A`
  - total `B`
  - invalid skipped
  - unseen emitted (`C`)

### `jobcli seen update`

Purpose: merge ranked/accepted unseen jobs into the seen history `B`.

Flags:

- `--seen` required, current `B.json` (missing file treated as empty list)
- `--input` required, source list to append (typically ranked `C.json`)
- `--out` required, output path for updated `B.json` (can match `--seen`)
- `--stats` optional, print counts

Merge rules:

- Use same normalized key (primary `title + company`, fallback `url`)
- Keep existing `B` entry on key collision
- Append only new unique entries from `--input`

## Search and Site Command Integration

Apply to both `jobcli search` and site commands (`linkedin`, `indeed`, etc.)
through shared `SearchOptions`.

Flow inside `runSearch`:

1. Scrape and aggregate jobs (`A`).
2. If `--seen` is provided, load `B` and compute `C = A - B`.
3. Export:

- `--new-only=true`: output `C`
- otherwise: output current behavior (`A`)

4. If `--new-out` is set, write `C` to JSON file regardless of display/output mode.

Examples:

```bash
jobcli search "hardware engineer" --location "Munich, Germany" --limit 30 \
  --seen jobs_seen.json --new-only --json --output jobs_new.json
```

```bash
jobcli linkedin "hardware engineer" --seen jobs_seen.json --new-only --json
```

## Implementation Plan

### 1) Add Seen Domain Package

Create `internal/seen`:

- `internal/seen/compare.go`
  - normalization
  - key generation
  - diff (`A - B`)
  - merge/update logic
- `internal/seen/io.go`
  - read/write `[]models.Job` JSON
  - missing-file-as-empty helper for `B`

### 2) Add CLI Commands

Create `internal/cmd/seen.go`:

- `SeenCmd` with `Diff` and `Update` subcommands
- parse flags and call `internal/seen` logic
- print optional stats

Update `internal/cmd/root.go`:

- register `Seen SeenCmd`

### 3) Wire SearchOptions

Update `internal/cmd/search.go`:

- extend `SearchOptions` with `Seen`, `NewOnly`, `NewOut`
- in `runSearch`, compute unseen when `Seen` provided
- preserve current behavior when `Seen` is not provided

### 4) Docs

Update:

- `docs/usage.md` command list and new flags
- `docs/spec.md` new `seen` command group and behavior

## Ranking Workflow Integration

Target flow for `skills/SKILL-jobcli-ranking.md`:

1. Search jobs to JSON (`A`).
2. Filter unseen:
   - `jobcli seen diff --new A.json --seen B.json --out C.json`
   - or use one-step search with `--seen --new-only`.
3. Rank only `C.json`.
4. Persist seen history:
   - `jobcli seen update --seen B.json --input C_ranked.json --out B.json`

This ensures already-seen jobs are not re-ranked in future runs.

## Test Plan

Unit tests (`internal/seen/*_test.go`):

- normalization and key generation
- diff with overlaps
- diff with duplicates in `A`/`B`
- invalid records handling
- update merge behavior
- idempotency (`seen update` repeated twice is stable)

Command tests:

- `seen diff` produces correct `C.json`
- `seen update` writes merged `B.json`
- `search/site --seen --new-only` outputs unseen only
- `search/site --new-out` writes unseen JSON file

## Acceptance Criteria

- `seen diff` correctly outputs unseen `C.json` from `A.json` and `B.json` using
  normalized keys (primary `title + company`, fallback `url`).
- `seen update` merges new ranked jobs into `B.json` without duplicates.
- `search` and site commands support `--seen` and can return unseen jobs in one
  command.
- All new file IO for this feature uses JSON arrays compatible with
  `jobcli --json`.
