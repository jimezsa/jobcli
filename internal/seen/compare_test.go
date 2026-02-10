package seen

import (
	"testing"

	"github.com/jimezsa/jobcli/internal/models"
)

func TestNormalize(t *testing.T) {
	got := Normalize("  Senior   Software\tEngineer  ")
	want := "senior software engineer"
	if got != want {
		t.Fatalf("Normalize() = %q, want %q", got, want)
	}
}

func TestKey(t *testing.T) {
	job := models.Job{Title: "  Senior Engineer ", Company: " ACME   Corp "}
	got, ok := Key(job)
	if !ok {
		t.Fatalf("expected valid key")
	}
	want := "senior engineer::acme corp"
	if got != want {
		t.Fatalf("Key() = %q, want %q", got, want)
	}
}

func TestDiff(t *testing.T) {
	newJobs := []models.Job{
		{Title: "Senior Engineer", Company: "Acme", URL: "https://example.com/new-1"},
		{Title: "Senior   Engineer", Company: " Acme ", URL: "https://example.com/new-1-dupe"},
		{Title: "Platform Engineer", Company: "Beta", URL: "https://example.com/new-2"},
		{Title: "", Company: "Invalid", URL: "https://example.com/invalid"},
	}
	seenJobs := []models.Job{
		{Title: "senior engineer", Company: "acme", URL: "https://example.com/seen-1"},
		{Title: "senior engineer", Company: "acme", URL: "https://example.com/seen-1-dupe"},
		{Title: "No Company", Company: "   ", URL: "https://example.com/seen-invalid"},
	}

	unseen, stats := Diff(newJobs, seenJobs)

	if len(unseen) != 1 {
		t.Fatalf("expected 1 unseen job, got %d", len(unseen))
	}
	if unseen[0].Title != "Platform Engineer" {
		t.Fatalf("unexpected unseen job: %+v", unseen[0])
	}

	if stats.TotalNew != 4 {
		t.Fatalf("TotalNew = %d, want 4", stats.TotalNew)
	}
	if stats.TotalSeen != 3 {
		t.Fatalf("TotalSeen = %d, want 3", stats.TotalSeen)
	}
	if stats.InvalidNew != 1 {
		t.Fatalf("InvalidNew = %d, want 1", stats.InvalidNew)
	}
	if stats.InvalidSeen != 1 {
		t.Fatalf("InvalidSeen = %d, want 1", stats.InvalidSeen)
	}
	if stats.InvalidSkipped() != 2 {
		t.Fatalf("InvalidSkipped = %d, want 2", stats.InvalidSkipped())
	}
	if stats.Unseen != 1 {
		t.Fatalf("Unseen = %d, want 1", stats.Unseen)
	}
}

func TestMergeAndIdempotency(t *testing.T) {
	existing := []models.Job{
		{Title: "Senior Engineer", Company: "Acme", URL: "https://example.com/seen-1"},
		{Title: "", Company: "Unknown", URL: "https://example.com/seen-invalid"},
	}
	input := []models.Job{
		{Title: "Senior Engineer", Company: "Acme", URL: "https://example.com/new-collision"},
		{Title: "Platform Engineer", Company: "Beta", URL: "https://example.com/new-2"},
		{Title: "", Company: "Broken", URL: "https://example.com/new-invalid"},
	}

	merged, stats := Merge(existing, input)
	if len(merged) != 3 {
		t.Fatalf("expected merged len=3, got %d", len(merged))
	}
	if stats.Added != 1 {
		t.Fatalf("Added = %d, want 1", stats.Added)
	}
	if stats.InvalidSeen != 1 {
		t.Fatalf("InvalidSeen = %d, want 1", stats.InvalidSeen)
	}
	if stats.InvalidInput != 1 {
		t.Fatalf("InvalidInput = %d, want 1", stats.InvalidInput)
	}
	if stats.TotalOut != 3 {
		t.Fatalf("TotalOut = %d, want 3", stats.TotalOut)
	}

	mergedAgain, statsAgain := Merge(merged, input)
	if len(mergedAgain) != len(merged) {
		t.Fatalf("expected idempotent merge length %d, got %d", len(merged), len(mergedAgain))
	}
	if statsAgain.Added != 0 {
		t.Fatalf("expected second merge Added=0, got %d", statsAgain.Added)
	}
}
