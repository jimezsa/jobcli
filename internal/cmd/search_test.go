package cmd

import (
	"io"
	"path/filepath"
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
