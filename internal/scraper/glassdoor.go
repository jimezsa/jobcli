package scraper

import (
	"context"
	"fmt"
	"math"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/jimezsa/jobcli/internal/models"
	"github.com/jimezsa/jobcli/internal/network"
)

type Glassdoor struct {
	client *network.Client
}

func NewGlassdoor(client *network.Client) *Glassdoor {
	return &Glassdoor{client: client}
}

func (g *Glassdoor) Name() string {
	return SiteGlassdoor
}

func (g *Glassdoor) Search(ctx context.Context, params models.SearchParams) ([]models.Job, error) {
	searchURL := buildGlassdoorURL(params)
	doc, err := fetchDocument(ctx, g.client, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("glassdoor: %w", err)
	}

	jobs := parseJSONLDJobs(doc, SiteGlassdoor)
	jobs = append(jobs, parseGlassdoorJobs(doc)...)
	jobs = dedupeJobs(jobs)

	if params.Remote {
		jobs = filterRemote(jobs)
	}
	if params.Limit > 0 && len(jobs) > params.Limit {
		jobs = jobs[:params.Limit]
	}
	return jobs, nil
}

func buildGlassdoorURL(params models.SearchParams) string {
	values := url.Values{}
	values.Set("sc.keyword", params.Query)
	if params.Location != "" {
		values.Set("locKeyword", params.Location)
	}
	if params.Hours > 0 {
		days := int(math.Ceil(float64(params.Hours) / 24.0))
		if days < 1 {
			days = 1
		}
		values.Set("fromAge", fmt.Sprintf("%d", days))
	}
	return fmt.Sprintf("https://www.glassdoor.com/Job/jobs.htm?%s", values.Encode())
}

func parseGlassdoorJobs(doc *goquery.Document) []models.Job {
	var jobs []models.Job

	doc.Find(".react-job-listing").Each(func(_ int, s *goquery.Selection) {
		title := cleanText(s.Find(".jobLink").First().Text())
		if title == "" {
			title = cleanText(s.Find("[data-test='job-title']").First().Text())
		}

		company := cleanText(s.Find(".jobEmployerName").First().Text())
		if company == "" {
			company = cleanText(s.Find(".jobEmpolyerName").First().Text())
		}
		if company == "" {
			company = cleanText(s.Find("[data-test='job-link']").First().Text())
		}

		location := cleanText(s.Find(".jobLocation").First().Text())
		if location == "" {
			location = cleanText(s.Find("[data-test='emp-location']").First().Text())
		}

		salary := cleanText(s.Find(".salarySnippet").First().Text())
		link := s.Find("a.jobLink").First().AttrOr("href", "")
		link = absoluteURL("https://www.glassdoor.com", link)

		if title == "" || link == "" {
			return
		}

		job := models.Job{
			Site:     SiteGlassdoor,
			Title:    title,
			Company:  company,
			Location: location,
			URL:      link,
			Salary:   salary,
			Remote:   isRemote(location, ""),
		}

		jobs = append(jobs, job)
	})

	return jobs
}
