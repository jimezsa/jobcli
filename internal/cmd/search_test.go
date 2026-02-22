package cmd

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/jimezsa/jobcli/internal/export"
	"github.com/jimezsa/jobcli/internal/models"
	"github.com/jimezsa/jobcli/internal/seen"
)

func TestResolveFormatWithOutputPathRespectsGlobalFlags(t *testing.T) {
	ctx := &Context{Out: io.Discard, JSONOutput: true}
	got, err := resolveFormat(ctx, SearchOptions{}, "jobs.json")
	if err != nil {
		t.Fatalf("resolveFormat() error = %v", err)
	}
	if got != export.FormatJSON {
		t.Fatalf("resolveFormat() = %q, want %q", got, export.FormatJSON)
	}

	ctx = &Context{Out: io.Discard, PlainText: true}
	got, err = resolveFormat(ctx, SearchOptions{}, "jobs.tsv")
	if err != nil {
		t.Fatalf("resolveFormat() error = %v", err)
	}
	if got != export.FormatTSV {
		t.Fatalf("resolveFormat() = %q, want %q", got, export.FormatTSV)
	}
}

func TestUpdateSeenHistoryCreatesFileAndMerges(t *testing.T) {
	dir := t.TempDir()
	seenPath := filepath.Join(dir, "jobs_seen.json")

	input := []models.Job{
		{Site: "test", Title: "Hardware Engineer", Company: "Acme", URL: "https://example.com/1"},
	}

	if err := updateSeenHistory(seenPath, input); err != nil {
		t.Fatalf("updateSeenHistory() error = %v", err)
	}

	got, err := seen.ReadJobs(seenPath)
	if err != nil {
		t.Fatalf("ReadJobs() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}

	// Calling it again with the same job should be idempotent.
	if err := updateSeenHistory(seenPath, input); err != nil {
		t.Fatalf("updateSeenHistory() (2nd) error = %v", err)
	}
	got, err = seen.ReadJobs(seenPath)
	if err != nil {
		t.Fatalf("ReadJobs() (2nd) error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) after 2nd update = %d, want 1", len(got))
	}

	// Add a second distinct job.
	input2 := []models.Job{
		{Site: "test", Title: "Hardware Engineer", Company: "Acme", URL: "https://example.com/1"},
		{Site: "test", Title: "Embedded Engineer", Company: "Beta", URL: "https://example.com/2"},
	}
	if err := updateSeenHistory(seenPath, input2); err != nil {
		t.Fatalf("updateSeenHistory() (3rd) error = %v", err)
	}
	got, err = seen.ReadJobs(seenPath)
	if err != nil {
		t.Fatalf("ReadJobs() (3rd) error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) after 3rd update = %d, want 2", len(got))
	}
}

func TestParseQueries(t *testing.T) {
	t.Run("single query", func(t *testing.T) {
		got, err := parseQueries("software engineer")
		if err != nil {
			t.Fatalf("parseQueries() error = %v", err)
		}
		want := []string{"software engineer"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("parseQueries() = %#v, want %#v", got, want)
		}
	})

	t.Run("multi query with spaces", func(t *testing.T) {
		got, err := parseQueries("software engineer, hardware engineer")
		if err != nil {
			t.Fatalf("parseQueries() error = %v", err)
		}
		want := []string{"software engineer", "hardware engineer"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("parseQueries() = %#v, want %#v", got, want)
		}
	})

	t.Run("empty tokens removed", func(t *testing.T) {
		got, err := parseQueries("software engineer, , Data Scientist")
		if err != nil {
			t.Fatalf("parseQueries() error = %v", err)
		}
		want := []string{"software engineer", "Data Scientist"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("parseQueries() = %#v, want %#v", got, want)
		}
	})

	t.Run("case-insensitive dedupe keeps first token", func(t *testing.T) {
		got, err := parseQueries("Backend,backend, BACKEND")
		if err != nil {
			t.Fatalf("parseQueries() error = %v", err)
		}
		want := []string{"Backend"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("parseQueries() = %#v, want %#v", got, want)
		}
	})

	t.Run("max query validation", func(t *testing.T) {
		input := strings.Join([]string{
			"q1", "q2", "q3", "q4", "q5",
			"q6", "q7", "q8", "q9", "q10", "q11",
		}, ",")
		_, err := parseQueries(input)
		if err == nil {
			t.Fatalf("parseQueries() error = nil, want error")
		}
		if err.Error() != "too many queries: max 10" {
			t.Fatalf("parseQueries() error = %q, want %q", err.Error(), "too many queries: max 10")
		}
	})

	t.Run("empty input validation", func(t *testing.T) {
		_, err := parseQueries(" ,  , ")
		if err == nil {
			t.Fatalf("parseQueries() error = nil, want error")
		}
		if err.Error() != "at least one non-empty query is required" {
			t.Fatalf("parseQueries() error = %q, want %q", err.Error(), "at least one non-empty query is required")
		}
	})
}

func TestLoadQueriesFromJSON(t *testing.T) {
	t.Run("top-level string array", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `["software engineer","  Data Scientist  ",""]`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got, err := loadQueriesFromJSON(path)
		if err != nil {
			t.Fatalf("loadQueriesFromJSON() error = %v", err)
		}
		want := []string{"software engineer", "Data Scientist"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("loadQueriesFromJSON() = %#v, want %#v", got, want)
		}
	})

	t.Run("object with job_titles", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `{"job_titles":["Backend Engineer","backend engineer","SRE"]}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got, err := loadQueriesFromJSON(path)
		if err != nil {
			t.Fatalf("loadQueriesFromJSON() error = %v", err)
		}
		want := []string{"Backend Engineer", "backend engineer", "SRE"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("loadQueriesFromJSON() = %#v, want %#v", got, want)
		}
	})

	t.Run("extended object with profile options", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `{
  "job_titles": ["Backend Engineer", "SRE"],
  "search_options": {"location":"Munich, Germany","limit":5},
  "global_options": {"json":true}
}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got, err := loadQueriesFromJSON(path)
		if err != nil {
			t.Fatalf("loadQueriesFromJSON() error = %v", err)
		}
		want := []string{"Backend Engineer", "SRE"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("loadQueriesFromJSON() = %#v, want %#v", got, want)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `{"job_titles":[`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		_, err := loadQueriesFromJSON(path)
		if err == nil {
			t.Fatalf("loadQueriesFromJSON() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "parse --query-file") {
			t.Fatalf("loadQueriesFromJSON() error = %q, want parse --query-file message", err.Error())
		}
	})

	t.Run("unsupported schema", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `{"queries":["backend"]}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		_, err := loadQueriesFromJSON(path)
		if err == nil {
			t.Fatalf("loadQueriesFromJSON() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "unknown field \"queries\"") {
			t.Fatalf("loadQueriesFromJSON() error = %q, want unknown-field message", err.Error())
		}
	})

	t.Run("non-string entry", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `{"job_titles":["backend",123]}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		_, err := loadQueriesFromJSON(path)
		if err == nil {
			t.Fatalf("loadQueriesFromJSON() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "job_titles[1] must be a string") {
			t.Fatalf("loadQueriesFromJSON() error = %q, want non-string index message", err.Error())
		}
	})
}

func TestResolveQueries(t *testing.T) {
	t.Run("query-file only", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `{"job_titles":["Backend","SRE"]}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got, err := resolveQueries("", path)
		if err != nil {
			t.Fatalf("resolveQueries() error = %v", err)
		}
		want := []string{"Backend", "SRE"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("resolveQueries() = %#v, want %#v", got, want)
		}
	})

	t.Run("positional plus query-file preserves first and dedupes case-insensitively", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `{"job_titles":["backend","ML Engineer","  "]}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got, err := resolveQueries("Backend,Data Engineer", path)
		if err != nil {
			t.Fatalf("resolveQueries() error = %v", err)
		}
		want := []string{"Backend", "Data Engineer", "ML Engineer"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("resolveQueries() = %#v, want %#v", got, want)
		}
	})

	t.Run("combined sources enforce max query validation", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `{"job_titles":["q7","q8","q9","q10","q11"]}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		_, err := resolveQueries("q1,q2,q3,q4,q5,q6", path)
		if err == nil {
			t.Fatalf("resolveQueries() error = nil, want error")
		}
		if err.Error() != "too many queries: max 10" {
			t.Fatalf("resolveQueries() error = %q, want %q", err.Error(), "too many queries: max 10")
		}
	})

	t.Run("both sources empty returns validation error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `{"job_titles":[" ",""]}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		_, err := resolveQueries(" , ", path)
		if err == nil {
			t.Fatalf("resolveQueries() error = nil, want error")
		}
		if err.Error() != "at least one non-empty query is required" {
			t.Fatalf("resolveQueries() error = %q, want %q", err.Error(), "at least one non-empty query is required")
		}
	})

	t.Run("query-file with options only and positional query works", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `{
  "search_options": {"location":"Munich, Germany"},
  "global_options": {"json":true}
}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got, err := resolveQueries("backend engineer", path)
		if err != nil {
			t.Fatalf("resolveQueries() error = %v", err)
		}
		want := []string{"backend engineer"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("resolveQueries() = %#v, want %#v", got, want)
		}
	})
}

func TestApplyQueryFileDefaults(t *testing.T) {
	t.Run("applies query-file defaults when CLI flags are not provided", func(t *testing.T) {
		location := "Munich, Germany"
		limit := 5
		hours := 28
		seenPath := "jobs_seen.json"
		output := "jobs_new.json"
		newOnly := true
		seenUpdate := true
		jsonOutput := true
		color := "never"

		ctx := &Context{
			Out:       io.Discard,
			Err:       io.Discard,
			Verbose:   false,
			ColorMode: "auto",
		}
		opts := SearchOptions{
			Location: "Berlin, Germany",
			Limit:    20,
		}
		cfg := queryFileConfig{
			Search: queryFileSearchOptions{
				Location:   &location,
				Limit:      &limit,
				Hours:      &hours,
				Seen:       &seenPath,
				Output:     &output,
				NewOnly:    &newOnly,
				SeenUpdate: &seenUpdate,
			},
			Global: queryFileGlobalOptions{
				JSON:  &jsonOutput,
				Color: &color,
			},
		}

		gotOpts, gotSites, err := applyQueryFileDefaults(ctx, opts, "all", cfg, true, map[string]bool{})
		if err != nil {
			t.Fatalf("applyQueryFileDefaults() error = %v", err)
		}

		if gotOpts.Location != location || gotOpts.Limit != limit || gotOpts.Hours != hours {
			t.Fatalf("search defaults were not applied: %+v", gotOpts)
		}
		if gotOpts.Seen != seenPath || gotOpts.Output != output || !gotOpts.NewOnly || !gotOpts.SeenUpdate {
			t.Fatalf("seen/output defaults were not applied: %+v", gotOpts)
		}
		if gotSites != "all" {
			t.Fatalf("sites changed unexpectedly: got %q", gotSites)
		}
		if !ctx.JSONOutput {
			t.Fatalf("expected ctx.JSONOutput=true from query-file defaults")
		}
		if ctx.ColorMode != "never" {
			t.Fatalf("expected ctx.ColorMode=never, got %q", ctx.ColorMode)
		}
	})

	t.Run("CLI flags override query-file defaults", func(t *testing.T) {
		location := "Munich, Germany"
		limit := 5
		output := "jobs_new.json"
		sites := "linkedin"
		jsonOutput := true
		verbose := true

		ctx := &Context{
			Out:        io.Discard,
			Err:        io.Discard,
			JSONOutput: false,
			Verbose:    false,
			ColorMode:  "auto",
		}
		opts := SearchOptions{
			Location: "Madrid, Spain",
			Limit:    25,
			Output:   "explicit.json",
		}
		cfg := queryFileConfig{
			Search: queryFileSearchOptions{
				Location: &location,
				Limit:    &limit,
				Output:   &output,
				Sites:    &sites,
			},
			Global: queryFileGlobalOptions{
				JSON:    &jsonOutput,
				Verbose: &verbose,
			},
		}
		provided := map[string]bool{
			"--location": true,
			"--limit":    true,
			"--output":   true,
			"--sites":    true,
			"--json":     true,
			"--verbose":  true,
		}

		gotOpts, gotSites, err := applyQueryFileDefaults(ctx, opts, "glassdoor", cfg, true, provided)
		if err != nil {
			t.Fatalf("applyQueryFileDefaults() error = %v", err)
		}

		if gotOpts.Location != "Madrid, Spain" || gotOpts.Limit != 25 || gotOpts.Output != "explicit.json" {
			t.Fatalf("CLI values should win, got %+v", gotOpts)
		}
		if gotSites != "glassdoor" {
			t.Fatalf("CLI sites should win, got %q", gotSites)
		}
		if ctx.JSONOutput {
			t.Fatalf("CLI json flag presence should prevent query-file override")
		}
		if ctx.Verbose {
			t.Fatalf("CLI verbose flag presence should prevent query-file override")
		}
	})

	t.Run("site command ignores search_options.sites", func(t *testing.T) {
		sites := "linkedin,indeed"
		ctx := &Context{
			Out:       io.Discard,
			Err:       io.Discard,
			ColorMode: "auto",
		}
		cfg := queryFileConfig{
			Search: queryFileSearchOptions{
				Sites: &sites,
			},
		}

		_, gotSites, err := applyQueryFileDefaults(ctx, SearchOptions{}, "stepstone", cfg, false, map[string]bool{})
		if err != nil {
			t.Fatalf("applyQueryFileDefaults() error = %v", err)
		}
		if gotSites != "stepstone" {
			t.Fatalf("site command should keep fixed site, got %q", gotSites)
		}
	})
}

func TestLoadQueryFileConfigValidation(t *testing.T) {
	t.Run("rejects conflicting global json/plain", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `{
  "job_titles": ["backend"],
  "global_options": {"json": true, "plain": true}
}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		_, err := loadQueryFileConfig(path)
		if err == nil {
			t.Fatalf("loadQueryFileConfig() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "cannot both be true") {
			t.Fatalf("loadQueryFileConfig() error = %q, want conflict error", err.Error())
		}
	})

	t.Run("rejects invalid enum values", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "queries.json")
		content := `{
  "job_titles": ["backend"],
  "search_options": {"links":"tiny"},
  "global_options": {"color":"blue"}
}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		_, err := loadQueryFileConfig(path)
		if err == nil {
			t.Fatalf("loadQueryFileConfig() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "search_options.links") {
			t.Fatalf("loadQueryFileConfig() error = %q, want links validation error", err.Error())
		}
	})
}

func TestMergeUniqueJobsDedupesAcrossQueries(t *testing.T) {
	existing := []models.Job{
		{Site: "linkedin", Title: "Backend Engineer", Company: "Acme", URL: "https://example.com/1"},
		{Site: "indeed", URL: "https://example.com/fallback"},
	}
	incoming := []models.Job{
		{Site: "glassdoor", Title: " backend engineer ", Company: "ACME", URL: "https://example.com/other"},
		{Site: "ziprecruiter", URL: "https://example.com/fallback"},
		{Site: "linkedin", Title: "Data Engineer", Company: "Acme", URL: "https://example.com/2"},
		{Site: "stepstone"},
	}

	got := mergeUniqueJobs(existing, incoming)
	if len(got) != 4 {
		t.Fatalf("len(got) = %d, want 4", len(got))
	}
	if got[0].Title != "Backend Engineer" || got[1].URL != "https://example.com/fallback" {
		t.Fatalf("existing jobs order/values changed: %#v", got[:2])
	}
	if got[2].Title != "Data Engineer" {
		t.Fatalf("expected unique incoming job at index 2, got %#v", got[2])
	}
	if got[3].Site != "stepstone" {
		t.Fatalf("expected invalid-key incoming job at index 3, got %#v", got[3])
	}
}

func TestMergeUniqueJobsKeepsSingleQueryDuplicates(t *testing.T) {
	incoming := []models.Job{
		{Site: "linkedin", Title: "Backend Engineer", Company: "Acme", URL: "https://example.com/1"},
		{Site: "indeed", Title: "Backend Engineer", Company: "Acme", URL: "https://example.com/2"},
	}

	got := mergeUniqueJobs(nil, incoming)
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
}

func TestLimitJobs(t *testing.T) {
	jobs := []models.Job{
		{Site: "linkedin", Title: "one"},
		{Site: "indeed", Title: "two"},
		{Site: "glassdoor", Title: "three"},
	}

	got := limitJobs(jobs, 2)
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}

	got = limitJobs(jobs, 0)
	if len(got) != 3 {
		t.Fatalf("len(got) with unlimited = %d, want 3", len(got))
	}
}

func TestFormatSearchSummary(t *testing.T) {
	t.Run("empty jobs", func(t *testing.T) {
		got := formatSearchSummary(nil)
		want := "summary: new_jobs=0 by_site=none"
		if got != want {
			t.Fatalf("formatSearchSummary() = %q, want %q", got, want)
		}
	})

	t.Run("counts by site sorted and normalized", func(t *testing.T) {
		jobs := []models.Job{
			{Site: "LinkedIn", Title: "A"},
			{Site: " indeed ", Title: "B"},
			{Site: "linkedin", Title: "C"},
			{Site: "", Title: "D"},
		}

		got := formatSearchSummary(jobs)
		want := "summary: new_jobs=4 by_site=indeed:1, linkedin:2, unknown:1"
		if got != want {
			t.Fatalf("formatSearchSummary() = %q, want %q", got, want)
		}
	})
}

func TestMultiQuerySeenWorkflowAndLimitPerQuery(t *testing.T) {
	dir := t.TempDir()
	seenPath := filepath.Join(dir, "jobs_seen.json")

	seenSeed := []models.Job{
		{Site: "linkedin", Title: "Platform Engineer", Company: "Acme", URL: "https://example.com/seed"},
	}
	if err := seen.WriteJobs(seenPath, seenSeed); err != nil {
		t.Fatalf("WriteJobs() seed error = %v", err)
	}

	queryOne := []models.Job{
		{Site: "linkedin", Title: "Platform Engineer", Company: "Acme", URL: "https://example.com/seed"},
		{Site: "indeed", Title: "Hardware Engineer", Company: "Beta", URL: "https://example.com/1"},
		{Site: "ziprecruiter", Title: "Embedded Engineer", Company: "Delta", URL: "https://example.com/extra-q1"},
	}
	queryTwo := []models.Job{
		{Site: "glassdoor", Title: "Hardware Engineer", Company: "beta", URL: "https://example.com/dup"},
		{Site: "ziprecruiter", Title: "Data Engineer", Company: "Gamma", URL: "https://example.com/2"},
		{Site: "linkedin", Title: "Ml Engineer", Company: "Epsilon", URL: "https://example.com/extra-q2"},
	}

	limit := 2
	limitedQ1 := limitJobs(queryOne, limit)
	limitedQ2 := limitJobs(queryTwo, limit)

	merged := mergeUniqueJobs(nil, limitedQ1)
	merged = mergeUniqueJobs(merged, limitedQ2)
	if len(merged) != 3 {
		t.Fatalf("len(merged) = %d, want 3", len(merged))
	}
	if len(merged) <= limit {
		t.Fatalf("final merged output should not be capped by per-query limit: len(merged) = %d, limit = %d", len(merged), limit)
	}

	seenJobs, err := seen.ReadJobsAllowMissing(seenPath)
	if err != nil {
		t.Fatalf("ReadJobsAllowMissing() error = %v", err)
	}
	unseenJobs, _ := seen.Diff(merged, seenJobs)
	if len(unseenJobs) != 2 {
		t.Fatalf("len(unseenJobs) = %d, want 2", len(unseenJobs))
	}

	if err := updateSeenHistory(seenPath, unseenJobs); err != nil {
		t.Fatalf("updateSeenHistory() error = %v", err)
	}
	updatedSeen, err := seen.ReadJobs(seenPath)
	if err != nil {
		t.Fatalf("ReadJobs() error = %v", err)
	}
	if len(updatedSeen) != 3 {
		t.Fatalf("len(updatedSeen) = %d, want 3", len(updatedSeen))
	}
}
