package scraper

import (
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jimezsa/jobcli/internal/models"
)

func TestParsePostedAt(t *testing.T) {
	cases := []struct {
		value  string
		layout string
	}{
		{"2024-01-02", "2006-01-02"},
		{"2024-01-02T15:04:05-0700", "2006-01-02T15:04:05-0700"},
		{time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC).Format(time.RFC3339), time.RFC3339},
	}

	for _, tc := range cases {
		parsed, err := parsePostedAt(tc.value)
		if err != nil {
			t.Fatalf("expected parse success for %s: %v", tc.value, err)
		}
		if parsed.IsZero() {
			t.Fatalf("parsed time should not be zero for %s", tc.value)
		}
	}
}

func TestAbsoluteURL(t *testing.T) {
	base := "https://example.com/path/page"
	cases := []struct {
		href string
		want string
	}{
		{"/jobs/1", "https://example.com/jobs/1"},
		{"https://other.com/a", "https://other.com/a"},
		{"//cdn.example.com/asset", "https://cdn.example.com/asset"},
	}

	for _, tc := range cases {
		got := absoluteURL(base, tc.href)
		if got != tc.want {
			t.Fatalf("absoluteURL(%q) = %q, want %q", tc.href, got, tc.want)
		}
	}
}

func TestParseJSONLDJobs(t *testing.T) {
	html := `
<!doctype html>
<html>
<head>
  <script type="application/ld+json">
  {
    "@context": "http://schema.org",
    "@type": "JobPosting",
    "title": "Go Developer",
    "hiringOrganization": {"name": "Acme"},
    "jobLocation": {"address": {"addressLocality": "Austin", "addressRegion": "TX", "addressCountry": "US"}},
    "url": "https://example.com/job1",
    "datePosted": "2024-01-15",
    "description": "Build APIs"
  }
  </script>
  <script type="application/ld+json">
  {
    "@graph": [
      {
        "@type": "JobPosting",
        "title": "Platform Engineer",
        "hiringOrganization": {"name": "Beta"},
        "jobLocation": {"address": {"addressLocality": "Remote"}},
        "url": "https://example.com/job2",
        "datePosted": "2024-01-16",
        "description": "Remote role"
      }
    ]
  }
  </script>
</head>
<body></body>
</html>`

	doc := mustDoc(t, html)
	jobs := parseJSONLDJobs(doc, SiteGoogleJobs)
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	if jobs[0].Title == "" || jobs[0].Company == "" || jobs[0].URL == "" {
		t.Fatalf("job missing required fields: %+v", jobs[0])
	}
}

func TestJSONLDSalaryAndLocation(t *testing.T) {
	job := jobFromJobPosting(map[string]any{
		"title":              "SRE",
		"hiringOrganization": map[string]any{"name": "Gamma"},
		"url":                "https://example.com/job3",
		"baseSalary": map[string]any{
			"currency": "USD",
			"value":    map[string]any{"minValue": 100000, "maxValue": 150000},
		},
		"jobLocation": map[string]any{
			"address": map[string]any{
				"streetAddress":   "1 Main",
				"addressLocality": "Denver",
				"addressRegion":   "CO",
				"postalCode":      "80202",
				"addressCountry":  "US",
			},
		},
		"description": strings.Repeat("a", 300),
	}, SiteLinkedIn)

	if job.Salary == "" || !strings.Contains(job.Salary, "USD") {
		t.Fatalf("expected salary with currency, got %q", job.Salary)
	}
	if !strings.Contains(job.Location, "Denver") {
		t.Fatalf("expected location to include city, got %q", job.Location)
	}
	if !strings.HasSuffix(job.Snippet, "...") {
		t.Fatalf("expected snippet to be truncated, got %q", job.Snippet)
	}
}

func TestDedupeAndFilterRemote(t *testing.T) {
	jobs := []models.Job{
		{Site: "x", Title: "A", Company: "C", Location: "Remote", URL: "https://example.com/a", Remote: true},
		{Site: "x", Title: "A", Company: "C", Location: "Remote", URL: "https://example.com/a", Remote: true},
		{Site: "x", Title: "B", Company: "C", Location: "NY", URL: "https://example.com/b", Remote: false},
	}

	deduped := dedupeJobs(jobs)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 jobs after dedupe, got %d", len(deduped))
	}

	remote := filterRemote(deduped)
	if len(remote) != 1 {
		t.Fatalf("expected 1 remote job, got %d", len(remote))
	}
}

func mustDoc(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse document: %v", err)
	}
	return doc
}
