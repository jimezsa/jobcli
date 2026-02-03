package models

// SearchParams captures the normalized search inputs used by scrapers.
type SearchParams struct {
	Query    string
	Location string
	Country  string
	Limit    int
	Offset   int
	Remote   bool
	JobType  string
	Hours    int
}
