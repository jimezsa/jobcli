package scraper

import "testing"

func TestParseGoogleJobsAnchors(t *testing.T) {
	html := `
<div>
  <a href="/search?htidocid=abc123">Site Reliability Engineer</a>
</div>`

	doc := mustDoc(t, html)
	jobs := parseGoogleJobsAnchors(doc)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Title != "Site Reliability Engineer" {
		t.Fatalf("unexpected title: %q", jobs[0].Title)
	}
	if jobs[0].URL == "" {
		t.Fatalf("expected URL to be set")
	}
}

func TestCountryToGoogleGL(t *testing.T) {
	cases := map[string]string{
		"usa": "us",
		"US":  "us",
		"uk":  "gb",
		"ca":  "ca",
	}
	for input, want := range cases {
		if got := countryToGoogleGL(input); got != want {
			t.Fatalf("countryToGoogleGL(%q) = %q, want %q", input, got, want)
		}
	}
}
