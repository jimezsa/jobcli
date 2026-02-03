package scraper

import (
	"strings"
	"testing"

	"github.com/MrJJimenez/jobcli/internal/models"
)

func TestBuildIndeedURL(t *testing.T) {
	params := defaultParams()
	params.Query = "golang"
	params.Location = "New York, NY"
	params.Country = "us"
	params.Offset = 20
	params.JobType = "fulltime"

	url := buildIndeedURL(params)
	if url == "" {
		t.Fatalf("expected URL to be built")
	}
	if !containsAll(url, []string{"q=golang", "l=New+York%2C+NY", "start=20", "jt=fulltime"}) {
		t.Fatalf("unexpected indeed url: %s", url)
	}
}

func TestNormalizeSnippet(t *testing.T) {
	input := "  hello  \n world  "
	got := normalizeSnippet(input)
	if got != "hello world" {
		t.Fatalf("expected normalized snippet, got %q", got)
	}
}

func defaultParams() models.SearchParams { return models.SearchParams{} }

func containsAll(value string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
