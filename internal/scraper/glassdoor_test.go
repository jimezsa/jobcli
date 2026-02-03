package scraper

import "testing"

func TestParseGlassdoorJobs(t *testing.T) {
	html := `
<div class="react-job-listing">
  <a class="jobLink" href="/Job/example-job.htm">Senior Dev</a>
  <div class="jobEmployerName">Glassdoor Inc</div>
  <div class="jobLocation">Austin, TX</div>
  <div class="salarySnippet">$120k</div>
</div>`

	doc := mustDoc(t, html)
	jobs := parseGlassdoorJobs(doc)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Company != "Glassdoor Inc" {
		t.Fatalf("unexpected company: %q", jobs[0].Company)
	}
	if jobs[0].URL == "" {
		t.Fatalf("expected URL to be set")
	}
}
