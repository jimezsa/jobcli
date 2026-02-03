package models

import "time"

// ScraperConfig contains runtime options shared by scrapers.
type ScraperConfig struct {
	Proxies    []string
	Timeout    time.Duration
	UserAgents []string
}
