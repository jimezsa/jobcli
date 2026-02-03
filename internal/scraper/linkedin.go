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

const linkedInPageSize = 10

// LinkedIn uses a guest endpoint that returns HTML job cards.
type LinkedIn struct {
	client *network.Client
}

func NewLinkedIn(client *network.Client) *LinkedIn {
	return &LinkedIn{client: client}
}

func (l *LinkedIn) Name() string {
	return SiteLinkedIn
}

func (l *LinkedIn) Search(ctx context.Context, params models.SearchParams) ([]models.Job, error) {
	var jobs []models.Job
	limit := params.Limit

	start := params.Offset
	for {
		if limit > 0 && len(jobs) >= limit {
			break
		}

		searchURL := buildLinkedInURL(params, start)
		doc, err := fetchDocument(ctx, l.client, searchURL, nil)
		if err != nil {
			return nil, fmt.Errorf("linkedin: %w", err)
		}

		pageJobs := parseLinkedInJobs(doc)
		if len(pageJobs) == 0 {
			break
		}

		for _, job := range pageJobs {
			if params.Remote && !job.Remote {
				continue
			}
			jobs = append(jobs, job)
			if limit > 0 && len(jobs) >= limit {
				break
			}
		}

		start += linkedInPageSize
	}

	return jobs, nil
}

func buildLinkedInURL(params models.SearchParams, start int) string {
	values := url.Values{}
	values.Set("keywords", params.Query)
	if params.Location != "" {
		values.Set("location", params.Location)
	}
	values.Set("start", fmt.Sprintf("%d", start))
	if params.Hours > 0 {
		values.Set("f_TPR", fmt.Sprintf("r%d", params.Hours*3600))
	}
	return fmt.Sprintf("https://www.linkedin.com/jobs-guest/jobs/api/seeMoreJobPostings/search?%s", values.Encode())
}

func parseLinkedInJobs(doc *goquery.Document) []models.Job {
	var jobs []models.Job
	seen := map[string]struct{}{}

	doc.Find("li").Each(func(_ int, s *goquery.Selection) {
		link := attrFirst(s, "a.base-card__full-link", "href")
		link = strings.TrimSpace(link)
		if link == "" {
			return
		}

		title := cleanText(s.Find("h3.base-search-card__title").First().Text())
		company := cleanText(s.Find("h4.base-search-card__subtitle").First().Text())
		location := cleanText(s.Find("span.job-search-card__location").First().Text())
		postedRaw := cleanText(s.Find("time").First().AttrOr("datetime", ""))
		if postedRaw == "" {
			postedRaw = cleanText(s.Find("time").First().Text())
		}

		job := models.Job{
			Site:        SiteLinkedIn,
			Title:       title,
			Company:     company,
			Location:    location,
			URL:         link,
			PostedAtRaw: postedRaw,
			Remote:      isRemote(location, ""),
		}
		if job.PostedAtRaw != "" {
			if ts, err := parsePostedAt(job.PostedAtRaw); err == nil {
				job.PostedAt = ts
			}
		}

		key := job.URL
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		jobs = append(jobs, job)
	})

	return jobs
}

func attrFirst(s *goquery.Selection, selector string, attr string) string {
	if s == nil {
		return ""
	}
	return s.Find(selector).First().AttrOr(attr, "")
}
