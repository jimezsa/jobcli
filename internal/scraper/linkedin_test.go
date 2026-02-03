package scraper

import "testing"

func TestParseLinkedInJobs(t *testing.T) {
	html := `
<ul>
  <li>
    <a class="base-card__full-link" href="https://www.linkedin.com/jobs/view/1"></a>
    <h3 class="base-search-card__title">Staff Engineer</h3>
    <h4 class="base-search-card__subtitle">Example Co</h4>
    <span class="job-search-card__location">Remote</span>
    <time datetime="2024-01-10"></time>
  </li>
</ul>`

	doc := mustDoc(t, html)
	jobs := parseLinkedInJobs(doc)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Title != "Staff Engineer" {
		t.Fatalf("unexpected title: %q", jobs[0].Title)
	}
	if !jobs[0].Remote {
		t.Fatalf("expected remote to be true")
	}
}
