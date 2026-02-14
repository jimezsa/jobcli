package scraper

import "testing"

func TestStepstoneParseCard_ExtractsTeaserSnippet(t *testing.T) {
	html := `
<article>
  <h2>Platform Engineer</h2>
  <div>Example GmbH</div>
  <div>Munich, Bavaria, Germany</div>
  <p data-at="job-item-teaser">Build distributed backend services.</p>
  <time>vor 2 Tagen</time>
</article>`

	doc := mustDoc(t, html)
	card := doc.Find("article").First()
	company, location, snippet, posted, remote := stepstoneParseCard(card, "Platform Engineer")

	if company != "Example GmbH" {
		t.Fatalf("unexpected company: %q", company)
	}
	if location != "Munich, Bavaria, Germany" {
		t.Fatalf("unexpected location: %q", location)
	}
	if snippet != "Build distributed backend services." {
		t.Fatalf("unexpected snippet: %q", snippet)
	}
	if posted != "vor 2 Tagen" {
		t.Fatalf("unexpected posted: %q", posted)
	}
	if remote {
		t.Fatalf("expected remote false")
	}
}

func TestStepstoneParseCard_DoesNotUseLocationAsSnippet(t *testing.T) {
	html := `
<article>
  <h2>Platform Engineer</h2>
  <div>Example GmbH</div>
  <div>Munich, Bavaria, Germany</div>
  <div data-testid="job-item-teaser">Munich, Bavaria, Germany</div>
</article>`

	doc := mustDoc(t, html)
	card := doc.Find("article").First()
	_, _, snippet, _, _ := stepstoneParseCard(card, "Platform Engineer")

	if snippet != "" {
		t.Fatalf("expected empty snippet, got %q", snippet)
	}
}

func TestParseStepstoneDescription(t *testing.T) {
	html := `<div data-at="jobad-description">Build APIs for enterprise integrations.</div>`
	doc := mustDoc(t, html)

	got := parseStepstoneDescription(doc)
	if got != "Build APIs for enterprise integrations." {
		t.Fatalf("unexpected description: %q", got)
	}
}

func TestParseStepstoneDescription_FallsBackToJSONLD(t *testing.T) {
	html := `
<script type="application/ld+json">
{
  "@context": "http://schema.org",
  "@type": "JobPosting",
  "title": "Platform Engineer",
  "hiringOrganization": {"name": "Example GmbH"},
  "url": "https://www.stepstone.de/stellenangebote--platform-engineer-example",
  "description": "Design and operate resilient services."
}
</script>`
	doc := mustDoc(t, html)

	got := parseStepstoneDescription(doc)
	if got != "Design and operate resilient services." {
		t.Fatalf("unexpected description: %q", got)
	}
}
