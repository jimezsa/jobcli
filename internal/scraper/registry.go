package scraper

import (
	"strings"

	"github.com/jimezsa/jobcli/internal/network"
)

const (
	SiteLinkedIn     = "linkedin"
	SiteIndeed       = "indeed"
	SiteGlassdoor    = "glassdoor"
	SiteZipRecruiter = "ziprecruiter"
	SiteGoogleJobs   = "google"
	SiteStepstone    = "stepstone"
)

func Registry(rotator *network.Rotator) (map[string]Scraper, error) {
	makeClient := func() (*network.Client, error) {
		return network.NewClient(rotator)
	}

	linkedIn, err := makeClient()
	if err != nil {
		return nil, err
	}
	indeed, err := makeClient()
	if err != nil {
		return nil, err
	}
	glassdoor, err := makeClient()
	if err != nil {
		return nil, err
	}
	zipRecruiter, err := makeClient()
	if err != nil {
		return nil, err
	}
	googleJobs, err := makeClient()
	if err != nil {
		return nil, err
	}
	stepstone, err := makeClient()
	if err != nil {
		return nil, err
	}

	return map[string]Scraper{
		SiteLinkedIn:     NewLinkedIn(linkedIn),
		SiteIndeed:       NewIndeed(indeed),
		SiteGlassdoor:    NewGlassdoor(glassdoor),
		SiteZipRecruiter: NewZipRecruiter(zipRecruiter),
		SiteGoogleJobs:   NewGoogleJobs(googleJobs),
		SiteStepstone:    NewStepstone(stepstone),
	}, nil
}

func NormalizeSites(sites []string) []string {
	out := make([]string, 0, len(sites))
	for _, site := range sites {
		site = strings.ToLower(strings.TrimSpace(site))
		if site == "" {
			continue
		}
		site = strings.TrimPrefix(site, "www.")
		out = append(out, site)
	}
	return out
}
