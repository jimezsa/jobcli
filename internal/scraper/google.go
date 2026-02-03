package scraper

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/MrJJimenez/jobcli/internal/models"
	"github.com/MrJJimenez/jobcli/internal/network"
	"github.com/PuerkitoBio/goquery"
)

type GoogleJobs struct {
	client *network.Client
}

func NewGoogleJobs(client *network.Client) *GoogleJobs {
	return &GoogleJobs{client: client}
}

func (g *GoogleJobs) Name() string {
	return SiteGoogleJobs
}

func (g *GoogleJobs) Search(ctx context.Context, params models.SearchParams) ([]models.Job, error) {
	searchURL := buildGoogleJobsURL(params)
	doc, err := fetchDocument(ctx, g.client, searchURL, map[string]string{
		"accept-language": "en-US,en;q=0.9",
	})
	if err != nil {
		return nil, fmt.Errorf("google jobs: %w", err)
	}

	jobs := parseJSONLDJobs(doc, SiteGoogleJobs)
	jobs = append(jobs, parseGoogleJobsAnchors(doc)...)
	jobs = dedupeJobs(jobs)

	if params.Remote {
		jobs = filterRemote(jobs)
	}
	if params.Limit > 0 && len(jobs) > params.Limit {
		jobs = jobs[:params.Limit]
	}
	return jobs, nil
}

func buildGoogleJobsURL(params models.SearchParams) string {
	values := url.Values{}
	values.Set("q", params.Query)
	values.Set("ibp", "htl;jobs")
	values.Set("hl", "en")
	if params.Country != "" {
		values.Set("gl", strings.ToUpper(countryToGoogleGL(params.Country)))
	}
	if params.Location != "" {
		values.Set("l", params.Location)
	}
	return fmt.Sprintf("https://www.google.com/search?%s", values.Encode())
}

func countryToGoogleGL(country string) string {
	country = strings.TrimSpace(strings.ToLower(country))
	switch country {
	case "us", "usa", "united states":
		return "us"
	case "uk", "gb", "united kingdom":
		return "gb"
	case "ca", "canada":
		return "ca"
	case "au", "australia":
		return "au"
	default:
		return country
	}
}

// parseGoogleJobsAnchors is a best-effort fallback for job cards that include htidocid links.
func parseGoogleJobsAnchors(doc *goquery.Document) []models.Job {
	var jobs []models.Job

	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		href := strings.TrimSpace(s.AttrOr("href", ""))
		if href == "" || !strings.Contains(href, "htidocid=") {
			return
		}

		title := cleanText(s.Text())
		if title == "" {
			title = cleanText(s.AttrOr("aria-label", ""))
		}
		if title == "" {
			return
		}

		job := models.Job{
			Site:  SiteGoogleJobs,
			Title: title,
			URL:   absoluteURL("https://www.google.com", href),
		}
		jobs = append(jobs, job)
	})

	return jobs
}
