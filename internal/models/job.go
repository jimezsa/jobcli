package models

import "time"

// Job is the normalized posting returned by scrapers.
type Job struct {
	ID          string    `json:"id,omitempty"`
	Site        string    `json:"site"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	URL         string    `json:"url"`
	Remote      bool      `json:"remote,omitempty"`
	JobType     string    `json:"job_type,omitempty"`
	Salary      string    `json:"salary,omitempty"`
	Description string    `json:"description,omitempty"`
	Snippet     string    `json:"snippet,omitempty"`
	PostedAt    time.Time `json:"posted_at,omitempty"`
	PostedAtRaw string    `json:"posted_at_raw,omitempty"`
}
