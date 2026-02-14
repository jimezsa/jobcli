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

func TestParseLinkedInJobs_ExtractsSnippet(t *testing.T) {
	html := `
<ul>
  <li>
    <a class="base-card__full-link" href="https://www.linkedin.com/jobs/view/2"></a>
    <h3 class="base-search-card__title">Platform Engineer</h3>
    <h4 class="base-search-card__subtitle">Remote Co</h4>
    <span class="job-search-card__location">Berlin, Germany</span>
    <div class="job-search-card__snippet">
      This is a remote-first position
    </div>
  </li>
</ul>`

	doc := mustDoc(t, html)
	jobs := parseLinkedInJobs(doc)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Snippet != "This is a remote-first position" {
		t.Fatalf("unexpected snippet: %q", jobs[0].Snippet)
	}
	if !jobs[0].Remote {
		t.Fatalf("expected remote to be true from snippet")
	}
}

func TestParseLinkedInJobs_DoesNotUseLocationAsSnippet(t *testing.T) {
	html := `
<ul>
  <li>
    <a class="base-card__full-link" href="https://www.linkedin.com/jobs/view/3"></a>
    <h3 class="base-search-card__title">Backend Engineer</h3>
    <h4 class="base-search-card__subtitle">Example Co</h4>
    <span class="job-search-card__location">Munich, Germany</span>
    <div class="base-search-card__metadata">
      <span>Munich, Germany</span>
    </div>
  </li>
</ul>`

	doc := mustDoc(t, html)
	jobs := parseLinkedInJobs(doc)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Snippet != "" {
		t.Fatalf("expected empty snippet, got %q", jobs[0].Snippet)
	}
}

func TestLinkedInDetailURL(t *testing.T) {
	got := linkedInDetailURL("https://de.linkedin.com/jobs/view/electronics-engineer-f-m-d-at-omnisent-4361039203?position=1&pageNum=0")
	want := "https://www.linkedin.com/jobs-guest/jobs/api/jobPosting/4361039203"
	if got != want {
		t.Fatalf("unexpected detail url: got %q want %q", got, want)
	}
}

func TestParseLinkedInDescription(t *testing.T) {
	html := `<div class="show-more-less-html__markup">Build APIs for distributed systems.</div>`
	doc := mustDoc(t, html)
	got := parseLinkedInDescription(doc)
	if got != "Build APIs for distributed systems." {
		t.Fatalf("unexpected description: %q", got)
	}
}
