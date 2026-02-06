package scraper

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/jimezsa/jobcli/internal/models"
	"github.com/jimezsa/jobcli/internal/network"
	"github.com/PuerkitoBio/goquery"
)

type Indeed struct {
	client *network.Client
}

func NewIndeed(client *network.Client) *Indeed {
	return &Indeed{client: client}
}

func (i *Indeed) Name() string {
	return SiteIndeed
}

func (i *Indeed) Search(ctx context.Context, params models.SearchParams) ([]models.Job, error) {
	searchURL := buildIndeedURL(params)
	doc, err := fetchDocument(ctx, i.client, searchURL, nil)
	if err != nil {
		return nil, err
	}

	var jobs []models.Job
	doc.Find("a.tapItem").Each(func(_ int, s *goquery.Selection) {
		if params.Limit > 0 && len(jobs) >= params.Limit {
			return
		}

		title := strings.TrimSpace(s.Find("h2.jobTitle span").First().Text())
		company := strings.TrimSpace(s.Find("span.companyName").First().Text())
		location := strings.TrimSpace(s.Find("div.companyLocation").First().Text())
		snippet := strings.TrimSpace(s.Find("div.job-snippet").Text())
		posted := strings.TrimSpace(s.Find("span.date").Text())

		link, _ := s.Attr("href")
		if link != "" && !strings.HasPrefix(link, "http") {
			link = baseIndeedURL(params.Country) + link
		}

		job := models.Job{
			Site:        SiteIndeed,
			Title:       title,
			Company:     company,
			Location:    location,
			URL:         link,
			Snippet:     normalizeSnippet(snippet),
			PostedAtRaw: posted,
			Remote:      isRemote(location, snippet),
			JobType:     params.JobType,
		}

		if params.Remote && !job.Remote {
			return
		}

		if job.Title == "" || job.URL == "" {
			return
		}

		jobs = append(jobs, job)
	})

	return jobs, nil
}

func buildIndeedURL(params models.SearchParams) string {
	base := baseIndeedURL(params.Country)
	values := url.Values{}
	values.Set("q", params.Query)
	if params.Location != "" {
		values.Set("l", params.Location)
	}
	if params.Offset > 0 {
		values.Set("start", fmt.Sprintf("%d", params.Offset))
	}
	if params.JobType != "" {
		values.Set("jt", params.JobType)
	}
	if params.Hours > 0 {
		days := int(math.Ceil(float64(params.Hours) / 24.0))
		if days < 1 {
			days = 1
		}
		values.Set("fromage", fmt.Sprintf("%d", days))
	}
	return fmt.Sprintf("%s/jobs?%s", base, values.Encode())
}

func baseIndeedURL(country string) string {
	country = strings.TrimSpace(strings.ToLower(country))
	if country == "" || country == "usa" || country == "us" {
		return "https://www.indeed.com"
	}
	return fmt.Sprintf("https://%s.indeed.com", country)
}

func normalizeSnippet(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func isRemote(location string, snippet string) bool {
	value := strings.ToLower(location + " " + snippet)
	return strings.Contains(value, "remote")
}
