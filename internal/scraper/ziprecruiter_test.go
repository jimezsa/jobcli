package scraper

import "testing"

func TestParseZipRecruiterJobs(t *testing.T) {
	html := `
<article class="job_result">
  <a class="job_link" href="/c/example/job/123">Platform Engineer</a>
  <a class="t_org_link">Zip Co</a>
  <div class="location">Remote</div>
  <div class="job_snippet">Build systems</div>
</article>`

	doc := mustDoc(t, html)
	jobs := parseZipRecruiterJobs(doc)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Title != "Platform Engineer" {
		t.Fatalf("unexpected title: %q", jobs[0].Title)
	}
	if !jobs[0].Remote {
		t.Fatalf("expected remote to be true")
	}
}
