package seen

import (
	"strings"

	"github.com/jimezsa/jobcli/internal/models"
)

const keySeparator = "::"

// DiffStats captures stats for A-B unseen filtering.
type DiffStats struct {
	TotalNew    int
	TotalSeen   int
	InvalidNew  int
	InvalidSeen int
	Unseen      int
}

// InvalidSkipped returns the total invalid records skipped during comparison.
func (s DiffStats) InvalidSkipped() int {
	return s.InvalidNew + s.InvalidSeen
}

// MergeStats captures stats for seen history updates.
type MergeStats struct {
	TotalSeen    int
	TotalInput   int
	InvalidSeen  int
	InvalidInput int
	Added        int
	TotalOut     int
}

// InvalidSkipped returns the total invalid records skipped during merge.
func (s MergeStats) InvalidSkipped() int {
	return s.InvalidSeen + s.InvalidInput
}

// Normalize applies the v1 key normalization.
func Normalize(value string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(value)))
	return strings.Join(fields, " ")
}

// Key builds the normalized title+company key for a job.
func Key(job models.Job) (string, bool) {
	title := Normalize(job.Title)
	company := Normalize(job.Company)
	if title == "" || company == "" {
		return "", false
	}
	return title + keySeparator + company, true
}

// Diff returns unseen jobs from newJobs using existing seenJobs keys.
func Diff(newJobs []models.Job, seenJobs []models.Job) ([]models.Job, DiffStats) {
	stats := DiffStats{
		TotalNew:  len(newJobs),
		TotalSeen: len(seenJobs),
	}

	seenKeys := make(map[string]struct{}, len(seenJobs))
	for _, job := range seenJobs {
		key, ok := Key(job)
		if !ok {
			stats.InvalidSeen++
			continue
		}
		if _, exists := seenKeys[key]; exists {
			continue
		}
		seenKeys[key] = struct{}{}
	}

	newKeys := make(map[string]struct{}, len(newJobs))
	unseen := make([]models.Job, 0, len(newJobs))
	for _, job := range newJobs {
		key, ok := Key(job)
		if !ok {
			stats.InvalidNew++
			continue
		}
		if _, exists := newKeys[key]; exists {
			continue
		}
		newKeys[key] = struct{}{}
		if _, exists := seenKeys[key]; exists {
			continue
		}
		unseen = append(unseen, job)
	}

	stats.Unseen = len(unseen)
	return unseen, stats
}

// Merge appends unique new jobs into the seen history.
// Existing seen entries win collisions.
func Merge(existingSeen []models.Job, inputJobs []models.Job) ([]models.Job, MergeStats) {
	stats := MergeStats{
		TotalSeen:  len(existingSeen),
		TotalInput: len(inputJobs),
	}

	keys := make(map[string]struct{}, len(existingSeen)+len(inputJobs))
	out := make([]models.Job, 0, len(existingSeen)+len(inputJobs))

	for _, job := range existingSeen {
		key, ok := Key(job)
		if !ok {
			stats.InvalidSeen++
			out = append(out, job)
			continue
		}
		if _, exists := keys[key]; exists {
			continue
		}
		keys[key] = struct{}{}
		out = append(out, job)
	}

	for _, job := range inputJobs {
		key, ok := Key(job)
		if !ok {
			stats.InvalidInput++
			continue
		}
		if _, exists := keys[key]; exists {
			continue
		}
		keys[key] = struct{}{}
		out = append(out, job)
		stats.Added++
	}

	stats.TotalOut = len(out)
	return out, stats
}
