package seen

import (
	"path/filepath"
	"testing"

	"github.com/jimezsa/jobcli/internal/models"
)

func TestReadWriteJobs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "jobs.json")

	jobs := []models.Job{{Title: "SRE", Company: "Acme", URL: "https://example.com/1"}}
	if err := WriteJobs(path, jobs); err != nil {
		t.Fatalf("WriteJobs() error = %v", err)
	}

	got, err := ReadJobs(path)
	if err != nil {
		t.Fatalf("ReadJobs() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected len=1, got %d", len(got))
	}
	if got[0].Title != jobs[0].Title || got[0].Company != jobs[0].Company {
		t.Fatalf("unexpected job read back: %+v", got[0])
	}
}

func TestReadJobsAllowMissing(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.json")

	got, err := ReadJobsAllowMissing(missing)
	if err != nil {
		t.Fatalf("ReadJobsAllowMissing() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty jobs for missing file, got %d", len(got))
	}
}
